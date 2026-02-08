package pegasus

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/game_tool_box/internal/pegasus/metadata"
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

	games := []GameViewModel{{Selected: true, Game: metadata.Game{GameName: "A"}}}
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
	games := []GameViewModel{{Selected: true, Game: metadata.Game{GameName: "A"}}}
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

func TestGenerateEmptyMediaFolders_CreatesDirsAndSkipsExisting(t *testing.T) {
	root := t.TempDir()

	// Pre-create media/A to ensure it is skipped.
	if err := os.MkdirAll(filepath.Join(root, "media", "A"), 0o755); err != nil {
		t.Fatalf("mkdir media/A: %v", err)
	}

	games := []GameViewModel{
		{Selected: true, Game: metadata.Game{GameName: "A"}},
		{Selected: true, Game: metadata.Game{GameName: "B"}},
		{Selected: false, Game: metadata.Game{GameName: "C"}},
	}
	res := GenerateEmptyMediaFolders(root, games)
	if len(res.Errors) != 0 {
		t.Fatalf("unexpected errors: %v", res.Errors)
	}
	if res.Created != 1 {
		t.Fatalf("expected Created=1, got %d", res.Created)
	}
	if res.Skipped != 1 {
		t.Fatalf("expected Skipped=1, got %d", res.Skipped)
	}
	if res.Failed != 0 {
		t.Fatalf("expected Failed=0, got %d", res.Failed)
	}

	if st, err := os.Stat(filepath.Join(root, "media", "B")); err != nil || !st.IsDir() {
		t.Fatalf("expected media/B dir exists, err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "media", "C")); !os.IsNotExist(err) {
		t.Fatalf("expected media/C not created")
	}
}

func TestGenerateEmptyMediaFolders_EmptyRootFails(t *testing.T) {
	res := GenerateEmptyMediaFolders("", []GameViewModel{{Selected: true, Game: metadata.Game{GameName: "A"}}})
	if res.Failed == 0 {
		t.Fatalf("expected Failed>0")
	}
	if len(res.Errors) == 0 {
		t.Fatalf("expected error")
	}
}
