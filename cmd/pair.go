package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
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

		peersFile := filepath.Join(homeDir, "peers.txt")

		// Check if already present.
		existing, _ := engine.ParsePeersFile(peersFile)
		for _, p := range existing {
			if p == peerID {
				fmt.Printf("Already paired with %s\n", peerID)
				return nil
			}
		}

		if err := engine.AppendPeersFile(peersFile, peerID); err != nil {
			return fmt.Errorf("writing peers.txt: %w", err)
		}

		fmt.Printf("Paired with %s\n", peerID)
		return nil
	},
}
