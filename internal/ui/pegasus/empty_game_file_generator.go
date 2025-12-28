package tmgui

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

// NewGeneratorView creates the Fyne UI for "游戏文件生成器".
//
//   - Checkbox column uses "✓" and toggles when selecting that column.
func NewGeneratorView(w fyne.Window) fyne.CanvasObject {
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

	selectAll := func(sel bool) {
		for i := range allGames {
			allGames[i].Selected = sel
		}
		table.Refresh()
	}

	loadGameData := func() {
		root := strings.TrimSpace(rootEntry.Text)
		logging.Infof("pegasus: click load data root=%s", root)
		if root == "" {
			dialog.ShowInformation("提示", "请先设置根目录", w)
			return
		}
		games, err := pegasus.LoadGamesFromRootDir(root)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		allGames = games
		applyFilter(searchEntry.Text)
		table.Refresh()
		loadedLabel.SetText(fmt.Sprintf("已加载 %d 个游戏", len(allGames)))
		dialog.ShowInformation("提示", "游戏数据加载完成", w)
	}

	generateSelected := func() {
		root := strings.TrimSpace(rootEntry.Text)
		logging.Infof("pegasus: click generate selected root=%s", root)
		if root == "" {
			dialog.ShowInformation("提示", "请先设置根目录", w)
			return
		}
		selected := 0
		for _, g := range allGames {
			if g.Selected {
				selected++
			}
		}
		if selected == 0 {
			dialog.ShowInformation("提示", "请选择要生成的游戏", w)
			return
		}

		res := pegasus.GenerateSelectedFiles(root, allGames)
		if len(res.Errors) > 0 {
			dialog.ShowError(fmt.Errorf("部分生成失败: %v", res.Errors[0]), w)
			return
		}
		dialog.ShowInformation("提示", fmt.Sprintf("文件生成完成\nCreated=%d, Skipped=%d", res.Created, res.Skipped), w)
		logging.Infof("pegasus: generate finished created=%d skipped=%d errors=%d", res.Created, res.Skipped, len(res.Errors))
	}

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

	buttonRow := container.NewHBox(
		widget.NewButton("加载/刷新数据", loadGameData),
		widget.NewButton("全选", func() {
			logging.Infof("pegasus: click select all")
			selectAll(true)
		}),
		widget.NewButton("取消全选", func() {
			logging.Infof("pegasus: click deselect all")
			selectAll(false)
		}),
		widget.NewButton("生成选中文件", generateSelected),
	)

	// Old: searchRow := container.NewHBox(widget.NewLabel("搜索:"), searchEntry, clearSearchBtn)
	searchRow := container.NewBorder(nil, nil, widget.NewLabel("搜索:"), clearSearchBtn, container.NewMax(searchEntry))

	left := container.NewBorder(nil, nil, nil, nil, table)
	split := container.NewHSplit(left, right)
	split.Offset = 0.45

	status := container.NewHBox(loadedLabel)

	return container.NewBorder(
		container.NewVBox(container.NewBorder(nil, nil, nil, chooseRootBtn, rootEntry), buttonRow, searchRow),
		status,
		nil,
		nil,
		split,
	)
}
