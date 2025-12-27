package main

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"

	"github.com/game_tool_box/internal/config"
	"github.com/game_tool_box/internal/i18n"
	"github.com/game_tool_box/internal/resources"
	aboutui "github.com/game_tool_box/internal/ui/about"
	"github.com/game_tool_box/internal/ui/pegasus"
	settingsui "github.com/game_tool_box/internal/ui/settings"
)

func main() {
	// i18n.Current() is initialized in i18n init() (including persisted config).
	i18n.SetCurrent(i18n.Current())

	a := app.New()
	w := a.NewWindow(i18n.T(i18n.Current(), "app.title"))

	// Apply persisted theme at startup.
	if c, err := config.Load(); err == nil {
		switch strings.ToLower(strings.TrimSpace(c.Theme)) {
		case "light":
			a.Settings().SetTheme(theme.LightTheme())
		case "dark":
			a.Settings().SetTheme(theme.DarkTheme())
		default:
			// "system" or empty => follow system preference.
			a.Settings().SetTheme(theme.DefaultTheme())
		}
	}

	w.SetIcon(resources.IconPng)

	// Default startup view: About page rendered from Markdown.
	aboutMD := strings.ReplaceAll(resources.AboutMarkdown(), "{VERSION}", resources.Version)
	mainView := aboutui.New(aboutMD)

	// Simple view router
	router := container.NewMax(mainView)
	w.SetContent(router)

	t := func(key string) string { return i18n.T(i18n.Current(), key) }

	showTodo := func(title string) {
		dialog.ShowInformation(title, t("todo.prefix")+title+"。", w)
	}

	showMain := func() {
		router.Objects = []fyne.CanvasObject{mainView}
		router.Refresh()
	}

	var rebuildMenu func()
	var showSettings func()

	rebuildMenu = func() {
		w.SetTitle(t("app.title"))

		pegasusGameFileGenerator := fyne.NewMenuItem(t("menuitem.pegasus.gameFileGen"), func() {
			view := tmgui.NewGeneratorView(w)
			// No Back button on pages entered from menu.
			router.Objects = []fyne.CanvasObject{view}
			router.Refresh()
		})

		showSettings = func() {
			view := settingsui.NewSettingsView(t, func(newLang i18n.Lang) {
				// Switch language and rebuild menus immediately.
				i18n.SetCurrentPersisted(newLang)
				rebuildMenu()
				// NOTE: do NOT call showSettings() here; the settings view refreshes itself.
			})
			// No Back button on pages entered from menu.
			router.Objects = []fyne.CanvasObject{container.NewPadded(view)}
			router.Refresh()
		}

		mSettings := fyne.NewMenu(t("menu.settings"),
			fyne.NewMenuItem(t("menuitem.settings.settings"), showSettings),
		)

		mPegasus := fyne.NewMenu(t("menu.pegasus"), pegasusGameFileGenerator)

		mHelp := fyne.NewMenu(t("menu.help"),
			fyne.NewMenuItem(t("menuitem.help.docs"), func() { showTodo(t("menuitem.help.docs")) }),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem(t("menuitem.help.feedback"), func() { showTodo(t("menuitem.help.feedback")) }),
			fyne.NewMenuItem(t("menuitem.help.update"), func() { showTodo(t("menuitem.help.update")) }),
			fyne.NewMenuItem(t("menuitem.help.contrib"), func() { showTodo(t("menuitem.help.contrib")) }),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem(t("menuitem.help.about"), showMain),
		)

		w.SetMainMenu(fyne.NewMainMenu(mSettings, mPegasus, mHelp))
	}

	rebuildMenu()

	w.Resize(fyne.NewSize(900, 650))
	w.ShowAndRun()
}
