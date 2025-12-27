package about

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// New returns a scrollable "About" view rendered from Markdown.
func New(aboutMarkdown string) fyne.CanvasObject {
	aboutMarkdown = strings.TrimSpace(aboutMarkdown)
	if aboutMarkdown == "" {
		aboutMarkdown = "(empty)"
	}

	rt := widget.NewRichTextFromMarkdown(aboutMarkdown)
	rt.Wrapping = fyne.TextWrapWord
	return container.NewScroll(container.NewPadded(rt))
}
