package romconverter

import (
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
// Future: Add binary path, version, etc.
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
// Note: Tool names are static as they are proper nouns/brand names.
// Descriptions are in English; localization can be added in UI layer if needed.
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
}

// NewConverter creates a new ROM converter.
func NewConverter(
tool ConversionTool,
sourceDir, targetDir string,
sourceFormat, targetFormat SwitchFormat,
progressCallback func(ConversionProgress),
) *Converter {
return &Converter{
tool:         tool,
sourceDir:    sourceDir,
targetDir:    targetDir,
sourceFormat: sourceFormat,
targetFormat: targetFormat,
progressCB:   progressCallback,
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
// This is a placeholder implementation that will need to be extended
// with actual tool integration.
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

// TODO: Implement actual conversion logic based on the selected tool
// For now, this is a placeholder that simulates conversion
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

// convertFile converts a single ROM file.
// This is a placeholder that needs to be implemented with actual tool integration.
func (c *Converter) convertFile(sourcePath string) error {
// Placeholder implementation
// In a real implementation, this would:
// 1. Check if the conversion tool is installed
// 2. Download the tool if not installed
// 3. Execute the tool with appropriate parameters
// 4. Handle the conversion output

// For now, return an error indicating this is not yet implemented
return fmt.Errorf("conversion not yet implemented - tool integration required")
}

// IsToolInstalled checks if the selected conversion tool is installed.
func (c *Converter) IsToolInstalled() bool {
// Placeholder - will check if tool binary exists
return false
}

// DownloadTool downloads and installs the selected conversion tool.
func (c *Converter) DownloadTool() error {
// Placeholder for tool download logic
// Will implement GitHub release download in the future
return fmt.Errorf("tool download not yet implemented")
}
