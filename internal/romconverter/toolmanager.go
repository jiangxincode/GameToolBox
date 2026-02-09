package romconverter

import (
"archive/zip"
"context"
"encoding/json"
"fmt"
"io"
"net/http"
"os"
"os/exec"
"path/filepath"
"runtime"
"strings"
"time"

"github.com/game_tool_box/internal/config"
"github.com/game_tool_box/internal/logging"
)

// ToolManager manages ROM conversion tools.
type ToolManager struct {
toolsDir string
}

// NewToolManager creates a new tool manager.
func NewToolManager() (*ToolManager, error) {
configDir, err := config.Dir()
if err != nil {
return nil, fmt.Errorf("failed to get config dir: %w", err)
}
toolsDir := filepath.Join(configDir, "tools")
if err := os.MkdirAll(toolsDir, 0755); err != nil {
return nil, fmt.Errorf("failed to create tools dir: %w", err)
}
return &ToolManager{toolsDir: toolsDir}, nil
}

// GetToolPath returns the path to a tool's executable.
func (tm *ToolManager) GetToolPath(tool ConversionTool) string {
toolDir := filepath.Join(tm.toolsDir, tool.ID)

// For nsz tool (Python-based)
if tool.ID == "nsz" {
if runtime.GOOS == "windows" {
return filepath.Join(toolDir, "nsz.bat")
}
return filepath.Join(toolDir, "nsz.py")
}

// For 4NXCI tool
if tool.ID == "4nxci" {
if runtime.GOOS == "windows" {
return filepath.Join(toolDir, "4nxci.exe")
}
return filepath.Join(toolDir, "4nxci")
}

// For FC/NES tools - return directory path for manual setup
if tool.ID == "nesromtool" || tool.ID == "fdsconv" {
return toolDir
}

return ""
}

// IsToolInstalled checks if a tool is installed.
func (tm *ToolManager) IsToolInstalled(tool ConversionTool) bool {
if tool.ID == "custom" {
// Custom tools are considered "installed" if user provides their own
return true
}

// For FC/NES tools, check if directory with README exists
if tool.ID == "nesromtool" || tool.ID == "fdsconv" {
toolDir := tm.GetToolPath(tool)
readmePath := filepath.Join(toolDir, "README.txt")
_, err := os.Stat(readmePath)
return err == nil
}

toolPath := tm.GetToolPath(tool)
if toolPath == "" {
return false
}

_, err := os.Stat(toolPath)
return err == nil
}

// GitHubRelease represents a GitHub release.
type GitHubRelease struct {
TagName string `json:"tag_name"`
Assets  []struct {
Name               string `json:"name"`
BrowserDownloadURL string `json:"browser_download_url"`
} `json:"assets"`
}

// DownloadTool downloads and installs a tool from GitHub.
func (tm *ToolManager) DownloadTool(ctx context.Context, tool ConversionTool, progressCallback func(string)) error {
if tool.ID == "custom" {
return fmt.Errorf("custom tool cannot be downloaded automatically")
}

if tool.GitHubRepo == "" {
return fmt.Errorf("no GitHub repository specified for tool %s", tool.Name)
}

logging.Infof("Downloading tool %s from %s", tool.Name, tool.GitHubRepo)
if progressCallback != nil {
progressCallback(fmt.Sprintf("Fetching release info for %s...", tool.Name))
}

// Get latest release info
release, err := tm.getLatestRelease(ctx, tool.GitHubRepo)
if err != nil {
// If no GitHub repo or release fails, check if it's an FC tool
if tool.ID == "nesromtool" || tool.ID == "fdsconv" {
return tm.downloadFCTool(ctx, tool, progressCallback)
}
return fmt.Errorf("failed to get latest release: %w", err)
}

// For nsz, we need to clone the repo instead of downloading a release
// since it's a Python script-based tool
if tool.ID == "nsz" {
return tm.downloadNszTool(ctx, tool, progressCallback)
}

// For 4NXCI, download the appropriate release asset
if tool.ID == "4nxci" {
return tm.download4NXCITool(ctx, tool, release, progressCallback)
}

// For FC/NES tools without GitHub releases
if tool.ID == "nesromtool" || tool.ID == "fdsconv" {
return tm.downloadFCTool(ctx, tool, progressCallback)
}

return fmt.Errorf("unsupported tool: %s", tool.ID)
}

