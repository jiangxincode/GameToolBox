package pegasus

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRemoveSelectedGames_CurrentlyMirrorsGenerator(t *testing.T) {
	root := t.TempDir()

	games := []GameModel{
		{ID: 1, GameName: "g1", FileName: "roms/g1.zip", Selected: true},
	}

	res := RemoveSelectedGames(root, games)
	if res.Failed != 0 {
		t.Fatalf("expected Failed=0, got %d (errs=%v)", res.Failed, res.Errors)
	}
	if res.Created != 1 {
		t.Fatalf("expected Created=1, got %d", res.Created)
	}

	if _, err := os.Stat(filepath.Join(root, "roms/g1.zip")); err != nil {
		t.Fatalf("expected file created: %v", err)
	}
}
