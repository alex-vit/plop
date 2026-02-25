package engine

import (
	"context"
	"crypto/tls"
	"path/filepath"

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
func New(homeDir string) (*Engine, error) {
	certFile := filepath.Join(homeDir, "cert.pem")
	keyFile := filepath.Join(homeDir, "key.pem")

	cert, err := LoadCert(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	evLogger := events.NewLogger()

	// The config wrapper and event logger are suture services that must
	// be running before App.Start() — Syncthing's startup() calls
	// cfg.Modify() which blocks until the wrapper's Serve() loop is active.
	earlyService := suture.New("early", suture.Spec{})
	ctx, cancel := context.WithCancel(context.Background())
	earlyService.ServeBackground(ctx)
	earlyService.Add(evLogger)

	cfgPath := filepath.Join(homeDir, "config.xml")
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
