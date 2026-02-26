package tray

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/alex-vit/plop/icon"
	"github.com/energye/systray"
)

type runtimeConfig struct {
	GUI struct {
		Address string `xml:"address"`
		APIKey  string `xml:"apikey"`
	} `xml:"gui"`
	Folders []struct {
		ID string `xml:"id,attr"`
	} `xml:"folder"`
}

type dbStatusResponse struct {
	State          string `json:"state"`
	NeedTotalItems int    `json:"needTotalItems"`
}

type systemConnectionsResponse struct {
	Connections map[string]struct {
		Connected bool `json:"connected"`
	} `json:"connections"`
}

type trayStatus struct {
	title     string
	tooltip   string
	iconState icon.StatusLight
}

func startStatusMonitor(homeDir string, item *systray.MenuItem) func() {
	stop := make(chan struct{})
	var once sync.Once

	go func() {
		client := &http.Client{Timeout: 2 * time.Second}
		ticker := time.NewTicker(4 * time.Second)
		defer ticker.Stop()

		current := trayStatus{iconState: icon.StatusLightSyncing}
		for {
			next := computeTrayStatus(client, homeDir)
			if next.title != current.title {
				item.SetTitle(next.title)
			}
			if next.tooltip != current.tooltip {
				systray.SetTooltip(next.tooltip)
			}
			if next.iconState != current.iconState {
				setTrayIcon(next.iconState)
			}
			current = next

			select {
			case <-stop:
				return
			case <-ticker.C:
			}
		}
	}()

	return func() {
		once.Do(func() {
			close(stop)
		})
	}
}

func computeTrayStatus(client *http.Client, homeDir string) trayStatus {
	cfg, err := readRuntimeConfig(homeDir)
	if err != nil {
		if os.IsNotExist(err) {
			return trayStatus{title: "Status: Starting...", tooltip: "plop - Starting...", iconState: icon.StatusLightSyncing}
		}
		return trayStatus{title: "Status: Config error", tooltip: "plop - Config error", iconState: icon.StatusLightAttention}
	}

	addr := resolveGUIAddress(homeDir, cfg.GUI.Address)
	if addr == "" || cfg.GUI.APIKey == "" {
		return trayStatus{title: "Status: Starting...", tooltip: "plop - Starting...", iconState: icon.StatusLightSyncing}
	}

	baseURL := addr
	if !strings.Contains(baseURL, "://") {
		baseURL = "http://" + baseURL
	}

	folderID := "default"
	if len(cfg.Folders) > 0 && cfg.Folders[0].ID != "" {
		folderID = cfg.Folders[0].ID
	}

	var dbStatus dbStatusResponse
	dbURL := baseURL + "/rest/db/status?folder=" + url.QueryEscape(folderID)
	if err := apiGet(client, dbURL, cfg.GUI.APIKey, &dbStatus); err != nil {
		return trayStatus{title: "Status: Unavailable", tooltip: "plop - Status unavailable", iconState: icon.StatusLightAttention}
	}

	connected, totalPeers := fetchConnectionCounts(client, baseURL, cfg.GUI.APIKey)
	state := strings.ToLower(strings.TrimSpace(dbStatus.State))

	switch {
	case strings.Contains(state, "error") || state == "unknown":
		return trayStatus{title: "Status: Error", tooltip: "plop - Sync error", iconState: icon.StatusLightAttention}
	case state == "idle":
		if dbStatus.NeedTotalItems > 0 {
			return trayStatus{title: "Status: Syncing...", tooltip: "plop - Syncing...", iconState: icon.StatusLightSyncing}
		}
		if totalPeers > 0 && connected == 0 {
			return trayStatus{
				title:     "Status: Waiting for peers",
				tooltip:   fmt.Sprintf("plop - Waiting for peers (0/%d connected)", totalPeers),
				iconState: icon.StatusLightAttention,
			}
		}
		if totalPeers > 0 {
			return trayStatus{
				title:     "Status: Synced",
				tooltip:   fmt.Sprintf("plop - Synced (%d/%d peers connected)", connected, totalPeers),
				iconState: icon.StatusLightSynced,
			}
		}
		return trayStatus{title: "Status: Synced", tooltip: "plop - Synced", iconState: icon.StatusLightSynced}
	case state == "":
		return trayStatus{title: "Status: Starting...", tooltip: "plop - Starting...", iconState: icon.StatusLightSyncing}
	default:
		return trayStatus{title: "Status: Syncing...", tooltip: "plop - Syncing...", iconState: icon.StatusLightSyncing}
	}
}

func fetchConnectionCounts(client *http.Client, baseURL, apiKey string) (connected int, total int) {
	var conns systemConnectionsResponse
	if err := apiGet(client, baseURL+"/rest/system/connections", apiKey, &conns); err != nil {
		return 0, 0
	}

	for _, peer := range conns.Connections {
		total++
		if peer.Connected {
			connected++
		}
	}
	return connected, total
}

func readRuntimeConfig(homeDir string) (runtimeConfig, error) {
	var cfg runtimeConfig

	data, err := os.ReadFile(filepath.Join(homeDir, "config.xml"))
	if err != nil {
		return cfg, err
	}
	if err := xml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func resolveGUIAddress(homeDir, cfgAddress string) string {
	addr := strings.TrimSpace(cfgAddress)
	if addr != "" && !strings.HasSuffix(addr, ":0") {
		return addr
	}

	fromLog, err := readRuntimeGUIAddress(filepath.Join(homeDir, "log.txt"))
	if err != nil {
		return ""
	}
	return fromLog
}

func readRuntimeGUIAddress(logPath string) (string, error) {
	data, err := os.ReadFile(logPath)
	if err != nil {
		return "", err
	}

	const marker = "GUI and API listening on "
	lines := strings.Split(string(data), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		idx := strings.Index(line, marker)
		if idx < 0 {
			continue
		}
		addr := strings.TrimSpace(line[idx+len(marker):])
		if addr != "" {
			return addr, nil
		}
	}
	return "", errors.New("runtime GUI address not found in log")
}

func apiGet(client *http.Client, endpoint, apiKey string, out any) error {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-API-Key", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, out)
}
