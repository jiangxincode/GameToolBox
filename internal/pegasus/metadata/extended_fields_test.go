package metadata

import (
	"strings"
	"testing"
)

// TestDocument_ParseExtendedFields verifies that all new Pegasus fields are correctly parsed
func TestDocument_ParseExtendedFields(t *testing.T) {
	in := strings.Join([]string{
		"game: Final Fantasy VII",
		"file: ff7.iso",
		"developer: Square",
		"publisher: Sony",
		"genre: RPG",
		"players: 1",
		"rating: 95",
		"release: 1997",
		"description: A classic RPG",
		"logo: media/ff7/logo.png",
		"video: media/ff7/video.mp4",
		"screenshot: media/ff7/screenshot.png",
		"boxFront: media/ff7/boxFront.png",
		"",
	}, "\n")

	doc := Parse(in)
	games := doc.Games()

	if len(games) != 1 {
		t.Fatalf("expected 1 game, got %d", len(games))
	}

	g := games[0]
	if g.GameName != "Final Fantasy VII" {
		t.Errorf("GameName: expected %q, got %q", "Final Fantasy VII", g.GameName)
	}
	if g.FileName != "ff7.iso" {
		t.Errorf("FileName: expected %q, got %q", "ff7.iso", g.FileName)
	}
	if g.Developer != "Square" {
		t.Errorf("Developer: expected %q, got %q", "Square", g.Developer)
	}
	if g.Publisher != "Sony" {
		t.Errorf("Publisher: expected %q, got %q", "Sony", g.Publisher)
	}
	if g.Genre != "RPG" {
		t.Errorf("Genre: expected %q, got %q", "RPG", g.Genre)
	}
	if g.Players != "1" {
		t.Errorf("Players: expected %q, got %q", "1", g.Players)
	}
	if g.Rating != "95" {
		t.Errorf("Rating: expected %q, got %q", "95", g.Rating)
	}
	if g.Release != "1997" {
		t.Errorf("Release: expected %q, got %q", "1997", g.Release)
	}
	if g.Description != "A classic RPG" {
		t.Errorf("Description: expected %q, got %q", "A classic RPG", g.Description)
	}
	if g.Logo != "media/ff7/logo.png" {
		t.Errorf("Logo: expected %q, got %q", "media/ff7/logo.png", g.Logo)
	}
	if g.Video != "media/ff7/video.mp4" {
		t.Errorf("Video: expected %q, got %q", "media/ff7/video.mp4", g.Video)
	}
	if g.Screenshot != "media/ff7/screenshot.png" {
		t.Errorf("Screenshot: expected %q, got %q", "media/ff7/screenshot.png", g.Screenshot)
	}
	if g.BoxFront != "media/ff7/boxFront.png" {
		t.Errorf("BoxFront: expected %q, got %q", "media/ff7/boxFront.png", g.BoxFront)
	}
}

// TestDocument_SetGamesWithExtendedFields verifies that extended fields are written when present
func TestDocument_SetGamesWithExtendedFields(t *testing.T) {
	doc := New()
	games := []Game{
		{
			GameName:    "Super Mario World",
			FileName:    "smw.smc",
			Developer:   "Nintendo",
			Publisher:   "Nintendo",
			Genre:       "Platformer",
			Players:     "2",
			Rating:      "94",
			Release:     "1990",
			Description: "Classic platformer",
			Logo:        "media/smw/logo.png",
			BoxFront:    "media/smw/boxFront.png",
		},
	}

	doc.SetGames(games)
	out := doc.Render()

	expectedFields := []string{
		"game:Super Mario World",
		"file:smw.smc",
		"developer:Nintendo",
		"publisher:Nintendo",
		"genre:Platformer",
		"players:2",
		"rating:94",
		"release:1990",
		"description:Classic platformer",
		"logo:media/smw/logo.png",
		"boxFront:media/smw/boxFront.png",
	}

	for _, field := range expectedFields {
		if !strings.Contains(out, field) {
			t.Errorf("expected output to contain %q, got:\n%s", field, out)
		}
	}
}

