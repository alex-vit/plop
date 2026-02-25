package engine

import (
	"context"
	"crypto/tls"
	"os"
	"path/filepath"

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
	earlyService     *suture.Supervisor
	earlyServiceStop context.CancelFunc
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

	// Create default config if none exists.
	cfgPath := filepath.Join(homeDir, "config.xml")
	configExisted := true
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		configExisted = false
		if folderPath == "" {
			userHome, err := os.UserHomeDir()
			if err != nil {
				return nil, err
			}
			folderPath = filepath.Join(userHome, "plop")
		}
		if err := os.MkdirAll(folderPath, 0o755); err != nil {
			return nil, err
		}
		cfg := NewConfig(myID, folderPath, peers)
		if err := SaveConfig(homeDir, cfg); err != nil {
			return nil, err
		}
	}

	// Add peers to existing config on subsequent runs.
	if len(peers) > 0 && configExisted {
		f, err := os.Open(cfgPath)
		if err != nil {
			return nil, err
		}
		cfg, _, err := config.ReadXML(f, myID)
		f.Close()
		if err != nil {
			return nil, err
		}
		for _, p := range peers {
			AddPeer(&cfg, p)
		}
		if err := SaveConfig(homeDir, cfg); err != nil {
			return nil, err
		}
	}

	evLogger := events.NewLogger()

	// The config wrapper and event logger are suture services that must
	// be running before App.Start() — Syncthing's startup() calls
	// cfg.Modify() which blocks until the wrapper's Serve() loop is active.
	earlyService := suture.New("early", suture.Spec{})
	ctx, cancel := context.WithCancel(context.Background())
	earlyService.ServeBackground(ctx)
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
		earlyService:     earlyService,
		earlyServiceStop: cancel,
		cert:             cert,
		homeDir:          homeDir,
	}, nil
}

func (e *Engine) Start() error {
	return e.app.Start()
}

func (e *Engine) Wait() svcutil.ExitStatus {
	return e.app.Wait()
}

func (e *Engine) Stop() {
	e.app.Stop(svcutil.ExitSuccess)
	e.earlyServiceStop()
}

func (e *Engine) DeviceID() protocol.DeviceID {
	return DeviceID(e.cert)
}

// SyncFolder returns the path of the first configured sync folder.
func (e *Engine) SyncFolder() string {
	cfgPath := filepath.Join(e.homeDir, "config.xml")
	f, err := os.Open(cfgPath)
	if err != nil {
		return ""
	}
	defer f.Close()
	cfg, _, err := config.ReadXML(f, e.DeviceID())
	if err != nil || len(cfg.Folders) == 0 {
		return ""
	}
	return cfg.Folders[0].Path
}
