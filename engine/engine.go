package engine

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/syncthing/notify"
	"github.com/syncthing/syncthing/lib/config"
	"github.com/syncthing/syncthing/lib/db/backend"
	"github.com/syncthing/syncthing/lib/events"
	"github.com/syncthing/syncthing/lib/protocol"
	"github.com/syncthing/syncthing/lib/svcutil"
	"github.com/syncthing/syncthing/lib/syncthing"
	suture "github.com/thejerf/suture/v4"
)

// Engine wraps syncthing.App with plop's simplified lifecycle.
type Engine struct {
	app              *syncthing.App
	cfgWrapper       config.Wrapper
	evLogger         events.Logger
	earlyService     *suture.Supervisor
	earlyServiceDone <-chan error
	earlyServiceStop context.CancelFunc
	peerWatchStop    context.CancelFunc
	statusSvc        *statusService
	cert             tls.Certificate
	homeDir          string
}

// New creates a new Engine. It loads the cert, config, opens the database,
// and starts the early services (event logger and config wrapper) that
// Syncthing's App.Start() depends on.
func New(homeDir string, folderPath string, peers []protocol.DeviceID) (*Engine, error) {
	if err := os.MkdirAll(homeDir, 0o700); err != nil {
		return nil, err
	}

	certFile := filepath.Join(homeDir, "cert.pem")
	keyFile := filepath.Join(homeDir, "key.pem")

	cert, err := LoadOrGenerateCert(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	myID := DeviceID(cert)

	// Write --peer flags to peers.txt so they persist.
	if len(peers) > 0 {
		peersFile := filepath.Join(homeDir, "peers.txt")
		existing, _ := ParsePeersFile(peersFile)
		seen := make(map[protocol.DeviceID]bool, len(existing))
		for _, p := range existing {
			seen[p] = true
		}
		for _, p := range peers {
			if !seen[p] {
				if err := AppendPeersFile(peersFile, p); err != nil {
					return nil, fmt.Errorf("writing peers.txt: %w", err)
				}
			}
		}
	}

	// Ensure peers.txt exists so "Open Settings" always has a file to open.
	peersFile := filepath.Join(homeDir, "peers.txt")
	if _, err := os.Stat(peersFile); os.IsNotExist(err) {
		os.WriteFile(peersFile, []byte("# Add one device ID per line\n"), 0o644)
	}

	// Collect all desired peers from peers.txt.
	allPeers, _ := ParsePeersFile(peersFile)

	// Load or create config.
	cfgPath := filepath.Join(homeDir, "config.xml")
	var cfg config.Configuration
	if f, err := os.Open(cfgPath); err == nil {
		cfg, _, err = config.ReadXML(f, myID)
		f.Close()
		if err != nil {
			return nil, err
		}
	} else {
		if folderPath == "" {
			userHome, err := os.UserHomeDir()
			if err != nil {
				return nil, err
			}
			folderPath = filepath.Join(userHome, "plop")
		}
		cfg = NewConfig(myID, folderPath, allPeers)
	}

	// Always reconcile peers.
	syncPeersConfig(&cfg, myID, allPeers)
	if err := ensureRuntimeGUIAddress(&cfg); err != nil {
		return nil, fmt.Errorf("resolving GUI address: %w", err)
	}
	if err := SaveConfig(homeDir, cfg); err != nil {
		return nil, err
	}

	// Ensure sync folders and .stignore exist.
	for _, folder := range cfg.Folders {
		os.MkdirAll(folder.Path, 0o755)
		writeDefaultStignore(folder.Path)
	}

	evLogger := events.NewLogger()

	// The config wrapper and event logger are suture services that must
	// be running before App.Start() — Syncthing's startup() calls
	// cfg.Modify() which blocks until the wrapper's Serve() loop is active.
	earlyService := suture.New("early", suture.Spec{})
	ctx, cancel := context.WithCancel(context.Background())
	earlyServiceDone := earlyService.ServeBackground(ctx)
	earlyService.Add(evLogger)

	cfgWrapper, err := syncthing.LoadConfigAtStartup(cfgPath, cert, evLogger, false, true, false)
	if err != nil {
		cancel()
		return nil, err
	}
	earlyService.Add(cfgWrapper)

	dbPath := filepath.Join(homeDir, "db")
	db, err := backend.Open(dbPath, backend.TuningAuto)
	if err != nil {
		cancel()
		return nil, err
	}

	app, err := syncthing.New(cfgWrapper, db, evLogger, cert, syncthing.Options{
		NoUpgrade: true,
	})
	if err != nil {
		db.Close()
		cancel()
		return nil, err
	}

	return &Engine{
		app:              app,
		cfgWrapper:       cfgWrapper,
		evLogger:         evLogger,
		earlyService:     earlyService,
		earlyServiceDone: earlyServiceDone,
		earlyServiceStop: cancel,
		cert:             cert,
		homeDir:          homeDir,
	}, nil
}

func (e *Engine) Start() error {
	if err := e.app.Start(); err != nil {
		return err
	}
	e.statusSvc = newStatusService(e.cfgWrapper, newSyncthingStatusRuntime(e.app.Internals), newStatusEventSource(e.evLogger), e.DeviceID())
	e.statusSvc.Start()

	ctx, cancel := context.WithCancel(context.Background())
	e.peerWatchStop = cancel
	go e.watchPeers(ctx)
	return nil
}

func (e *Engine) Wait() svcutil.ExitStatus {
	return e.app.Wait()
}

func (e *Engine) Stop() {
	if e.peerWatchStop != nil {
		e.peerWatchStop()
	}
	if e.statusSvc != nil {
		e.statusSvc.Stop()
	}
	e.app.Stop(svcutil.ExitSuccess)
	e.earlyServiceStop()
	if e.earlyServiceDone != nil {
		<-e.earlyServiceDone
	}
}

func (e *Engine) DeviceID() protocol.DeviceID {
	return DeviceID(e.cert)
}

// StatusSnapshot returns the latest cached internal status snapshot.
func (e *Engine) StatusSnapshot() StatusSnapshot {
	if e.statusSvc == nil {
		return StatusSnapshot{
			State:     StatusStateUnavailable,
			Error:     "status service not started",
			UpdatedAt: time.Now().UTC(),
		}
	}
	return e.statusSvc.Snapshot()
}

// StatusUpdates returns a channel with status snapshot updates.
func (e *Engine) StatusUpdates() <-chan StatusSnapshot {
	if e.statusSvc == nil {
		return nil
	}
	return e.statusSvc.Updates()
}

// PeersFilePath returns the path to the peers.txt file.
func (e *Engine) PeersFilePath() string {
	return filepath.Join(e.homeDir, "peers.txt")
}

// SyncFolder returns the path of the first configured sync folder.
func (e *Engine) SyncFolder() string {
	cfg := e.cfgWrapper.RawCopy()
	if len(cfg.Folders) == 0 {
		return ""
	}
	return cfg.Folders[0].Path
}

// syncPeers reads peers.txt and updates the running config to match.
func (e *Engine) syncPeers() {
	peersFile := e.PeersFilePath()
	desired, err := ParsePeersFile(peersFile)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("peers: reading %s: %v", peersFile, err)
		}
		return
	}

	myID := e.DeviceID()
	if _, err := e.cfgWrapper.Modify(func(cfg *config.Configuration) {
		syncPeersConfig(cfg, myID, desired)
	}); err != nil {
		log.Printf("peers: updating config: %v", err)
	}
}

