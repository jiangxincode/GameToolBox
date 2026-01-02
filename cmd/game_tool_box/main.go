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

	i18n.SetCurrent(i18n.Current())

	fyneApp := app.New()
	fyneWindow := fyneApp.NewWindow(i18n.T(i18n.Current(), "app.title"))

	appTitle := func() string { return i18n.T(i18n.Current(), "app.title") }
	setBreadcrumb := func(crumb string) {
		crumb = strings.TrimSpace(crumb)
		if crumb == "" {
			fyneWindow.SetTitle(appTitle())
			return
		}
		fyneWindow.SetTitle(appTitle() + " - " + crumb)
	}

	if cfg, err := config.Load(); err == nil {
		switch strings.ToLower(strings.TrimSpace(cfg.Theme)) {
		case "light":
			fyneApp.Settings().SetTheme(theme.LightTheme())
		case "dark":
			fyneApp.Settings().SetTheme(theme.DarkTheme())
		default:
			// "system" or empty => follow system preference.
			fyneApp.Settings().SetTheme(theme.DefaultTheme())
		}
	}

	fyneWindow.SetIcon(resources.IconPng)

	// Default startup view: About page rendered from Markdown.
	aboutMD := strings.ReplaceAll(resources.AboutMarkdown(), "{VERSION}", resources.Version)
	mainView := aboutui.New(aboutMD)

	router := container.NewStack(mainView)
	fyneWindow.SetContent(router)

	t := func(key string) string { return i18n.T(i18n.Current(), key) }

	var rebuildMenu func()
	var showSettings func()

	rebuildMenu = func() {
		// Keep current breadcrumb if any by default: menu rebuild updates app title only.
		fyneWindow.SetTitle(appTitle())

		mPegasus := fyne.NewMenu(t("menu.pegasus"))
		mSettings := fyne.NewMenu(t("menu.settings"))
		mHelp := fyne.NewMenu(t("menu.help"))

		pegasusRomManager := fyne.NewMenuItem(t("menuitem.pegasus.romManager"), func() {
			logging.Infof("menu click: pegasus.romManager")
			setBreadcrumb(t("menu.pegasus") + " / " + t("menuitem.pegasus.romManager"))
			view := pegasusui.NewRomeManagerView(fyneWindow)
			router.Objects = []fyne.CanvasObject{view}
			router.Refresh()
		})
		mPegasus.Items = append(mPegasus.Items, pegasusRomManager)

		pegasusMediaManager := fyne.NewMenuItem(t("menuitem.pegasus.mediaManager"), func() {
			logging.Infof("menu click: pegasus.mediaManager")
			setBreadcrumb(t("menu.pegasus") + " / " + t("menuitem.pegasus.mediaManager"))
			view := pegasusui.NewMediaManagerView(fyneWindow)
			router.Objects = []fyne.CanvasObject{view}
			router.Refresh()
		})
		mPegasus.Items = append(mPegasus.Items, pegasusMediaManager)

		pegasusConfigManager := fyne.NewMenuItem(t("menuitem.pegasus.configManager"), func() {
			logging.Infof("menu click: pegasus.configManager")
			setBreadcrumb(t("menu.pegasus") + " / " + t("menuitem.pegasus.configManager"))
			view := pegasusui.NewConfigManagerView(fyneWindow)
			router.Objects = []fyne.CanvasObject{view}
			router.Refresh()
		})
		mPegasus.Items = append(mPegasus.Items, pegasusConfigManager)

		pegasusGameRemover := fyne.NewMenuItem(t("menuitem.pegasus.gameRemover"), func() {
			logging.Infof("menu click: pegasus.gameRemover")
			setBreadcrumb(t("menu.pegasus") + " / " + t("menuitem.pegasus.gameRemover"))
			view := pegasusui.NewGameRemoverView(fyneWindow)
			router.Objects = []fyne.CanvasObject{view}
			router.Refresh()
		})
		mPegasus.Items = append(mPegasus.Items, pegasusGameRemover)

		showSettings = func() {
			logging.Infof("menu click: settings.settings")
			setBreadcrumb(t("menu.settings") + " / " + t("menuitem.settings.settings"))
			view := settingsui.NewSettingsView(
				t,
				func(newLang i18n.Lang) {
					logging.Infof("settings change: lang=%s", newLang)
					i18n.SetCurrentPersisted(newLang)
				},
				func() {
					// Rebuild and re-set menus after theme/lang changes.
					rebuildMenu()
					setBreadcrumb(t("menu.settings") + " / " + t("menuitem.settings.settings"))
				},
			)
			router.Objects = []fyne.CanvasObject{container.NewPadded(view)}
			router.Refresh()
		}
		mSettings.Items = append(mSettings.Items, fyne.NewMenuItem(t("menuitem.settings.settings"), showSettings))

		goDocs := fyne.NewMenuItem(t("menuitem.help.docs"), func() {
			logging.Infof("menu click: help.docs")
			u, err := url.Parse("https://jiangxincode.github.io/GameToolBox")
			if err != nil {
				dialog.ShowError(err, fyneWindow)
				return
			}
			_ = fyneApp.OpenURL(u)
		})
		mHelp.Items = append(mHelp.Items, goDocs)
		mHelp.Items = append(mHelp.Items, fyne.NewMenuItemSeparator())

		feedback := fyne.NewMenuItem(t("menuitem.help.feedback"), func() {
			logging.Infof("menu click: help.feedback")
			u, err := url.Parse("https://github.com/jiangxincode/GameToolBox/issues/new")
			if err != nil {
				dialog.ShowError(err, fyneWindow)
				return
			}
			_ = fyneApp.OpenURL(u)
		})
		mHelp.Items = append(mHelp.Items, feedback)

		checkUpdate := func() {
			logging.Infof("menu click: help.update")
			progress := dialog.NewProgress(t("menuitem.help.update"), "...", fyneWindow)
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
						dialog.ShowError(err, fyneWindow)
						return
					}
					dialog.ShowInformation(t("menuitem.help.update"), msg, fyneWindow)
				})
			}()
		}
		updateItem := fyne.NewMenuItem(t("menuitem.help.update"), checkUpdate)
		mHelp.Items = append(mHelp.Items, updateItem)

		contrib := fyne.NewMenuItem(t("menuitem.help.contrib"), func() {
			logging.Infof("menu click: help.contrib")
			u, err := url.Parse("https://github.com/jiangxincode/GameToolBox")
			if err != nil {
				dialog.ShowError(err, fyneWindow)
				return
			}
			_ = fyneApp.OpenURL(u)
		})
		mHelp.Items = append(mHelp.Items, contrib)
		mHelp.Items = append(mHelp.Items, fyne.NewMenuItemSeparator())

		showAbout := func() {
			logging.Infof("menu click: help.about")
			setBreadcrumb(t("menu.help") + " / " + t("menuitem.help.about"))
			router.Objects = []fyne.CanvasObject{mainView}
			router.Refresh()
		}
		about := fyne.NewMenuItem(t("menuitem.help.about"), showAbout)
		mHelp.Items = append(mHelp.Items, about)

		// IMPORTANT: Set the main menu only after items are fully populated.
		fyneWindow.SetMainMenu(fyne.NewMainMenu(mSettings, mPegasus, mHelp))
	}

	rebuildMenu()
	setBreadcrumb(t("menu.help") + " / " + t("menuitem.help.about"))

	resizeAndCenter := func(size fyne.Size) {
		fyneWindow.Resize(size)
		fyneWindow.CenterOnScreen()
	}

	resizeAndCenter(fyne.NewSize(900, 650))
	fyneWindow.ShowAndRun()
}
