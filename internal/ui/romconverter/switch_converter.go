package romconverterui

import (
"fmt"
"path/filepath"
"strings"

"fyne.io/fyne/v2"
"fyne.io/fyne/v2/container"
"fyne.io/fyne/v2/dialog"
"fyne.io/fyne/v2/widget"

"github.com/game_tool_box/internal/config"
"github.com/game_tool_box/internal/i18n"
"github.com/game_tool_box/internal/logging"
"github.com/game_tool_box/internal/romconverter"
)

// NewSwitchConverterView creates the Switch ROM converter UI.
func NewSwitchConverterView(w fyne.Window) fyne.CanvasObject {
logPrefix := logging.PrefixFromCallerSkip(0)
t := func(key string) string {
return i18n.T(i18n.Current(), key)
}

// Tool selection
tools := romconverter.GetSwitchTools()
toolNames := make([]string, len(tools))
for i, tool := range tools {
toolNames[i] = tool.Name
}
toolSelect := widget.NewSelect(toolNames, nil)
toolSelect.PlaceHolder = t("romConverter.selectTool")

// Format selection
formats := romconverter.GetSwitchFormats()
formatNames := make([]string, len(formats))
for i, format := range formats {
formatNames[i] = string(format)
}

sourceFormatSelect := widget.NewSelect(formatNames, nil)
sourceFormatSelect.PlaceHolder = t("romConverter.sourceFormat")
sourceFormatSelect.Selected = string(romconverter.FormatNSP) // Default

targetFormatSelect := widget.NewSelect(formatNames, nil)
targetFormatSelect.PlaceHolder = t("romConverter.targetFormat")
targetFormatSelect.Selected = string(romconverter.FormatXCI) // Default

// Directory entries
sourceEntry := widget.NewEntry()
sourceEntry.SetPlaceHolder(t("romConverter.sourceDir"))
targetEntry := widget.NewEntry()
targetEntry.SetPlaceHolder(t("romConverter.targetDir"))

// Load saved directories
if cfg, err := config.Load(); err == nil {
if strings.TrimSpace(cfg.RomConverterSourceDir) != "" {
sourceEntry.SetText(cfg.RomConverterSourceDir)
}
if strings.TrimSpace(cfg.RomConverterTargetDir) != "" {
targetEntry.SetText(cfg.RomConverterTargetDir)
}
}

// Persist directory changes
persistSourceDir := func(path string) {
if strings.TrimSpace(path) == "" {
return
}
c, _ := config.Load()
c.RomConverterSourceDir = path
_ = config.Save(c)
logging.Infof("%s persist source dir=%s", logPrefix, path)
}

persistTargetDir := func(path string) {
if strings.TrimSpace(path) == "" {
return
}
c, _ := config.Load()
c.RomConverterTargetDir = path
_ = config.Save(c)
logging.Infof("%s persist target dir=%s", logPrefix, path)
}

sourceEntry.OnChanged = persistSourceDir
targetEntry.OnChanged = persistTargetDir

// Choose directory buttons
chooseSourceBtn := widget.NewButton(t("romConverter.chooseDir"), func() {
logging.Infof("%s click choose source dir", logPrefix)
fd := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
if err != nil {
dialog.ShowError(err, w)
return
}
if uri != nil {
path := filepath.FromSlash(uri.Path())
sourceEntry.SetText(path)
persistSourceDir(path)
}
}, w)
fd.Show()
})

chooseTargetBtn := widget.NewButton(t("romConverter.chooseDir"), func() {
logging.Infof("%s click choose target dir", logPrefix)
fd := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
if err != nil {
dialog.ShowError(err, w)
return
}
if uri != nil {
path := filepath.FromSlash(uri.Path())
targetEntry.SetText(path)
persistTargetDir(path)
}
}, w)
fd.Show()
})

// Status labels
statusLabel := widget.NewLabel(t("romConverter.status") + ": " + t("romConverter.idle"))
currentFileLabel := widget.NewLabel(t("romConverter.currentFile") + ": -")
progressLabel := widget.NewLabel(fmt.Sprintf("%s: 0 | %s: 0 | %s: 0",
t("romConverter.successCount"),
t("romConverter.failureCount"),
t("romConverter.totalCount")))

