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

	// Common placeholder for not-yet-implemented menu items
	showTodo := func(title string) {
		dialog.ShowInformation(title, "这里将来会放置："+title+"。", w)
	}

	// Java: tmg.game.file.gererator.title = "天马G-游戏文件生成器"
	tianmaGGameFileGenerator := fyne.NewMenuItem("天马G-游戏文件生成器", func() {
		showTodo("天马G-游戏文件生成器")
	})

	// 关键点：提供一个显式的“设置”菜单（可先留空），避免框架在某些平台将默认 Quit/Exit
	// 菜单项“塞进”第一个菜单（你这里就是“天马G”）。
	mSettings := fyne.NewMenu("设置")

	mTianmaG := fyne.NewMenu("天马G", tianmaGGameFileGenerator)

	// Java help.menu structure:
	// Help -> Document / Settings / Feedback / Check for Updates / Contribute / About
	// 说明：你要求将依赖路径/显示语言/外观/置顶/开机启动放到设置界面内部，因此不在菜单展示。
	mHelp := fyne.NewMenu("帮助",
		fyne.NewMenuItem("文档", func() { showTodo("文档") }),
		fyne.NewMenuItem("设置", func() { showTodo("设置") }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("意见反馈", func() { showTodo("意见反馈") }),
		fyne.NewMenuItem("检查更新", func() { showTodo("检查更新") }),
		fyne.NewMenuItem("贡献", func() { showTodo("贡献") }),
		fyne.NewMenuItemSeparator(),
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
