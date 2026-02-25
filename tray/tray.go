package tray

import (
	"encoding/xml"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/alex-vit/plop/icon"
	"github.com/energye/systray"
)

// Run blocks on the calling goroutine, showing the system tray icon.
func Run(version, homeDir string) {
	systray.Run(func() { onReady(version, homeDir) }, onExit)
}

func onReady(version, homeDir string) {
	systray.SetTemplateIcon(icon.Data, icon.Data)
	systray.SetTooltip("plop")
	systray.SetOnClick(func(menu systray.IMenu) { menu.ShowMenu() })
	systray.SetOnRClick(func(menu systray.IMenu) { menu.ShowMenu() })

	mTitle := systray.AddMenuItem("plop "+displayVersion(version), "")
	mTitle.Disable()

	systray.AddSeparator()

	mFolder := systray.AddMenuItem("Open Sync Folder", "Open synced folder in file manager")
	mFolder.Click(func() { openSyncFolder(homeDir) })

	mSettings := systray.AddMenuItem("Open Settings", "Open config in text editor")
	mSettings.Click(func() { openInEditor(filepath.Join(homeDir, "config.xml")) })

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
