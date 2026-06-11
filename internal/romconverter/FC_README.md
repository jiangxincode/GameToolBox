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

### 1. TNES2INES (Longestboi/TNES2INES)
TNES to iNES converter (Python-based).

**GitHub:** https://github.com/Longestboi/TNES2INES

**Features:**
- TNES to iNES 1.0 conversion
- Extract PRG/CHR ROMs
- Extract FDS BIOS and .qd files

**Requirements:**
- Python 3.x

**Status:** ✅ Auto-download + Execute

**Setup:**
The tool will be automatically downloaded from GitHub when first used. No manual setup required.

### 2. inestool (dsedivec/inestool)
iNES header reader/writer for NES ROMs (Python-based).

**GitHub:** https://github.com/dsedivec/inestool

**Features:**
- Read iNES headers from NES ROMs
- Write/fix iNES headers using a database (NstDatabase.xml)
- Support for headerless ROMs

**Requirements:**
- Python 3.x
- Optional: pylzma for 7-Zip archive support

**Status:** ✅ Auto-download + Execute

**Setup:**
The tool will be automatically downloaded from GitHub when first used. No manual setup required.

### 3. FDS Header Cleaner (cturczynskyj/fds-header-remover)
FDS header remover for Famicom Disk System ROMs (Python-based).

**GitHub:** https://github.com/cturczynskyj/fds-header-remover

**Features:**
- Remove Famicom Disk System (FDS) headers from ROM files
- Process single files or directories recursively
- Supports nested folder structures

**Requirements:**
- Python 3.x

**Status:** ✅ Auto-download + Execute

**Setup:**
The tool will be automatically downloaded from GitHub when first used. No manual setup required.

## Conversion Matrix

| Source Format | Target Format | Tool | Status |
|--------------|---------------|------|--------|
| TNES | NES | TNES2INES | ✅ Auto-download + Execute |
| NES | NES | inestool | ✅ Auto-download + Execute (iNES header fix/write) |
| FDS | FDS | FDS Header Cleaner | ✅ Auto-download + Execute (header removal) |

**Note:** All tools are automatically downloaded from GitHub when first used. No manual setup required.

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
- ✅ **All tools auto-download and execute**
- ✅ **TNES2INES auto-download and execution**
- ✅ **inestool auto-download and execution**
- ✅ **FDS Header Cleaner auto-download and execution**

## Notes

- **All tools auto-download**: Every conversion tool is automatically downloaded from GitHub when first used
- **No manual setup required**: Tools are downloaded and configured automatically
- **Python required**: All FC/NES conversion tools require Python 3.x to be installed
- **Format Support**: All major FC/NES formats are supported in the UI framework

## Future Enhancements

- [x] Automatic tool download from GitHub (TNES2INES)
- [x] Automatic tool download from GitHub (inestool)
- [x] Automatic tool download from GitHub (FDS Header Cleaner)
- [ ] NES ROM header editor integration
- [ ] FDS disk manipulation tools
- [ ] Batch conversion optimization
- [ ] Format validation and repair
- [ ] Metadata preservation during conversion
- [ ] Support for more TNES/FC conversion tools

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
