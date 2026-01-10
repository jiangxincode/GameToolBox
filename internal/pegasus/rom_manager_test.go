package pegasus

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/game_tool_box/internal/pegasus/metadata"
)

func TestGenerateSelectedFiles(t *testing.T) {
	root := t.TempDir()
	games := []GameViewModel{
		{Selected: true, Game: metadata.Game{GameName: "A", FileName: "a.zip"}},
		{Selected: false, Game: metadata.Game{GameName: "B", FileName: "b.zip"}},
		{Selected: true, Game: metadata.Game{GameName: "C", FileName: "nested/c.zip"}},
	}

	res := GenerateSelectedFiles(root, games)
	if res.Created != 2 {
		t.Fatalf("expected Created=2 got %d (res=%+v)", res.Created, res)
	}

	if _, err := os.Stat(filepath.Join(root, "a.zip")); err != nil {
		t.Fatalf("expected a.zip created: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "nested", "c.zip")); err != nil {
		t.Fatalf("expected nested/c.zip created: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "b.zip")); !os.IsNotExist(err) {
		t.Fatalf("expected b.zip not created")
	}

	// second run should skip
	res2 := GenerateSelectedFiles(root, games)
	if res2.Skipped != 2 {
		t.Fatalf("expected Skipped=2 got %d (res=%+v)", res2.Skipped, res2)
	}
}

func TestDeleteRomsNotInConfig_DeletesOnlyMissingInConfig(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "roms"), 0o755); err != nil {
		t.Fatalf("mkdir roms: %v", err)
	}

	// Disk has A.zip and B.zip
	if err := os.WriteFile(filepath.Join(root, "roms", "A.zip"), []byte("a"), 0o644); err != nil {
		t.Fatalf("write A.zip: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "roms", "B.zip"), []byte("b"), 0o644); err != nil {
		t.Fatalf("write B.zip: %v", err)
	}

	// metadata contains only A.zip
	meta := strings.Join([]string{
		"game: A",
		"file: roms/A.zip",
		"",
	}, "\n")
	if err := os.WriteFile(filepath.Join(root, "metadata.pegasus.txt"), []byte(meta), 0o644); err != nil {
		t.Fatalf("write metadata: %v", err)
	}

	res := DeleteRomsNotInConfig(root)
	if len(res.Errors) > 0 {
		t.Fatalf("expected no errors, got %v", res.Errors[0])
	}
	if res.Deleted != 1 {
		t.Fatalf("expected Deleted=1 got %d (res=%+v)", res.Deleted, res)
	}

	// A should remain, B should be deleted
	if _, err := os.Stat(filepath.Join(root, "roms", "A.zip")); err != nil {
		t.Fatalf("expected A.zip remains: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "roms", "B.zip")); !os.IsNotExist(err) {
		t.Fatalf("expected B.zip deleted")
	}
}
