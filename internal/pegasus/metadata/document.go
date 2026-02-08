package metadata

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// Document is a lightly-normalizing, round-trippable representation of
// Pegasus metadata files (typically `metadata.pegasus.txt`).
//
// This package intentionally focuses on the subset of the Pegasus meta-files
// spec that GameToolBox needs, while keeping the original file as intact as
// possible (comments, unknown keys, collections, and formatting).
//
// Goals:
//   - Preserve non-game content (comments, blank lines, collection blocks, unknown keys)
//   - Provide helpers to list/remove game entries (by name or by `file:`)
//   - Write back using an atomic replace + light whitespace normalization
//
// Parsing behavior (aligned with https://pegasus-frontend.org/docs/user-guide/meta-files/):
//   - A new game entry starts at a line whose trimmed key is `game` (case-insensitive).
//     Example accepted forms: `game: Title`, `game : Title`, `GAME: Title`.
//   - Inside a game entry, we only *extract* a small subset of keys used by this app:
//       `file`, `sort-by`, `developer`, `description` (case-insensitive).
//     Any additional lines are preserved as raw lines within the entry.
//   - Collection blocks (e.g. `collection: ...`) are treated as *non-game* sections.
//     When such a key is encountered while parsing a game entry, the game entry is
//     closed and the line becomes part of the raw section.
//
// Known limitations / intentional trade-offs:
//   - We don't fully model every Pegasus key; unknown keys are preserved but not
//     exposed via the structured Game fields.
//   - Duplicate keys inside a game entry are stored as-is in lines; the structured
//     fields keep the last parsed value.
//   - Render() collapses multiple consecutive blank lines and always ensures the
//     file ends with a single newline.

type Document struct {
	items   []Item
	newline string // "\n" or "\r\n"; used when rendering
}

type ItemKind int

const (
	ItemRaw ItemKind = iota
	ItemGame
	ItemCollection
)

type Item interface {
	Kind() ItemKind
	Lines() []string
}

type RawItem struct {
	lines []string
}

func (r RawItem) Kind() ItemKind  { return ItemRaw }
func (r RawItem) Lines() []string { return r.lines }

// Collection represents a Pegasus collection block with its metadata.
// Collections group games together and define how they should be launched.
type Collection struct {
	Name       string // collection: name
	ShortName  string // shortname: name
	SortBy     string // sort-by: number
	Launch     string // launch: command template
	Extensions string // extension: comma-separated list
	Files      string // files: path patterns
}

type CollectionItem struct {
	lines      []string
	collection Collection
}

func (c CollectionItem) Kind() ItemKind      { return ItemCollection }
func (c CollectionItem) Lines() []string     { return c.lines }
func (c CollectionItem) Collection() Collection { return c.collection }

type Game struct {
	// Core fields
	GameName    string
	FileName    string
	SortBy      string
	Developer   string
	Description string

	// Additional Pegasus standard fields
	Publisher string
	Genre     string
	Players   string
	Rating    string
	Release   string

	// Asset fields
	Logo       string
	Video      string
	Screenshot string
	BoxFront   string
	BoxBack    string
	BoxSpine   string
	BoxFull    string
	Background string
	Music      string

	// List fields (stored as comma-separated strings for now)
	Files  string // for games with multiple files
	Genres string // for multiple genres
}

type GameItem struct {
	lines []string
	game  Game
}

func (g GameItem) Kind() ItemKind  { return ItemGame }
func (g GameItem) Lines() []string { return g.lines }
func (g GameItem) Game() Game      { return g.game }

// NewGameItem constructs a GameItem from raw lines and parsed game fields.
// This is primarily used by internal tooling that needs to rewrite a game block
// while preserving all other document items.
func NewGameItem(lines []string, game Game) GameItem {
	return GameItem{lines: append([]string(nil), lines...), game: game}
}

// LoadFromRootDir reads <rootDir>/metadata.pegasus.txt and parses it into a Document.
func LoadFromRootDir(rootDir string) (*Document, error) {
	path := filepath.Join(rootDir, "metadata.pegasus.txt")
	return ReadFile(path)
}

// ReadFile reads and parses a metadata.pegasus.txt file.
func ReadFile(path string) (*Document, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(string(b)), nil
}