// syncPeersConfig reconciles cfg to have exactly the desired set of peers.
func syncPeersConfig(cfg *config.Configuration, myID protocol.DeviceID, desired []protocol.DeviceID) {
	want := make(map[protocol.DeviceID]bool, len(desired))
	for _, p := range desired {
		want[p] = true
	}

	// Remove peers not in desired set (skip self).
	var toRemove []protocol.DeviceID
	for _, d := range cfg.Devices {
		if d.DeviceID == myID {
			continue
		}
		if !want[d.DeviceID] {
			toRemove = append(toRemove, d.DeviceID)
		}
	}
	for _, id := range toRemove {
		RemovePeer(cfg, id)
	}

	// Add missing peers.
	for _, p := range desired {
		AddPeer(cfg, p)
	}
}

// watchPeers watches peers.txt for changes and re-syncs the config.
func (e *Engine) watchPeers(ctx context.Context) {
	peersFile := e.PeersFilePath()

	c := make(chan notify.EventInfo, 1)
	// Watch the parent directory — watching a single file that may not exist
	// yet or gets recreated (editor save) is unreliable on some platforms.
	dir := filepath.Dir(peersFile)
	if err := notify.Watch(dir, c, notify.Create|notify.Write|notify.Rename); err != nil {
		log.Printf("peers: watching %s: %v", dir, err)
		return
	}
	defer notify.Stop(c)

	base := filepath.Base(peersFile)
	// Debounce: coalesce rapid edits into a single sync.
	var timer *time.Timer
	for {
		select {
		case <-ctx.Done():
			if timer != nil {
				timer.Stop()
			}
			return
		case ei := <-c:
			if filepath.Base(ei.Path()) != base {
				continue
			}
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(500*time.Millisecond, e.syncPeers)
		}
	}
}
