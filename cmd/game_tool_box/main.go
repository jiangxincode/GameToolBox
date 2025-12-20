package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"example.com/game_tool_box/internal/resources"
)

func main() {
	a := app.New()
	w := a.NewWindow("game_tool_box")

	// Set main window icon
	w.SetIcon(resources.IconPng)

	tianmaGEmptyGameFileGenerator := fyne.NewMenuItem("天马G空游戏文件生成器", func() {
		dialog.ShowInformation(
			"天马G空游戏文件生成器",
			"这里将来会放置：天马G空游戏文件生成器。",
			w,
		)
	})

	// 关键点：提供一个显式的“设置”菜单（可先留空），避免框架在某些平台将默认 Quit/Exit
	// 菜单项“塞进”第一个菜单（你这里就是“天马G”）。
	mSettings := fyne.NewMenu("设置")

	mTianmaG := fyne.NewMenu("天马G", tianmaGEmptyGameFileGenerator)
	mHelp := fyne.NewMenu("帮助",
		fyne.NewMenuItem("关于", func() {
			dialog.ShowInformation("关于", "game_tool_box (Fyne)", w)
		}),
	)

	w.SetMainMenu(fyne.NewMainMenu(mSettings, mTianmaG, mHelp))

	w.SetContent(container.NewVBox(
		widget.NewLabel("game_tool_box"),
		widget.NewLabel("Fyne GUI skeleton is ready."),
	))

	w.Resize(fyne.NewSize(480, 320))
	w.ShowAndRun()
}
