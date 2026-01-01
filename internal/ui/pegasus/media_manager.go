package pegasusui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/game_tool_box/internal/logging"
	"github.com/game_tool_box/internal/pegasus"
)

func NewMediaManagerView(w fyne.Window) fyne.CanvasObject {
	logPrefix := logging.PrefixFromCallerSkip(0)

	ui := newGameListUI(w, logPrefix)

	rootEntry, persistRootDir := newRootEntryWithConfig(w, logPrefix)
	ui.rootEntry = rootEntry
	chooseRootBtn := newChooseRootButton(w, rootEntry, persistRootDir, logPrefix)

	loadGameData := func() {
		root := ui.rootDir()
		logging.Infof("%s click load data root=%s", logPrefix, root)
		if root == "" {
			dialog.ShowInformation("提示", "请先设置根目录", w)
			return
		}
		games, err := pegasus.LoadGamesFromRootDir(root)
		if err != nil {
			logging.Errorf("%s load games failed root=%s err=%v", logPrefix, root, err)
			dialog.ShowError(err, w)
			return
		}
		ui.setGames(games)
		dialog.ShowInformation("提示", "游戏数据加载完成", w)
	}

	generateSelected := func() {
		root := ui.rootDir()
		logging.Infof("%s click generate selected root=%s", logPrefix, root)
		if root == "" {
			dialog.ShowInformation("提示", "请先设置根目录", w)
			return
		}
		if ui.selectedCount() == 0 {
			dialog.ShowInformation("提示", "请选择要生成的游戏", w)
			return
		}

		// IMPORTANT: do NOT call pegasus.GenerateSelectedFiles here.
		res := pegasus.GenerateSelectedFilesForOssHandheld(root, ui.allGames)
		if len(res.Errors) > 0 {
			dialog.ShowError(fmt.Errorf("部分生成失败: %v", res.Errors[0]), w)
			logging.Infof("%s generate failed created=%d skipped=%d errors=%d", logPrefix, res.Created, res.Skipped, len(res.Errors))
			return
		}
		dialog.ShowInformation("提示", fmt.Sprintf("文件生成完成\nCreated=%d, Skipped=%d", res.Created, res.Skipped), w)
		logging.Infof("%s generate finished created=%d skipped=%d errors=%d", logPrefix, res.Created, res.Skipped, len(res.Errors))
	}

	generateEmptyMediaDirs := func() {
		root := ui.rootDir()
		logging.Infof("%s click generate empty media dirs root=%s", logPrefix, root)
		if root == "" {
			dialog.ShowInformation("提示", "请先设置根目录", w)
			return
		}
		if ui.selectedCount() == 0 {
			dialog.ShowInformation("提示", "请选择要生成媒体目录的游戏", w)
			return
		}

		res := pegasus.GenerateEmptyMediaFolders(root, ui.allGames)
		if len(res.Errors) > 0 {
			dialog.ShowError(fmt.Errorf("部分生成失败: %v", res.Errors[0]), w)
			logging.Infof("%s generate empty media dirs failed created=%d skipped=%d errors=%d", logPrefix, res.Created, res.Skipped, len(res.Errors))
			return
		}
		dialog.ShowInformation("提示", fmt.Sprintf("媒体目录生成完成\nCreated=%d, Skipped=%d", res.Created, res.Skipped), w)
		logging.Infof("%s generate empty media dirs finished created=%d skipped=%d errors=%d", logPrefix, res.Created, res.Skipped, len(res.Errors))
	}

	buttonRow := container.NewHBox(
		widget.NewButton("加载/刷新数据", loadGameData),
		widget.NewButton("全选", func() {
			logging.Infof("%s click select all", logPrefix)
			ui.selectAll(true)
		}),
		widget.NewButton("取消全选", func() {
			logging.Infof("%s click deselect all", logPrefix)
			ui.selectAll(false)
		}),
		widget.NewButton("生成对应游戏空媒体文件夹", generateEmptyMediaDirs),
		widget.NewButton("转换为开源掌机格式", generateSelected),
	)

	searchRow := newSearchRow(ui.search, logPrefix)

	left := container.NewBorder(nil, nil, nil, nil, ui.table)
	split := container.NewHSplit(left, ui.right)
	split.Offset = 0.45

	status := container.NewHBox(ui.loadedLabel)

	content := container.NewBorder(
		container.NewVBox(container.NewBorder(nil, nil, nil, chooseRootBtn, rootEntry), buttonRow, searchRow),
		status,
		nil,
		nil,
		split,
	)
	return content
}