// Conversion button - define early so it can be referenced in the click handler
convertBtn := widget.NewButton(t("romConverter.convert"), nil)
convertBtn.OnTapped = func() {
logging.Infof("%s click convert", logPrefix)

// Validate inputs
if toolSelect.Selected == "" {
dialog.ShowInformation(t("romConverter.error.noTool"), t("romConverter.error.noTool"), w)
return
}
if strings.TrimSpace(sourceEntry.Text) == "" {
dialog.ShowInformation(t("romConverter.error.noSourceDir"), t("romConverter.error.noSourceDir"), w)
return
}
if strings.TrimSpace(targetEntry.Text) == "" {
dialog.ShowInformation(t("romConverter.error.noTargetDir"), t("romConverter.error.noTargetDir"), w)
return
}
if sourceEntry.Text == targetEntry.Text {
dialog.ShowInformation(t("romConverter.error.sameDir"), t("romConverter.error.sameDir"), w)
return
}

// Get selected tool
var selectedTool romconverter.ConversionTool
for _, tool := range tools {
if tool.Name == toolSelect.Selected {
selectedTool = tool
break
}
}

sourceFormat := romconverter.SwitchFormat(sourceFormatSelect.Selected)
targetFormat := romconverter.SwitchFormat(targetFormatSelect.Selected)

// Create converter with progress callback
converter := romconverter.NewConverter(
selectedTool,
sourceEntry.Text,
targetEntry.Text,
sourceFormat,
targetFormat,
func(progress romconverter.ConversionProgress) {
// Update UI on progress
fyne.Do(func() {
currentFileLabel.SetText(t("romConverter.currentFile") + ": " + progress.CurrentFile)
progressLabel.SetText(fmt.Sprintf("%s: %d | %s: %d | %s: %d",
t("romConverter.successCount"), progress.SuccessCount,
t("romConverter.failureCount"), progress.FailureCount,
t("romConverter.totalCount"), progress.TotalCount))

if progress.IsRunning {
statusLabel.SetText(t("romConverter.status") + ": " + t("romConverter.converting"))
} else {
statusLabel.SetText(t("romConverter.status") + ": " + t("romConverter.complete"))
}
})
},
)

// Disable button during conversion
convertBtn.Disable()
statusLabel.SetText(t("romConverter.status") + ": " + t("romConverter.converting"))

// Run conversion in background
go func() {
result := converter.Convert()

// Update UI with final result
fyne.Do(func() {
convertBtn.Enable()
statusLabel.SetText(t("romConverter.status") + ": " + t("romConverter.complete"))

msg := fmt.Sprintf("%s: %d\n%s: %d",
t("romConverter.successCount"), result.SuccessCount,
t("romConverter.failureCount"), result.FailureCount)

if len(result.Errors) > 0 {
// Show first few errors
errorMsg := msg + "\n\n错误信息:\n"
maxErrors := 5
for i, err := range result.Errors {
if i >= maxErrors {
errorMsg += fmt.Sprintf("\n...还有 %d 个错误", len(result.Errors)-maxErrors)
break
}
errorMsg += fmt.Sprintf("\n%s", err.Error())
}
dialog.ShowInformation(t("romConverter.complete"), errorMsg, w)
} else {
dialog.ShowInformation(t("romConverter.complete"), msg, w)
}

logging.Infof("%s conversion complete success=%d failure=%d errors=%d",
logPrefix, result.SuccessCount, result.FailureCount, len(result.Errors))
})
}()
}

// Layout
toolRow := container.NewBorder(nil, nil, widget.NewLabel(t("romConverter.tool")+":"), nil, toolSelect)
formatRow := container.NewHBox(
widget.NewLabel(t("romConverter.sourceFormat")+":"),
sourceFormatSelect,
widget.NewLabel("→"),
widget.NewLabel(t("romConverter.targetFormat")+":"),
targetFormatSelect,
)
sourceRow := container.NewBorder(nil, nil, nil, chooseSourceBtn, sourceEntry)
targetRow := container.NewBorder(nil, nil, nil, chooseTargetBtn, targetEntry)

form := container.NewVBox(
toolRow,
formatRow,
widget.NewLabel(t("romConverter.sourceDir")+":"),
sourceRow,
widget.NewLabel(t("romConverter.targetDir")+":"),
targetRow,
widget.NewSeparator(),
convertBtn,
)

statusBox := container.NewVBox(
widget.NewSeparator(),
statusLabel,
currentFileLabel,
progressLabel,
)

content := container.NewBorder(form, statusBox, nil, nil, container.NewMax())
return container.NewPadded(content)
}
