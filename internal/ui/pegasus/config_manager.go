package pegasusui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/game_tool_box/internal/logging"
	"github.com/game_tool_box/internal/pegasus"
)

func NewConfigManagerView(w fyne.Window) fyne.CanvasObject {
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

	listMissing := func() {
		root := ui.rootDir()
		logging.Infof("%s click list missing root=%s", logPrefix, root)
		if root == "" {
			dialog.ShowInformation("提示", "请先设置根目录", w)
			return
		}
		diff, err := pegasus.DiffConfigAgainstRomFiles(root)
		if err != nil {
			logging.Errorf("%s diff config vs rom files failed root=%s err=%v", logPrefix, root, err)
			dialog.ShowError(err, w)
			return
		}
		if len(diff.MissingInConfig) == 0 {
			dialog.ShowInformation("提示", "配置文件中没有缺失的游戏", w)
			return
		}
		lines := make([]string, 0, len(diff.MissingInConfig))
		for _, g := range diff.MissingInConfig {
			lines = append(lines, fmt.Sprintf("%s (%s)", g.GameName, g.FileName))
		}
		dialog.ShowInformation("缺失的游戏", strings.Join(lines, "\n"), w)
	}

	listExtra := func() {
		root := ui.rootDir()
		logging.Infof("%s click list extra root=%s", logPrefix, root)
		if root == "" {
			dialog.ShowInformation("提示", "请先设置根目录", w)
			return
		}
		diff, err := pegasus.DiffConfigAgainstRomFiles(root)
		if err != nil {
			logging.Errorf("%s diff config vs rom files failed root=%s err=%v", logPrefix, root, err)
			dialog.ShowError(err, w)
			return
		}
		if len(diff.ExtraInConfig) == 0 {
			dialog.ShowInformation("提示", "配置文件中没有多余的游戏", w)
			return
		}
		lines := make([]string, 0, len(diff.ExtraInConfig))
		for _, g := range diff.ExtraInConfig {
			lines = append(lines, fmt.Sprintf("%s (%s)", g.GameName, g.FileName))
		}
		dialog.ShowInformation("多余的游戏", strings.Join(lines, "\n"), w)
	}

	deleteExtra := func() {
		root := ui.rootDir()
		logging.Infof("%s click delete extra root=%s", logPrefix, root)
		if root == "" {
			dialog.ShowInformation("提示", "请先设置根目录", w)
			return
		}
		diff, err := pegasus.DiffConfigAgainstRomFiles(root)
		if err != nil {
			logging.Errorf("%s diff config vs rom files failed root=%s err=%v", logPrefix, root, err)
			dialog.ShowError(err, w)
			return
		}
		if len(diff.ExtraInConfig) == 0 {
			dialog.ShowInformation("提示", "配置文件中没有多余的游戏", w)
			return
		}
		confirmMsg := fmt.Sprintf("将从 metadata.pegasus.txt 中删除 %d 条多余的游戏记录，是否继续？", len(diff.ExtraInConfig))
		dialog.ShowConfirm("确认删除", confirmMsg, func(ok bool) {
			if !ok {
				return
			}
			gameNames := make([]string, 0, len(diff.ExtraInConfig))
			for _, g := range diff.ExtraInConfig {
				gameNames = append(gameNames, g.GameName)
			}
			removed, err := pegasus.RemoveGamesFromMetadata(root, gameNames)
			if err != nil {
				logging.Errorf("%s remove games from metadata failed root=%s err=%v", logPrefix, root, err)
				dialog.ShowError(err, w)
				return
			}
			dialog.ShowInformation("提示", fmt.Sprintf("删除完成，已删除 %d 条记录", removed), w)
		}, w)
	}

	listDuplicates := func() {
		root := ui.rootDir()
		logging.Infof("%s click list duplicates root=%s", logPrefix, root)
		if root == "" {
			dialog.ShowInformation("提示", "请先设置根目录", w)
			return
		}
		diff, err := pegasus.DiffConfigAgainstRomFiles(root)
		if err != nil {
			logging.Errorf("%s diff config vs rom files failed root=%s err=%v", logPrefix, root, err)
			dialog.ShowError(err, w)
			return
		}
		if len(diff.DuplicateInConfig) == 0 {
			dialog.ShowInformation("提示", "配置文件中没有重复的游戏", w)
			return
		}
		dialog.ShowInformation("重复的游戏", strings.Join(diff.DuplicateInConfig, "\n"), w)
	}

	fillMissing := func() {
		root := ui.rootDir()
		logging.Infof("%s click fill missing root=%s", logPrefix, root)
		if root == "" {
			dialog.ShowInformation("提示", "请先设置根目录", w)
			return
		}
		diff, err := pegasus.DiffConfigAgainstRomFiles(root)
		if err != nil {
			logging.Errorf("%s diff config vs rom files failed root=%s err=%v", logPrefix, root, err)
			dialog.ShowError(err, w)
			return
		}
		if len(diff.MissingInConfig) == 0 {
			dialog.ShowInformation("提示", "配置文件中没有缺失的游戏", w)
			return
		}
		confirmMsg := fmt.Sprintf("将向 metadata.pegasus.txt 追加 %d 条缺失的游戏记录，是否继续？", len(diff.MissingInConfig))
		dialog.ShowConfirm("确认补齐", confirmMsg, func(ok bool) {
			if !ok {
				return
			}
			if err := pegasus.AppendMissingGamesToMetadata(root, diff.MissingInConfig); err != nil {
				logging.Errorf("%s append missing games to metadata failed root=%s err=%v", logPrefix, root, err)
				dialog.ShowError(err, w)
				return
			}
			dialog.ShowInformation("提示", "补齐完成", w)
		}, w)
	}

	// 按钮太多并排会导致界面过宽，这里按功能分组分成 3 行。
	buttonRow1 := container.NewHBox(
		widget.NewButton("加载/刷新数据", loadGameData),
		widget.NewButton("全选", func() {
			logging.Infof("%s click select all", logPrefix)
			ui.selectAll(true)
		}),
		widget.NewButton("取消全选", func() {
			logging.Infof("%s click deselect all", logPrefix)
			ui.selectAll(false)
		}),
		widget.NewButton("列出配置中重复的游戏", listDuplicates),
	)
	buttonRow2 := container.NewHBox(
		widget.NewButton("列出无配置有ROM的游戏", listMissing),
		widget.NewButton("补齐对应游戏的配置", fillMissing),
	)
	buttonRow3 := container.NewHBox(
		widget.NewButton("列出有配置无ROM的游戏", listExtra),
		widget.NewButton("删除对应游戏的配置", deleteExtra),
	)
	buttonRows := container.NewVBox(buttonRow1, buttonRow2, buttonRow3)

	searchRow := newSearchRow(ui.search, logPrefix)

	left := container.NewBorder(nil, nil, nil, nil, ui.table)
	split := container.NewHSplit(left, ui.right)
	split.Offset = 0.45

	status := container.NewHBox(ui.loadedLabel)

	content := container.NewBorder(
		container.NewVBox(container.NewBorder(nil, nil, nil, chooseRootBtn, rootEntry), buttonRows, searchRow),
		status,
		nil,
		nil,
		split,
	)
	return content
}
