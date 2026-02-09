# FC/NES ROM Converter Support

This document describes the FC (Famicom) / NES (Nintendo Entertainment System) ROM format conversion support.

## Supported Formats

- **NES**: iNES format (most common, .nes extension)
- **FDS**: Famicom Disk System format (.fds extension)
- **UNF**: UNIF format (.unf or .unif extension)
- **NSF**: NES Sound Format (.nsf extension, music/sound files)

## Format Overview

### NES (iNES)
The most common NES ROM format, using the iNES header structure.
- **Extension**: .nes
- **Header**: 16-byte iNES header
- **Usage**: Widely supported by emulators

### FDS (Famicom Disk System)
Format for Famicom Disk System games.
- **Extension**: .fds
- **Type**: Disk image format
- **Usage**: Requires FDS BIOS in emulators

### UNF (UNIF)
Universal NES Image Format, an alternative to iNES.
- **Extension**: .unf or .unif
- **Features**: More flexible header structure
- **Usage**: Less common, but supported by many modern emulators

### NSF (NES Sound Format)
Music/sound file format extracted from NES games.
- **Extension**: .nsf
- **Type**: Audio-only format
- **Usage**: Music players and emulators

## Available Tools

### 1. NES ROM Tool
General-purpose NES ROM header editor and converter.

**Features:**
- Header editing
- Format validation
- Basic conversions

**Status:** To be integrated

### 2. FDS Converter
Specialized tool for FDS format conversions.

**Features:**
- FDS to NES conversion
- Disk image manipulation

**Status:** To be integrated

### 3. Custom Tool
Users can configure their own conversion tools.

**Usage:**
- Select "Custom Tool" from the tool dropdown
- Configure the tool path and parameters
- Execute custom conversions

## Conversion Matrix

| Source Format | Target Format | Tool | Status |
|--------------|---------------|------|--------|
| NES | FDS | FDS Converter | 🔄 Planned |
| FDS | NES | FDS Converter | 🔄 Planned |
| NES | UNF | NES ROM Tool | 🔄 Planned |
| UNF | NES | NES ROM Tool | 🔄 Planned |
| Custom | Custom | Custom Tool | ✅ Available |

## Usage

### From UI
1. Select "ROM文件格式转换" (ROM Format Converter) → "FC (Famicom/NES)" from menu
2. Choose conversion tool
3. Select source format (NES, FDS, UNF, NSF)
4. Select target format
5. Choose source ROM directory
6. Choose target save directory
7. Click "开始转换" (Start Conversion)

### Current Status

**Framework Complete:**
- ✅ UI interface for FC conversion
- ✅ Format selection (NES, FDS, UNF, NSF)
- ✅ Tool selection dropdown
- ✅ Directory pickers with memory
- ✅ Progress tracking
- ✅ Error handling

**To Be Implemented:**
- ⏳ Actual tool integration (download & execution)
- ⏳ Format-specific conversion logic
- ⏳ Tool configuration UI

## Notes

- **Custom Tool Recommended**: For immediate use, configure your own conversion tools
- **Tool Integration**: Automatic tool download and execution will be added in future updates
- **Manual Conversion**: Users can currently use external tools and this UI for file management
- **Format Support**: All major FC/NES formats are supported in the UI framework

## Future Enhancements

- [ ] Automatic tool download from GitHub
- [ ] NES ROM header editor integration
- [ ] FDS disk manipulation tools
- [ ] Batch conversion optimization
- [ ] Format validation and repair
- [ ] Metadata preservation during conversion

## Technical Details

### File Extensions
The converter automatically detects files based on extensions:
- `.nes` → NES format
- `.fds` → FDS format
- `.unf`, `.unif` → UNF format
- `.nsf` → NSF format

### Tool Manager Integration
FC conversion uses the same `ToolManager` infrastructure as Switch conversion, enabling:
- Consistent tool management
- Automatic downloads (when implemented)
- Progress tracking
- Error reporting

### Directory Memory
Source and target directories are automatically remembered and restored on next use, shared with Switch converter for convenience.
