package pegasus

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/game_tool_box/internal/pegasus/metadata"
)

func TestRemoveSelectedGames_CurrentlyMirrorsGenerator(t *testing.T) {
	root := t.TempDir()

	// create the rom file to be removed
	if err := os.MkdirAll(filepath.Join(root, "roms"), 0o755); err != nil {
		t.Fatalf("mkdir roms: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "roms", "g1.zip"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write rom: %v", err)
	}

	games := []GameModel{
		{ID: 1, Selected: true, Game: metadata.Game{GameName: "g1", FileName: "roms/g1.zip"}},
	}

	res := RemoveSelectedGames(root, games)
	if res.Failed != 0 {
		t.Fatalf("expected Failed=0, got %d (errs=%v)", res.Failed, res.Errors)
	}

	if _, err := os.Stat(filepath.Join(root, "roms", "g1.zip")); !os.IsNotExist(err) {
		t.Fatalf("expected rom removed, stat err=%v", err)
	}
}

func TestRemoveSelectedGames_RemovesMetadataMediaAndRom(t *testing.T) {
	root := t.TempDir()

	// metadata with two games
	meta := strings.Join([]string{
		"game: A",
		"file: A.zip",
		"developer: devA",
		"",
		"game: B",
		"file: B.zip",
		"developer: devB",
		"",
	}, "\n")
	if err := os.WriteFile(filepath.Join(root, "metadata.pegasus.txt"), []byte(meta), 0o644); err != nil {
		t.Fatalf("write metadata: %v", err)
	}

	// roms
	if err := os.WriteFile(filepath.Join(root, "A.zip"), []byte("a"), 0o644); err != nil {
		t.Fatalf("write A.zip: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "B.zip"), []byte("b"), 0o644); err != nil {
		t.Fatalf("write B.zip: %v", err)
	}

	// media dirs
	if err := os.MkdirAll(filepath.Join(root, "media", "A"), 0o755); err != nil {
		t.Fatalf("mkdir media/A: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "media", "A", "boxFront.png"), []byte("img"), 0o644); err != nil {
		t.Fatalf("write media/A/boxFront.png: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "media", "B"), 0o755); err != nil {
		t.Fatalf("mkdir media/B: %v", err)
	}

	games := []GameModel{
		{ID: 1, Selected: true, Game: metadata.Game{GameName: "A", FileName: "A.zip"}},
		{ID: 2, Selected: false, Game: metadata.Game{GameName: "B", FileName: "B.zip"}},
	}

	res := RemoveSelectedGames(root, games)
	if res.Failed != 0 {
		t.Fatalf("expected Failed=0 got %d (errs=%v)", res.Failed, res.Errors)
	}

	// A rom removed
	if _, err := os.Stat(filepath.Join(root, "A.zip")); !os.IsNotExist(err) {
		t.Fatalf("expected A.zip removed, stat err=%v", err)
	}
	// B rom remains
	if _, err := os.Stat(filepath.Join(root, "B.zip")); err != nil {
		t.Fatalf("expected B.zip exists: %v", err)
	}

	// media/A removed
	if _, err := os.Stat(filepath.Join(root, "media", "A")); !os.IsNotExist(err) {
		t.Fatalf("expected media/A removed, stat err=%v", err)
	}
	// media/B remains
	if _, err := os.Stat(filepath.Join(root, "media", "B")); err != nil {
		t.Fatalf("expected media/B exists: %v", err)
	}

	// metadata rewritten without A block
	b, err := os.ReadFile(filepath.Join(root, "metadata.pegasus.txt"))
	if err != nil {
		t.Fatalf("read metadata: %v", err)
	}
	out := string(b)
	if strings.Contains(out, "game: A") {
		t.Fatalf("expected metadata does not contain game A, got:\n%s", out)
	}
	if !strings.Contains(out, "game: B") {
		t.Fatalf("expected metadata still contains game B, got:\n%s", out)
	}
}

func TestRemoveSelectedGames_MissingRomCountsSkipped(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "metadata.pegasus.txt"), []byte("game: A\nfile: A.zip\n"), 0o644); err != nil {
		t.Fatalf("write metadata: %v", err)
	}

	games := []GameModel{{Selected: true, Game: metadata.Game{GameName: "A", FileName: "A.zip"}}}
	res := RemoveSelectedGames(root, games)
	if res.Failed != 0 {
		t.Fatalf("expected Failed=0 got %d (errs=%v)", res.Failed, res.Errors)
	}
	if res.Skipped == 0 {
		t.Fatalf("expected Skipped>0 when rom missing, got %d", res.Skipped)
	}
}
