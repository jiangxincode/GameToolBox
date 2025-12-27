package settingsui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/game_tool_box/internal/i18n"
)

// NewSettingsView returns the settings page.
//
// Contract:
//   - onLangChanged: will be called after user selects a language.
//   - t: translation function for current language.
func NewSettingsView(t func(key string) string, onLangChanged func(lang i18n.Lang)) fyne.CanvasObject {
	// Widgets we need to update on language switch.
	title := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	langSelect := widget.NewSelect(nil, nil)
	languageItem := widget.NewFormItem("", langSelect)
	form := widget.NewForm(languageItem)
	themeLabel := widget.NewLabel("")
	themeTodo := widget.NewLabel("")
	themeRow := container.NewHBox(themeLabel, layout.NewSpacer(), themeTodo)

	updating := false

	var refresh func()
	refresh = func() {
		updating = true
		defer func() { updating = false }()

		// Rebuild language options using current UI language.
		supported := i18n.Supported()
		options := make([]string, 0, len(supported))
		langByLabel := map[string]i18n.Lang{}
		labelByLang := map[i18n.Lang]string{}
		for _, l := range supported {
			label := i18n.LangName(i18n.Current(), l)
			options = append(options, label)
			langByLabel[label] = l
			labelByLang[l] = label
		}

		langSelect.Options = options
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
			// Caller should have switched i18n.Current(); now update labels/options.
			refresh()
		}

		// Select current without triggering change logic.
		if curLabel, ok := labelByLang[i18n.Current()]; ok {
			langSelect.Selected = curLabel
			langSelect.Refresh()
		}

		title.SetText(t("page.settings.title"))
		languageItem.Text = t("settings.language")
		themeLabel.SetText(t("settings.theme"))
		themeTodo.SetText(t("settings.theme.todo"))
		form.Refresh()
	}

	refresh()

	return container.NewVBox(
		title,
		widget.NewSeparator(),
		form,
		themeRow,
	)
}
