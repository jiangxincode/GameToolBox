package pegasus

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReorderGamesInMetadata_ReordersAndRenumbersSortBy(t *testing.T) {
	root := t.TempDir()
	meta := filepath.Join(root, "metadata.pegasus.txt")
	in := strings.Join([]string{
		"game:A",
		"file:A.zip",
		"sort-by:003",
		"developer:A",
		"description:A",
		"",
		"game:B",
		"file:B.zip",
		"sort-by:001",
		"developer:B",
		"description:B",
		"",
		"game:C",
		"file:C.zip",
		"sort-by:002",
		"developer:C",
		"description:C",
		"",
	}, "\n")
	if err := os.WriteFile(meta, []byte(in), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	_, err := ReorderGamesInMetadata(root, []string{"B", "C", "A"})
	if err != nil {
		t.Fatalf("ReorderGamesInMetadata: %v", err)
	}

	b, err := os.ReadFile(meta)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	out := string(b)
	// Ensure order.
	idxB := strings.Index(out, "game:B")
	idxC := strings.Index(out, "game:C")
	idxA := strings.Index(out, "game:A")
	if !(idxB >= 0 && idxC > idxB && idxA > idxC) {
		t.Fatalf("unexpected order:\n%s", out)
	}
	// Ensure renumbering.
	if !strings.Contains(out, "game:B\nfile:B.zip\nsort-by:001") {
		t.Fatalf("expected B sort-by 001, got:\n%s", out)
	}
}

func TestReorderGamesInMetadata_PreservesCollectionBlock(t *testing.T) {
	root := t.TempDir()
	meta := filepath.Join(root, "metadata.pegasus.txt")

	in := strings.Join([]string{
		"collection: PSVITA",
		"sort-by: 103",
		"shortname:aloys_psv",
		"launch: X",
		"",
		"game:A",
		"file:A.zip",
		"sort-by:003",
		"developer:A",
		"description:A",
		"",
		"game:B",
		"file:B.zip",
		"sort-by:001",
		"developer:B",
		"description:B",
		"",
	}, "\n")
	if err := os.WriteFile(meta, []byte(in), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	_, err := ReorderGamesInMetadata(root, []string{"B", "A"})
	if err != nil {
		t.Fatalf("ReorderGamesInMetadata: %v", err)
	}

	b, err := os.ReadFile(meta)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	out := string(b)

	// Collection block should remain at the top with same content.
	wantPrefix := strings.Join([]string{
		"collection: PSVITA",
		"sort-by: 103",
		"shortname:aloys_psv",
		"launch: X",
		"",
	}, "\n")
	if !strings.HasPrefix(out, wantPrefix) {
		t.Fatalf("collection block changed/unexpected prefix:\n%s", out)
	}

	// Ensure game order was updated.
	idxB := strings.Index(out, "game:B")
	idxA := strings.Index(out, "game:A")
	if !(idxB >= 0 && idxA > idxB) {
		t.Fatalf("unexpected game order:\n%s", out)
	}
}

// (order list now uses game titles rather than file names)
