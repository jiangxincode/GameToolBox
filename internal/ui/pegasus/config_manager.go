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

	loadFromRomFiles := func() {
		root := ui.rootDir()
		logging.Infof("%s click load from rom files root=%s", logPrefix, root)
		if root == "" {
			dialog.ShowInformation("提示", "请先设置根目录", w)
			return
		}
		games, err := pegasus.LoadGamesFromRomFiles(root)
		if err != nil {
			logging.Errorf("%s load from rom files failed root=%s err=%v", logPrefix, root, err)
			dialog.ShowError(err, w)
			return
		}
		ui.setGames(games)
		dialog.ShowInformation("提示", "ROM 文件扫描完成", w)
	}

	generateSelected := func() {
		root := ui.rootDir()
		logging.Infof("%s click generate selected config root=%s", logPrefix, root)
		if root == "" {
			dialog.ShowInformation("提示", "请先设置根目录", w)
			return
		}
		if ui.selectedCount() == 0 {
			dialog.ShowInformation("提示", "请选择要生成配置的游戏", w)
			return
		}

		res := pegasus.GenerateSelectedConfig(root, ui.allGames)
		if len(res.Errors) > 0 {
			dialog.ShowError(fmt.Errorf("部分生成失败: %v", res.Errors[0]), w)
			logging.Infof("%s generate config failed written=%d skipped=%d failed=%d errors=%d", logPrefix, res.Written, res.Skipped, res.Failed, len(res.Errors))
			return
		}
		dialog.ShowInformation("提示", fmt.Sprintf("配置生成完成\nWritten=%d, Skipped=%d", res.Written, res.Skipped), w)
		logging.Infof("%s generate config finished written=%d skipped=%d failed=%d errors=%d", logPrefix, res.Written, res.Skipped, res.Failed, len(res.Errors))
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
			files := make([]string, 0, len(diff.ExtraInConfig))
			for _, g := range diff.ExtraInConfig {
				files = append(files, g.FileName)
			}
			removed, err := pegasus.RemoveGamesFromMetadata(root, files)
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
		widget.NewButton("从ROM文件加载数据", loadFromRomFiles),
		widget.NewButton("全选", func() {
			logging.Infof("%s click select all", logPrefix)
			ui.selectAll(true)
		}),
		widget.NewButton("取消全选", func() {
			logging.Infof("%s click deselect all", logPrefix)
			ui.selectAll(false)
		}),
		widget.NewButton("生成选中游戏的配置", generateSelected),
		widget.NewButton("列出配置文件中重复的游戏", listDuplicates),
	)
	buttonRow2 := container.NewHBox(
		widget.NewButton("列出配置文件中缺失的游戏", listMissing),
		widget.NewButton("补齐配置文件中缺失的游戏", fillMissing),
	)
	buttonRow3 := container.NewHBox(
		widget.NewButton("列出配置文件中多余的游戏", listExtra),
		widget.NewButton("删除配置文件中多余的游戏", deleteExtra),
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
