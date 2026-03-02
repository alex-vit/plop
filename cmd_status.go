package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alex-vit/plop/engine"
	"github.com/syncthing/syncthing/lib/config"
)

const internalStatusMaxAge = 15 * time.Second

var errInternalStatusStale = errors.New("internal status is stale")

func runStatus(args []string) error {
	fs := flag.NewFlagSet("plop status", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("status takes no arguments")
	}

	cfg, err := readStatusConfig(homeDir)
	if err != nil {
		return err
	}

	snapshot, err := readInternalStatus(homeDir, time.Now())
	if err != nil {
		return fmt.Errorf("reading internal status (is 'plop run' active?): %w", err)
	}
	printInternalStatus(snapshot, cfg)
	return nil
}

func readStatusConfig(homeDir string) (config.Configuration, error) {
	cfgPath := filepath.Join(homeDir, "config.xml")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return config.Configuration{}, fmt.Errorf("reading config (have you run 'plop init'?): %w", err)
	}

	var cfg config.Configuration
	if err := xml.Unmarshal(data, &cfg); err != nil {
		return config.Configuration{}, fmt.Errorf("parsing config: %w", err)
	}
	return cfg, nil
}

func readInternalStatus(homeDir string, now time.Time) (engine.StatusSnapshot, error) {
	var snapshot engine.StatusSnapshot

	data, err := os.ReadFile(filepath.Join(homeDir, engine.StatusFileName))
	if err != nil {
		return snapshot, err
	}
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return snapshot, err
	}
	if snapshot.State == "" {
		return snapshot, fmt.Errorf("missing state in %s", engine.StatusFileName)
	}
	if snapshot.UpdatedAt.IsZero() {
		return snapshot, fmt.Errorf("missing updatedAt in %s", engine.StatusFileName)
	}
	if now.Sub(snapshot.UpdatedAt) > internalStatusMaxAge {
		return snapshot, errInternalStatusStale
	}
	return snapshot, nil
}

func printInternalStatus(snapshot engine.StatusSnapshot, cfg config.Configuration) {
	deviceID := snapshot.DeviceID
	if deviceID == "" {
		deviceID = "unknown"
	}
	fmt.Printf("Device ID: %s\n", deviceID)

	if len(cfg.Folders) > 0 {
		folderID := cfg.Folders[0].ID
		folderPath := cfg.Folders[0].Path
		fmt.Printf("Folder:    %s (%s)\n", folderPath, folderID)
	}

	state := string(snapshot.State)
	folderState := strings.TrimSpace(snapshot.FolderState)
	if folderState != "" && !strings.EqualFold(folderState, state) {
		state = fmt.Sprintf("%s (%s)", state, folderState)
	}
	fmt.Printf("State:     %s\n", state)

	if snapshot.NeedTotalItems > 0 {
		fmt.Printf("Need:      %d items\n", snapshot.NeedTotalItems)
	}

	if snapshot.TotalPeers > 0 {
		fmt.Printf("Peers:     %d/%d connected\n", snapshot.ConnectedPeers, snapshot.TotalPeers)
	} else {
		fmt.Printf("Peers:     %d connected\n", snapshot.ConnectedPeers)
	}
}
