package cmd

import (
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	stconfig "github.com/syncthing/syncthing/lib/config"
	"github.com/syncthing/syncthing/lib/protocol"

	"github.com/alex-vit/plop/engine"
)

var pairSyncthing bool

func init() {
	pairCmd.Flags().BoolVar(&pairSyncthing, "syncthing", false, "print setup guidance for Syncthing (optionally also add a peer ID)")
	rootCmd.AddCommand(pairCmd)
}

var pairCmd = &cobra.Command{
	Use:   "pair [device-id]",
	Short: "Add a peer device",
	Long:  "Add a peer device ID to peers.txt. Use --syncthing for a guided setup flow with Syncthing Android/Desktop.",
	Example: strings.TrimSpace(`
plop pair SYNCTHING_DEVICE_ID
plop pair --syncthing
plop pair --syncthing SYNCTHING_DEVICE_ID
`),
	Args: func(cmd *cobra.Command, args []string) error {
		if pairSyncthing {
			return cobra.MaximumNArgs(1)(cmd, args)
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		addedPeer := false
		if len(args) == 1 {
			var peerID protocol.DeviceID
			if err := peerID.UnmarshalText([]byte(args[0])); err != nil {
				return fmt.Errorf("invalid device ID: %w", err)
			}
			var err error
			addedPeer, err = pairPeer(peerID)
			if err != nil {
				return err
			}
		}

		if pairSyncthing {
			printSyncthingGuide(len(args) == 1, addedPeer)
		}
		return nil
	},
}

func pairPeer(peerID protocol.DeviceID) (bool, error) {
	peersFile := filepath.Join(homeDir, "peers.txt")

	// Check if already present.
	existing, _ := engine.ParsePeersFile(peersFile)
	for _, p := range existing {
		if p.ID == peerID {
			fmt.Printf("Already paired with %s\n", peerID)
			return false, nil
		}
	}

	if err := engine.AppendPeersFile(peersFile, peerID); err != nil {
		return false, fmt.Errorf("writing peers.txt: %w", err)
	}

	fmt.Printf("Paired with %s\n", peerID)
	return true, nil
}

func printSyncthingGuide(gotPeerArg, addedPeer bool) {
	folderID, folderPath := loadPrimaryFolderInfo(homeDir)
	localID, hasLocalID, err := tryLoadLocalDeviceID(homeDir)
	if err != nil {
		fmt.Printf("Warning: couldn't load local device ID: %v\n", err)
	}

	fmt.Println()
	fmt.Println("Syncthing interop guide (plop <-> Syncthing):")
	if hasLocalID {
		fmt.Printf("  plop device ID: %s\n", localID)
	} else {
		fmt.Println("  plop device ID: not initialized yet (run `plop id`)")
	}
	fmt.Printf("  plop folder ID: %s\n", folderID)
	fmt.Printf("  plop folder:    %s\n", folderPath)
	fmt.Println()

	if !gotPeerArg {
		fmt.Println("1) Add the Syncthing device to plop:")
		fmt.Println("   plop pair SYNCTHING_DEVICE_ID")
	} else if addedPeer {
		fmt.Println("1) Syncthing device ID was added to peers.txt.")
	} else {
		fmt.Println("1) Syncthing device ID is already present in peers.txt.")
	}

	fmt.Println("2) In Syncthing (Android/Desktop), add this plop device as a remote device.")
	if hasLocalID {
		fmt.Printf("   Use device ID: %s\n", localID)
	} else {
		fmt.Println("   Use the output of: plop id")
	}
	fmt.Println("   Android: Devices tab -> + -> Enter Device ID -> Save.")

	fmt.Printf("3) In Syncthing, create/share exactly one folder with ID %q.\n", folderID)
	fmt.Println("   Android: first copy your phone ID from Devices -> This Device (phone).")
	fmt.Println("   If that folder ID already exists on Android, share it with plop.")
	fmt.Println("   Otherwise create it explicitly (Syncthing defaults to a generated ID).")
	fmt.Println("   Use Send & Receive and choose any local path on that device.")

	fmt.Println("4) Share/accept on both sides, then add a small test file to verify sync.")
}

func loadPrimaryFolderInfo(homeDir string) (string, string) {
	folderID := engine.DefaultFolderID
	folderPath := defaultSyncFolderPath()

	cfgPath := filepath.Join(homeDir, "config.xml")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return folderID, folderPath
	}

	var cfg stconfig.Configuration
	if err := xml.Unmarshal(data, &cfg); err != nil {
		return folderID, folderPath
	}
	if len(cfg.Folders) == 0 {
		return folderID, folderPath
	}
	if cfg.Folders[0].ID != "" {
		folderID = cfg.Folders[0].ID
	}
	if cfg.Folders[0].Path != "" {
		folderPath = cfg.Folders[0].Path
	}
	return folderID, folderPath
}

func defaultSyncFolderPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "plop"
	}
	return filepath.Join(home, "plop")
}

func tryLoadLocalDeviceID(homeDir string) (string, bool, error) {
	certFile := filepath.Join(homeDir, "cert.pem")
	keyFile := filepath.Join(homeDir, "key.pem")

	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		return "", false, nil
	}
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		return "", false, nil
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return "", false, err
	}
	return engine.DeviceID(cert).String(), true, nil
}