// Parse parses the text content and returns a Document.
func Parse(text string) *Document {
	newline := "\n"
	if strings.Contains(text, "\r\n") {
		newline = "\r\n"
	}

	scanner := bufio.NewScanner(strings.NewReader(text))
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 2*1024*1024)

	d := &Document{newline: newline}

	var rawBuf []string
	var gameLines []string
	var collectionLines []string
	var currentGame *Game
	var currentCollection *Collection
	gameID := 1
	_ = gameID // kept for future stable IDs; currently unused

	flushRaw := func() {
		if len(rawBuf) == 0 {
			return
		}
		d.items = append(d.items, RawItem{lines: append([]string(nil), rawBuf...)})
		rawBuf = rawBuf[:0]
	}
	flushGame := func() {
		if currentGame == nil {
			return
		}
		d.items = append(d.items, GameItem{lines: append([]string(nil), gameLines...), game: *currentGame})
		currentGame = nil
		gameLines = gameLines[:0]
	}
	flushCollection := func() {
		if currentCollection == nil {
			return
		}
		d.items = append(d.items, CollectionItem{lines: append([]string(nil), collectionLines...), collection: *currentCollection})
		currentCollection = nil
		collectionLines = collectionLines[:0]
	}

	parseKV := func(line string) (key, val string, ok bool) {
		trim := strings.TrimSpace(line)
		if trim == "" {
			return "", "", false
		}
		idx := strings.IndexByte(trim, ':')
		if idx < 0 {
			return "", "", false
		}
		key = strings.ToLower(strings.TrimSpace(trim[:idx]))
		val = strings.TrimSpace(trim[idx+1:])
		if key == "" {
			return "", "", false
		}
		return key, val, true
	}

	isGameField := func(key string) bool {
		switch key {
		case "file", "sort-by", "developer", "description",
			"publisher", "genre", "genres", "players", "rating", "release",
			"logo", "video", "screenshot", "boxfront", "boxback", "boxspine", "boxfull",
			"background", "music", "files":
			return true
		default:
			return false
		}
	}

	isCollectionField := func(key string) bool {
		switch key {
		case "shortname", "sort-by", "launch", "extension", "files":
			return true
		default:
			return false
		}
	}

	for scanner.Scan() {
		line := scanner.Text()
		key, val, ok := parseKV(line)

		// New collection block
		if ok && key == "collection" {
			flushRaw()
			flushGame()
			flushCollection()

			c := Collection{Name: val}
			currentCollection = &c
			collectionLines = append(collectionLines, line)
			continue
		}

		// New game block
		if ok && key == "game" {
			flushRaw()
			flushGame()
			flushCollection()

			g := Game{GameName: val}
			currentGame = &g
			gameLines = append(gameLines, line)
			continue
		}

		// Inside a collection block
		if currentCollection != nil {
			if ok && key != "" && !isCollectionField(key) {
				// Unknown field that's not a collection field - ends collection block
				flushCollection()
				rawBuf = append(rawBuf, line)
				continue
			}

			// Parse collection fields
			if ok {
				switch key {
				case "shortname":
					currentCollection.ShortName = val
				case "sort-by":
					currentCollection.SortBy = val
				case "launch":
					currentCollection.Launch = val
				case "extension":
					currentCollection.Extensions = val
				case "files":
					currentCollection.Files = val
				}
			}
			collectionLines = append(collectionLines, line)
			continue
		}

		// Inside a game block
		if currentGame != nil {
			// If we encounter a non-empty key/value that isn't a recognized game field,
			// it likely starts another block type (e.g. collection), so we close the
			// game block and treat this line as raw.
			if ok && key != "" && !isGameField(key) {
				flushGame()
				rawBuf = append(rawBuf, line)
				continue
			}

			// Parse known game fields
			if ok {
				switch key {
				case "file":
					currentGame.FileName = val
				case "sort-by":
					currentGame.SortBy = val
				case "developer":
					currentGame.Developer = val
				case "description":
					currentGame.Description = val
				case "publisher":
					currentGame.Publisher = val
				case "genre":
					currentGame.Genre = val
				case "genres":
					currentGame.Genres = val
				case "players":
					currentGame.Players = val
				case "rating":
					currentGame.Rating = val
				case "release":
					currentGame.Release = val
				case "logo":
					currentGame.Logo = val
				case "video":
					currentGame.Video = val
				case "screenshot":
					currentGame.Screenshot = val
				case "boxfront":
					currentGame.BoxFront = val
				case "boxback":
					currentGame.BoxBack = val
				case "boxspine":
					currentGame.BoxSpine = val
				case "boxfull":
					currentGame.BoxFull = val
				case "background":
					currentGame.Background = val
				case "music":
					currentGame.Music = val
				case "files":
					currentGame.Files = val
				}
			}
			gameLines = append(gameLines, line)
			continue
		}

		// Not in any block - treat as raw
		rawBuf = append(rawBuf, line)
	}

	flushRaw()
	flushGame()
	flushCollection()
	return d
}

