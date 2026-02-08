# Pegasus Metadata Format Support

This document describes the GameToolBox implementation of the Pegasus Frontend metadata file format.

## Overview

GameToolBox supports the Pegasus metadata file format specification, which allows defining game collections and individual game metadata in plain text files named `metadata.pegasus.txt`.

Reference: https://pegasus-frontend.org/docs/user-guide/meta-files/

## Supported Features

### Game Blocks

Game blocks define individual games with their metadata. A game block starts with `game:` and includes various fields:

#### Core Fields
- `game:` - Game title (required)
- `file:` - Path to game file (required)
- `sort-by:` - Sort order number
- `developer:` - Developer name
- `description:` - Game description

#### Extended Fields
- `publisher:` - Publisher name
- `genre:` - Game genre (single)
- `genres:` - Multiple genres (comma-separated)
- `players:` - Number of players (e.g., "1", "1-4")
- `rating:` - Rating score
- `release:` - Release year or date

#### Asset Fields
- `logo:` - Path to logo image
- `video:` - Path to video file
- `screenshot:` - Path to screenshot image
- `boxFront:` - Path to front box art
- `boxBack:` - Path to back box art
- `boxSpine:` - Path to spine box art
- `boxFull:` - Path to full box art
- `background:` - Path to background image
- `music:` - Path to music file
- `files:` - Additional file paths

### Collection Blocks

Collection blocks define game collections and how they are launched. A collection block starts with `collection:`:

- `collection:` - Collection name (required)
- `shortname:` - Short identifier
- `sort-by:` - Sort order number
- `launch:` - Launch command template (e.g., `retroarch -L "core.so" {file.path}`)
- `extension:` - Comma-separated file extensions
- `files:` - Path patterns for game files

### Example

```
# Super Nintendo Collection
collection: Super Nintendo
shortname: snes
extension: smc, sfc, snes
launch: snes9x "{file.path}"

game: Super Mario World
file: Super Mario World (USA).smc
developer: Nintendo
publisher: Nintendo
genre: Platformer
players: 2
rating: 94
release: 1990
description: Classic 2D platformer
logo: media/Super Mario World/logo.png
boxFront: media/Super Mario World/boxFront.png

game: The Legend of Zelda: A Link to the Past
file: Zelda - A Link to the Past (USA).smc
developer: Nintendo
publisher: Nintendo
genre: Action-Adventure
players: 1
rating: 95
release: 1991
description: Epic adventure in Hyrule
logo: media/Zelda ALTTP/logo.png
screenshot: media/Zelda ALTTP/screenshot.png
```

## Implementation Details

### Parsing

The parser (`metadata.Parse()`) handles:
- Case-insensitive field names
- Optional spaces around colons (both `game:Name` and `game: Name` work)
- Comments (lines starting with `#` or `;`)
- Blank lines for separation
- Preservation of unknown fields and comments

### Generation

When generating metadata files (`SetGames()`, `AppendGames()`):
- Uses format without space after colon for backward compatibility: `game:Name`
- Automatically adds `sort-by` fields with zero-padding
- Defaults developer/description to game name if not provided
- Only includes optional fields when they have values

### Round-Trip Preservation

The implementation preserves:
- Collection blocks and their fields
- Comments and blank lines (with normalization)
- Unknown/custom fields
- Original file structure as much as possible

## API Usage

### Reading Metadata

```go
// Load from standard location
doc, err := metadata.LoadFromRootDir("/path/to/rom/dir")

// Read specific file
doc, err := metadata.ReadFile("/path/to/metadata.pegasus.txt")

// Parse from string
doc := metadata.Parse(content)

// Get games and collections
games := doc.Games()
collections := doc.Collections()
```

### Creating/Modifying Metadata

```go
// Create new document
doc := metadata.New()

// Set all games (replaces content)
doc.SetGames([]metadata.Game{
    {
        GameName:  "Test Game",
        FileName:  "test.rom",
        Developer: "Developer",
        Genre:     "Action",
        Players:   "1-2",
    },
})

// Append games to existing document
doc.AppendGames(newGames)

// Upsert (update if exists, insert if not)
changed, found, err := doc.UpsertGameByFile(game, metadata.UpsertOptions{})

// Write to file atomically
err = doc.WriteFileAtomic("/path/to/metadata.pegasus.txt")
```

### Removing Games

```go
// Remove by game names
removed := doc.RemoveByGameNames(map[string]struct{}{
    "Game to Remove": {},
})
```

## Limitations and Future Work

### Current Limitations

1. **Multiline Values**: Description and other fields don't yet support multiline values with indentation
2. **List Syntax**: The `files[]` array syntax is not yet supported (stored as comma-separated strings)
3. **Collection Writing**: Only collection parsing/preservation is supported; no API for creating collections yet

### Backward Compatibility

- Format uses no space after colon (`game:Name`) to match existing files
- Parser accepts both formats (with/without space)
- All existing tests and functionality preserved

## Testing

Comprehensive test coverage includes:
- Extended field parsing and generation
- Collection block parsing and preservation
- Round-trip preservation
- Multiple games and collections
- Upsert operations with extended fields
- Edge cases and error handling

Run tests:
```bash
go test ./internal/pegasus/metadata/... -v
```
