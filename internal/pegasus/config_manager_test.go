package pegasus

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateSelectedConfigFiles_MirrorsRomManagerBehavior(t *testing.T) {
	root := t.TempDir()

	games := []GameModel{
		{ID: 1, GameName: "g1", FileName: "roms/g1.zip", Selected: true},
		{ID: 2, GameName: "g2", FileName: "roms/g2.zip", Selected: false},
	}

	res := GenerateSelectedConfigFiles(root, games)
	if res.Failed != 0 {
		t.Fatalf("expected Failed=0, got %d (errs=%v)", res.Failed, res.Errors)
	}
	if res.Created != 1 {
		t.Fatalf("expected Created=1, got %d", res.Created)
	}
	if res.Skipped != 0 {
		t.Fatalf("expected Skipped=0, got %d", res.Skipped)
	}

	if _, err := os.Stat(filepath.Join(root, "roms/g1.zip")); err != nil {
		t.Fatalf("expected file created: %v", err)
	}

	// second run should skip
	res2 := GenerateSelectedConfigFiles(root, games)
	if res2.Skipped != 1 {
		t.Fatalf("expected Skipped=1 on second run, got %d", res2.Skipped)
	}
}
