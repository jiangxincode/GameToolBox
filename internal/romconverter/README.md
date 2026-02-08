# ROM Converter Module

This module provides ROM file format conversion functionality for Nintendo Switch games.

## Supported Formats

- **XCI**: Nintendo Switch cartridge image format
- **NSP**: Nintendo Submission Package (eShop format)
- **NSZ**: Compressed NSP format
- **XCZ**: Compressed XCI format

## Supported Tools

### 1. nsz (nicoboss/nsz)
Python-based compression/decompression tool for Switch ROMs.

**Supported Conversions:**
- NSP ↔ NSZ (compress/decompress)
- XCI ↔ XCZ (compress/decompress)

**Requirements:**
- Python 3.x
- Dependencies: `pycryptodomex`, `zstandard`

**Installation:**
The tool will be automatically downloaded when first needed. After download, install dependencies:
```bash
pip install pycryptodomex zstandard
```

### 2. 4NXCI (The-4n/4NXCI)
Binary tool for converting XCI to NSP format.

**Supported Conversions:**
- XCI → NSP

**Installation:**
The tool will be automatically downloaded for your platform when first needed.

### 3. Custom Tool
Users can integrate their own conversion tools.

## Tool Management

### Automatic Download
When a conversion is initiated and the required tool is not installed:
1. The tool manager checks if the tool exists locally
2. If not found, it automatically downloads from GitHub
3. For nsz: Downloads the Python script
4. For 4NXCI: Downloads the appropriate binary for your OS

### Tool Storage
All tools are stored in: `~/.gametoolbox/tools/`

Directory structure:
```
~/.gametoolbox/tools/
├── nsz/
│   ├── nsz.py
│   └── README.txt
└── 4nxci/
    └── 4nxci (or 4nxci.exe on Windows)
```

## Usage

### From UI
1. Select "ROM文件格式转换" → "Switch" from menu
2. Choose conversion tool
3. Select source and target formats
4. Choose source ROM directory
5. Choose target save directory
6. Click "开始转换" (Start Conversion)

### Programmatic Usage

```go
import "github.com/game_tool_box/internal/romconverter"

// Create a converter
tool := romconverter.ConversionTool{
    Name: "nsz (nicoboss/nsz)",
    ID: "nsz",
    GitHubRepo: "nicoboss/nsz",
}

converter := romconverter.NewConverter(
    tool,
    "/path/to/source",
    "/path/to/target",
    romconverter.FormatNSP,
    romconverter.FormatNSZ,
    func(progress romconverter.ConversionProgress) {
        // Handle progress updates
        fmt.Printf("Converting: %s\n", progress.CurrentFile)
    },
)

// Execute conversion (tool will be downloaded if needed)
result := converter.Convert()

fmt.Printf("Success: %d, Failed: %d\n", 
    result.SuccessCount, 
    result.FailureCount)
```

## Architecture

### Key Components

1. **Converter**: Main conversion orchestrator
   - Validates inputs
   - Coordinates tool manager
   - Handles batch processing
   - Reports progress

2. **ToolManager**: Tool lifecycle management
   - Downloads tools from GitHub
   - Detects installed tools
   - Executes conversions
   - Manages tool storage

3. **ConversionTool**: Tool metadata
   - Name, ID, description
   - GitHub repository info
   - Tool-specific configuration

## Error Handling

The module provides detailed error reporting:
- Download failures
- Tool execution errors
- Invalid format combinations
- Missing dependencies

All errors are collected and reported at the end of batch conversions.

## Future Enhancements

- [ ] Support for additional game consoles (FC, etc.)
- [ ] Retry mechanism for failed conversions
- [ ] Parallel processing of multiple files
- [ ] Custom tool configuration UI
- [ ] Tool version management
- [ ] Conversion presets/profiles
