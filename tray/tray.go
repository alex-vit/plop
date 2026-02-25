package tray

import (
	"github.com/alex-vit/plop/icon"
	"github.com/energye/systray"
)

// Run blocks on the calling goroutine, showing the system tray icon.
func Run(version string) {
	systray.Run(func() { onReady(version) }, onExit)
}

func onReady(version string) {
	systray.SetTemplateIcon(icon.Data, icon.Data)
	systray.SetTooltip("plop")
	systray.SetOnClick(func(menu systray.IMenu) { menu.ShowMenu() })
	systray.SetOnRClick(func(menu systray.IMenu) { menu.ShowMenu() })

	mTitle := systray.AddMenuItem("plop "+displayVersion(version), "")
	mTitle.Disable()

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
