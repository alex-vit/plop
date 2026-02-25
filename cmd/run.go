package cmd

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/syncthing/syncthing/lib/protocol"

	"github.com/alex-vit/plop/engine"
	"github.com/alex-vit/plop/paths"
)

var peerStrs []string

func init() {
	runCmd.Flags().StringArrayVar(&peerStrs, "peer", nil, "device ID of a peer to sync with (repeatable)")
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run [folder]",
	Short: "Start the sync daemon",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		home := homeDir
		folderPath := ""

		if len(args) == 1 {
			abs, err := filepath.Abs(args[0])
			if err != nil {
				return fmt.Errorf("resolving folder path: %w", err)
			}
			folderPath = abs

			configDir, err := paths.ConfigDir()
			if err != nil {
				return fmt.Errorf("config dir: %w", err)
			}
			hash := sha256.Sum256([]byte(abs))
			home = filepath.Join(configDir, "instances", fmt.Sprintf("%x", hash[:4]))
		}

		var peers []protocol.DeviceID
		for _, s := range peerStrs {
			id, err := protocol.DeviceIDFromString(s)
			if err != nil {
				return fmt.Errorf("invalid peer device ID %q: %w", s, err)
			}
			peers = append(peers, id)
		}

		eng, err := engine.New(home, folderPath, peers)
		if err != nil {
			return fmt.Errorf("creating engine: %w", err)
		}

		if err := eng.Start(); err != nil {
			return fmt.Errorf("starting engine: %w", err)
		}

		fmt.Printf("plop running as %s\n", eng.DeviceID())
		fmt.Printf("Syncing: %s\n", eng.SyncFolder())
		fmt.Printf("Config: %s\n", filepath.Join(home, "config.xml"))
		fmt.Println("Press Ctrl-C to stop.")

		// Handle shutdown signals.
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

		done := make(chan struct{})
		go func() {
			eng.Wait()
			close(done)
		}()

		select {
		case <-sig:
			fmt.Println("\nShutting down...")
			eng.Stop()
			<-done
		case <-done:
		}

		return nil
	},
}
