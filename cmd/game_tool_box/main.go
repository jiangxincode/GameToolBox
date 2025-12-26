package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/game_tool_box/internal/resources"
	"github.com/game_tool_box/internal/ui/pegasus"
)

func main() {
	a := app.New()
	w := a.NewWindow("game_tool_box")

	w.SetIcon(resources.IconPng)

	// Simple view router
	mainView := container.NewVBox(
		widget.NewLabel("game_tool_box"),
		widget.NewLabel("Fyne GUI skeleton is ready."),
	)
	router := container.NewMax(mainView)
	w.SetContent(router)

	showTodo := func(title string) {
		dialog.ShowInformation(title, "这里将来会放置："+title+"。", w)
	}

	showMain := func() {
		router.Objects = []fyne.CanvasObject{mainView}
		router.Refresh()
	}

	// Java: pegasus.game.file.gererator.title = "天马G-游戏文件生成器"
	pegasusGameFileGenerator := fyne.NewMenuItem("天马G-游戏文件生成器", func() {
		view := tmgui.NewGeneratorView(w)
		back := widget.NewButton("返回", showMain)
		page := container.NewBorder(container.NewHBox(back), nil, nil, nil, view)
		router.Objects = []fyne.CanvasObject{page}
		router.Refresh()
	})

	// 保留一个显式的“设置”菜单（可先留空），避免框架在某些平台将默认 Quit/Exit 注入第一个菜单。
	mSettings := fyne.NewMenu("设置")

	mPegasus := fyne.NewMenu("天马G", pegasusGameFileGenerator)

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

	w.SetMainMenu(fyne.NewMainMenu(mSettings, mPegasus, mHelp))

	w.Resize(fyne.NewSize(900, 650))
	w.ShowAndRun()
}
