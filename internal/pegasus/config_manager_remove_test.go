package pegasus

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRemoveGamesFromMetadata_PreservesCollectionAndNormalizesBlankLines(t *testing.T) {
	root := t.TempDir()
	metaPath := filepath.Join(root, "metadata.pegasus.txt")

	// Intentionally include extra blank lines; after removal we expect them to be collapsed.
	meta := strings.Join([]string{
		"game:A",
		"file:roms/A.zip",
		"",
		"",
		"collection:Favorites",
		"sort-by:999",
		"",
		"game:B",
		"file:roms/B.zip",
		"",
		"",
	}, "\n")
	if err := os.WriteFile(metaPath, []byte(meta), 0o644); err != nil {
		t.Fatalf("write metadata: %v", err)
	}

	removed, err := RemoveGamesFromMetadata(root, []string{"roms/A.zip"})
	if err != nil {
		t.Fatalf("RemoveGamesFromMetadata: %v", err)
	}
	if removed != 1 {
		t.Fatalf("expected removed=1 got %d", removed)
	}

	b, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("read metadata: %v", err)
	}
	out := string(b)

	if strings.Contains(out, "file:roms/A.zip") {
		t.Fatalf("expected A removed, got:\n%s", out)
	}
	if !strings.Contains(out, "collection:Favorites") {
		t.Fatalf("expected collection preserved, got:\n%s", out)
	}
	if !strings.Contains(out, "file:roms/B.zip") {
		t.Fatalf("expected B preserved, got:\n%s", out)
	}
	if strings.Contains(out, "\n\n\n") {
		t.Fatalf("expected blank lines normalized (no triple newline), got:\n%s", out)
	}
}
