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
	home, _ := os.UserHomeDir()
	defaultHome := filepath.Join(home, ".gosync")
	rootCmd.PersistentFlags().StringVar(&homeDir, "home", defaultHome, "gosync data directory")
}
