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
//   - Format uses no space after colon for backward compatibility: "key:value"
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

		gameData := Game{
			GameName:    g.GameName,
			FileName:    g.FileName,
			SortBy:      sortStr,
			Developer:   dev,
			Description: desc,
		}

		// Add optional fields if present
		if g.Publisher != "" {
			lines = append(lines, "publisher:"+g.Publisher)
			gameData.Publisher = g.Publisher
		}
		if g.Genre != "" {
			lines = append(lines, "genre:"+g.Genre)
			gameData.Genre = g.Genre
		}
		if g.Genres != "" {
			lines = append(lines, "genres:"+g.Genres)
			gameData.Genres = g.Genres
		}
		if g.Players != "" {
			lines = append(lines, "players:"+g.Players)
			gameData.Players = g.Players
		}
		if g.Rating != "" {
			lines = append(lines, "rating:"+g.Rating)
			gameData.Rating = g.Rating
		}
		if g.Release != "" {
			lines = append(lines, "release:"+g.Release)
			gameData.Release = g.Release
		}
		if g.Logo != "" {
			lines = append(lines, "logo:"+g.Logo)
			gameData.Logo = g.Logo
		}
		if g.Video != "" {
			lines = append(lines, "video:"+g.Video)
			gameData.Video = g.Video
		}
		if g.Screenshot != "" {
			lines = append(lines, "screenshot:"+g.Screenshot)
			gameData.Screenshot = g.Screenshot
		}
		if g.BoxFront != "" {
			lines = append(lines, "boxFront:"+g.BoxFront)
			gameData.BoxFront = g.BoxFront
		}
		if g.BoxBack != "" {
			lines = append(lines, "boxBack:"+g.BoxBack)
			gameData.BoxBack = g.BoxBack
		}
		if g.BoxSpine != "" {
			lines = append(lines, "boxSpine:"+g.BoxSpine)
			gameData.BoxSpine = g.BoxSpine
		}
		if g.BoxFull != "" {
			lines = append(lines, "boxFull:"+g.BoxFull)
			gameData.BoxFull = g.BoxFull
		}
		if g.Background != "" {
			lines = append(lines, "background:"+g.Background)
			gameData.Background = g.Background
		}
		if g.Music != "" {
			lines = append(lines, "music:"+g.Music)
			gameData.Music = g.Music
		}
		if g.Files != "" {
			lines = append(lines, "files:"+g.Files)
			gameData.Files = g.Files
		}

		d.items = append(d.items, GameItem{lines: lines, game: gameData})
	}

	// Ensure the file ends with a blank line.
	d.items = append(d.items, RawItem{lines: []string{""}})
}
