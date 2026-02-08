package pegasusui

import (
	"fmt"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/game_tool_box/internal/logging"
	"github.com/game_tool_box/internal/pegasus"
)

func NewManualAddGameView(w fyne.Window) fyne.CanvasObject {
	logPrefix := logging.PrefixFromCallerSkip(0)

	rootEntry, persistRootDir := newRootEntryWithConfig(w, logPrefix)
	chooseRootBtn := newChooseRootButton(w, rootEntry, persistRootDir, logPrefix)

	romEntry := widget.NewEntry()
	romEntry.Disable()
	romEntry.SetPlaceHolder("Rom 文件路径")

	gameNameEntry := widget.NewEntry()
	gameNameEntry.SetPlaceHolder("游戏名称")

	developerEntry := widget.NewEntry()
	developerEntry.SetPlaceHolder("开发商（可选）")

	descEntry := widget.NewMultiLineEntry()
	descEntry.SetPlaceHolder("游戏简介（可选）")
	descEntry.SetMinRowsVisible(3)

	romRelEntry := widget.NewEntry()
	romRelEntry.Hide()

	boxEntry := widget.NewEntry()
	boxEntry.Disable()
	logoEntry := widget.NewEntry()
	logoEntry.Disable()
	videoEntry := widget.NewEntry()
	videoEntry.Disable()

	overwriteCheck := widget.NewCheck("如果存在则覆盖", nil)

	chooseRomBtn := widget.NewButton("选择 Rom 文件", func() {
		fd := dialog.NewFileOpen(func(uri fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if uri == nil {
				return
			}
			p := filepath.FromSlash(uri.URI().Path())
			_ = uri.Close()
			romEntry.SetText(p)
			if strings.TrimSpace(gameNameEntry.Text) == "" {
				base := filepath.Base(p)
				gameNameEntry.SetText(strings.TrimSuffix(base, filepath.Ext(base)))
			}
		}, w)
		fd.Show()
	})

	chooseMedia := func(target *widget.Entry) {
		fd := dialog.NewFileOpen(func(uri fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if uri == nil {
				return
			}
			p := filepath.FromSlash(uri.URI().Path())
			_ = uri.Close()
			target.SetText(p)
		}, w)
		fd.Show()
	}

	chooseCoverBtn := widget.NewButton("选择封面图片", func() { chooseMedia(boxEntry) })
	chooseLogoBtn := widget.NewButton("选择logo图片", func() { chooseMedia(logoEntry) })
	chooseVideoBtn := widget.NewButton("选择视频文件", func() { chooseMedia(videoEntry) })

	var submitBtn *widget.Button
	submitBtn = widget.NewButton("添加游戏", func() {
		root := strings.TrimSpace(rootEntry.Text)
		if root == "" {
			dialog.ShowInformation("缺少根目录", "请为游戏文件指定根目录.", w)
			return
		}
		romPath := strings.TrimSpace(romEntry.Text)
		if romPath == "" {
			dialog.ShowInformation("缺少 Rom 文件", "请为游戏选择 Rom 文件.", w)
			return
		}

		req := pegasus.ManualAddGameRequest{
			GameName:      gameNameEntry.Text,
			Developer:     developerEntry.Text,
			Description:   descEntry.Text,
			SourceRomPath: romPath,
			BoxFrontPath:  boxEntry.Text,
			LogoPath:      logoEntry.Text,
			VideoPath:     videoEntry.Text,
			Overwrite:     overwriteCheck.Checked,
		}

		gname := strings.TrimSpace(gameNameEntry.Text)
		if gname == "" {
			gname = strings.TrimSuffix(filepath.Base(romPath), filepath.Ext(romPath))
		}
		destName := filepath.Base(romPath)
		confirmMsg := fmt.Sprintf("您确定要添加游戏 '%s' 吗？\nRom 文件: '%s'\n将被复制到根目录: '%s'", gname, filepath.Base(romPath), destName)
		dialog.ShowConfirm("确认添加游戏", confirmMsg, func(ok bool) {
			if !ok {
				return
			}

			submitBtn.Disable()
			progress := dialog.NewProgress("添加游戏中...", "...", w)
			progress.Show()
			progress.SetValue(-1)

			go func() {
				res, err := pegasus.ManualAddGame(root, req)
				fyne.Do(func() {
					submitBtn.Enable()
					progress.Hide()
					if err != nil {
						logging.Errorf("%s manual add game failed err=%v", logPrefix, err)
						dialog.ShowError(err, w)
						return
					}

					msg := fmt.Sprintf("游戏添加成功！\nRom 路径: %s\n媒体文件已复制: %v", res.RomRelPath, res.MediaCopied)
					dialog.ShowInformation("成功", msg, w)
				})
			}()
		}, w)
	})

	// 真实标签。
	romEntry.SetPlaceHolder("Rom 文件路径")
	gameNameEntry.SetPlaceHolder("游戏名称")
	romRelEntry.SetPlaceHolder("相对 Rom 路径")
	boxEntry.SetPlaceHolder("封面图片路径")
	logoEntry.SetPlaceHolder("Logo 图片路径")
	videoEntry.SetPlaceHolder("视频文件路径")
	overwriteCheck.SetText("如果存在则覆盖")

	form := container.NewVBox(
		container.NewBorder(nil, nil, nil, chooseRootBtn, rootEntry),
		widget.NewSeparator(),

		// 游戏名称放在 ROM 选择之上
		container.NewBorder(nil, nil, widget.NewLabel("游戏名称"), nil, container.NewMax(gameNameEntry)),
		container.NewBorder(nil, nil, widget.NewLabel("Rom 文件"), chooseRomBtn, container.NewMax(romEntry)),
		container.NewBorder(nil, nil, widget.NewLabel("开发商"), nil, container.NewMax(developerEntry)),
		widget.NewLabel("游戏简介"),
		descEntry,
		overwriteCheck,

		widget.NewSeparator(),
		widget.NewLabel("媒体文件"),
		container.NewBorder(nil, nil, widget.NewLabel("封面图片"), chooseCoverBtn, container.NewMax(boxEntry)),
		container.NewBorder(nil, nil, widget.NewLabel("Logo 图片"), chooseLogoBtn, container.NewMax(logoEntry)),
		container.NewBorder(nil, nil, widget.NewLabel("视频文件"), chooseVideoBtn, container.NewMax(videoEntry)),

		widget.NewSeparator(),
		submitBtn,
	)

	// 现在设置实际字符串（避免某些平台上的 NUL 占位符渲染问题）。
	romEntry.SetPlaceHolder("Rom 文件路径")
	gameNameEntry.SetPlaceHolder("游戏名称")
	romRelEntry.SetPlaceHolder("相对 Rom 路径")
	overwriteCheck.SetText("如果存在则覆盖")
	chooseRomBtn.SetText("选择 Rom 文件")
	chooseCoverBtn.SetText("选择封面图片")
	chooseLogoBtn.SetText("选择Logo图片")
	chooseVideoBtn.SetText("选择视频文件")
	submitBtn.SetText("添加游戏")

	return container.NewPadded(form)
}
