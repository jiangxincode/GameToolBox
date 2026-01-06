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

func NewGameAdderView(w fyne.Window) fyne.CanvasObject {
	logPrefix := logging.PrefixFromCallerSkip(0)

	rootEntry, persistRootDir := newRootEntryWithConfig(w, logPrefix)
	chooseRootBtn := newChooseRootButton(w, rootEntry, persistRootDir, logPrefix)

	// Input fields for game information
	gameNameEntry := widget.NewEntry()
	gameNameEntry.SetPlaceHolder("输入游戏名称（必填）")

	fileNameEntry := widget.NewEntry()
	fileNameEntry.SetPlaceHolder("输入ROM文件名（必填，相对路径）")

	developerEntry := widget.NewEntry()
	developerEntry.SetPlaceHolder("输入开发者（可选）")

	descriptionEntry := widget.NewMultiLineEntry()
	descriptionEntry.SetPlaceHolder("输入游戏描述（可选）")
	descriptionEntry.SetMinRowsVisible(3)

	// Media file entries
	logoEntry := widget.NewEntry()
	logoEntry.SetPlaceHolder("Logo图片路径（可选）")

	boxFrontEntry := widget.NewEntry()
	boxFrontEntry.SetPlaceHolder("封面图片路径（可选）")

	videoEntry := widget.NewEntry()
	videoEntry.SetPlaceHolder("视频文件路径（可选）")

	// File chooser button for ROM file
	chooseRomBtn := widget.NewButton("选择ROM文件", func() {
		root := strings.TrimSpace(rootEntry.Text)
		if root == "" {
			dialog.ShowInformation("提示", "请先设置根目录", w)
			return
		}

		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if reader == nil {
				return
			}
			defer reader.Close()

			absPath := reader.URI().Path()
			logging.Infof("%s selected rom file: %s", logPrefix, absPath)

			// Calculate relative path from root directory
			relPath, err := filepath.Rel(root, absPath)
			if err != nil {
				// If we can't get relative path, warn user and use absolute path
				dialog.ShowInformation("提示",
					fmt.Sprintf("所选文件不在根目录内，将使用绝对路径:\n%s", absPath), w)
				relPath = absPath
			} else {
				// Convert to forward slashes for consistency
				relPath = filepath.ToSlash(relPath)
			}

			fileNameEntry.SetText(relPath)

			// Auto-fill game name if empty
			if strings.TrimSpace(gameNameEntry.Text) == "" {
				baseName := filepath.Base(absPath)
				gameName := strings.TrimSuffix(baseName, filepath.Ext(baseName))
				gameNameEntry.SetText(gameName)
			}
		}, w)
	})

	// Logo file chooser button
	chooseLogoBtn := widget.NewButton("选择Logo", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if reader == nil {
				return
			}
			defer reader.Close()

			absPath := reader.URI().Path()
			logging.Infof("%s selected logo file: %s", logPrefix, absPath)
			logoEntry.SetText(absPath)
		}, w)
	})

	// BoxFront file chooser button
	chooseBoxFrontBtn := widget.NewButton("选择封面", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if reader == nil {
				return
			}
			defer reader.Close()

			absPath := reader.URI().Path()
			logging.Infof("%s selected boxFront file: %s", logPrefix, absPath)
			boxFrontEntry.SetText(absPath)
		}, w)
	})

	// Video file chooser button
	chooseVideoBtn := widget.NewButton("选择视频", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if reader == nil {
				return
			}
			defer reader.Close()

			absPath := reader.URI().Path()
			logging.Infof("%s selected video file: %s", logPrefix, absPath)
			videoEntry.SetText(absPath)
		}, w)
	})

	// Add game button
	addGameBtn := widget.NewButton("添加游戏", func() {
		root := strings.TrimSpace(rootEntry.Text)
		logging.Infof("%s click add game root=%s", logPrefix, root)

		if root == "" {
			dialog.ShowInformation("提示", "请先设置根目录", w)
			return
		}

		gameName := strings.TrimSpace(gameNameEntry.Text)
		fileName := strings.TrimSpace(fileNameEntry.Text)

		if gameName == "" {
			dialog.ShowInformation("提示", "请输入游戏名称", w)
			return
		}

		if fileName == "" {
			dialog.ShowInformation("提示", "请输入ROM文件名", w)
			return
		}

		game := pegasus.GameModel{
			GameName:          gameName,
			FileName:          fileName,
			Developer:         strings.TrimSpace(developerEntry.Text),
			Description:       strings.TrimSpace(descriptionEntry.Text),
			LogoImagePath:     strings.TrimSpace(logoEntry.Text),
			BoxFrontImagePath: strings.TrimSpace(boxFrontEntry.Text),
			VideoFilePath:     strings.TrimSpace(videoEntry.Text),
		}

		res := pegasus.AddSingleGame(root, game)
		if len(res.Errors) > 0 {
			errMsg := fmt.Sprintf("添加失败: %s", res.Errors[0].Error())
			dialog.ShowError(fmt.Errorf("%s", errMsg), w)
			return
		}

		if res.Skipped > 0 {
			dialog.ShowInformation("提示", "游戏已存在，已跳过", w)
			logging.Infof("%s game skipped (already exists): %s", logPrefix, gameName)
			return
		}

		dialog.ShowInformation("提示", fmt.Sprintf("游戏添加成功!\n已添加: %d", res.Added), w)
		logging.Infof("%s game added successfully: %s", logPrefix, gameName)

		// Clear input fields after successful addition
		gameNameEntry.SetText("")
		fileNameEntry.SetText("")
		developerEntry.SetText("")
		descriptionEntry.SetText("")
		logoEntry.SetText("")
		boxFrontEntry.SetText("")
		videoEntry.SetText("")
	})

	// Clear button
	clearBtn := widget.NewButton("清空表单", func() {
		logging.Infof("%s click clear form", logPrefix)
		gameNameEntry.SetText("")
		fileNameEntry.SetText("")
		developerEntry.SetText("")
		descriptionEntry.SetText("")
		logoEntry.SetText("")
		boxFrontEntry.SetText("")
		videoEntry.SetText("")
	})

	// Create form layout
	form := container.NewVBox(
		widget.NewLabel("游戏名称:"),
		gameNameEntry,
		widget.NewLabel("ROM文件:"),
		container.NewBorder(nil, nil, nil, chooseRomBtn, fileNameEntry),
		widget.NewLabel("开发者:"),
		developerEntry,
		widget.NewLabel("描述:"),
		descriptionEntry,
		widget.NewSeparator(),
		widget.NewLabel("媒体文件 (可选):"),
		widget.NewLabel("Logo图片:"),
		container.NewBorder(nil, nil, nil, chooseLogoBtn, logoEntry),
		widget.NewLabel("封面图片:"),
		container.NewBorder(nil, nil, nil, chooseBoxFrontBtn, boxFrontEntry),
		widget.NewLabel("视频文件:"),
		container.NewBorder(nil, nil, nil, chooseVideoBtn, videoEntry),
		container.NewHBox(addGameBtn, clearBtn),
	)

	// Instructions
	instructions := widget.NewLabel(
		"说明:\n" +
			"1. 设置Pegasus根目录\n" +
			"2. 填写游戏信息（游戏名称和ROM文件名为必填项）\n" +
			"3. 可以点击\"选择ROM文件\"按钮自动填充文件路径\n" +
			"4. 可选：添加媒体文件（Logo图、封面图、视频）\n" +
			"5. 点击\"添加游戏\"将游戏添加到metadata.pegasus.txt\n" +
			"6. 媒体文件将自动复制到media/<游戏名称>/目录\n" +
			"7. 添加成功后表单会自动清空，可继续添加下一个游戏",
	)
	instructions.Wrapping = fyne.TextWrapWord

	content := container.NewBorder(
		container.NewVBox(
			container.NewBorder(nil, nil, nil, chooseRootBtn, rootEntry),
			widget.NewSeparator(),
		),
		container.NewVBox(
			widget.NewSeparator(),
			instructions,
		),
		nil,
		nil,
		container.NewPadded(form),
	)

	return content
}
