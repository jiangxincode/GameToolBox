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

func NewRomeManagerView(w fyne.Window) fyne.CanvasObject {
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

		res := pegasus.GenerateSelectedFiles(root, ui.allGames)
		if len(res.Errors) > 0 {
			dialog.ShowError(fmt.Errorf("部分生成失败: %v", res.Errors[0]), w)
			return
		}
		dialog.ShowInformation("提示", fmt.Sprintf("文件生成完成\nCreated=%d, Skipped=%d", res.Created, res.Skipped), w)
		logging.Infof("%s generate finished created=%d skipped=%d errors=%d", logPrefix, res.Created, res.Skipped, len(res.Errors))
	}

	// listMissing: metadata 中存在但磁盘 ROM 不存在（即 diff.ExtraInConfig）
	listMissing := func() {
		root := ui.rootDir()
		logging.Infof("%s click list missing rom files root=%s", logPrefix, root)
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
			dialog.ShowInformation("提示", "ROM 文件中没有缺失的游戏", w)
			return
		}

		lines := make([]string, 0, len(diff.ExtraInConfig))
		for _, g := range diff.ExtraInConfig {
			// g.FileName might be absolute; show a nicer display when possible.
			display := g.FileName
			if !filepath.IsAbs(display) {
				display = filepath.ToSlash(display)
			}
			if strings.TrimSpace(g.GameName) != "" {
				lines = append(lines, fmt.Sprintf("%s (%s)", g.GameName, display))
			} else {
				lines = append(lines, display)
			}
		}
		dialog.ShowInformation("ROM 缺失的游戏", strings.Join(lines, "\n"), w)
	}

	// generateMissing: 为 metadata 中存在但 ROM 不存在的条目生成空 ROM 文件
	generateMissing := func() {
		root := ui.rootDir()
		logging.Infof("%s click generate missing rom files root=%s", logPrefix, root)
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
			dialog.ShowInformation("提示", "ROM 文件中没有缺失的游戏", w)
			return
		}

		// 只补齐“选中”的缺失项：在当前表格选中列表里且确实缺失。
		selected := map[string]bool{}
		for _, g := range ui.allGames {
			if !g.Selected {
				continue
			}
			k := strings.ToLower(filepath.ToSlash(strings.TrimSpace(g.FileName)))
			if k != "" {
				selected[k] = true
			}
		}

		toGen := make([]pegasus.GameViewModel, 0, len(diff.ExtraInConfig))
		for _, g := range diff.ExtraInConfig {
			fileKey := strings.ToLower(filepath.ToSlash(strings.TrimSpace(g.FileName)))
			if len(selected) > 0 {
				if !selected[fileKey] {
					continue
				}
			}
			// GenerateSelectedFiles expects relative paths under root.
			// If metadata uses absolute paths, we can't safely create under root; skip.
			if filepath.IsAbs(g.FileName) {
				continue
			}
			g2 := g
			g2.Selected = true
			toGen = append(toGen, g2)
		}

		if len(toGen) == 0 {
			// Most likely: user selected none, or all missing entries are absolute paths.
			dialog.ShowInformation("提示", "没有需要补齐的缺失 ROM（请先在列表中勾选要补齐的游戏，且 file 路径需为相对路径）", w)
			return
		}

		confirmMsg := fmt.Sprintf("将生成 %d 个缺失的空 ROM 文件，是否继续？", len(toGen))
		dialog.ShowConfirm("确认补齐", confirmMsg, func(ok bool) {
			if !ok {
				return
			}
			res := pegasus.GenerateSelectedFiles(root, toGen)
			if len(res.Errors) > 0 {
				dialog.ShowError(fmt.Errorf("部分生成失败: %v", res.Errors[0]), w)
				return
			}
			dialog.ShowInformation("提示", fmt.Sprintf("补齐完成\nCreated=%d, Skipped=%d", res.Created, res.Skipped), w)
			logging.Infof("%s generate missing finished created=%d skipped=%d errors=%d", logPrefix, res.Created, res.Skipped, len(res.Errors))
		}, w)
	}

	// listExtra: 磁盘 ROM 存在但 metadata 中不存在（即 diff.MissingInConfig）
	listExtra := func() {
		root := ui.rootDir()
		logging.Infof("%s click list extra rom files root=%s", logPrefix, root)
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
			dialog.ShowInformation("提示", "ROM 文件中没有多余的游戏", w)
			return
		}

		lines := make([]string, 0, len(diff.MissingInConfig))
		for _, g := range diff.MissingInConfig {
			lines = append(lines, g.FileName)
		}
		dialog.ShowInformation("ROM 多余的游戏", strings.Join(lines, "\n"), w)
	}

	deleteRomsNotInConfig := func() {
		root := ui.rootDir()
		logging.Infof("%s click delete roms not in config root=%s", logPrefix, root)
		if root == "" {
			dialog.ShowInformation("提示", "请先设置根目录", w)
			return
		}

		diff, err := pegasus.DiffConfigAgainstRomFiles(root)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		if len(diff.MissingInConfig) == 0 {
			dialog.ShowInformation("提示", "没有需要删除的 ROM（配置文件中不存在）", w)
			return
		}

		confirmMsg := fmt.Sprintf("将删除 %d 个配置文件中不存在的 ROM 文件，是否继续？\n(这些文件在磁盘存在，但 metadata.pegasus.txt 中没有对应记录)", len(diff.MissingInConfig))
		dialog.ShowConfirm("确认删除", confirmMsg, func(ok bool) {
			if !ok {
				return
			}
			res := pegasus.DeleteRomsNotInConfig(root)
			if len(res.Errors) > 0 {
				dialog.ShowError(fmt.Errorf("部分删除失败: %v", res.Errors[0]), w)
				return
			}
			dialog.ShowInformation("提示", fmt.Sprintf("删除完成\nDeleted=%d, Skipped=%d", res.Deleted, res.Skipped), w)
			logging.Infof("%s delete roms finished deleted=%d skipped=%d failed=%d errors=%d", logPrefix, res.Deleted, res.Skipped, res.Failed, len(res.Errors))
		}, w)
	}

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
		widget.NewButton("生成选中游戏的空ROM", generateSelected),
	)

	buttonRow2 := container.NewHBox(
		widget.NewButton("查找有配置无ROM的游戏", listMissing),
		widget.NewButton("补齐对应游戏的空ROM", generateMissing),
	)
	buttonRow3 := container.NewHBox(
		widget.NewButton("查找无配置有ROM的游戏", listExtra),
		widget.NewButton("删除对应游戏的ROM", deleteRomsNotInConfig),
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