// getLatestRelease fetches the latest release from GitHub.
func (tm *ToolManager) getLatestRelease(ctx context.Context, repo string) (*GitHubRelease, error) {
url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)

req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
if err != nil {
return nil, err
}

client := &http.Client{Timeout: 30 * time.Second}
resp, err := client.Do(req)
if err != nil {
return nil, err
}
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
}

var release GitHubRelease
if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
return nil, err
}

return &release, nil
}

// downloadNszTool downloads the nsz Python tool.
func (tm *ToolManager) downloadNszTool(ctx context.Context, tool ConversionTool, progressCallback func(string)) error {
toolDir := filepath.Join(tm.toolsDir, tool.ID)
if err := os.MkdirAll(toolDir, 0755); err != nil {
return err
}

if progressCallback != nil {
progressCallback("Downloading nsz tool (Python scripts)...")
}

// Download the main nsz.py script
nszURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/master/nsz.py", tool.GitHubRepo)
nszPath := filepath.Join(toolDir, "nsz.py")

if err := tm.downloadFile(ctx, nszURL, nszPath, progressCallback); err != nil {
return fmt.Errorf("failed to download nsz.py: %w", err)
}

// Make executable on Unix
if runtime.GOOS != "windows" {
if err := os.Chmod(nszPath, 0755); err != nil {
return err
}
}

// Create a simple README
readmePath := filepath.Join(toolDir, "README.txt")
readme := `NSZ Tool Downloaded Successfully

This is a Python-based tool. To use it:
1. Ensure Python 3.x is installed
2. Install required dependencies: pip install pycryptodomex zstandard

The tool will be executed through Python when you run conversions.
`
if err := os.WriteFile(readmePath, []byte(readme), 0644); err != nil {
logging.Infof("Failed to create README: %v", err)
}

logging.Infof("nsz tool installed successfully at %s", toolDir)
return nil
}

// download4NXCITool downloads the 4NXCI tool.
func (tm *ToolManager) download4NXCITool(ctx context.Context, tool ConversionTool, release *GitHubRelease, progressCallback func(string)) error {
toolDir := filepath.Join(tm.toolsDir, tool.ID)
if err := os.MkdirAll(toolDir, 0755); err != nil {
return err
}

// Find the appropriate asset for current OS
var assetURL string
var assetName string

osName := runtime.GOOS
for _, asset := range release.Assets {
name := strings.ToLower(asset.Name)

if osName == "windows" && strings.Contains(name, "win") {
assetURL = asset.BrowserDownloadURL
assetName = asset.Name
break
} else if osName == "linux" && strings.Contains(name, "linux") {
assetURL = asset.BrowserDownloadURL
assetName = asset.Name
break
} else if osName == "darwin" && (strings.Contains(name, "mac") || strings.Contains(name, "darwin")) {
assetURL = asset.BrowserDownloadURL
assetName = asset.Name
break
}
}

if assetURL == "" {
return fmt.Errorf("no suitable release found for %s", osName)
}

if progressCallback != nil {
progressCallback(fmt.Sprintf("Downloading %s...", assetName))
}

// Download the asset
tempFile := filepath.Join(toolDir, assetName)
if err := tm.downloadFile(ctx, assetURL, tempFile, progressCallback); err != nil {
return fmt.Errorf("failed to download asset: %w", err)
}

// If it's a zip, extract it
if strings.HasSuffix(strings.ToLower(assetName), ".zip") {
if progressCallback != nil {
progressCallback("Extracting archive...")
}
if err := tm.extractZip(tempFile, toolDir); err != nil {
return fmt.Errorf("failed to extract zip: %w", err)
}
os.Remove(tempFile) // Clean up zip file
}

// Make executable on Unix
if runtime.GOOS != "windows" {
exePath := filepath.Join(toolDir, "4nxci")
if err := os.Chmod(exePath, 0755); err != nil {
logging.Infof("Failed to make executable: %v", err)
}
}

logging.Infof("4NXCI tool installed successfully at %s", toolDir)
return nil
}

