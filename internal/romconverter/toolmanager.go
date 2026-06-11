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

	// For TNES2INES tool (Python-based)
	if tool.ID == "tnes2ines" {
		return filepath.Join(toolDir, "tnes2ines.py")
	}

	// For NES ROM Tool (inestool - Python-based)
	if tool.ID == "nesromtool" {
		return filepath.Join(toolDir, "inestool.py")
	}

	// For FDS Converter (fds-header-remover - Python-based)
	if tool.ID == "fdsconv" {
		return filepath.Join(toolDir, "fds_header_cleaner.py")
	}

	return ""
}

// IsToolInstalled checks if a tool is installed.
func (tm *ToolManager) IsToolInstalled(tool ConversionTool) bool {
	// For Python-based tools, check if the script file exists
	if tool.ID == "tnes2ines" || tool.ID == "nesromtool" || tool.ID == "fdsconv" {
		toolPath := tm.GetToolPath(tool)
		_, err := os.Stat(toolPath)
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
		// For Python-based tools, try direct script download when no release is found
		if tool.ID == "nsz" {
			return tm.downloadNszTool(ctx, tool, progressCallback)
		}
		if tool.ID == "tnes2ines" {
			return tm.downloadTNES2INESTool(ctx, tool, progressCallback)
		}
		if tool.ID == "nesromtool" {
			return tm.downloadInestoolTool(ctx, tool, progressCallback)
		}
		if tool.ID == "fdsconv" {
			return tm.downloadFDSHeaderCleanerTool(ctx, tool, progressCallback)
		}
		return fmt.Errorf("failed to get latest release: %w", err)
	}

	// For nsz, download Python script
	if tool.ID == "nsz" {
		return tm.downloadNszTool(ctx, tool, progressCallback)
	}

	// For 4NXCI, download the appropriate release asset
	if tool.ID == "4nxci" {
		return tm.download4NXCITool(ctx, tool, release, progressCallback)
	}

	// For TNES2INES, download the Python script
	if tool.ID == "tnes2ines" {
		return tm.downloadTNES2INESTool(ctx, tool, progressCallback)
	}

	// For NES ROM Tool (inestool), download the Python script
	if tool.ID == "nesromtool" {
		return tm.downloadInestoolTool(ctx, tool, progressCallback)
	}

	// For FDS Converter (fds-header-remover), download the Python script
	if tool.ID == "fdsconv" {
		return tm.downloadFDSHeaderCleanerTool(ctx, tool, progressCallback)
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

// downloadTNES2INESTool downloads the TNES2INES Python tool.
func (tm *ToolManager) downloadTNES2INESTool(ctx context.Context, tool ConversionTool, progressCallback func(string)) error {
	toolDir := filepath.Join(tm.toolsDir, tool.ID)
	if err := os.MkdirAll(toolDir, 0755); err != nil {
		return err
	}

	if progressCallback != nil {
		progressCallback("Downloading TNES2INES tool (Python script)...")
	}

	// Download the main tnes2ines.py script
	tnes2inesURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/main/tnes2ines.py", tool.GitHubRepo)
	tnes2inesPath := filepath.Join(toolDir, "tnes2ines.py")

	if err := tm.downloadFile(ctx, tnes2inesURL, tnes2inesPath, progressCallback); err != nil {
		return fmt.Errorf("failed to download tnes2ines.py: %w", err)
	}

	// Make executable on Unix
	if runtime.GOOS != "windows" {
		if err := os.Chmod(tnes2inesPath, 0755); err != nil {
			return err
		}
	}

	// Create a simple README
	readmePath := filepath.Join(toolDir, "README.txt")
	readme := `TNES2INES Tool Downloaded Successfully

This is a Python-based tool for converting TNES ROMs to iNES format.

Supported conversions:
- TNES -> iNES 1.0
- Extract PRG/CHR ROMs
- Extract FDS BIOS and .qd files

To use it:
1. Ensure Python 3.x is installed
2. Run: python tnes2ines.py -c <input_file>

The tool will be executed through Python when you run conversions.
`
	if err := os.WriteFile(readmePath, []byte(readme), 0644); err != nil {
		logging.Infof("Failed to create README: %v", err)
	}

	logging.Infof("TNES2INES tool installed successfully at %s", toolDir)
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
	if !tm.IsToolInstalled(tool) {
		return fmt.Errorf("tool %s is not installed", tool.Name)
	}

	// Execute based on tool type
	switch tool.ID {
	case "tnes2ines":
		return tm.executeTNES2INES(sourcePath, targetDir, sourceFormat, targetFormat)
	case "nesromtool":
		return tm.executeNESROMTool(sourcePath, targetDir, sourceFormat, targetFormat)
	case "fdsconv":
		return tm.executeFDSConverter(sourcePath, targetDir, sourceFormat, targetFormat)
	default:
		return fmt.Errorf("unsupported FC/NES tool: %s", tool.ID)
	}
}

// executeTNES2INES executes the TNES2INES tool.
func (tm *ToolManager) executeTNES2INES(sourcePath, targetDir string, sourceFormat, targetFormat FCFormat) error {
	tnes2inesPath := tm.GetToolPath(ConversionTool{ID: "tnes2ines"})

	// TNES2INES converts TNES to iNES format
	// Check if conversion is supported (TNES -> NES)
	validConversions := map[string]bool{
		"TNES->NES": true,
	}

	conversionKey := fmt.Sprintf("%s->%s", sourceFormat, targetFormat)
	if !validConversions[conversionKey] {
		return fmt.Errorf("TNES2INES does not support %s to %s conversion", sourceFormat, targetFormat)
	}

	// Build output file path
	baseName := filepath.Base(sourcePath)
	ext := filepath.Ext(baseName)
	outputName := strings.TrimSuffix(baseName, ext) + ".nes"
	outputPath := filepath.Join(targetDir, outputName)

	// Build command: python tnes2ines.py -c <input> > <output>
	// TNES2INES outputs to stdout, so we redirect to file
	args := []string{tnes2inesPath, "-c", sourcePath}

	// Execute
	cmd := exec.Command("python3", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		logging.Errorf("TNES2INES execution failed: %s", string(output))
		return fmt.Errorf("TNES2INES failed: %w (output: %s)", err, string(output))
	}

	// Write output to file
	if err := os.WriteFile(outputPath, output, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	logging.Infof("TNES2INES conversion successful: %s -> %s", sourcePath, outputPath)
	return nil
}

// executeNESROMTool executes inestool for NES ROM header operations.
func (tm *ToolManager) executeNESROMTool(sourcePath, targetDir string, sourceFormat, targetFormat FCFormat) error {
	inestoolPath := tm.GetToolPath(ConversionTool{ID: "nesromtool"})

	// inestool handles iNES header read/write operations
	validConversions := map[string]bool{
		"NES->NES": true, // Fix/write iNES header
	}

	conversionKey := fmt.Sprintf("%s->%s", sourceFormat, targetFormat)
	if !validConversions[conversionKey] {
		return fmt.Errorf("inestool does not support %s to %s conversion, it handles iNES header read/write/fix for .nes files", sourceFormat, targetFormat)
	}

	// Build output file path
	baseName := filepath.Base(sourcePath)
	outputPath := filepath.Join(targetDir, baseName)

	// Copy source file to target directory first
	srcData, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}
	if err := os.WriteFile(outputPath, srcData, 0644); err != nil {
		return fmt.Errorf("failed to copy file to target: %w", err)
	}

	// Execute inestool write to fix/add iNES header
	args := []string{inestoolPath, "write", outputPath}

	cmd := exec.Command("python3", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		logging.Errorf("inestool execution failed: %s", string(output))
		return fmt.Errorf("inestool failed: %w (output: %s)", err, string(output))
	}

	logging.Infof("inestool header fix successful: %s -> %s", sourcePath, outputPath)
	return nil
}

// executeFDSConverter executes fds_header_cleaner for FDS header operations.
func (tm *ToolManager) executeFDSConverter(sourcePath, targetDir string, sourceFormat, targetFormat FCFormat) error {
	// fds_header_cleaner handles FDS header removal
	validConversions := map[string]bool{
		"FDS->FDS": true, // Remove FDS header (output is headerless FDS)
	}

	conversionKey := fmt.Sprintf("%s->%s", sourceFormat, targetFormat)
	if !validConversions[conversionKey] {
		return fmt.Errorf("FDS Header Cleaner does not support %s to %s conversion, it removes FDS headers from .fds files", sourceFormat, targetFormat)
	}

	// Build output file path
	baseName := filepath.Base(sourcePath)
	outputPath := filepath.Join(targetDir, baseName)

	// Read source file
	srcData, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Check if file has FDS header (starts with "FDS")
	hasHeader := len(srcData) >= 16 && string(srcData[:3]) == "FDS"

	if hasHeader {
		// Remove the 16-byte header
		if err := os.WriteFile(outputPath, srcData[16:], 0644); err != nil {
			return fmt.Errorf("failed to write headerless file: %w", err)
		}
		logging.Infof("FDS header removed: %s -> %s", sourcePath, outputPath)
	} else {
		// Just copy the file as-is
		if err := os.WriteFile(outputPath, srcData, 0644); err != nil {
			return fmt.Errorf("failed to copy file: %w", err)
		}
		logging.Infof("FDS file copied (no header found): %s -> %s", sourcePath, outputPath)
	}

	return nil
}

// downloadInestoolTool downloads the inestool Python tool from GitHub.
func (tm *ToolManager) downloadInestoolTool(ctx context.Context, tool ConversionTool, progressCallback func(string)) error {
	toolDir := filepath.Join(tm.toolsDir, tool.ID)
	if err := os.MkdirAll(toolDir, 0755); err != nil {
		return err
	}

	if progressCallback != nil {
		progressCallback("Downloading inestool (iNES header reader/writer)...")
	}

	// Download inestool.py from dsedivec/inestool
	inestoolURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/master/inestool.py", tool.GitHubRepo)
	inestoolPath := filepath.Join(toolDir, "inestool.py")

	if err := tm.downloadFile(ctx, inestoolURL, inestoolPath, progressCallback); err != nil {
		return fmt.Errorf("failed to download inestool.py: %w", err)
	}

	// Make executable on Unix
	if runtime.GOOS != "windows" {
		if err := os.Chmod(inestoolPath, 0755); err != nil {
			return err
		}
	}

	// Create a simple README
	readmePath := filepath.Join(toolDir, "README.txt")
	readme := `inestool - iNES Header Reader/Writer

Downloaded from: https://github.com/dsedivec/inestool

Features:
- Read iNES headers from NES ROMs
- Write/fix iNES headers using a database (NstDatabase.xml)
- Support for headerless ROMs

Requirements:
- Python 3.x
- Optional: pylzma for 7-Zip archive support (pip install pylzma)

Usage:
  python inestool.py read <rom_file>    - Read and display iNES header
  python inestool.py write <rom_file>   - Fix/write iNES header from database

For the write command, you need a NES database file (NstDatabase.xml) in the current directory.
Download it from: https://github.com/rdanbrook/nestopia
`
	if err := os.WriteFile(readmePath, []byte(readme), 0644); err != nil {
		logging.Infof("Failed to create README: %v", err)
	}

	logging.Infof("inestool installed successfully at %s", toolDir)
	return nil
}

// downloadFDSHeaderCleanerTool downloads the FDS Header Cleaner Python tool from GitHub.
func (tm *ToolManager) downloadFDSHeaderCleanerTool(ctx context.Context, tool ConversionTool, progressCallback func(string)) error {
	toolDir := filepath.Join(tm.toolsDir, tool.ID)
	if err := os.MkdirAll(toolDir, 0755); err != nil {
		return err
	}

	if progressCallback != nil {
		progressCallback("Downloading FDS Header Cleaner...")
	}

	// Download fds_header_cleaner.py from cturczynskyj/fds-header-remover
	fdsCleanerURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/master/fds_header_cleaner.py", tool.GitHubRepo)
	fdsCleanerPath := filepath.Join(toolDir, "fds_header_cleaner.py")

	if err := tm.downloadFile(ctx, fdsCleanerURL, fdsCleanerPath, progressCallback); err != nil {
		return fmt.Errorf("failed to download fds_header_cleaner.py: %w", err)
	}

	// Make executable on Unix
	if runtime.GOOS != "windows" {
		if err := os.Chmod(fdsCleanerPath, 0755); err != nil {
			return err
		}
	}

	// Create a simple README
	readmePath := filepath.Join(toolDir, "README.txt")
	readme := `FDS Header Cleaner

Downloaded from: https://github.com/cturczynskyj/fds-header-remover

Features:
- Remove Famicom Disk System (FDS) headers from ROM files
- Process single files or directories recursively
- Supports nested folder structures

Requirements:
- Python 3.x

Usage:
  python fds_header_cleaner.py

Place FDS ROM files in the fds_roms folder and run the script.
Headerless ROMs will be saved to the headerless_roms folder.
`
	if err := os.WriteFile(readmePath, []byte(readme), 0644); err != nil {
		logging.Infof("Failed to create README: %v", err)
	}

	logging.Infof("FDS Header Cleaner installed successfully at %s", toolDir)
	return nil
}
