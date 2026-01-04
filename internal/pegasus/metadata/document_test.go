package metadata

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocument_RemoveByGameNames_PreservesOtherContent(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "metadata.pegasus.txt")

	in := strings.Join([]string{
		"# comment",
		"",
		"game: A",
		"file: A.zip",
		"",
		"collection: Favorites",
		"sort-by: 1",
		"",
		"game: B",
		"file: B.zip",
		"",
		"",
	}, "\n")
	if err := os.WriteFile(path, []byte(in), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	doc, err := ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	_ = doc.Games() // smoke: ensure Games() is callable

	names := map[string]struct{}{"A": {}}
	if got := doc.RemoveByGameNames(names); got != 1 {
		t.Fatalf("expected removed=1, got %d", got)
	}
	if err := doc.WriteFileAtomic(path); err != nil {
		t.Fatalf("WriteFileAtomic: %v", err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	out := string(b)
	if strings.Contains(out, "game: A") {
		t.Fatalf("expected A removed, got:\n%s", out)
	}
	if !strings.Contains(out, "collection: Favorites") {
		t.Fatalf("expected collection preserved, got:\n%s", out)
	}
	if !strings.Contains(out, "game: B") {
		t.Fatalf("expected B preserved, got:\n%s", out)
	}
	// light normalization collapses multiple blank lines
	if strings.Contains(out, "\n\n\n") {
		t.Fatalf("expected no triple blank lines, got:\n%s", out)
	}
}
