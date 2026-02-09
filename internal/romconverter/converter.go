package romconverter

import (
"context"
"fmt"
"os"
"path/filepath"
"strings"
)

// ConversionTool represents a ROM conversion tool.
type ConversionTool struct {
Name        string // Display name
ID          string // Unique identifier
Description string // Tool description
GitHubRepo  string // GitHub repository (owner/repo)
}

// SwitchFormat represents a Nintendo Switch ROM format.
type SwitchFormat string

const (
FormatXCI SwitchFormat = "XCI"
FormatNSP SwitchFormat = "NSP"
FormatNSZ SwitchFormat = "NSZ"
FormatXCZ SwitchFormat = "XCZ"
)

// GetSwitchFormats returns all supported Switch ROM formats.
func GetSwitchFormats() []SwitchFormat {
return []SwitchFormat{FormatXCI, FormatNSP, FormatNSZ, FormatXCZ}
}

// GetSwitchTools returns available Switch ROM conversion tools.
func GetSwitchTools() []ConversionTool {
return []ConversionTool{
{
Name:        "nsz (nicoboss/nsz)",
ID:          "nsz",
Description: "NSZ/XCZ compression/decompression tool",
GitHubRepo:  "nicoboss/nsz",
},
{
Name:        "4NXCI",
ID:          "4nxci",
Description: "XCI to NSP converter",
GitHubRepo:  "The-4n/4NXCI",
},
{
Name:        "Custom Tool",
ID:          "custom",
Description: "User-defined conversion tool",
GitHubRepo:  "",
},
}
}

// ConversionProgress represents the current state of a conversion operation.
type ConversionProgress struct {
CurrentFile  string
SuccessCount int
FailureCount int
TotalCount   int
IsRunning    bool
ErrorMessage string
}

// ConversionResult represents the final result of a conversion.
type ConversionResult struct {
SuccessCount int
FailureCount int
Errors       []error
}

// Converter handles ROM format conversion operations.
type Converter struct {
tool         ConversionTool
sourceDir    string
targetDir    string
sourceFormat SwitchFormat
targetFormat SwitchFormat
progressCB   func(ConversionProgress)
toolManager  *ToolManager
ctx          context.Context
}

// NewConverter creates a new ROM converter.
func NewConverter(
tool ConversionTool,
sourceDir, targetDir string,
sourceFormat, targetFormat SwitchFormat,
progressCallback func(ConversionProgress),
) *Converter {
toolManager, _ := NewToolManager()
return &Converter{
tool:         tool,
sourceDir:    sourceDir,
targetDir:    targetDir,
sourceFormat: sourceFormat,
targetFormat: targetFormat,
progressCB:   progressCallback,
toolManager:  toolManager,
ctx:          context.Background(),
}
}

// Validate checks if the converter configuration is valid.
func (c *Converter) Validate() error {
if c.sourceDir == "" {
return fmt.Errorf("source directory is empty")
}
if c.targetDir == "" {
return fmt.Errorf("target directory is empty")
}
if c.sourceDir == c.targetDir {
return fmt.Errorf("source and target directories cannot be the same")
}
if _, err := os.Stat(c.sourceDir); os.IsNotExist(err) {
return fmt.Errorf("source directory does not exist: %s", c.sourceDir)
}
return nil
}

