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

	"github.com/game_tool_box/internal/pegasus"
)

// NewGeneratorView creates the Fyne UI for "天马G-游戏文件生成器".
//
// Notes:
//   - Video preview in Swing uses JavaFX; Go/Fyne version shows a placeholder for now.
//   - Checkbox column uses "✓" and toggles when selecting that column.
func NewGeneratorView(w fyne.Window) fyne.CanvasObject {
	rootEntry := widget.NewEntry()
	rootEntry.SetPlaceHolder("选择根目录（包含 metadata.pegasus.txt）")

	var allGames []pegasus.GameModel
	filteredIdx := []int{}

	loadedLabel := widget.NewLabel("已加载 0 个游戏")

	// right side
	coverTitle := widget.NewLabelWithStyle("封面图片", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	coverImg := canvas.NewImageFromResource(nil)
	coverImg.FillMode = canvas.ImageFillContain
	coverImg.SetMinSize(fyne.NewSize(300, 400))
	coverBox := container.NewBorder(coverTitle, nil, nil, nil, container.New(layout.NewMaxLayout(), coverImg))

	gameDetail := widget.NewMultiLineEntry()
	gameDetail.Disable()
	gameDetail.Wrapping = fyne.TextWrapWord
	gameDetailBox := widget.NewCard("游戏详情", "", container.New(layout.NewMaxLayout(), gameDetail))

	mediaTabs := container.NewAppTabs(
		container.NewTabItem("封面图片", coverBox),
		container.NewTabItem("视频预览", widget.NewLabel("Go 版暂不支持视频预览")),
	)
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

	headers := []string{"选择状态", "序号", "游戏名称", "文件名称"}
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

		detail := strings.Builder{}
		detail.WriteString("游戏名称: ")
		detail.WriteString(g.GameName)
		detail.WriteString("\n")
		detail.WriteString("文件名称: ")
		detail.WriteString(g.FileName)
		detail.WriteString("\n")
		detail.WriteString("排序编号: ")
		detail.WriteString(g.SortBy)
		detail.WriteString("\n")
		detail.WriteString("开发商: ")
		detail.WriteString(g.Developer)
		detail.WriteString("\n")
		detail.WriteString("游戏简介: \n")
		detail.WriteString(g.Description)
		detail.WriteString("\n")
		gameDetail.SetText(detail.String())
		gameDetail.Refresh()
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
		applyFilter(s)
		table.Refresh()
		selectedFilteredRow = -1
	}

	clearSearchBtn := widget.NewButton("清除搜索", func() {
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
	}

	chooseRootBtn := widget.NewButton("设置根目录", func() {
		fd := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if uri == nil {
				return
			}
			rootEntry.SetText(filepath.FromSlash(uri.Path()))
		}, w)
		fd.Show()
	})

	buttonRow := container.NewHBox(
		widget.NewButton("加载/刷新数据", loadGameData),
		widget.NewButton("全选", func() { selectAll(true) }),
		widget.NewButton("取消全选", func() { selectAll(false) }),
		widget.NewButton("生成选中文件", generateSelected),
	)

	searchRow := container.NewHBox(widget.NewLabel("搜索:"), searchEntry, clearSearchBtn)

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
