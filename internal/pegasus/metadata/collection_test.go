package metadata

import (
	"strings"
	"testing"
)

// TestDocument_ParseCollection verifies that collection blocks are correctly parsed
func TestDocument_ParseCollection(t *testing.T) {
	in := strings.Join([]string{
		"collection: ALOYS_PSV",
		"sort-by: 103",
		"shortname: aloys_psv",
		"launch: vita3k \"{file.path}\"",
		"",
		"game: Test Game",
		"file: test.zip",
		"",
	}, "\n")

	doc := Parse(in)
	collections := doc.Collections()

	if len(collections) != 1 {
		t.Fatalf("expected 1 collection, got %d", len(collections))
	}

	c := collections[0]
	if c.Name != "ALOYS_PSV" {
		t.Errorf("Name: expected %q, got %q", "ALOYS_PSV", c.Name)
	}
	if c.SortBy != "103" {
		t.Errorf("SortBy: expected %q, got %q", "103", c.SortBy)
	}
	if c.ShortName != "aloys_psv" {
		t.Errorf("ShortName: expected %q, got %q", "aloys_psv", c.ShortName)
	}
	if c.Launch != "vita3k \"{file.path}\"" {
		t.Errorf("Launch: expected %q, got %q", "vita3k \"{file.path}\"", c.Launch)
	}

	// Verify game is still parsed after collection
	games := doc.Games()
	if len(games) != 1 {
		t.Fatalf("expected 1 game, got %d", len(games))
	}
	if games[0].GameName != "Test Game" {
		t.Errorf("Game name: expected %q, got %q", "Test Game", games[0].GameName)
	}
}

// TestDocument_PreserveCollectionInRoundTrip verifies collections are preserved
func TestDocument_PreserveCollectionInRoundTrip(t *testing.T) {
	in := strings.Join([]string{
		"# Comment",
		"collection: Super Nintendo",
		"extension: smc, sfc",
		"launch: snes9x \"{file.path}\"",
		"",
		"game: Super Mario World",
		"file: smw.smc",
		"",
	}, "\n")

	doc := Parse(in)
	out := doc.Render()

	// Verify collection is preserved in output
	if !strings.Contains(out, "collection: Super Nintendo") {
		t.Errorf("expected collection line preserved, got:\n%s", out)
	}
	if !strings.Contains(out, "extension: smc, sfc") {
		t.Errorf("expected extension line preserved, got:\n%s", out)
	}
	if !strings.Contains(out, "launch: snes9x") {
		t.Errorf("expected launch line preserved, got:\n%s", out)
	}

	// Verify game is preserved
	if !strings.Contains(out, "game: Super Mario World") {
		t.Errorf("expected game preserved, got:\n%s", out)
	}

	// Verify comment is preserved
	if !strings.Contains(out, "# Comment") {
		t.Errorf("expected comment preserved, got:\n%s", out)
	}
}

// TestDocument_MultipleCollectionsAndGames verifies multiple collections work
func TestDocument_MultipleCollectionsAndGames(t *testing.T) {
	in := strings.Join([]string{
		"collection: SNES",
		"extension: smc",
		"",
		"game: Mario",
		"file: mario.smc",
		"",
		"collection: PSX",
		"extension: iso",
		"",
		"game: FF7",
		"file: ff7.iso",
		"",
	}, "\n")

	doc := Parse(in)
	collections := doc.Collections()
	games := doc.Games()

	if len(collections) != 2 {
		t.Fatalf("expected 2 collections, got %d", len(collections))
	}

	if collections[0].Name != "SNES" {
		t.Errorf("collection 0: expected SNES, got %q", collections[0].Name)
	}
	if collections[0].Extensions != "smc" {
		t.Errorf("collection 0 extensions: expected smc, got %q", collections[0].Extensions)
	}

	if collections[1].Name != "PSX" {
		t.Errorf("collection 1: expected PSX, got %q", collections[1].Name)
	}
	if collections[1].Extensions != "iso" {
		t.Errorf("collection 1 extensions: expected iso, got %q", collections[1].Extensions)
	}

	if len(games) != 2 {
		t.Fatalf("expected 2 games, got %d", len(games))
	}

	if games[0].GameName != "Mario" {
		t.Errorf("game 0: expected Mario, got %q", games[0].GameName)
	}
	if games[1].GameName != "FF7" {
		t.Errorf("game 1: expected FF7, got %q", games[1].GameName)
	}
}

// TestDocument_CollectionFieldsPreserved verifies all collection fields work
func TestDocument_CollectionFieldsPreserved(t *testing.T) {
	in := strings.Join([]string{
		"collection: TestCollection",
		"shortname: test_coll",
		"sort-by: 100",
		"launch: emulator \"{file.path}\"",
		"extension: iso, bin, cue",
		"files: /path/to/roms",
		"",
	}, "\n")

	doc := Parse(in)
	collections := doc.Collections()

	if len(collections) != 1 {
		t.Fatalf("expected 1 collection, got %d", len(collections))
	}

	c := collections[0]
	if c.Name != "TestCollection" {
		t.Errorf("Name: expected %q, got %q", "TestCollection", c.Name)
	}
	if c.ShortName != "test_coll" {
		t.Errorf("ShortName: expected %q, got %q", "test_coll", c.ShortName)
	}
	if c.SortBy != "100" {
		t.Errorf("SortBy: expected %q, got %q", "100", c.SortBy)
	}
	if c.Launch != "emulator \"{file.path}\"" {
		t.Errorf("Launch: expected %q, got %q", "emulator \"{file.path}\"", c.Launch)
	}
	if c.Extensions != "iso, bin, cue" {
		t.Errorf("Extensions: expected %q, got %q", "iso, bin, cue", c.Extensions)
	}
	if c.Files != "/path/to/roms" {
		t.Errorf("Files: expected %q, got %q", "/path/to/roms", c.Files)
	}
}
