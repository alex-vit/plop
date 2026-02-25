package cmd

import (
	"os"

	"github.com/spf13/cobra"

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
	Run: func(cmd *cobra.Command, args []string) {
		tray.Run(Version)
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
