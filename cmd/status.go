package cmd

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/syncthing/syncthing/lib/config"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show sync status",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath := filepath.Join(homeDir, "config.xml")
		data, err := os.ReadFile(cfgPath)
		if err != nil {
			return fmt.Errorf("reading config (have you run 'gosync init'?): %w", err)
		}

		var cfg config.Configuration
		if err := xml.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("parsing config: %w", err)
		}

		addr := cfg.GUI.RawAddress
		apiKey := cfg.GUI.APIKey

		if addr == "" || addr == "127.0.0.1:0" {
			return fmt.Errorf("daemon not running or GUI address not configured (is 'gosync run' active?)")
		}

		baseURL := "http://" + addr

		// System status.
		sysStatus, err := apiGet(baseURL+"/rest/system/status", apiKey)
		if err != nil {
			return fmt.Errorf("querying status (is 'gosync run' active?): %w", err)
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
	},
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
