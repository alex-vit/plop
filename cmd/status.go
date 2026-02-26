package cmd

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alex-vit/plop/engine"
	"github.com/spf13/cobra"
	"github.com/syncthing/syncthing/lib/config"
)

const internalStatusMaxAge = 15 * time.Second

var errInternalStatusStale = errors.New("internal status is stale")

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show sync status",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := readStatusConfig(homeDir)
		if err != nil {
			return err
		}

		snapshot, err := readInternalStatus(homeDir, time.Now())
		if err == nil {
			printInternalStatus(snapshot, cfg)
			return nil
		}
		// Keep REST mode for one release as fallback while internal status
		// rollout settles.
		return printRESTStatus(cfg)
	},
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

func printRESTStatus(cfg config.Configuration) error {
	addr := cfg.GUI.RawAddress
	apiKey := cfg.GUI.APIKey

	if addr == "" || addr == "127.0.0.1:0" {
		return fmt.Errorf("daemon not running or GUI address not configured (is 'plop run' active?)")
	}

	baseURL := "http://" + addr

	// System status.
	sysStatus, err := apiGet(baseURL+"/rest/system/status", apiKey)
	if err != nil {
		return fmt.Errorf("querying status (is 'plop run' active?): %w", err)
	}

	fmt.Printf("Device ID: %s\n", sysStatus["myID"])

	// Folder status.
	if len(cfg.Folders) > 0 {
		folderID := cfg.Folders[0].ID
		folderPath := cfg.Folders[0].Path
		fmt.Printf("Folder:    %s (%s)\n", folderPath, folderID)

		dbStatus, err := apiGet(baseURL+"/rest/db/status?folder="+folderID, apiKey)
		if err == nil {
			fmt.Printf("State:     %s\n", dbStatus["state"])
			if global, ok := dbStatus["globalFiles"]; ok {
				fmt.Printf("Files:     %.0f global\n", global)
			}
		}
	}

	// Connections.
	conns, err := apiGet(baseURL+"/rest/system/connections", apiKey)
	if err == nil {
		if connections, ok := conns["connections"].(map[string]interface{}); ok {
			connected := 0
			for _, v := range connections {
				if peer, ok := v.(map[string]interface{}); ok {
					if c, _ := peer["connected"].(bool); c {
						connected++
					}
				}
			}
			fmt.Printf("Peers:     %d connected\n", connected)
		}
	}
	return nil
}

func apiGet(url, apiKey string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result, nil
}