// Convert performs the ROM format conversion.
func (c *Converter) Convert() ConversionResult {
result := ConversionResult{}

if err := c.Validate(); err != nil {
result.Errors = append(result.Errors, err)
return result
}

// Create target directory if it doesn't exist
if err := os.MkdirAll(c.targetDir, 0755); err != nil {
result.Errors = append(result.Errors, fmt.Errorf("failed to create target directory: %w", err))
return result
}

// Check if tool is installed, if not download it
if c.toolManager != nil && !c.toolManager.IsToolInstalled(c.tool) {
if c.tool.ID != "custom" {
if c.progressCB != nil {
c.progressCB(ConversionProgress{
CurrentFile:  "Downloading tool...",
SuccessCount: 0,
FailureCount: 0,
TotalCount:   0,
IsRunning:    true,
})
}

err := c.toolManager.DownloadTool(c.ctx, c.tool, func(msg string) {
if c.progressCB != nil {
c.progressCB(ConversionProgress{
CurrentFile:  msg,
SuccessCount: 0,
FailureCount: 0,
TotalCount:   0,
IsRunning:    true,
})
}
})

if err != nil {
result.Errors = append(result.Errors, fmt.Errorf("failed to download tool: %w", err))
return result
}
}
}

// Find all ROM files with the source format
romFiles, err := c.findROMFiles()
if err != nil {
result.Errors = append(result.Errors, err)
return result
}

if len(romFiles) == 0 {
result.Errors = append(result.Errors, fmt.Errorf("no ROM files found with format %s", c.sourceFormat))
return result
}

// Process each ROM file
for i, romFile := range romFiles {
if c.progressCB != nil {
c.progressCB(ConversionProgress{
CurrentFile:  filepath.Base(romFile),
SuccessCount: result.SuccessCount,
FailureCount: result.FailureCount,
TotalCount:   len(romFiles),
IsRunning:    true,
})
}

err := c.convertFile(romFile)
if err != nil {
result.FailureCount++
result.Errors = append(result.Errors, fmt.Errorf("%s: %w", filepath.Base(romFile), err))
} else {
result.SuccessCount++
}

// Progress update after each file
if c.progressCB != nil && i == len(romFiles)-1 {
c.progressCB(ConversionProgress{
CurrentFile:  filepath.Base(romFile),
SuccessCount: result.SuccessCount,
FailureCount: result.FailureCount,
TotalCount:   len(romFiles),
IsRunning:    false,
})
}
}

return result
}

// findROMFiles finds all ROM files in the source directory with the source format.
func (c *Converter) findROMFiles() ([]string, error) {
var files []string
ext := "." + strings.ToLower(string(c.sourceFormat))

err := filepath.Walk(c.sourceDir, func(path string, info os.FileInfo, err error) error {
if err != nil {
return err
}
if !info.IsDir() && strings.ToLower(filepath.Ext(path)) == ext {
files = append(files, path)
}
return nil
})

return files, err
}

// convertFile converts a single ROM file using the tool manager.
func (c *Converter) convertFile(sourcePath string) error {
if c.toolManager == nil {
return fmt.Errorf("tool manager not initialized")
}

// Use the tool manager to execute the conversion
return c.toolManager.ExecuteConversion(c.tool, sourcePath, c.targetDir, c.sourceFormat, c.targetFormat)
}

// IsToolInstalled checks if the selected conversion tool is installed.
func (c *Converter) IsToolInstalled() bool {
if c.toolManager == nil {
return false
}
return c.toolManager.IsToolInstalled(c.tool)
}

// DownloadTool downloads and installs the selected conversion tool.
func (c *Converter) DownloadTool() error {
if c.toolManager == nil {
return fmt.Errorf("tool manager not initialized")
}
return c.toolManager.DownloadTool(c.ctx, c.tool, nil)
}

// FCFormat represents a Famicom/NES ROM format.
type FCFormat string

const (
FormatNES FCFormat = "NES"
FormatFDS FCFormat = "FDS"
FormatUNF FCFormat = "UNF"
FormatNSF FCFormat = "NSF"
)

// GetFCFormats returns all supported FC/NES ROM formats.
func GetFCFormats() []FCFormat {
return []FCFormat{FormatNES, FormatFDS, FormatUNF, FormatNSF}
}

// GetFCTools returns available FC/NES ROM conversion tools.
func GetFCTools() []ConversionTool {
return []ConversionTool{
{
Name:        "NES ROM Tool",
ID:          "nesromtool",
Description: "NES ROM header editor and converter",
GitHubRepo:  "",
},
{
Name:        "FDS Converter",
ID:          "fdsconv",
Description: "FDS to NES converter",
GitHubRepo:  "",
},
{
Name:        "Custom Tool",
ID:          "custom",
Description: "User-defined conversion tool",
GitHubRepo:  "",
},
}
}

// FCConverter handles FC/NES ROM format conversion operations.
type FCConverter struct {
tool         ConversionTool
sourceDir    string
targetDir    string
sourceFormat FCFormat
targetFormat FCFormat
progressCB   func(ConversionProgress)
toolManager  *ToolManager
ctx          context.Context
}

