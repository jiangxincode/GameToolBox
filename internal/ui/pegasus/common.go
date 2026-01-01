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

// gameListUI bundles shared widgets/state used by multiple Pegasus screens.
// It keeps behavior consistent across screens while letting each screen provide
// its own action buttons and load/operate logic.
type gameListUI struct {
	w         fyne.Window
	logPrefix string

	rootEntry   *widget.Entry
	search      *widget.Entry
	loadedLabel *widget.Label

	allGames    []pegasus.GameModel
	filteredIdx []int

	table            *widget.Table
	coverImg         *canvas.Image
	gameDetail       *widget.RichText
	gameDetailScroll *container.Scroll
	right            fyne.CanvasObject
}

func newRootEntryWithConfig(w fyne.Window, logPrefix string) (*widget.Entry, func(string)) {
	_ = w // kept for symmetry/future expansion

	rootEntry := widget.NewEntry()
	rootEntry.SetPlaceHolder("选择根目录（包含 metadata.pegasus.txt）")

	// Restore last root directory
	if c, err := config.Load(); err == nil {
		if strings.TrimSpace(c.RootDir) != "" {
			rootEntry.SetText(c.RootDir)
			logging.Infof("%s restore rootDir=%s", logPrefix, c.RootDir)
		}
	}

	persist := func(p string) {
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
		logging.Infof("%s persist rootDir=%s", logPrefix, p)
	}

	// Persist on manual edit (debounced enough for small config writes)
	rootEntry.OnChanged = func(s string) {
		persist(s)
	}

	return rootEntry, persist
}

func newGameListUI(w fyne.Window, logPrefix string) *gameListUI {
	ui := &gameListUI{w: w, logPrefix: logPrefix}

	ui.loadedLabel = widget.NewLabel("已加载 0 个游戏")

	ui.coverImg = canvas.NewImageFromResource(nil)
	ui.coverImg.FillMode = canvas.ImageFillContain
	ui.coverImg.SetMinSize(fyne.NewSize(300, 400))
	coverBox := container.New(layout.NewMaxLayout(), ui.coverImg)

	ui.gameDetail = widget.NewRichTextFromMarkdown("")
	ui.gameDetail.Wrapping = fyne.TextWrapWord
	ui.gameDetailScroll = container.NewVScroll(ui.gameDetail)
	ui.gameDetailScroll.SetMinSize(fyne.NewSize(320, 220))
	gameDetailBox := widget.NewCard("游戏详情", "", ui.gameDetailScroll)

	mediaTabs := container.NewAppTabs(
		container.NewTabItem("封面图片", coverBox),
		container.NewTabItem("视频预览", widget.NewLabel("Go 版暂不支持视频预览")),
	)
	ui.right = container.NewBorder(nil, gameDetailBox, nil, nil, mediaTabs)

	headers := []string{"选择", "序号", "游戏名称", "文件名称"}
	ui.table = widget.NewTable(
		func() (int, int) { return len(ui.filteredIdx) + 1, len(headers) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			lbl := obj.(*widget.Label)
			if id.Row == 0 {
				lbl.TextStyle = fyne.TextStyle{Bold: true}
				lbl.SetText(headers[id.Col])
				return
			}
			row := id.Row - 1
			if row < 0 || row >= len(ui.filteredIdx) {
				lbl.SetText("")
				return
			}
			g := ui.allGames[ui.filteredIdx[row]]
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
	ui.table.SetColumnWidth(0, 60)
	ui.table.SetColumnWidth(1, 40)
	ui.table.SetColumnWidth(2, 180)
	ui.table.SetColumnWidth(3, 180)

	ui.table.OnSelected = func(id widget.TableCellID) {
		if id.Row == 0 {
			return
		}
		row := id.Row - 1
		if row < 0 || row >= len(ui.filteredIdx) {
			return
		}
		idx := ui.filteredIdx[row]
		if id.Col == 0 {
			ui.allGames[idx].Selected = !ui.allGames[idx].Selected
			ui.table.Refresh()
		}
		ui.showDetailFor(ui.allGames[idx])
	}

	ui.search = widget.NewEntry()
	ui.search.SetPlaceHolder("搜索游戏名称或文件名称")
	ui.search.OnChanged = func(s string) {
		logging.Infof("%s search changed q=%s", logPrefix, s)
		ui.applyFilter(s)
		ui.table.Refresh()
	}

	ui.applyFilter("")
	return ui
}

func (ui *gameListUI) rootDir() string {
	if ui.rootEntry == nil {
		return ""
	}
	return strings.TrimSpace(ui.rootEntry.Text)
}

func (ui *gameListUI) applyFilter(query string) {
	q := strings.ToLower(strings.TrimSpace(query))
	ui.filteredIdx = ui.filteredIdx[:0]
	for i, g := range ui.allGames {
		if q == "" || strings.Contains(strings.ToLower(g.GameName), q) || strings.Contains(strings.ToLower(g.FileName), q) {
			ui.filteredIdx = append(ui.filteredIdx, i)
		}
	}
}

func (ui *gameListUI) setGames(games []pegasus.GameModel) {
	ui.allGames = games
	ui.applyFilter(ui.search.Text)
	ui.table.Refresh()
	ui.loadedLabel.SetText(fmt.Sprintf("已加载 %d 个游戏", len(ui.allGames)))
}

func (ui *gameListUI) selectAll(sel bool) {
	for i := range ui.allGames {
		ui.allGames[i].Selected = sel
	}
	ui.table.Refresh()
}

func (ui *gameListUI) selectedCount() int {
	selected := 0
	for _, g := range ui.allGames {
		if g.Selected {
			selected++
		}
	}
	return selected
}

func (ui *gameListUI) showDetailFor(g pegasus.GameModel) {
	boxFront := g.BoxFrontImagePath
	if boxFront != "" {
		if _, err := os.Stat(boxFront); err == nil {
			ui.coverImg.File = boxFront
			ui.coverImg.Resource = nil
			ui.coverImg.Refresh()
		} else {
			ui.coverImg.File = ""
			ui.coverImg.Resource = nil
			ui.coverImg.Refresh()
		}
	} else {
		ui.coverImg.File = ""
		ui.coverImg.Resource = nil
		ui.coverImg.Refresh()
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

	ui.gameDetail.ParseMarkdown(md.String())
	ui.gameDetail.Refresh()
	ui.gameDetailScroll.ScrollToTop()
}

func newChooseRootButton(w fyne.Window, rootEntry *widget.Entry, persist func(string), logPrefix string) *widget.Button {
	return widget.NewButton("设置根目录", func() {
		logging.Infof("%s click choose root", logPrefix)
		fd := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				logging.Errorf("%s open folder dialog failed err=%v", logPrefix, err)
				dialog.ShowError(err, w)
				return
			}
			if uri == nil {
				return
			}
			p := filepath.FromSlash(uri.Path())
			logging.Infof("%s chose rootDir=%s", logPrefix, p)
			rootEntry.SetText(p)
			persist(p)
		}, w)
		fd.Show()
	})
}

func newSearchRow(searchEntry *widget.Entry, logPrefix string) fyne.CanvasObject {
	clearSearchBtn := widget.NewButton("清除搜索", func() {
		logging.Infof("%s click clear search", logPrefix)
		searchEntry.SetText("")
	})
	return container.NewBorder(nil, nil, widget.NewLabel("搜索:"), clearSearchBtn, container.NewMax(searchEntry))
}
