package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/energye/systray"
	"github.com/spf13/cobra"
	stlogger "github.com/syncthing/syncthing/lib/logger"

	"github.com/alex-vit/plop/engine"
	"github.com/alex-vit/plop/paths"
	"github.com/alex-vit/plop/tray"
)

var (
	homeDir string
	Version = ""
)

var rootCmd = &cobra.Command{
	Use:   "plop",
	Short: "Peer-to-peer file sync",
	Long:  "A minimal P2P file sync tool powered by Syncthing.",
	RunE: func(cmd *cobra.Command, args []string) error {
		logFile := setupLogFile(homeDir)
		if logFile != nil {
			defer logFile.Close()
		}

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

		tray.Run(Version, homeDir, eng.DeviceID().String(), eng.StatusUpdates())
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.MousetrapHelpText = "" // Allow launching from Explorer (GUI app, not a CLI tool).
	defaultHome, _ := paths.ConfigDir()
	rootCmd.PersistentFlags().StringVar(&homeDir, "home", defaultHome, "plop data directory")
}

// setupLogFile redirects log output and stdout/stderr to log.txt in the
// config directory. Used in GUI mode where there's no terminal to see output.
func setupLogFile(homeDir string) *os.File {
	os.MkdirAll(homeDir, 0o700)
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

	// Syncthing's DefaultLogger captures os.Stdout at package init time, so
	// reassigning os.Stdout won't affect it. Hook in via AddHandler instead.
	stlogger.DefaultLogger.AddHandler(stlogger.LevelInfo, func(level stlogger.LogLevel, msg string) {
		var prefix string
		switch level {
		case stlogger.LevelDebug:
			prefix = "DEBUG"
		case stlogger.LevelVerbose:
			prefix = "VERBOSE"
		case stlogger.LevelInfo:
			prefix = "INFO"
		case stlogger.LevelWarn:
			prefix = "WARNING"
		default:
			prefix = "INFO"
		}
		fmt.Fprintf(f, "%s %s: %s\n", time.Now().Format("2006/01/02 15:04:05"), prefix, msg)
	})

	return f
}