// TestDocument_AppendGamesWithExtendedFields verifies extended fields work with AppendGames
func TestDocument_AppendGamesWithExtendedFields(t *testing.T) {
	doc := Parse("game:Existing\nfile:existing.rom\n\n")

	newGames := []Game{
		{
			GameName:  "New Game",
			FileName:  "new.rom",
			Developer: "DevTeam",
			Publisher: "PubCorp",
			Genre:     "Action",
			Players:   "4",
			Rating:    "85",
		},
	}

	if err := doc.AppendGames(newGames); err != nil {
		t.Fatalf("AppendGames: %v", err)
	}

	out := doc.Render()

	if !strings.Contains(out, "game:New Game") {
		t.Errorf("expected new game, got:\n%s", out)
	}
	if !strings.Contains(out, "publisher:PubCorp") {
		t.Errorf("expected publisher field, got:\n%s", out)
	}
	if !strings.Contains(out, "genre:Action") {
		t.Errorf("expected genre field, got:\n%s", out)
	}
	if !strings.Contains(out, "players:4") {
		t.Errorf("expected players field, got:\n%s", out)
	}
	if !strings.Contains(out, "rating:85") {
		t.Errorf("expected rating field, got:\n%s", out)
	}

	// Verify old game is still there
	if !strings.Contains(out, "game:Existing") {
		t.Errorf("expected existing game preserved, got:\n%s", out)
	}
}

// TestDocument_UpsertWithExtendedFields verifies upsert handles extended fields
func TestDocument_UpsertWithExtendedFields(t *testing.T) {
	in := strings.Join([]string{
		"game: Old Title",
		"file: test.rom",
		"developer: OldDev",
		"",
	}, "\n")

	doc := Parse(in)
	game := Game{
		GameName:  "New Title",
		FileName:  "test.rom",
		Developer: "NewDev",
		Publisher: "NewPub",
		Genre:     "Strategy",
		Players:   "1-2",
	}

	changed, found, err := doc.UpsertGameByFile(game, UpsertOptions{})
	if err != nil {
		t.Fatalf("UpsertGameByFile: %v", err)
	}
	if !changed {
		t.Errorf("expected changed=true")
	}
	if !found {
		t.Errorf("expected found=true")
	}

	out := doc.Render()

	if !strings.Contains(out, "game:New Title") {
		t.Errorf("expected game name updated, got:\n%s", out)
	}
	if !strings.Contains(out, "developer:NewDev") {
		t.Errorf("expected developer updated, got:\n%s", out)
	}
	if !strings.Contains(out, "publisher:NewPub") {
		t.Errorf("expected publisher added, got:\n%s", out)
	}
	if !strings.Contains(out, "genre:Strategy") {
		t.Errorf("expected genre added, got:\n%s", out)
	}
	if !strings.Contains(out, "players:1-2") {
		t.Errorf("expected players added, got:\n%s", out)
	}
}

// TestDocument_ParseMultipleGamesWithVariedFields verifies parsing works with mixed field sets
func TestDocument_ParseMultipleGamesWithVariedFields(t *testing.T) {
	in := strings.Join([]string{
		"# Collection header",
		"collection: SNES",
		"",
		"game: Mario",
		"file: mario.smc",
		"genre: Platformer",
		"",
		"game: Zelda",
		"file: zelda.smc",
		"developer: Nintendo",
		"publisher: Nintendo",
		"players: 1",
		"rating: 98",
		"",
	}, "\n")

	doc := Parse(in)
	games := doc.Games()

	if len(games) != 2 {
		t.Fatalf("expected 2 games, got %d", len(games))
	}

	// First game should have genre but not other fields
	if games[0].GameName != "Mario" {
		t.Errorf("game 0: expected Mario, got %q", games[0].GameName)
	}
	if games[0].Genre != "Platformer" {
		t.Errorf("game 0: expected genre Platformer, got %q", games[0].Genre)
	}

	// Second game should have multiple fields
	if games[1].GameName != "Zelda" {
		t.Errorf("game 1: expected Zelda, got %q", games[1].GameName)
	}
	if games[1].Developer != "Nintendo" {
		t.Errorf("game 1: expected developer Nintendo, got %q", games[1].Developer)
	}
	if games[1].Publisher != "Nintendo" {
		t.Errorf("game 1: expected publisher Nintendo, got %q", games[1].Publisher)
	}
	if games[1].Players != "1" {
		t.Errorf("game 1: expected players 1, got %q", games[1].Players)
	}
	if games[1].Rating != "98" {
		t.Errorf("game 1: expected rating 98, got %q", games[1].Rating)
	}

	// Verify collection is preserved (not parsed as a game)
	out := doc.Render()
	if !strings.Contains(out, "collection: SNES") {
		t.Errorf("expected collection preserved, got:\n%s", out)
	}
}
