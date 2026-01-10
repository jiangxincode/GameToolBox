package metadata

import (
	"path/filepath"
	"strings"
)

type UpsertOptions struct {
	// If true, when an existing game block matches by file:, we also update its
	// sort-by value from g.SortBy (or auto-generated if empty).
	// Default: false (preserve existing sort-by).
	UpdateSortBy bool
}

// UpsertGameByFile inserts a new game block if no existing block has a matching
// `file:` (after NormalizeFileKey). If a match exists, it updates the existing
// block's known fields (game/file/developer/description and optionally sort-by)
// while preserving unknown lines.
//
// Returns (changed, found, err).
func (d *Document) UpsertGameByFile(g Game, opts UpsertOptions) (changed bool, found bool, err error) {
	key := NormalizeFileKey(g.FileName)
	if key == "" {
		return false, false, ErrInvalidFileName
	}

	name := strings.TrimSpace(g.GameName)
	if name == "" {
		base := filepath.Base(filepath.FromSlash(strings.TrimSpace(g.FileName)))
		name = strings.TrimSuffix(base, filepath.Ext(base))
	}

	dev := strings.TrimSpace(g.Developer)
	if dev == "" {
		dev = name
	}
	desc := strings.TrimSpace(g.Description)
	if desc == "" {
		desc = name
	}

	// First pass: try update existing.
	for i, it := range d.items {
		gi, ok := it.(GameItem)
		if !ok {
			continue
		}
		if NormalizeFileKey(gi.game.FileName) != key {
			continue
		}
		found = true

		newGame := gi.game
		newGame.GameName = name
		newGame.FileName = strings.TrimSpace(g.FileName)
		newGame.Developer = dev
		newGame.Description = desc

		// sort-by: preserve unless explicitly asked to update.
		if opts.UpdateSortBy {
			if strings.TrimSpace(g.SortBy) != "" {
				newGame.SortBy = strings.TrimSpace(g.SortBy)
			} else {
				info := d.InferSortByFormat()
				newGame.SortBy = formatSortBy(info.Width, info.MaxValue+1)
			}
		}

		newLines, linesChanged := updateGameLines(gi.lines, newGame, opts.UpdateSortBy)
		if linesChanged {
			changed = true
			d.items[i] = GameItem{lines: newLines, game: newGame}
		}
		return changed, found, nil
	}

	// Not found: append as new block.
	changed = true
	if err := d.AppendGames([]Game{{
		GameName:    name,
		FileName:    strings.TrimSpace(g.FileName),
		Developer:   dev,
		Description: desc,
	}}); err != nil {
		return false, false, err
	}
	return true, false, nil
}

// ErrInvalidFileName is returned when the file name is empty.
var ErrInvalidFileName = &MetadataError{"fileName is empty"}

type MetadataError struct{ msg string }

func (e *MetadataError) Error() string { return e.msg }

func formatSortBy(width, n int) string {
	if width <= 1 {
		return strconvItoa(n)
	}
	s := strconvItoa(n)
	if len(s) >= width {
		return s
	}
	return strings.Repeat("0", width-len(s)) + s
}

func strconvItoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	var b [32]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

// updateGameLines rewrites known keys in a game block while preserving the rest.
func updateGameLines(lines []string, newG Game, updateSort bool) (out []string, changed bool) {
	out = make([]string, 0, len(lines)+4)
	seen := map[string]bool{}

	setLine := func(key, val string) string {
		return key + ":" + val
	}

	for _, ln := range lines {
		k, _, ok := parseKVLoose(ln)
		if !ok {
			out = append(out, ln)
			continue
		}
		switch k {
		case "game":
			seen[k] = true
			nl := setLine("game", newG.GameName)
			if strings.TrimSpace(ln) != strings.TrimSpace(nl) {
				changed = true
			}
			out = append(out, nl)
		case "file":
			seen[k] = true
			nl := setLine("file", newG.FileName)
			if strings.TrimSpace(ln) != strings.TrimSpace(nl) {
				changed = true
			}
			out = append(out, nl)
		case "developer":
			seen[k] = true
			nl := setLine("developer", newG.Developer)
			if strings.TrimSpace(ln) != strings.TrimSpace(nl) {
				changed = true
			}
			out = append(out, nl)
		case "description":
			seen[k] = true
			nl := setLine("description", newG.Description)
			if strings.TrimSpace(ln) != strings.TrimSpace(nl) {
				changed = true
			}
			out = append(out, nl)
		case "sort-by":
			seen[k] = true
			if !updateSort {
				out = append(out, ln)
				continue
			}
			nl := setLine("sort-by", newG.SortBy)
			if strings.TrimSpace(ln) != strings.TrimSpace(nl) {
				changed = true
			}
			out = append(out, nl)
		default:
			out = append(out, ln)
		}
	}

	if !seen["game"] {
		changed = true
		out = append([]string{setLine("game", newG.GameName)}, out...)
	}
	if !seen["file"] {
		changed = true
		out = append(out, setLine("file", newG.FileName))
	}
	if updateSort && !seen["sort-by"] {
		changed = true
		out = append(out, setLine("sort-by", newG.SortBy))
	}
	if !seen["developer"] {
		changed = true
		out = append(out, setLine("developer", newG.Developer))
	}
	if !seen["description"] {
		changed = true
		out = append(out, setLine("description", newG.Description))
	}

	return out, changed
}

func parseKVLoose(line string) (key, val string, ok bool) {
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
