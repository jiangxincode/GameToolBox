package pegasus

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAddSingleGame(t *testing.T) {
	dir := t.TempDir()
	metaPath := filepath.Join(dir, "metadata.pegasus.txt")

	// Test adding to empty metadata
	game := GameModel{
		GameName:    "Test Game",
		FileName:    "test.rom",
		Developer:   "Test Dev",
		Description: "Test Description",
	}

	res := AddSingleGame(dir, game)
	if res.Added != 1 {
		t.Errorf("expected Added=1, got %d", res.Added)
	}
	if res.Failed != 0 {
		t.Errorf("expected Failed=0, got %d", res.Failed)
	}
	if len(res.Errors) > 0 {
		t.Errorf("expected no errors, got %v", res.Errors)
	}

	// Verify file was created and contains the game
	content, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("failed to read metadata: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "game:Test Game") {
		t.Errorf("metadata should contain game name")
	}
	if !strings.Contains(text, "file:test.rom") {
		t.Errorf("metadata should contain file name")
	}
	if !strings.Contains(text, "developer:Test Dev") {
		t.Errorf("metadata should contain developer")
	}
	if !strings.Contains(text, "description:Test Description") {
		t.Errorf("metadata should contain description")
	}

	// Test adding duplicate game - should be skipped
	res2 := AddSingleGame(dir, game)
	if res2.Added != 0 {
		t.Errorf("expected Added=0 for duplicate, got %d", res2.Added)
	}
	if res2.Skipped != 1 {
		t.Errorf("expected Skipped=1 for duplicate, got %d", res2.Skipped)
	}
}

func TestAddGames_MultipleGames(t *testing.T) {
	dir := t.TempDir()

	games := []GameModel{
		{GameName: "Game 1", FileName: "game1.rom"},
		{GameName: "Game 2", FileName: "game2.rom"},
		{GameName: "Game 3", FileName: "game3.rom"},
	}

	res := AddGames(dir, games)
	if res.Added != 3 {
		t.Errorf("expected Added=3, got %d", res.Added)
	}
	if res.Failed != 0 {
		t.Errorf("expected Failed=0, got %d", res.Failed)
	}

	// Verify all games were added
	loaded, err := LoadGamesFromRootDir(dir)
	if err != nil {
		t.Fatalf("failed to load games: %v", err)
	}
	if len(loaded) != 3 {
		t.Errorf("expected 3 games, got %d", len(loaded))
	}
}

func TestAddGames_EmptyGameName(t *testing.T) {
	dir := t.TempDir()

	game := GameModel{
		GameName: "",
		FileName: "test.rom",
	}

	res := AddSingleGame(dir, game)
	if res.Failed != 1 {
		t.Errorf("expected Failed=1 for empty game name, got %d", res.Failed)
	}
	if res.Added != 0 {
		t.Errorf("expected Added=0, got %d", res.Added)
	}
	if len(res.Errors) == 0 {
		t.Error("expected error for empty game name")
	}
}

func TestAddGames_EmptyFileName(t *testing.T) {
	dir := t.TempDir()

	game := GameModel{
		GameName: "Test Game",
		FileName: "",
	}

	res := AddSingleGame(dir, game)
	if res.Failed != 1 {
		t.Errorf("expected Failed=1 for empty file name, got %d", res.Failed)
	}
	if res.Added != 0 {
		t.Errorf("expected Added=0, got %d", res.Added)
	}
	if len(res.Errors) == 0 {
		t.Error("expected error for empty file name")
	}
}

func TestAddGames_EmptyRootDir(t *testing.T) {
	game := GameModel{
		GameName: "Test Game",
		FileName: "test.rom",
	}

	res := AddSingleGame("", game)
	if res.Failed != 1 {
		t.Errorf("expected Failed=1 for empty root dir, got %d", res.Failed)
	}
	if len(res.Errors) == 0 {
		t.Error("expected error for empty root dir")
	}
}

func TestAddGames_DuplicateByFileName(t *testing.T) {
	dir := t.TempDir()

	game1 := GameModel{
		GameName: "Game 1",
		FileName: "test.rom",
	}

	game2 := GameModel{
		GameName: "Game 2",
		FileName: "test.rom", // Same file name
	}

	res1 := AddSingleGame(dir, game1)
	if res1.Added != 1 {
		t.Errorf("expected Added=1 for first game, got %d", res1.Added)
	}

	res2 := AddSingleGame(dir, game2)
	if res2.Skipped != 1 {
		t.Errorf("expected Skipped=1 for duplicate file, got %d", res2.Skipped)
	}
	if res2.Added != 0 {
		t.Errorf("expected Added=0 for duplicate file, got %d", res2.Added)
	}
}

func TestAddGames_DuplicateByGameName(t *testing.T) {
	dir := t.TempDir()

	game1 := GameModel{
		GameName: "Test Game",
		FileName: "game1.rom",
	}

	game2 := GameModel{
		GameName: "Test Game", // Same game name
		FileName: "game2.rom",
	}

	res1 := AddSingleGame(dir, game1)
	if res1.Added != 1 {
		t.Errorf("expected Added=1 for first game, got %d", res1.Added)
	}

	res2 := AddSingleGame(dir, game2)
	if res2.Skipped != 1 {
		t.Errorf("expected Skipped=1 for duplicate name, got %d", res2.Skipped)
	}
	if res2.Added != 0 {
		t.Errorf("expected Added=0 for duplicate name, got %d", res2.Added)
	}
}

func TestAddGames_AppendToExisting(t *testing.T) {
	dir := t.TempDir()

	// First add some games
	game1 := GameModel{
		GameName: "Game 1",
		FileName: "game1.rom",
	}
	res1 := AddSingleGame(dir, game1)
	if res1.Added != 1 {
		t.Fatalf("expected Added=1, got %d", res1.Added)
	}

	// Then add more games
	game2 := GameModel{
		GameName: "Game 2",
		FileName: "game2.rom",
	}
	res2 := AddSingleGame(dir, game2)
	if res2.Added != 1 {
		t.Errorf("expected Added=1 for second game, got %d", res2.Added)
	}

	// Verify both games are in metadata
	loaded, err := LoadGamesFromRootDir(dir)
	if err != nil {
		t.Fatalf("failed to load games: %v", err)
	}
	if len(loaded) != 2 {
		t.Errorf("expected 2 games, got %d", len(loaded))
	}
}