// downloadFile downloads a file from a URL.
func (tm *ToolManager) downloadFile(ctx context.Context, url, destPath string, progressCallback func(string)) error {
req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
if err != nil {
return err
}

client := &http.Client{Timeout: 5 * time.Minute}
resp, err := client.Do(req)
if err != nil {
return err
}
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
return fmt.Errorf("download failed with status %d", resp.StatusCode)
}

out, err := os.Create(destPath)
if err != nil {
return err
}
defer out.Close()

_, err = io.Copy(out, resp.Body)
return err
}

// extractZip extracts a zip file to a destination directory.
func (tm *ToolManager) extractZip(zipPath, destDir string) error {
r, err := zip.OpenReader(zipPath)
if err != nil {
return err
}
defer r.Close()

for _, f := range r.File {
fpath := filepath.Join(destDir, f.Name)

if f.FileInfo().IsDir() {
os.MkdirAll(fpath, os.ModePerm)
continue
}

if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
return err
}

outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
if err != nil {
return err
}

rc, err := f.Open()
if err != nil {
outFile.Close()
return err
}

_, err = io.Copy(outFile, rc)
outFile.Close()
rc.Close()

if err != nil {
return err
}
}

return nil
}

// ExecuteConversion executes the conversion tool on a file.
func (tm *ToolManager) ExecuteConversion(tool ConversionTool, sourcePath, targetDir string, sourceFormat, targetFormat SwitchFormat) error {
if tool.ID == "custom" {
return fmt.Errorf("custom tool execution must be implemented by user")
}

if !tm.IsToolInstalled(tool) {
return fmt.Errorf("tool %s is not installed", tool.Name)
}

// Execute based on tool type
switch tool.ID {
case "nsz":
return tm.executeNsz(sourcePath, targetDir, sourceFormat, targetFormat)
case "4nxci":
return tm.execute4NXCI(sourcePath, targetDir, sourceFormat, targetFormat)
default:
return fmt.Errorf("unsupported tool: %s", tool.ID)
}
}

// executeNsz executes the nsz tool.
func (tm *ToolManager) executeNsz(sourcePath, targetDir string, sourceFormat, targetFormat SwitchFormat) error {
nszPath := tm.GetToolPath(ConversionTool{ID: "nsz"})

// nsz handles NSZ <-> NSP and XCZ <-> XCI conversions
// Check if conversion is supported
validConversions := map[string]bool{
"NSZ->NSP": true,
"NSP->NSZ": true,
"XCZ->XCI": true,
"XCI->XCZ": true,
}

conversionKey := fmt.Sprintf("%s->%s", sourceFormat, targetFormat)
if !validConversions[conversionKey] {
return fmt.Errorf("nsz tool does not support %s to %s conversion", sourceFormat, targetFormat)
}

// Build command
args := []string{nszPath}

// Compression or decompression?
if targetFormat == FormatNSZ || targetFormat == FormatXCZ {
args = append(args, "--compress")
} else {
args = append(args, "--decompress")
}

args = append(args, "-o", targetDir, sourcePath)

// Execute
cmd := exec.Command("python3", args...)
output, err := cmd.CombinedOutput()

if err != nil {
logging.Errorf("nsz execution failed: %s", string(output))
return fmt.Errorf("nsz failed: %w (output: %s)", err, string(output))
}

logging.Infof("nsz conversion successful: %s", string(output))
return nil
}

// execute4NXCI executes the 4NXCI tool.
func (tm *ToolManager) execute4NXCI(sourcePath, targetDir string, sourceFormat, targetFormat SwitchFormat) error {
toolPath := tm.GetToolPath(ConversionTool{ID: "4nxci"})

// 4NXCI converts XCI to NSP
if sourceFormat != FormatXCI || targetFormat != FormatNSP {
return fmt.Errorf("4NXCI only supports XCI to NSP conversion")
}

// Build command
args := []string{"-o", targetDir, sourcePath}

// Execute
cmd := exec.Command(toolPath, args...)
output, err := cmd.CombinedOutput()

if err != nil {
logging.Errorf("4NXCI execution failed: %s", string(output))
return fmt.Errorf("4NXCI failed: %w (output: %s)", err, string(output))
}

logging.Infof("4NXCI conversion successful: %s", string(output))
return nil
}

