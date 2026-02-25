package engine

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"

	"github.com/syncthing/syncthing/lib/config"
	"github.com/syncthing/syncthing/lib/events"
	"github.com/syncthing/syncthing/lib/protocol"
)

// NewConfig creates a minimal Syncthing configuration for plop:
// single folder, LAN + WAN discovery, relay-capable.
// TODO: RawGlobalAnnServers could be made configurable for self-hosted discovery servers.
func NewConfig(myID protocol.DeviceID, folderPath string, peers []protocol.DeviceID) config.Configuration {
	cfg := config.New(myID)

	folder := config.FolderConfiguration{
		ID:               "default",
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

func generateAPIKey() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
