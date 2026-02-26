package tray

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alex-vit/plop/autostart"
	"github.com/alex-vit/plop/engine"
	"github.com/alex-vit/plop/icon"
	"github.com/energye/systray"
)

var stopStatusMonitor func()

// Run blocks on the calling goroutine, showing the system tray icon.
func Run(version, homeDir, deviceID string, statusUpdates <-chan engine.StatusSnapshot) {
	systray.Run(func() { onReady(version, homeDir, deviceID, statusUpdates) }, onExit)
}

func onReady(version, homeDir, deviceID string, statusUpdates <-chan engine.StatusSnapshot) {
	setTrayIcon(icon.StatusLightSyncing)
	systray.SetTooltip("plop")
	systray.SetOnClick(func(menu systray.IMenu) { menu.ShowMenu() })
	systray.SetOnRClick(func(menu systray.IMenu) { menu.ShowMenu() })
	if runtime.GOOS == "windows" || runtime.GOOS == "linux" {
		systray.SetOnDClick(func(menu systray.IMenu) { openSyncFolder(homeDir) })
	}

	mTitle := systray.AddMenuItem("plop "+displayVersion(version), "")
	mTitle.Disable()

	mStatus := systray.AddMenuItem("Status: Starting...", "Current sync status")
	mStatus.Disable()
	stopStatusMonitor = startStatusMonitor(statusUpdates, mStatus)

	systray.AddSeparator()

	mFolder := systray.AddMenuItem("Open Plop Folder", "Open synced folder in file manager")
	mFolder.Click(func() { openSyncFolder(homeDir) })

	systray.AddSeparator()

	mCopyID := systray.AddMenuItem("Copy My ID", "Copy this device's ID to clipboard")
	mCopyID.Click(func() { copyToClipboard(deviceID) })

	mPairAndroid := systray.AddMenuItem("Pair Android (Syncthing)...", "Copy your ID and open a pairing checklist")
	mPairAndroid.Click(func() { openAndroidPairingGuide(homeDir, deviceID) })

	mPeers := systray.AddMenuItem("Add or Edit Peers", "Open peers.txt in text editor")
	mPeers.Click(func() { openInEditor(filepath.Join(homeDir, "peers.txt")) })

	mConfig := systray.AddMenuItem("Open Config Folder", "Open config directory in file manager")
	mConfig.Click(func() { openPath(homeDir) })

	if label := autostart.MenuLabel(); label != "" {
		systray.AddSeparator()

		mAutostart := systray.AddMenuItem(label, "Launch plop automatically when you sign in")
		if autostart.IsEnabled(homeDir) {
			mAutostart.Check()
		}
		mAutostart.Click(func() { toggleAutostart(mAutostart, homeDir) })
	}

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Exit", "Quit plop")
	mQuit.Click(func() { systray.Quit() })
}

func onExit() {
	if stopStatusMonitor != nil {
		stopStatusMonitor()
		stopStatusMonitor = nil
	}
}

func setTrayIcon(state icon.StatusLight) {
	pngData, icoData := icon.BytesForStatusLight(state)
	if runtime.GOOS == "darwin" {
		systray.SetTemplateIcon(pngData, pngData)
		return
	}
	systray.SetIcon(icoData)
}

func toggleAutostart(item *systray.MenuItem, homeDir string) {
	if item.Checked() {
		if err := autostart.Disable(homeDir); err != nil {
			log.Printf("autostart: disable failed: %v", err)
			return
		}
		item.Uncheck()
		return
	}

	if err := autostart.Enable(homeDir); err != nil {
		log.Printf("autostart: enable failed: %v", err)
		return
	}
	item.Check()
}

func displayVersion(version string) string {
	if version == "" {
		return "dev"
	}
	return version
}

func openSyncFolder(homeDir string) {
	data, err := os.ReadFile(filepath.Join(homeDir, "config.xml"))
	if err != nil {
		return
	}
	var cfg struct {
		Folders []struct {
			Path string `xml:"path,attr"`
		} `xml:"folder"`
	}
	if err := xml.Unmarshal(data, &cfg); err != nil || len(cfg.Folders) == 0 {
		return
	}
	openPath(cfg.Folders[0].Path)
}

func openPath(path string) {
	switch runtime.GOOS {
	case "darwin":
		exec.Command("open", path).Start()
	case "windows":
		exec.Command("explorer", path).Start()
	default:
		exec.Command("xdg-open", path).Start()
	}
}

func copyToClipboard(text string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "windows":
		cmd = exec.Command("cmd", "/c", "clip")
	default:
		cmd = exec.Command("xclip", "-selection", "clipboard")
	}
	cmd.Stdin = strings.NewReader(text)
	cmd.Run()
}

func openInEditor(path string) {
	switch runtime.GOOS {
	case "darwin":
		exec.Command("open", "-t", path).Start()
	case "windows":
		exec.Command("notepad", path).Start()
	default:
		exec.Command("xdg-open", path).Start()
	}
}

func openAndroidPairingGuide(homeDir, deviceID string) {
	folderID, folderPath := loadPrimaryFolder(homeDir)
	guidePath := filepath.Join(homeDir, "pair-android-syncthing.txt")

	// Make pairing action useful immediately: ID is ready to paste on phone.
	copyToClipboard(deviceID)

	_ = os.WriteFile(guidePath, []byte(renderAndroidPairingGuide(deviceID, folderID, folderPath, filepath.Join(homeDir, "peers.txt"))), 0o644)
	openInEditor(guidePath)
}

func loadPrimaryFolder(homeDir string) (string, string) {
	folderID := "default"
	folderPath := defaultSyncFolderPath()

	data, err := os.ReadFile(filepath.Join(homeDir, "config.xml"))
	if err != nil {
		return folderID, folderPath
	}

	var cfg struct {
		Folders []struct {
			ID   string `xml:"id,attr"`
			Path string `xml:"path,attr"`
		} `xml:"folder"`
	}
	if err := xml.Unmarshal(data, &cfg); err != nil || len(cfg.Folders) == 0 {
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

func renderAndroidPairingGuide(deviceID, folderID, folderPath, peersPath string) string {
	return fmt.Sprintf(`Plop + Syncthing Android Pairing
=================================

Your plop device ID (already copied to clipboard):
%s

Plop shared folder:
- Folder ID: %s
- Folder path: %s

On Android (Syncthing app):
1. Find and copy your phone's Device ID:
   Devices tab -> This Device (phone) -> copy Device ID.
2. Add plop as a remote device:
   Devices tab -> + -> Enter Device ID -> paste the plop ID above -> Save.
3. Folder setup:
   - If you already have a folder with ID "%s":
     open it and share it with the plop device.
   - If you do NOT have that folder ID:
     Folders tab -> + -> create one with Folder ID "%s".
4. Ensure folder type is Send & Receive.
5. Choose any local folder path on phone.
6. Accept share prompts on both devices.

On this computer:
1. Add the Android Device ID (step 1 above) to:
   %s
2. One device ID per line.

Then verify:
- In tray: Open Plop Folder.
- Add a small test file and confirm it appears on Android.
`, deviceID, folderID, folderPath, folderID, folderID, peersPath)
}
