package appui

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"

	"github.com/game_tool_box/internal/config"
	"github.com/game_tool_box/internal/i18n"
	"github.com/game_tool_box/internal/logging"
	"github.com/game_tool_box/internal/resources"
	aboutui "github.com/game_tool_box/internal/ui/about"
	pegasusui "github.com/game_tool_box/internal/ui/pegasus"
	settingsui "github.com/game_tool_box/internal/ui/settings"
	"github.com/game_tool_box/internal/update"
)

// Shell wires up the main window content router and menus.
//
// Contract:
//   - NewShell does not call ShowAndRun; caller owns app lifecycle.
//   - Shell methods are expected to be called from the main UI thread.
type Shell struct {
	app fyne.App
	w   fyne.Window

	router   *fyne.Container
	mainView fyne.CanvasObject

	t func(key string) string
}

func NewShell(app fyne.App, w fyne.Window) *Shell {
	// Default startup view: About page rendered from Markdown.
	aboutMD := strings.ReplaceAll(resources.AboutMarkdown(), "{VERSION}", resources.Version)
	mainView := aboutui.New(aboutMD)

	router := container.NewStack(mainView)
	w.SetContent(router)
	w.SetIcon(resources.IconPng)

	s := &Shell{
		app:      app,
		w:        w,
		router:   router,
		mainView: mainView,
		t: func(key string) string {
			return i18n.T(i18n.Current(), key)
		},
	}

	s.applyPersistedTheme()
	s.rebuildMenu()
	s.setBreadcrumb(s.t("menu.help") + " / " + s.t("menuitem.help.about"))

	return s
}

func (s *Shell) applyPersistedTheme() {
	if cfg, err := config.Load(); err == nil {
		switch strings.ToLower(strings.TrimSpace(cfg.Theme)) {
		case "light":
			s.app.Settings().SetTheme(theme.LightTheme())
		case "dark":
			s.app.Settings().SetTheme(theme.DarkTheme())
		default:
			// "system" or empty => follow system preference.
			s.app.Settings().SetTheme(theme.DefaultTheme())
		}
	}
}

func (s *Shell) appTitle() string {
	return s.t("app.title")
}

func (s *Shell) setBreadcrumb(crumb string) {
	crumb = strings.TrimSpace(crumb)
	if crumb == "" {
		s.w.SetTitle(s.appTitle())
		return
	}
	s.w.SetTitle(s.appTitle() + " - " + crumb)
}

func (s *Shell) navigate(crumb string, view fyne.CanvasObject) {
	s.setBreadcrumb(crumb)
	s.router.Objects = []fyne.CanvasObject{view}
	s.router.Refresh()
}

func (s *Shell) openURL(raw string) {
	u, err := url.Parse(raw)
	if err != nil {
		dialog.ShowError(err, s.w)
		return
	}
	_ = s.app.OpenURL(u)
}

func (s *Shell) rebuildMenu() {
	// Reset to base title when rebuilding menus; navigation will set breadcrumb.
	s.w.SetTitle(s.appTitle())

	mPegasus := s.buildPegasusMenu()
	mSettings := s.buildSettingsMenu()
	mHelp := s.buildHelpMenu()

	// IMPORTANT: Set the main menu only after items are fully populated.
	s.w.SetMainMenu(fyne.NewMainMenu(mSettings, mPegasus, mHelp))
}

