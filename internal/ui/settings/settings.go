package settingsui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/game_tool_box/internal/config"
	"github.com/game_tool_box/internal/i18n"
)

// NewSettingsView returns the settings page.
//
// Contract:
//   - onLangChanged: called after user selects a new language (you should persist it).
//   - onAppUIRefresh: optional; called after language/theme changes to refresh window-level UI (e.g. menus).
//   - t: translation function for current language.
func NewSettingsView(
	t func(key string) string,
	onLangChanged func(lang i18n.Lang),
	onAppUIRefresh func(),
) fyne.CanvasObject {
	// Widgets we need to update on language/theme switch.
	title := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	langSelect := widget.NewSelect(nil, nil)
	languageItem := widget.NewFormItem("", langSelect)

	themeSelect := widget.NewSelect(nil, nil)
	themeItem := widget.NewFormItem("", themeSelect)

	// --- Media scraping settings ---
	// ScreenScraper credentials.
	scDevID := widget.NewEntry()
	scDevID.SetPlaceHolder("ScreenScraper DevID")
	scDevPass := widget.NewPasswordEntry()
	scDevPass.SetPlaceHolder("ScreenScraper DevPassword")
	scUser := widget.NewEntry()
	scUser.SetPlaceHolder("(optional) ScreenScraper user")
	scPass := widget.NewPasswordEntry()
	scPass.SetPlaceHolder("(optional) ScreenScraper password")

	scDevIDItem := widget.NewFormItem("SS DevID", scDevID)
	scDevPassItem := widget.NewFormItem("SS DevPassword", scDevPass)
	scUserItem := widget.NewFormItem("SS 用户", scUser)
	scPassItem := widget.NewFormItem("SS 密码", scPass)

	mediaOverwriteCheck := widget.NewCheck("覆盖已存在文件", nil)
	mediaOverwriteItem := widget.NewFormItem("媒体刮削", mediaOverwriteCheck)

	form := widget.NewForm(
		languageItem,
		themeItem,
		widget.NewFormItem("", widget.NewSeparator()),
		scDevIDItem,
		scDevPassItem,
		scUserItem,
		scPassItem,
		mediaOverwriteItem,
	)

	updating := false

	var refresh func()
	refresh = func() {
		updating = true
		defer func() { updating = false }()

		// --- Language options ---
		supported := i18n.Supported()
		langOptions := make([]string, 0, len(supported))
		langByLabel := map[string]i18n.Lang{}
		labelByLang := map[i18n.Lang]string{}
		for _, l := range supported {
			label := i18n.LangName(i18n.Current(), l)
			langOptions = append(langOptions, label)
			langByLabel[label] = l
			labelByLang[l] = label
		}

		langSelect.Options = langOptions
		langSelect.OnChanged = func(selected string) {
			if updating {
				return
			}
			lang, ok := langByLabel[selected]
			if !ok || lang == i18n.Current() {
				return
			}
			onLangChanged(lang)
			if onAppUIRefresh != nil {
				onAppUIRefresh()
			}
			refresh()
		}

		if curLabel, ok := labelByLang[i18n.Current()]; ok {
			langSelect.Selected = curLabel
			langSelect.Refresh()
		}

		// --- Theme options ---
		themeOptions := []string{t("theme.system"), t("theme.light"), t("theme.dark")}
		themeKeyByLabel := map[string]string{
			t("theme.system"): "system",
			t("theme.light"):  "light",
			t("theme.dark"):   "dark",
		}
		labelByThemeKey := map[string]string{
			"system": t("theme.system"),
			"light":  t("theme.light"),
			"dark":   t("theme.dark"),
		}

		themeSelect.Options = themeOptions
		themeSelect.OnChanged = func(selected string) {
			if updating {
				return
			}
			key, ok := themeKeyByLabel[selected]
			if !ok {
				return
			}

			// Apply theme.
			switch key {
			case "light":
				fyne.CurrentApp().Settings().SetTheme(theme.LightTheme())
			case "dark":
				fyne.CurrentApp().Settings().SetTheme(theme.DarkTheme())
			default:
				fyne.CurrentApp().Settings().SetTheme(theme.DefaultTheme())
				key = "system"
			}

			// Persist.
			c, _ := config.Load()
			c.Theme = key
			_ = config.Save(c)

			if onAppUIRefresh != nil {
				onAppUIRefresh()
			}
		}

		// Load persisted config selections.
		persistedTheme := "system"
		if c, err := config.Load(); err == nil {
			if v := strings.ToLower(strings.TrimSpace(c.Theme)); v != "" {
				persistedTheme = v
			}
			// ScreenScraper persisted settings.
			scDevID.SetText(strings.TrimSpace(c.ScreenScraperDevID))
			scDevPass.SetText(strings.TrimSpace(c.ScreenScraperDevPassword))
			scUser.SetText(strings.TrimSpace(c.ScreenScraperUser))
			scPass.SetText(strings.TrimSpace(c.ScreenScraperPassword))
			mediaOverwriteCheck.SetChecked(c.MediaScrapeOverwrite)
		}
		if lbl, ok := labelByThemeKey[persistedTheme]; ok {
			themeSelect.Selected = lbl
			themeSelect.Refresh()
		}

		scDevID.OnChanged = func(s string) {
			if updating {
				return
			}
			c, _ := config.Load()
			c.ScreenScraperDevID = strings.TrimSpace(s)
			_ = config.Save(c)
		}
		scDevPass.OnChanged = func(s string) {
			if updating {
				return
			}
			c, _ := config.Load()
			c.ScreenScraperDevPassword = strings.TrimSpace(s)
			_ = config.Save(c)
		}
		scUser.OnChanged = func(s string) {
			if updating {
				return
			}
			c, _ := config.Load()
			c.ScreenScraperUser = strings.TrimSpace(s)
			_ = config.Save(c)
		}
		scPass.OnChanged = func(s string) {
			if updating {
				return
			}
			c, _ := config.Load()
			c.ScreenScraperPassword = strings.TrimSpace(s)
			_ = config.Save(c)
		}
		mediaOverwriteCheck.OnChanged = func(b bool) {
			if updating {
				return
			}
			c, _ := config.Load()
			c.MediaScrapeOverwrite = b
			_ = config.Save(c)
		}

		// --- Labels ---
		title.SetText(t("page.settings.title"))
		languageItem.Text = t("settings.language")
		themeItem.Text = t("settings.theme")
		form.Refresh()
	}

	refresh()

	return container.NewVBox(
		title,
		widget.NewSeparator(),
		form,
	)
}
