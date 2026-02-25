package tray

import (
	"encoding/xml"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alex-vit/plop/icon"
	"github.com/energye/systray"
)

// Run blocks on the calling goroutine, showing the system tray icon.
func Run(version, homeDir, deviceID string) {
	systray.Run(func() { onReady(version, homeDir, deviceID) }, onExit)
}

func onReady(version, homeDir, deviceID string) {
	if runtime.GOOS == "darwin" {
		systray.SetTemplateIcon(icon.Data, icon.Data)
	} else {
		systray.SetIcon(icon.DataICO)
	}
	systray.SetTooltip("plop")
	systray.SetOnClick(func(menu systray.IMenu) { menu.ShowMenu() })
	systray.SetOnRClick(func(menu systray.IMenu) { menu.ShowMenu() })
	if runtime.GOOS == "windows" || runtime.GOOS == "linux" {
		systray.SetOnDClick(func(menu systray.IMenu) { openSyncFolder(homeDir) })
	}

	mTitle := systray.AddMenuItem("plop "+displayVersion(version), "")
	mTitle.Disable()

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

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Exit", "Quit plop")
	mQuit.Click(func() { systray.Quit() })
}

func onExit() {}

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

