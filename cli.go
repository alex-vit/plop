package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/energye/systray"

	"github.com/alex-vit/plop/engine"
	"github.com/alex-vit/plop/paths"
	"github.com/alex-vit/plop/tray"
)

var (
	homeDir string
	version = ""
)

// run is the main entry point. It parses global flags, dispatches to the
// appropriate subcommand, and returns any error.
func run(args []string) error {
	defaultHome, _ := paths.ConfigDir()

	// Global flags — parsed before the subcommand.
	global := flag.NewFlagSet("plop", flag.ContinueOnError)
	global.StringVar(&homeDir, "home", defaultHome, "plop data directory")
	global.Usage = printUsage
	if err := global.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	subcmd := global.Arg(0)
	subargs := global.Args()
	if len(subargs) > 0 {
		subargs = subargs[1:]
	}

	var err error
	switch subcmd {
	case "":
		err = runRoot(subargs)
	case "init":
		err = runInit(subargs)
	case "pair":
		err = runPair(subargs)
	case "run":
		err = runRun(subargs)
	case "status":
		err = runStatus(subargs)
	case "id":
		err = runID(subargs)
	default:
		return fmt.Errorf("unknown command: %s\nRun 'plop --help' for usage", subcmd)
	}
	if errors.Is(err, flag.ErrHelp) {
		return nil
	}
	return err
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `plop — peer-to-peer file sync

Usage:
  plop [flags]              Start sync daemon with system tray
  plop <command> [flags]    Run a subcommand

Commands:
  init      Initialize plop (syncs ~/plop by default)
  pair      Add a peer device
  run       Start the sync daemon (headless)
  status    Show sync status
  id        Print this device's ID

Flags:
  --home string   plop data directory (default %q)
  --help          Show this help
`, homeDir)
}

func runRoot(args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("unexpected arguments: %v", args)
	}

	logFile := setupLogFile(homeDir)
	if logFile != nil {
		defer func() { _ = logFile.Close() }()
	}

	cleanOldBinary()

	eng, err := engine.New(homeDir, "", nil)
	if err != nil {
		return fmt.Errorf("creating engine: %w", err)
	}
	if err := eng.Start(); err != nil {
		return fmt.Errorf("starting engine: %w", err)
	}
	defer eng.Stop()

	fmt.Printf("plop running as %s\n", eng.DeviceID())
	fmt.Printf("Syncing: %s\n", eng.SyncFolder())
	fmt.Printf("Config: %s\n", filepath.Join(homeDir, "config.xml"))

	// Quit tray on signal so systray.Run unblocks.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		systray.Quit()
	}()

	// Quit tray if engine exits on its own.
	go func() {
		eng.Wait()
		systray.Quit()
	}()

	go autoUpdate()

	tray.Run(version, homeDir, eng.DeviceID().String(), eng.StatusUpdates())
	return nil
}

// setupLogFile redirects log output and stdout/stderr to log.txt in the
// config directory. Used in GUI mode where there's no terminal to see output.
func setupLogFile(homeDir string) *os.File {
	_ = os.MkdirAll(homeDir, 0o700)
	logPath := filepath.Join(homeDir, "log.txt")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		log.Printf("failed to open log file: %v", err)
		return nil
	}
	redirectFd(f)
	log.SetOutput(f)
	os.Stdout = f
	os.Stderr = f

	// Syncthing v2 uses log/slog. Set a global handler that writes to the
	// log file with a format matching the old Syncthing log output.
	slog.SetDefault(slog.New(slog.NewTextHandler(f, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	// Also set slog as the default logger backend so log.Printf goes to the
	// same destination with consistent formatting.
	slog.SetLogLoggerLevel(slog.LevelInfo)

	return f
}
