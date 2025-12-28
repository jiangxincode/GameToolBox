package pegasus

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateSelectedFilesForOssHandheld_RecreatesImagesAndCopiesBoxFront(t *testing.T) {
	root := t.TempDir()

	// Pre-create images dir with a junk file to ensure it's removed.
	if err := os.MkdirAll(filepath.Join(root, "images"), 0o755); err != nil {
		t.Fatalf("mkdir images: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "images", "junk.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write junk: %v", err)
	}

	// Create media/A/boxFront.png
	if err := os.MkdirAll(filepath.Join(root, "media", "A"), 0o755); err != nil {
		t.Fatalf("mkdir media/A: %v", err)
	}
	src := filepath.Join(root, "media", "A", "boxFront.png")
	if err := os.WriteFile(src, []byte("pngdata"), 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	games := []GameModel{{Selected: true, GameName: "A"}}
	res := GenerateSelectedFilesForOssHandheld(root, games)
	if len(res.Errors) != 0 {
		t.Fatalf("unexpected errors: %v", res.Errors)
	}
	if res.Created != 1 {
		t.Fatalf("Created mismatch: got %d", res.Created)
	}

	// junk should be removed
	if _, err := os.Stat(filepath.Join(root, "images", "junk.txt")); !os.IsNotExist(err) {
		t.Fatalf("expected junk removed, err=%v", err)
	}

	// A.png should exist
	dst := filepath.Join(root, "images", "A.png")
	b, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(b) != "pngdata" {
		t.Fatalf("dst content mismatch: %q", string(b))
	}
}

func TestGenerateSelectedFilesForOssHandheld_MissingSourceCountsFailed(t *testing.T) {
	root := t.TempDir()
	games := []GameModel{{Selected: true, GameName: "A"}}
	res := GenerateSelectedFilesForOssHandheld(root, games)
	if res.Created != 0 {
		t.Fatalf("expected Created=0, got %d", res.Created)
	}
	if res.Failed == 0 {
		t.Fatalf("expected Failed>0")
	}
	if len(res.Errors) == 0 {
		t.Fatalf("expected error")
	}
}