// NewFCConverter creates a new FC/NES ROM converter.
func NewFCConverter(
tool ConversionTool,
sourceDir, targetDir string,
sourceFormat, targetFormat FCFormat,
progressCallback func(ConversionProgress),
) *FCConverter {
toolManager, _ := NewToolManager()
return &FCConverter{
tool:         tool,
sourceDir:    sourceDir,
targetDir:    targetDir,
sourceFormat: sourceFormat,
targetFormat: targetFormat,
progressCB:   progressCallback,
toolManager:  toolManager,
ctx:          context.Background(),
}
}

// Validate checks if the FC converter configuration is valid.
func (c *FCConverter) Validate() error {
if c.sourceDir == "" {
return fmt.Errorf("source directory is empty")
}
if c.targetDir == "" {
return fmt.Errorf("target directory is empty")
}
if c.sourceDir == c.targetDir {
return fmt.Errorf("source and target directories cannot be the same")
}
if _, err := os.Stat(c.sourceDir); os.IsNotExist(err) {
return fmt.Errorf("source directory does not exist: %s", c.sourceDir)
}
return nil
}

// Convert performs the FC/NES ROM format conversion.
func (c *FCConverter) Convert() ConversionResult {
result := ConversionResult{}

if err := c.Validate(); err != nil {
result.Errors = append(result.Errors, err)
return result
}

// Create target directory if it doesn't exist
if err := os.MkdirAll(c.targetDir, 0755); err != nil {
result.Errors = append(result.Errors, fmt.Errorf("failed to create target directory: %w", err))
return result
}

// Find all ROM files with the source format
romFiles, err := c.findROMFiles()
if err != nil {
result.Errors = append(result.Errors, err)
return result
}

if len(romFiles) == 0 {
result.Errors = append(result.Errors, fmt.Errorf("no ROM files found with format %s", c.sourceFormat))
return result
}

// Process each ROM file
for i, romFile := range romFiles {
if c.progressCB != nil {
c.progressCB(ConversionProgress{
CurrentFile:  filepath.Base(romFile),
SuccessCount: result.SuccessCount,
FailureCount: result.FailureCount,
TotalCount:   len(romFiles),
IsRunning:    true,
})
}

err := c.convertFile(romFile)
if err != nil {
result.FailureCount++
result.Errors = append(result.Errors, fmt.Errorf("%s: %w", filepath.Base(romFile), err))
} else {
result.SuccessCount++
}

// Progress update after each file
if c.progressCB != nil && i == len(romFiles)-1 {
c.progressCB(ConversionProgress{
CurrentFile:  filepath.Base(romFile),
SuccessCount: result.SuccessCount,
FailureCount: result.FailureCount,
TotalCount:   len(romFiles),
IsRunning:    false,
})
}
}

return result
}

// findROMFiles finds all ROM files in the source directory with the source format.
func (c *FCConverter) findROMFiles() ([]string, error) {
var files []string
ext := "." + strings.ToLower(string(c.sourceFormat))

err := filepath.Walk(c.sourceDir, func(path string, info os.FileInfo, err error) error {
if err != nil {
return err
}
if !info.IsDir() && strings.ToLower(filepath.Ext(path)) == ext {
files = append(files, path)
}
return nil
})

return files, err
}

// convertFile converts a single FC/NES ROM file.
func (c *FCConverter) convertFile(sourcePath string) error {
if c.toolManager == nil {
return fmt.Errorf("tool manager not initialized")
}

// Use the tool manager to execute the conversion
return c.toolManager.ExecuteFCConversion(c.tool, sourcePath, c.targetDir, c.sourceFormat, c.targetFormat)
}

// IsToolInstalled checks if the selected conversion tool is installed.
func (c *FCConverter) IsToolInstalled() bool {
if c.toolManager == nil {
return false
}
return c.toolManager.IsToolInstalled(c.tool)
}

// DownloadTool downloads and installs the selected conversion tool.
func (c *FCConverter) DownloadTool() error {
if c.toolManager == nil {
return fmt.Errorf("tool manager not initialized")
}
return c.toolManager.DownloadTool(c.ctx, c.tool, nil)
}
