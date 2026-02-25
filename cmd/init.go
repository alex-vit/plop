package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"gosync/engine"
)

var folderPath string

func init() {
	home, _ := os.UserHomeDir()
	initCmd.Flags().StringVar(&folderPath, "folder", filepath.Join(home, "Sync"), "sync folder path")
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize gosync (syncs ~/Sync by default)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		absFolder, err := filepath.Abs(folderPath)
		if err != nil {
			return err
		}

		// Create sync folder if it doesn't exist.
		if err := os.MkdirAll(absFolder, 0o755); err != nil {
			return fmt.Errorf("creating sync folder: %w", err)
		}

		// Create data directory.
		if err := os.MkdirAll(homeDir, 0o700); err != nil {
			return fmt.Errorf("creating data directory: %w", err)
		}

		// Generate TLS certificate.
		certFile := filepath.Join(homeDir, "cert.pem")
		keyFile := filepath.Join(homeDir, "key.pem")
		cert, err := engine.GenerateCert(certFile, keyFile)
		if err != nil {
			return fmt.Errorf("generating certificate: %w", err)
		}

		myID := engine.DeviceID(cert)

		// Generate and save config.
		cfg := engine.NewConfig(myID, absFolder, nil)
		if err := engine.SaveConfig(homeDir, cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		fmt.Println("Initialized gosync")
		fmt.Printf("  Folder: %s\n", absFolder)
		fmt.Printf("  Data:   %s\n", homeDir)
		fmt.Printf("  Device ID: %s\n", myID)

		return nil
	},
}
