package metadata

import "fmt"

// New returns an empty Document with Unix newlines.
// Callers that want to preserve existing newline style should Parse/ReadFile instead.
func New() *Document {
	return &Document{newline: "\n"}
}

// SetGames replaces the document contents with a canonical set of game blocks.
//
// Canonical format:
//   - Exactly one blank line between game blocks.
//   - A trailing blank line at EOF.
//   - sort-by is zero-padded to 3 digits starting at 001.
func (d *Document) SetGames(games []Game) {
	d.items = d.items[:0]

	for i, g := range games {
		if i > 0 {
			d.items = append(d.items, RawItem{lines: []string{""}})
		}

		dev := g.Developer
		if dev == "" {
			dev = g.GameName
		}
		desc := g.Description
		if desc == "" {
			desc = g.GameName
		}

		sortStr := fmt.Sprintf("%03d", i+1)

		lines := []string{
			"game:" + g.GameName,
			"file:" + g.FileName,
			"sort-by:" + sortStr,
			"developer:" + dev,
			"description:" + desc,
		}
		d.items = append(d.items, GameItem{lines: lines, game: Game{GameName: g.GameName, FileName: g.FileName, SortBy: sortStr, Developer: dev, Description: desc}})
	}

	// Ensure the file ends with a blank line.
	d.items = append(d.items, RawItem{lines: []string{""}})
}
