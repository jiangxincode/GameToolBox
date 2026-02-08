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

		// Update all new fields if present
		if g.Publisher != "" {
			newGame.Publisher = strings.TrimSpace(g.Publisher)
		}
		if g.Genre != "" {
			newGame.Genre = strings.TrimSpace(g.Genre)
		}
		if g.Genres != "" {
			newGame.Genres = strings.TrimSpace(g.Genres)
		}
		if g.Players != "" {
			newGame.Players = strings.TrimSpace(g.Players)
		}
		if g.Rating != "" {
			newGame.Rating = strings.TrimSpace(g.Rating)
		}
		if g.Release != "" {
			newGame.Release = strings.TrimSpace(g.Release)
		}
		if g.Logo != "" {
			newGame.Logo = strings.TrimSpace(g.Logo)
		}
		if g.Video != "" {
			newGame.Video = strings.TrimSpace(g.Video)
		}
		if g.Screenshot != "" {
			newGame.Screenshot = strings.TrimSpace(g.Screenshot)
		}
		if g.BoxFront != "" {
			newGame.BoxFront = strings.TrimSpace(g.BoxFront)
		}
		if g.BoxBack != "" {
			newGame.BoxBack = strings.TrimSpace(g.BoxBack)
		}
		if g.BoxSpine != "" {
			newGame.BoxSpine = strings.TrimSpace(g.BoxSpine)
		}
		if g.BoxFull != "" {
			newGame.BoxFull = strings.TrimSpace(g.BoxFull)
		}
		if g.Background != "" {
			newGame.Background = strings.TrimSpace(g.Background)
		}
		if g.Music != "" {
			newGame.Music = strings.TrimSpace(g.Music)
		}
		if g.Files != "" {
			newGame.Files = strings.TrimSpace(g.Files)
		}

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
		Publisher:   strings.TrimSpace(g.Publisher),
		Genre:       strings.TrimSpace(g.Genre),
		Genres:      strings.TrimSpace(g.Genres),
		Players:     strings.TrimSpace(g.Players),
		Rating:      strings.TrimSpace(g.Rating),
		Release:     strings.TrimSpace(g.Release),
		Logo:        strings.TrimSpace(g.Logo),
		Video:       strings.TrimSpace(g.Video),
		Screenshot:  strings.TrimSpace(g.Screenshot),
		BoxFront:    strings.TrimSpace(g.BoxFront),
		BoxBack:     strings.TrimSpace(g.BoxBack),
		BoxSpine:    strings.TrimSpace(g.BoxSpine),
		BoxFull:     strings.TrimSpace(g.BoxFull),
		Background:  strings.TrimSpace(g.Background),
		Music:       strings.TrimSpace(g.Music),
		Files:       strings.TrimSpace(g.Files),
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
		case "publisher":
			seen[k] = true
			if newG.Publisher != "" {
				nl := setLine("publisher", newG.Publisher)
				if strings.TrimSpace(ln) != strings.TrimSpace(nl) {
					changed = true
				}
				out = append(out, nl)
			} else {
				out = append(out, ln)
			}
		case "genre":
			seen[k] = true
			if newG.Genre != "" {
				nl := setLine("genre", newG.Genre)
				if strings.TrimSpace(ln) != strings.TrimSpace(nl) {
					changed = true
				}
				out = append(out, nl)
			} else {
				out = append(out, ln)
			}
		case "genres":
			seen[k] = true
			if newG.Genres != "" {
				nl := setLine("genres", newG.Genres)
				if strings.TrimSpace(ln) != strings.TrimSpace(nl) {
					changed = true
				}
				out = append(out, nl)
			} else {
				out = append(out, ln)
			}
		case "players":
			seen[k] = true
			if newG.Players != "" {
				nl := setLine("players", newG.Players)
				if strings.TrimSpace(ln) != strings.TrimSpace(nl) {
					changed = true
				}
				out = append(out, nl)
			} else {
				out = append(out, ln)
			}
		case "rating":
			seen[k] = true
			if newG.Rating != "" {
				nl := setLine("rating", newG.Rating)
				if strings.TrimSpace(ln) != strings.TrimSpace(nl) {
					changed = true
				}
				out = append(out, nl)
			} else {
				out = append(out, ln)
			}
		case "release":
			seen[k] = true
			if newG.Release != "" {
				nl := setLine("release", newG.Release)
				if strings.TrimSpace(ln) != strings.TrimSpace(nl) {
					changed = true
				}
				out = append(out, nl)
			} else {
				out = append(out, ln)
			}
		case "logo":
			seen[k] = true
			if newG.Logo != "" {
				nl := setLine("logo", newG.Logo)
				if strings.TrimSpace(ln) != strings.TrimSpace(nl) {
					changed = true
				}
				out = append(out, nl)
			} else {
				out = append(out, ln)
			}
		case "video":
			seen[k] = true
			if newG.Video != "" {
				nl := setLine("video", newG.Video)
				if strings.TrimSpace(ln) != strings.TrimSpace(nl) {
					changed = true
				}
				out = append(out, nl)
			} else {
				out = append(out, ln)
			}
		case "screenshot":
			seen[k] = true
			if newG.Screenshot != "" {
				nl := setLine("screenshot", newG.Screenshot)
				if strings.TrimSpace(ln) != strings.TrimSpace(nl) {
					changed = true
				}
				out = append(out, nl)
			} else {
				out = append(out, ln)
			}
		case "boxfront":
			seen[k] = true
			if newG.BoxFront != "" {
				nl := setLine("boxFront", newG.BoxFront)
				if strings.TrimSpace(ln) != strings.TrimSpace(nl) {
					changed = true
				}
				out = append(out, nl)
			} else {
				out = append(out, ln)
			}
		case "boxback":
			seen[k] = true
			if newG.BoxBack != "" {
				nl := setLine("boxBack", newG.BoxBack)
				if strings.TrimSpace(ln) != strings.TrimSpace(nl) {
					changed = true
				}
				out = append(out, nl)
			} else {
				out = append(out, ln)
			}
		case "boxspine":
			seen[k] = true
			if newG.BoxSpine != "" {
				nl := setLine("boxSpine", newG.BoxSpine)
				if strings.TrimSpace(ln) != strings.TrimSpace(nl) {
					changed = true
				}
				out = append(out, nl)
			} else {
				out = append(out, ln)
			}
		case "boxfull":
			seen[k] = true
			if newG.BoxFull != "" {
				nl := setLine("boxFull", newG.BoxFull)
				if strings.TrimSpace(ln) != strings.TrimSpace(nl) {
					changed = true
				}
				out = append(out, nl)
			} else {
				out = append(out, ln)
			}
		case "background":
			seen[k] = true
			if newG.Background != "" {
				nl := setLine("background", newG.Background)
				if strings.TrimSpace(ln) != strings.TrimSpace(nl) {
					changed = true
				}
				out = append(out, nl)
			} else {
				out = append(out, ln)
			}
		case "music":
			seen[k] = true
			if newG.Music != "" {
				nl := setLine("music", newG.Music)
				if strings.TrimSpace(ln) != strings.TrimSpace(nl) {
					changed = true
				}
				out = append(out, nl)
			} else {
				out = append(out, ln)
			}
		case "files":
			seen[k] = true
			if newG.Files != "" {
				nl := setLine("files", newG.Files)
				if strings.TrimSpace(ln) != strings.TrimSpace(nl) {
					changed = true
				}
				out = append(out, nl)
			} else {
				out = append(out, ln)
			}
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

	// Add new optional fields if not present and have values
	if !seen["publisher"] && newG.Publisher != "" {
		changed = true
		out = append(out, setLine("publisher", newG.Publisher))
	}
	if !seen["genre"] && newG.Genre != "" {
		changed = true
		out = append(out, setLine("genre", newG.Genre))
	}
	if !seen["genres"] && newG.Genres != "" {
		changed = true
		out = append(out, setLine("genres", newG.Genres))
	}
	if !seen["players"] && newG.Players != "" {
		changed = true
		out = append(out, setLine("players", newG.Players))
	}
	if !seen["rating"] && newG.Rating != "" {
		changed = true
		out = append(out, setLine("rating", newG.Rating))
	}
	if !seen["release"] && newG.Release != "" {
		changed = true
		out = append(out, setLine("release", newG.Release))
	}
	if !seen["logo"] && newG.Logo != "" {
		changed = true
		out = append(out, setLine("logo", newG.Logo))
	}
	if !seen["video"] && newG.Video != "" {
		changed = true
		out = append(out, setLine("video", newG.Video))
	}
	if !seen["screenshot"] && newG.Screenshot != "" {
		changed = true
		out = append(out, setLine("screenshot", newG.Screenshot))
	}
	if !seen["boxfront"] && newG.BoxFront != "" {
		changed = true
		out = append(out, setLine("boxFront", newG.BoxFront))
	}
	if !seen["boxback"] && newG.BoxBack != "" {
		changed = true
		out = append(out, setLine("boxBack", newG.BoxBack))
	}
	if !seen["boxspine"] && newG.BoxSpine != "" {
		changed = true
		out = append(out, setLine("boxSpine", newG.BoxSpine))
	}
	if !seen["boxfull"] && newG.BoxFull != "" {
		changed = true
		out = append(out, setLine("boxFull", newG.BoxFull))
	}
	if !seen["background"] && newG.Background != "" {
		changed = true
		out = append(out, setLine("background", newG.Background))
	}
	if !seen["music"] && newG.Music != "" {
		changed = true
		out = append(out, setLine("music", newG.Music))
	}
	if !seen["files"] && newG.Files != "" {
		changed = true
		out = append(out, setLine("files", newG.Files))
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
