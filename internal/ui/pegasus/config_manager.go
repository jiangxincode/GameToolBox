package pegasusui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/game_tool_box/internal/config"
	"github.com/game_tool_box/internal/logging"
	"github.com/game_tool_box/internal/pegasus"
)

// NewConfigManagerView is a copy of NewRomeManagerView for now.
// We'll adjust the config manager behavior later without affecting rom manager.
func NewConfigManagerView(w fyne.Window) fyne.CanvasObject {
	rootEntry := widget.NewEntry()
	rootEntry.SetPlaceHolder("选择根目录（包含 metadata.pegasus.txt）")

	// Restore last root directory
	if c, err := config.Load(); err == nil {
		if strings.TrimSpace(c.RootDir) != "" {
			rootEntry.SetText(c.RootDir)
			logging.Infof("pegasus: restore rootDir=%s", c.RootDir)
		}
	}

	persistRootDir := func(p string) {
		p = strings.TrimSpace(p)
		if p == "" {
			return
		}
		c, _ := config.Load()
		if c.RootDir == p {
			return
		}
		c.RootDir = p
		_ = config.Save(c)
		logging.Infof("pegasus: persist rootDir=%s", p)
	}

	// Persist on manual edit (debounced enough for small config writes)
	rootEntry.OnChanged = func(s string) {
		persistRootDir(s)
	}

	var allGames []pegasus.GameModel
	filteredIdx := []int{}

	loadedLabel := widget.NewLabel("已加载 0 个游戏")

	// right side
	// Removed extra title above the image (tab title already shows it).
	coverImg := canvas.NewImageFromResource(nil)
	coverImg.FillMode = canvas.ImageFillContain
	coverImg.SetMinSize(fyne.NewSize(300, 400))
	coverBox := container.New(layout.NewMaxLayout(), coverImg)

	gameDetail := widget.NewRichTextFromMarkdown("")
	gameDetail.Wrapping = fyne.TextWrapWord
	gameDetailScroll := container.NewVScroll(gameDetail)
	gameDetailScroll.SetMinSize(fyne.NewSize(320, 220))
	gameDetailBox := widget.NewCard("游戏详情", "", gameDetailScroll)

	mediaTabs := container.NewAppTabs(
		container.NewTabItem("封面图片", coverBox),
		container.NewTabItem("视频预览", widget.NewLabel("Go 版暂不支持视频预览")),
	)
	// Give details a fixed bottom area so it's always readable.
	right := container.NewBorder(nil, gameDetailBox, nil, nil, mediaTabs)

	// helper: sync filtered indices based on search
	applyFilter := func(query string) {
		q := strings.ToLower(strings.TrimSpace(query))
		filteredIdx = filteredIdx[:0]
		for i, g := range allGames {
			if q == "" || strings.Contains(strings.ToLower(g.GameName), q) || strings.Contains(strings.ToLower(g.FileName), q) {
				filteredIdx = append(filteredIdx, i)
			}
		}
	}
	applyFilter("")

	headers := []string{"选择", "序号", "游戏名称", "文件名称"}
	table := widget.NewTable(
		func() (int, int) { return len(filteredIdx) + 1, len(headers) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			lbl := obj.(*widget.Label)
			if id.Row == 0 {
				lbl.TextStyle = fyne.TextStyle{Bold: true}
				lbl.SetText(headers[id.Col])
				return
			}
			g := allGames[filteredIdx[id.Row-1]]
			switch id.Col {
			case 0:
				if g.Selected {
					lbl.SetText("✓")
				} else {
					lbl.SetText("")
				}
			case 1:
				lbl.SetText(fmt.Sprintf("%d", g.ID))
			case 2:
				lbl.SetText(g.GameName)
			case 3:
				lbl.SetText(g.FileName)
			}
		},
	)
	// sizing close to swing
	table.SetColumnWidth(0, 60)
	table.SetColumnWidth(1, 40)
	table.SetColumnWidth(2, 180)
	table.SetColumnWidth(3, 180)

	selectedFilteredRow := -1
	selectedCol := 0

	showDetailFor := func(g pegasus.GameModel) {
		// cover
		boxFront := g.BoxFrontImagePath
		if boxFront != "" {
			if _, err := os.Stat(boxFront); err == nil {
				coverImg.File = boxFront
				coverImg.Resource = nil
				coverImg.Refresh()
			} else {
				coverImg.File = ""
				coverImg.Resource = nil
				coverImg.Refresh()
			}
		} else {
			coverImg.File = ""
			coverImg.Resource = nil
			coverImg.Refresh()
		}

		md := strings.Builder{}
		md.WriteString("**游戏名称**：")
		md.WriteString(g.GameName)
		md.WriteString("\n\n")
		md.WriteString("**文件名称**：")
		md.WriteString(g.FileName)
		md.WriteString("\n\n")
		md.WriteString("**排序编号**：")
		md.WriteString(g.SortBy)
		md.WriteString("\n\n")
		md.WriteString("**开发商**：")
		md.WriteString(g.Developer)
		md.WriteString("\n\n")
		md.WriteString("**游戏简介**\n\n")
		if strings.TrimSpace(g.Description) == "" {
			md.WriteString("（无）")
		} else {
			md.WriteString(g.Description)
		}

		gameDetail.ParseMarkdown(md.String())
		gameDetail.Refresh()
		gameDetailScroll.ScrollToTop()
	}

	// Track which cell was selected; we'll use it for checkbox toggle.
	table.OnSelected = func(id widget.TableCellID) {
		if id.Row == 0 {
			return
		}
		selectedFilteredRow = id.Row - 1
		selectedCol = id.Col

		idx := filteredIdx[selectedFilteredRow]
		if selectedCol == 0 {
			allGames[idx].Selected = !allGames[idx].Selected
			table.Refresh()
		}

		showDetailFor(allGames[idx])
	}

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("搜索游戏名称或文件名称")
	searchEntry.OnChanged = func(s string) {
		logging.Infof("pegasus: search changed q=%s", s)
		applyFilter(s)
		table.Refresh()
		selectedFilteredRow = -1
	}

	clearSearchBtn := widget.NewButton("清除搜索", func() {
		logging.Infof("pegasus: click clear search")
		searchEntry.SetText("")
	})

	chooseRootBtn := widget.NewButton("设置根目录", func() {
		logging.Infof("pegasus: click choose root")
		fd := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if uri == nil {
				return
			}
			p := filepath.FromSlash(uri.Path())
			logging.Infof("pegasus: chose rootDir=%s", p)
			rootEntry.SetText(p)
			persistRootDir(p)
		}, w)
		fd.Show()
	})

	loadFromRomFiles := func() {
		root := strings.TrimSpace(rootEntry.Text)
		logging.Infof("pegasus: click load from rom files root=%s", root)
		if root == "" {
			dialog.ShowInformation("提示", "请先设置根目录", w)
			return
		}
		games, err := pegasus.LoadGamesFromRomFiles(root)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		allGames = games
		applyFilter(searchEntry.Text)
		table.Refresh()
		loadedLabel.SetText(fmt.Sprintf("已加载 %d 个游戏", len(allGames)))
		dialog.ShowInformation("提示", "ROM 文件扫描完成", w)
	}

	listMissing := func() {
		root := strings.TrimSpace(rootEntry.Text)
		logging.Infof("pegasus: click list missing root=%s", root)
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
		root := strings.TrimSpace(rootEntry.Text)
		logging.Infof("pegasus: click list extra root=%s", root)
		if root == "" {
			dialog.ShowInformation("提示", "请先设置根目录", w)
			return
		}
		diff, err := pegasus.DiffConfigAgainstRomFiles(root)
		if err != nil {
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
		root := strings.TrimSpace(rootEntry.Text)
		logging.Infof("pegasus: click delete extra root=%s", root)
		if root == "" {
			dialog.ShowInformation("提示", "请先设置根目录", w)
			return
		}
		diff, err := pegasus.DiffConfigAgainstRomFiles(root)
		if err != nil {
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
				dialog.ShowError(err, w)
				return
			}
			dialog.ShowInformation("提示", fmt.Sprintf("删除完成，已删除 %d 条记录", removed), w)
		}, w)
	}

	listDuplicates := func() {
		root := strings.TrimSpace(rootEntry.Text)
		logging.Infof("pegasus: click list duplicates root=%s", root)
		if root == "" {
			dialog.ShowInformation("提示", "请先设置根目录", w)
			return
		}
		diff, err := pegasus.DiffConfigAgainstRomFiles(root)
		if err != nil {
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
		root := strings.TrimSpace(rootEntry.Text)
		logging.Infof("pegasus: click fill missing root=%s", root)
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
			dialog.ShowInformation("提示", "配置文件中没有缺失的游戏", w)
			return
		}
		confirmMsg := fmt.Sprintf("将向 metadata.pegasus.txt 追加 %d 条缺失的游戏记录，是否继续？", len(diff.MissingInConfig))
		dialog.ShowConfirm("确认补齐", confirmMsg, func(ok bool) {
			if !ok {
				return
			}
			if err := pegasus.AppendMissingGamesToMetadata(root, diff.MissingInConfig); err != nil {
				dialog.ShowError(err, w)
				return
			}
			dialog.ShowInformation("提示", "补齐完成", w)
		}, w)
	}

	buttonRow := container.NewHBox(
		widget.NewButton("从ROM文件加载数据", loadFromRomFiles),
		widget.NewButton("列出配置文件中缺失的游戏", listMissing),
		widget.NewButton("补齐配置文件中缺失的游戏", fillMissing),
		widget.NewButton("列出配置文件中多余的游戏", listExtra),
		widget.NewButton("删除配置文件中多余的游戏", deleteExtra),
		widget.NewButton("列出配置文件中重复的游戏", listDuplicates),
	)

	// Old: searchRow := container.NewHBox(widget.NewLabel("搜索:"), searchEntry, clearSearchBtn)
	searchRow := container.NewBorder(nil, nil, widget.NewLabel("搜索:"), clearSearchBtn, container.NewMax(searchEntry))

	left := container.NewBorder(nil, nil, nil, nil, table)
	split := container.NewHSplit(left, right)
	split.Offset = 0.45

	status := container.NewHBox(loadedLabel)

	content := container.NewBorder(
		container.NewVBox(container.NewBorder(nil, nil, nil, chooseRootBtn, rootEntry), buttonRow, searchRow),
		status,
		nil,
		nil,
		split,
	)
	return content
}
