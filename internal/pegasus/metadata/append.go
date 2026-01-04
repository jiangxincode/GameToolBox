package metadata

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

type SortByFormat struct {
	MaxValue int
	Width    int
}

// InferSortByFormat scans the document and returns the maximum sort-by value
// among game blocks and the width of the max sort-by token.
//
// Notes:
//   - Only considers sort-by values inside game blocks.
//   - Collection blocks and any raw content are ignored.
//   - Width preserves the raw digit width of the max sort-by (e.g. "010" => 3).
func (d *Document) InferSortByFormat() SortByFormat {
	maxVal := 0
	maxWidth := 0

	for _, it := range d.items {
		gi, ok := it.(GameItem)
		if !ok {
			continue
		}
		svRaw := strings.TrimSpace(gi.game.SortBy)
		if svRaw == "" {
			continue
		}
		// Mirror old behavior: parse integer after stripping leading zeros.
		sv := strings.TrimLeft(svRaw, "0")
		if sv == "" {
			sv = "0"
		}
		n, err := strconv.Atoi(sv)
		if err != nil {
			continue
		}
		if n > maxVal {
			maxVal = n
			maxWidth = len(svRaw)
		} else if n == maxVal {
			// Keep the widest representation.
			if len(svRaw) > maxWidth {
				maxWidth = len(svRaw)
			}
		}
	}

	if maxWidth <= 0 {
		maxWidth = 1
	}
	return SortByFormat{MaxValue: maxVal, Width: maxWidth}
}

// AppendGames appends games as new game blocks.
//
// Formatting contract (matches existing pegasus.AppendMissingGamesToMetadata tests):
//   - Each appended game block has fields: game/file/sort-by/developer/description
//   - developer/description default to game name.
//   - sort-by increments from existing maximum and is zero-padded to inferred width.
//   - Exactly one blank line between blocks, and file ends with a blank line.
func (d *Document) AppendGames(games []Game) error {
	if len(games) == 0 {
		return nil
	}

	fmtInfo := d.InferSortByFormat()
	sortNo := fmtInfo.MaxValue

	// Ensure we start with an explicit blank line separator if the last item isn't already blank.
	// We'll rely on Render() to collapse excessive blanks.
	lastLineIsBlank := false
	if len(d.items) > 0 {
		lns := d.items[len(d.items)-1].Lines()
		if len(lns) > 0 && strings.TrimSpace(lns[len(lns)-1]) == "" {
			lastLineIsBlank = true
		}
	}
	if !lastLineIsBlank {
		d.items = append(d.items, RawItem{lines: []string{""}})
	}

	for i, g := range games {
		name := strings.TrimSpace(g.GameName)
		if name == "" {
			// best-effort name from file
			base := filepath.Base(filepath.FromSlash(strings.TrimSpace(g.FileName)))
			name = strings.TrimSuffix(base, filepath.Ext(base))
		}
		file := strings.TrimSpace(g.FileName)
		if file == "" {
			return fmt.Errorf("append game %q: fileName is empty", name)
		}
		sortNo++
		sortStr := fmt.Sprintf("%0*d", fmtInfo.Width, sortNo)

		dev := strings.TrimSpace(g.Developer)
		if dev == "" {
			dev = name
		}
		desc := strings.TrimSpace(g.Description)
		if desc == "" {
			desc = name
		}

		lines := []string{
			"game:" + name,
			"file:" + file,
			"sort-by:" + sortStr,
			"developer:" + dev,
			"description:" + desc,
		}

		// Add exactly one blank line between appended game blocks.
		if i > 0 {
			d.items = append(d.items, RawItem{lines: []string{""}})
		}
		d.items = append(d.items, GameItem{lines: lines, game: Game{GameName: name, FileName: file, SortBy: sortStr, Developer: dev, Description: desc}})
	}

	// Ensure trailing blank line.
	d.items = append(d.items, RawItem{lines: []string{""}})
	return nil
}
