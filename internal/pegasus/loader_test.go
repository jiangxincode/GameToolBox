package pegasus

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestLoadGamesFromRootDir(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(thisFile), "testdata", "PSVITA")

	games, err := LoadGamesFromRootDir(root)
	if err != nil {
		t.Fatalf("LoadGamesFromRootDir: %v", err)
	}
	if len(games) == 0 {
		t.Fatalf("expected games > 0")
	}

	// basic sanity check against the real testdata
	if games[0].GameName == "" || games[0].FileName == "" {
		t.Fatalf("expected first game has name and file: %#v", games[0])
	}
}
