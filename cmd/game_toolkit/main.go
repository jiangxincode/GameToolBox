package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("game_toolkit")

	w.SetContent(container.NewVBox(
		widget.NewLabel("game_toolkit"),
		widget.NewLabel("Fyne GUI skeleton is ready."),
	))

	w.Resize(fyne.NewSize(480, 320))
	w.ShowAndRun()
}
