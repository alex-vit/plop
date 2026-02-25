package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var homeDir string

var rootCmd = &cobra.Command{
	Use:   "gosync",
	Short: "Peer-to-peer file sync",
	Long:  "A minimal P2P file sync tool powered by Syncthing.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	configDir, _ := os.UserConfigDir()
	defaultHome := filepath.Join(configDir, "gosync")
	rootCmd.PersistentFlags().StringVar(&homeDir, "home", defaultHome, "gosync data directory")
}