// Items returns a shallow copy of the document items in file order.
// This is intended for internal tooling that needs to reorder game blocks while
// preserving non-game content (collections/comments/raw formatting).
func (d *Document) Items() []Item {
	return append([]Item(nil), d.items...)
}

// SetItems replaces the document items.
// Callers must ensure items are well-formed.
func (d *Document) SetItems(items []Item) {
	d.items = append([]Item(nil), items...)
}

// Games returns all parsed games (in file order).
func (d *Document) Games() []Game {
	out := make([]Game, 0, len(d.items))
	for _, it := range d.items {
		gi, ok := it.(GameItem)
		if !ok {
			continue
		}
		out = append(out, gi.game)
	}
	return out
}

// Collections returns all parsed collections (in file order).
func (d *Document) Collections() []Collection {
	out := make([]Collection, 0, len(d.items))
	for _, it := range d.items {
		ci, ok := it.(CollectionItem)
		if !ok {
			continue
		}
		out = append(out, ci.collection)
	}
	return out
}

// RemoveByGameNames removes any game block whose GameName matches (after TrimSpace) one of names.
// Returns number of removed blocks.
func (d *Document) RemoveByGameNames(names map[string]struct{}) int {
	if len(names) == 0 {
		return 0
	}
	removed := 0
	filtered := d.items[:0]
	for _, it := range d.items {
		gi, ok := it.(GameItem)
		if ok {
			k := strings.TrimSpace(gi.game.GameName)
			if _, hit := names[k]; hit {
				removed++
				continue
			}
		}
		filtered = append(filtered, it)
	}
	d.items = filtered
	return removed
}

// Render renders the document with light normalization:
//   - unifies newlines to d.newline
//   - collapses runs of blank lines to at most one blank line
//   - ensures file ends with exactly one newline
func (d *Document) Render() string {
	var lines []string
	for _, it := range d.items {
		lines = append(lines, it.Lines()...)
	}

	out := make([]string, 0, len(lines))
	blankRun := 0
	for _, ln := range lines {
		if strings.TrimSpace(ln) == "" {
			blankRun++
			if blankRun > 1 {
				continue
			}
			out = append(out, "")
			continue
		}
		blankRun = 0
		out = append(out, ln)
	}

	return strings.Join(out, d.newline) + d.newline
}

// WriteFileAtomic renders the document to a temp file and atomically replaces dstPath.
func (d *Document) WriteFileAtomic(dstPath string) (err error) {
	if strings.TrimSpace(dstPath) == "" {
		return fmt.Errorf("dst path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return err
	}

	tmp := dstPath + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = f.Close()
			_ = os.Remove(tmp)
		}
	}()

	w := bufio.NewWriter(f)
	if _, werr := w.WriteString(d.Render()); werr != nil {
		err = werr
		return err
	}
	if werr := w.Flush(); werr != nil {
		err = werr
		return err
	}
	if cerr := f.Close(); cerr != nil {
		err = cerr
		return err
	}

	if repErr := ReplaceFileAtomic(dstPath, tmp); repErr != nil {
		err = repErr
		return err
	}
	return nil
}

// NormalizeFileKey normalizes a metadata `file:` value for stable matching.
// It mirrors legacy behavior used across the project:
//   - trims spaces
//   - converts to forward slashes
//   - removes a leading "./"
//   - lowercases
func NormalizeFileKey(fileName string) string {
	f := strings.TrimSpace(fileName)
	if f == "" {
		return ""
	}
	f = filepath.ToSlash(f)
	f = strings.TrimPrefix(f, "./")
	return strings.ToLower(f)
}

// Backward-compatible alias for internal callers.
func normalizeFileKey(fileName string) string { return NormalizeFileKey(fileName) }

// NormalizeGameNameKey normalizes a metadata `game:` value (title) for stable matching.
//
// Rules:
//   - trims leading/trailing spaces
//   - collapses any whitespace runs (spaces/tabs/newlines) into a single space
//   - lowercases
func NormalizeGameNameKey(name string) string {
	s := strings.TrimSpace(name)
	if s == "" {
		return ""
	}

	// Collapse all unicode whitespace to single ASCII spaces.
	b := strings.Builder{}
	b.Grow(len(s))
	prevSpace := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			if prevSpace {
				continue
			}
			b.WriteByte(' ')
			prevSpace = true
			continue
		}
		b.WriteRune(r)
		prevSpace = false
	}

	return strings.ToLower(b.String())
}
