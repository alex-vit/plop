package cmd

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/syncthing/syncthing/lib/config"
	"github.com/syncthing/syncthing/lib/protocol"

	"github.com/alex-vit/plop/engine"
)

func init() {
	rootCmd.AddCommand(pairCmd)
}

var pairCmd = &cobra.Command{
	Use:   "pair <device-id>",
	Short: "Add a peer device",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var peerID protocol.DeviceID
		if err := peerID.UnmarshalText([]byte(args[0])); err != nil {
			return fmt.Errorf("invalid device ID: %w", err)
		}

		// Load existing config.
		cfgPath := filepath.Join(homeDir, "config.xml")
		data, err := os.ReadFile(cfgPath)
		if err != nil {
			return fmt.Errorf("reading config: %w", err)
		}

		var cfg config.Configuration
		if err := xml.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("parsing config: %w", err)
		}

		engine.AddPeer(&cfg, peerID)

		if err := engine.SaveConfig(homeDir, cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		fmt.Printf("Paired with %s\n", peerID)
		return nil
	},
}
