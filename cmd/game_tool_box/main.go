package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/game_tool_box/internal/i18n"
	"github.com/game_tool_box/internal/logging"
	appui "github.com/game_tool_box/internal/ui/app"
)

func main() {
	logging.Init()
	logging.Infof("app start")
	defer func() {
		logging.Infof("app exit")
		logging.Close()
	}()

	i18n.SetCurrent(i18n.Current())

	fyneApp := app.New()
	fyneWindow := fyneApp.NewWindow(i18n.T(i18n.Current(), "app.title"))

	_ = appui.NewShell(fyneApp, fyneWindow)

	fyneWindow.Resize(fyne.NewSize(900, 650))
	fyneWindow.CenterOnScreen()
	fyneWindow.ShowAndRun()
}
