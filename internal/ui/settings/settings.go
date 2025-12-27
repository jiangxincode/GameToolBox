package settingsui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/game_tool_box/internal/config"
	"github.com/game_tool_box/internal/i18n"
)

// NewSettingsView returns the settings page.
//
// Contract:
//   - onLangChanged: will be called after user selects a language.
//   - t: translation function for current language.
func NewSettingsView(t func(key string) string, onLangChanged func(lang i18n.Lang)) fyne.CanvasObject {
	// Widgets we need to update on language/theme switch.
	title := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	langSelect := widget.NewSelect(nil, nil)
	languageItem := widget.NewFormItem("", langSelect)

	themeSelect := widget.NewSelect(nil, nil)
	themeItem := widget.NewFormItem("", themeSelect)

	form := widget.NewForm(languageItem, themeItem)

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
			if !ok {
				return
			}
			if lang == i18n.Current() {
				return
			}
			onLangChanged(lang)
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
				// "system": use default theme (follows system preference).
				fyne.CurrentApp().Settings().SetTheme(theme.DefaultTheme())
				key = "system"
			}

			// Persist.
			c, _ := config.Load()
			c.Theme = key
			_ = config.Save(c)
		}

		// Load persisted theme selection.
		persistedTheme := "system"
		if c, err := config.Load(); err == nil {
			if v := strings.ToLower(strings.TrimSpace(c.Theme)); v != "" {
				persistedTheme = v
			}
		}
		if lbl, ok := labelByThemeKey[persistedTheme]; ok {
			themeSelect.Selected = lbl
			themeSelect.Refresh()
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
		container.NewHBox(widget.NewLabel(t("settings.theme")), layout.NewSpacer(), widget.NewLabel("")),
	)
}
