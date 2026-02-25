package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/energye/systray"
	"github.com/spf13/cobra"

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

		tray.Run(Version, homeDir, eng.DeviceID().String())
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	defaultHome, _ := paths.ConfigDir()
	rootCmd.PersistentFlags().StringVar(&homeDir, "home", defaultHome, "plop data directory")
}
