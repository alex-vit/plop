package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alex-vit/plop/engine"
)

func runInit(args []string) error {
	home, _ := os.UserHomeDir()

	fs := flag.NewFlagSet("plop init", flag.ContinueOnError)
	folderPath := fs.String("folder", filepath.Join(home, "plop"), "sync folder path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("init takes no arguments")
	}

	absFolder, err := filepath.Abs(*folderPath)
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

	fmt.Println("Initialized plop")
	fmt.Printf("  Folder: %s\n", absFolder)
	fmt.Printf("  Data:   %s\n", homeDir)
	fmt.Printf("  Device ID: %s\n", myID)

	return nil
}
