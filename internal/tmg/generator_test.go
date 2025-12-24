package tmg

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateSelectedFiles(t *testing.T) {
	root := t.TempDir()
	games := []GameModel{
		{Selected: true, GameName: "A", FileName: "a.zip"},
		{Selected: false, GameName: "B", FileName: "b.zip"},
		{Selected: true, GameName: "C", FileName: "nested/c.zip"},
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
