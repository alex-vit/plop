package engine

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/syncthing/syncthing/lib/config"
	"github.com/syncthing/syncthing/lib/events"
	"github.com/syncthing/syncthing/lib/protocol"
)

const DefaultFolderID = "default"

// NewConfig creates a minimal Syncthing configuration for plop:
// single folder, LAN + WAN discovery, relay-capable.
// TODO: RawGlobalAnnServers could be made configurable for self-hosted discovery servers.
func NewConfig(myID protocol.DeviceID, folderPath string, peers []protocol.DeviceID) config.Configuration {
	cfg := config.New(myID)

	folder := config.FolderConfiguration{
		ID:               DefaultFolderID,
		Label:            "Sync",
		Path:             folderPath,
		Type:             config.FolderTypeSendReceive,
		FSWatcherEnabled: true,
		FSWatcherDelayS:  10,
		RescanIntervalS:  60,
		AutoNormalize:    true,
	}

	folder.Devices = append(folder.Devices, config.FolderDeviceConfiguration{DeviceID: myID})
	for _, p := range peers {
		folder.Devices = append(folder.Devices, config.FolderDeviceConfiguration{DeviceID: p})
		cfg.Devices = append(cfg.Devices, config.DeviceConfiguration{
			DeviceID:  p,
			Addresses: []string{"dynamic"},
		})
	}
	cfg.Folders = []config.FolderConfiguration{folder}

	// Use port 0 (OS-assigned) so multiple instances don't collide.
	// LAN discovery broadcasts the actual listen addresses to peers.
	// The relay URL enables WAN connectivity through Syncthing's relay pool.
	cfg.Options.RawListenAddresses = []string{
		"tcp://0.0.0.0:0",
		"quic://0.0.0.0:0",
		"dynamic+https://relays.syncthing.net/endpoint",
	}
	cfg.Options.URAccepted = -1

	cfg.GUI.Enabled = true
	cfg.GUI.RawAddress = "127.0.0.1:0"
	cfg.GUI.APIKey = generateAPIKey()

	return cfg
}

// ensureRuntimeGUIAddress assigns a concrete localhost port when GUI uses
// host:0, so external status checks can reliably find the REST API address.
func ensureRuntimeGUIAddress(cfg *config.Configuration) error {
	if !cfg.GUI.Enabled {
		return nil
	}

	addr := strings.TrimSpace(cfg.GUI.RawAddress)
	if addr == "" {
		addr = "127.0.0.1:0"
	}

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		// Non-TCP forms (for example UNIX sockets) are left unchanged.
		return nil
	}
	if port != "0" {
		return nil
	}

	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}

	resolved, err := pickFreeAddress(host)
	if err != nil {
		return err
	}
	cfg.GUI.RawAddress = resolved
	return nil
}

func pickFreeAddress(host string) (string, error) {
	ln, err := net.Listen("tcp", net.JoinHostPort(host, "0"))
	if err != nil {
		return "", fmt.Errorf("allocating GUI listener on %s: %w", host, err)
	}
	defer ln.Close()

	_, port, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		return "", fmt.Errorf("parsing allocated GUI listener address: %w", err)
	}
	return net.JoinHostPort(host, port), nil
}

// LoadConfig loads an existing config from disk.
func LoadConfig(homeDir string, myID protocol.DeviceID, evLogger events.Logger) (config.Wrapper, error) {
	path := filepath.Join(homeDir, "config.xml")
	w, _, err := config.Load(path, myID, evLogger)
	return w, err
}

// SaveConfig writes a configuration to config.xml in the home directory.
func SaveConfig(homeDir string, cfg config.Configuration) error {
	path := filepath.Join(homeDir, "config.xml")
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return cfg.WriteXML(f)
}

// AddPeer adds a device to the config's device list and folder sharing.
func AddPeer(cfg *config.Configuration, peerID protocol.DeviceID) {
	// Add to device list if not already present.
	for _, d := range cfg.Devices {
		if d.DeviceID == peerID {
			goto addToFolder
		}
	}
	cfg.Devices = append(cfg.Devices, config.DeviceConfiguration{
		DeviceID:  peerID,
		Addresses: []string{"dynamic"},
	})

addToFolder:
	if len(cfg.Folders) == 0 {
		return
	}
	for _, fd := range cfg.Folders[0].Devices {
		if fd.DeviceID == peerID {
			return
		}
	}
	cfg.Folders[0].Devices = append(cfg.Folders[0].Devices, config.FolderDeviceConfiguration{
		DeviceID: peerID,
	})
}

// RemovePeer removes a device from the config's device list and folder sharing.
func RemovePeer(cfg *config.Configuration, peerID protocol.DeviceID) {
	// Remove from device list.
	for i, d := range cfg.Devices {
		if d.DeviceID == peerID {
			cfg.Devices = append(cfg.Devices[:i], cfg.Devices[i+1:]...)
			break
		}
	}

	// Remove from folder device lists.
	for fi := range cfg.Folders {
		devs := cfg.Folders[fi].Devices
		for i, fd := range devs {
			if fd.DeviceID == peerID {
				cfg.Folders[fi].Devices = append(devs[:i], devs[i+1:]...)
				break
			}
		}
	}
}

// writeDefaultStignore creates a .stignore file with sensible defaults
// if one doesn't already exist.
func writeDefaultStignore(folderPath string) {
	p := filepath.Join(folderPath, ".stignore")
	if _, err := os.Stat(p); err == nil {
		return
	}
	os.WriteFile(p, []byte("// Syncthing internals\n.stfolder\n\n// OS junk\n.DS_Store\nThumbs.db\ndesktop.ini\n"), 0o644)
}

func generateAPIKey() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