func (s *Shell) buildPegasusMenu() *fyne.Menu {
	m := fyne.NewMenu(s.t("menu.pegasus"))
	m.Items = append(m.Items,
		fyne.NewMenuItem(s.t("menuitem.pegasus.romManager"), func() {
			logging.Infof("menu click: pegasus.romManager")
			view := pegasusui.NewRomeManagerView(s.w)
			s.navigate(s.t("menu.pegasus")+" / "+s.t("menuitem.pegasus.romManager"), view)
		}),
		fyne.NewMenuItem(s.t("menuitem.pegasus.mediaManager"), func() {
			logging.Infof("menu click: pegasus.mediaManager")
			view := pegasusui.NewMediaManagerView(s.w)
			s.navigate(s.t("menu.pegasus")+" / "+s.t("menuitem.pegasus.mediaManager"), view)
		}),
		fyne.NewMenuItem(s.t("menuitem.pegasus.configManager"), func() {
			logging.Infof("menu click: pegasus.configManager")
			view := pegasusui.NewConfigManagerView(s.w)
			s.navigate(s.t("menu.pegasus")+" / "+s.t("menuitem.pegasus.configManager"), view)
		}),
		fyne.NewMenuItem(s.t("menuitem.pegasus.gameRemover"), func() {
			logging.Infof("menu click: pegasus.gameRemover")
			view := pegasusui.NewGameRemoverView(s.w)
			s.navigate(s.t("menu.pegasus")+" / "+s.t("menuitem.pegasus.gameRemover"), view)
		}),
		fyne.NewMenuItem(s.t("menuitem.pegasus.gameAdder"), func() {
			logging.Infof("menu click: pegasus.gameAdder")
			view := pegasusui.NewGameAdderView(s.w)
			s.navigate(s.t("menu.pegasus")+" / "+s.t("menuitem.pegasus.gameAdder"), view)
		}),
	)
	return m
}

func (s *Shell) buildSettingsMenu() *fyne.Menu {
	m := fyne.NewMenu(s.t("menu.settings"))

	showSettings := func() {
		logging.Infof("menu click: settings.settings")
		view := settingsui.NewSettingsView(
			s.t,
			func(newLang i18n.Lang) {
				logging.Infof("settings change: lang=%s", newLang)
				i18n.SetCurrentPersisted(newLang)
			},
			func() {
				// Rebuild and re-set menus after theme/lang changes.
				s.rebuildMenu()
				// Keep user on settings page.
				s.setBreadcrumb(s.t("menu.settings") + " / " + s.t("menuitem.settings.settings"))
			},
		)

		s.navigate(s.t("menu.settings")+" / "+s.t("menuitem.settings.settings"), container.NewPadded(view))
	}

	m.Items = append(m.Items, fyne.NewMenuItem(s.t("menuitem.settings.settings"), showSettings))
	return m
}

func (s *Shell) buildHelpMenu() *fyne.Menu {
	m := fyne.NewMenu(s.t("menu.help"))

	m.Items = append(m.Items,
		fyne.NewMenuItem(s.t("menuitem.help.docs"), func() {
			logging.Infof("menu click: help.docs")
			s.openURL("https://jiangxincode.github.io/GameToolBox")
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem(s.t("menuitem.help.feedback"), func() {
			logging.Infof("menu click: help.feedback")
			s.openURL("https://github.com/jiangxincode/GameToolBox/issues/new")
		}),
	)

	m.Items = append(m.Items, fyne.NewMenuItem(s.t("menuitem.help.update"), s.menuCheckUpdate()))

	m.Items = append(m.Items,
		fyne.NewMenuItem(s.t("menuitem.help.contrib"), func() {
			logging.Infof("menu click: help.contrib")
			s.openURL("https://github.com/jiangxincode/GameToolBox")
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem(s.t("menuitem.help.about"), func() {
			logging.Infof("menu click: help.about")
			s.navigate(s.t("menu.help")+" / "+s.t("menuitem.help.about"), s.mainView)
		}),
	)

	return m
}

func (s *Shell) menuCheckUpdate() func() {
	return func() {
		logging.Infof("menu click: help.update")
		progress := dialog.NewProgress(s.t("menuitem.help.update"), "...", s.w)
		progress.Show()
		progress.SetValue(-1)

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
					dialog.ShowError(err, s.w)
					return
				}
				dialog.ShowInformation(s.t("menuitem.help.update"), msg, s.w)
			})
		}()
	}
}
