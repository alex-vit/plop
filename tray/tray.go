package tray

import (
	"encoding/xml"
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
	if runtime.GOOS != "windows" {
		systray.SetOnClick(func(menu systray.IMenu) { menu.ShowMenu() })
	}
	systray.SetOnRClick(func(menu systray.IMenu) { menu.ShowMenu() })
	if runtime.GOOS == "windows" || runtime.GOOS == "linux" {
		systray.SetOnDClick(func(menu systray.IMenu) { openSyncFolder(homeDir) })
	}

	mTitle := systray.AddMenuItem("plop "+displayVersion(version), "")
	mTitle.Disable()

	mStatus := systray.AddMenuItem("Status: Starting...", "Current sync status")
	mStatus.Disable()

	// Pre-allocate peer slots immediately after the status item so they appear in-line.
	// The status monitor shows/hides/updates them as snapshots arrive.
	const maxPeerItems = 10
	peerItems := make([]*systray.MenuItem, maxPeerItems)
	for i := range peerItems {
		item := systray.AddMenuItem("", "")
		item.Disable()
		item.Hide()
		peerItems[i] = item
	}

	stopStatusMonitor = startStatusMonitor(statusUpdates, mStatus, peerItems)

	systray.AddSeparator()

	mFolder := systray.AddMenuItem("Open Plop Folder", "Open synced folder in file manager")
	mFolder.Click(func() { openSyncFolder(homeDir) })

	systray.AddSeparator()

	mCopyID := systray.AddMenuItem("Copy My ID", "Copy this device's ID to clipboard")
	mCopyID.Click(func() { copyToClipboard(deviceID) })

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
		systray.SetIcon(pngData)
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