// ExecuteFCConversion executes the FC/NES conversion tool on a file.
func (tm *ToolManager) ExecuteFCConversion(tool ConversionTool, sourcePath, targetDir string, sourceFormat FCFormat, targetFormat FCFormat) error {
if tool.ID == "custom" {
return fmt.Errorf("custom tool execution must be implemented by user")
}

if !tm.IsToolInstalled(tool) {
return fmt.Errorf("tool %s is not installed", tool.Name)
}

// Execute based on tool type
switch tool.ID {
case "nesromtool":
return tm.executeNESROMTool(sourcePath, targetDir, sourceFormat, targetFormat)
case "fdsconv":
return tm.executeFDSConverter(sourcePath, targetDir, sourceFormat, targetFormat)
default:
return fmt.Errorf("unsupported FC/NES tool: %s", tool.ID)
}
}

// executeNESROMTool executes NES ROM Tool for conversions.
func (tm *ToolManager) executeNESROMTool(sourcePath, targetDir string, sourceFormat, targetFormat FCFormat) error {
// NES ROM Tool handles NES <-> UNF conversions
validConversions := map[string]bool{
"NES->UNF": true,
"UNF->NES": true,
}

conversionKey := fmt.Sprintf("%s->%s", sourceFormat, targetFormat)
if !validConversions[conversionKey] {
return fmt.Errorf("NES ROM Tool does not support %s to %s conversion", sourceFormat, targetFormat)
}

// For now, return a helpful error message since these tools require specific setup
return fmt.Errorf("NES ROM Tool conversion requires manual tool setup. Please use custom tool option and configure your preferred NES ROM converter")
}

// executeFDSConverter executes FDS Converter for FDS <-> NES conversions.
func (tm *ToolManager) executeFDSConverter(sourcePath, targetDir string, sourceFormat, targetFormat FCFormat) error {
// FDS Converter handles FDS <-> NES conversions
validConversions := map[string]bool{
"FDS->NES": true,
"NES->FDS": true,
}

conversionKey := fmt.Sprintf("%s->%s", sourceFormat, targetFormat)
if !validConversions[conversionKey] {
return fmt.Errorf("FDS Converter does not support %s to %s conversion", sourceFormat, targetFormat)
}

// For now, return a helpful error message since these tools require specific setup
return fmt.Errorf("FDS Converter requires manual tool setup. Please use custom tool option and configure your preferred FDS converter")
}

// downloadFCTool downloads FC/NES conversion tools.
func (tm *ToolManager) downloadFCTool(ctx context.Context, tool ConversionTool, progressCallback func(string)) error {
toolDir := filepath.Join(tm.toolsDir, tool.ID)
if err := os.MkdirAll(toolDir, 0755); err != nil {
return err
}

if progressCallback != nil {
progressCallback(fmt.Sprintf("Preparing %s tool...", tool.Name))
}

// Create a README with instructions for manual setup
readmePath := filepath.Join(toolDir, "README.txt")
readme := fmt.Sprintf(`%s Tool Setup Instructions

This tool requires manual setup as there is no single standard FC/NES conversion tool
with automatic download support.

Recommended Tools:
1. For NES ROM header editing: Use online tools or hex editors
2. For FDS conversion: Use FDS2NES converters available online
3. For UNIF conversion: Use specialized NES ROM converters

To use custom tools:
1. Select "Custom Tool" from the dropdown
2. Place your converter in this directory: %s
3. Configure and test your conversion manually

Common Tools:
- NES ROM utilities (various online tools)
- FDS conversion utilities
- Hex editors for manual header modification

For more information, visit NES ROM hacking communities and forums.
`, tool.Name, toolDir)

if err := os.WriteFile(readmePath, []byte(readme), 0644); err != nil {
logging.Errorf("Failed to create README: %v", err)
}

logging.Infof("FC/NES tool directory created with instructions at %s", toolDir)
return nil
}
