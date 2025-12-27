package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"

	"github.com/game_tool_box/internal/config"
	"github.com/game_tool_box/internal/i18n"
	"github.com/game_tool_box/internal/logging"
	"github.com/game_tool_box/internal/resources"
	aboutui "github.com/game_tool_box/internal/ui/about"
	"github.com/game_tool_box/internal/ui/pegasus"
	settingsui "github.com/game_tool_box/internal/ui/settings"
	"github.com/game_tool_box/internal/update"
)

func main() {
	logging.Init()
	logging.Infof("app start")
	defer func() {
		logging.Infof("app exit")
		logging.Close()
	}()

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
	router := container.NewStack(mainView)
	w.SetContent(router)

	t := func(key string) string { return i18n.T(i18n.Current(), key) }

	showMain := func() {
		router.Objects = []fyne.CanvasObject{mainView}
		router.Refresh()
	}

	var rebuildMenu func()
	var showSettings func()

	rebuildMenu = func() {
		w.SetTitle(t("app.title"))

		pegasusGameFileGenerator := fyne.NewMenuItem(t("menuitem.pegasus.gameFileGen"), func() {
			logging.Infof("menu click: pegasus.gameFileGen")
			view := tmgui.NewGeneratorView(w)
			// No Back button on pages entered from menu.
			router.Objects = []fyne.CanvasObject{view}
			router.Refresh()
		})

		showSettings = func() {
			logging.Infof("menu click: settings.settings")
			view := settingsui.NewSettingsView(t, func(newLang i18n.Lang) {
				logging.Infof("settings change: lang=%s", newLang)
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

		checkUpdate := func() {
			logging.Infof("menu click: help.update")
			progress := dialog.NewProgress(t("menuitem.help.update"), "...", w)
			progress.Show()
			progressInfinite := progress
			progressInfinite.SetValue(-1)

			go func() {
				info, err := update.LatestRelease(context.Background(), "jiangxincode", "GameToolBox")

				current := strings.TrimSpace(resources.Version)

				var msg string
				if err == nil {
					msg = fmt.Sprintf("Current: %s\nLatest:  %s", current, info.TagName)
					if info.HTMLURL != "" {
						msg += "\n\n" + info.HTMLURL
					}
				}

				fyne.Do(func() {
					progress.Hide()
					if err != nil {
						dialog.ShowError(err, w)
						return
					}
					dialog.ShowInformation(t("menuitem.help.update"), msg, w)
				})
			}()
		}

		mHelp := fyne.NewMenu(t("menu.help"),
			fyne.NewMenuItem(t("menuitem.help.docs"), func() {
				logging.Infof("menu click: help.docs")
				u, err := url.Parse("https://jiangxincode.github.io/GameToolBox")
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				_ = a.OpenURL(u)
			}),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem(t("menuitem.help.feedback"), func() {
				logging.Infof("menu click: help.feedback")
				u, err := url.Parse("https://github.com/jiangxincode/GameToolBox/issues/new")
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				_ = a.OpenURL(u)
			}),
			fyne.NewMenuItem(t("menuitem.help.update"), checkUpdate),
			fyne.NewMenuItem(t("menuitem.help.contrib"), func() {
				logging.Infof("menu click: help.contrib")
				u, err := url.Parse("https://github.com/jiangxincode/GameToolBox")
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				_ = a.OpenURL(u)
			}),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem(t("menuitem.help.about"), func() {
				logging.Infof("menu click: help.about")
				showMain()
			}),
		)

		w.SetMainMenu(fyne.NewMainMenu(mSettings, mPegasus, mHelp))
	}

	rebuildMenu()

	resizeAndCenter := func(size fyne.Size) {
		w.Resize(size)
		w.CenterOnScreen()
	}

	// Initial window size and center.
	resizeAndCenter(fyne.NewSize(900, 650))
	w.ShowAndRun()
}
