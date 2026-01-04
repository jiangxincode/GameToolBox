package metadata

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Document is a lightly-normalizing, round-trippable representation of
// metadata.pegasus.txt.
//
// Goals:
//   - Preserve unknown lines/blocks (comments/collection/unknown keys)
//   - Light normalization on write (newline normalization + collapse excessive blank lines)
//   - Provide helpers to list/remove game blocks
//
// Parsing contract (compatible with existing code):
//   - A game block starts at a line whose TrimSpace has prefix "game:".
//   - The game name is the trimmed remainder after "game:".
//   - file/sort-by/developer/description are recognized within a game block.
//   - Everything else is preserved as raw lines.

type Document struct {
	items   []Item
	newline string // "\n" or "\r\n"; used when rendering
}

type ItemKind int

const (
	ItemRaw ItemKind = iota
	ItemGame
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

type Game struct {
	GameName    string
	FileName    string
	SortBy      string
	Developer   string
	Description string
}

type GameItem struct {
	lines []string
	game  Game
}

func (g GameItem) Kind() ItemKind  { return ItemGame }
func (g GameItem) Lines() []string { return g.lines }
func (g GameItem) Game() Game      { return g.game }

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
	var current *Game
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
		if current == nil {
			return
		}
		d.items = append(d.items, GameItem{lines: append([]string(nil), gameLines...), game: *current})
		current = nil
		gameLines = gameLines[:0]
	}

	for scanner.Scan() {
		line := scanner.Text()
		trim := strings.TrimSpace(line)

		if strings.HasPrefix(trim, "game:") {
			flushRaw()
			flushGame()

			g := Game{GameName: strings.TrimSpace(trim[len("game:"):])}
			current = &g
			gameLines = append(gameLines, line)
			continue
		}

		if current == nil {
			rawBuf = append(rawBuf, line)
			continue
		}

		// If we're inside a game block and encounter a non-blank line that is NOT
		// a recognized game field, treat it as the start of a raw section (e.g.
		// collection blocks), matching existing behavior that only cares about
		// known keys.
		if trim != "" &&
			!strings.HasPrefix(trim, "file:") &&
			!strings.HasPrefix(trim, "sort-by:") &&
			!strings.HasPrefix(trim, "developer:") &&
			!strings.HasPrefix(trim, "description:") {
			// Close game block and re-process this line as raw.
			flushGame()
			rawBuf = append(rawBuf, line)
			continue
		}

		// inside game block
		switch {
		case strings.HasPrefix(trim, "file:"):
			current.FileName = strings.TrimSpace(trim[len("file:"):])
		case strings.HasPrefix(trim, "sort-by:"):
			current.SortBy = strings.TrimSpace(trim[len("sort-by:"):])
		case strings.HasPrefix(trim, "developer:"):
			current.Developer = strings.TrimSpace(trim[len("developer:"):])
		case strings.HasPrefix(trim, "description:"):
			current.Description = strings.TrimSpace(trim[len("description:"):])
		}
		gameLines = append(gameLines, line)
	}

	flushRaw()
	flushGame()
	return d
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

// RemoveByFileKeys removes any game block whose normalized file key matches keys.
// Returns number of removed blocks.
func (d *Document) RemoveByFileKeys(keys map[string]struct{}) int {
	if len(keys) == 0 {
		return 0
	}
	removed := 0
	filtered := d.items[:0]
	for _, it := range d.items {
		gi, ok := it.(GameItem)
		if ok {
			k := NormalizeFileKey(gi.game.FileName)
			if _, hit := keys[k]; hit {
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
