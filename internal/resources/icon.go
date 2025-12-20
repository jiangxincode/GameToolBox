package resources

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

// IconPng is the application/window icon.
//
//go:embed icon.png
var iconPngBytes []byte

var IconPng fyne.Resource = fyne.NewStaticResource("icon.png", iconPngBytes)
