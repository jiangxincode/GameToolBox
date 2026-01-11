package pegasus

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestSafeJoinUnderRoot_AllowsNormalRelative(t *testing.T) {
	root := t.TempDir()
	p, err := SafeJoinUnderRoot(root, "roms/a.zip")
	if err != nil {
		t.Fatalf("SafeJoinUnderRoot: %v", err)
	}
	want := filepath.Clean(filepath.Join(root, "roms", "a.zip"))
	if filepath.Clean(p) != want {
		t.Fatalf("path mismatch: got=%q want=%q", p, want)
	}
}

func TestSafeJoinUnderRoot_RejectsTraversal(t *testing.T) {
	root := t.TempDir()
	if _, err := SafeJoinUnderRoot(root, "../a.zip"); err == nil {
		t.Fatalf("expected error for traversal")
	}
	if _, err := SafeJoinUnderRoot(root, "roms/../../a.zip"); err == nil {
		t.Fatalf("expected error for traversal")
	}
}

func TestSafeJoinUnderRoot_RejectsAbs(t *testing.T) {
	root := t.TempDir()

	// Platform-specific absolute paths.
	if runtime.GOOS == "windows" {
		if _, err := SafeJoinUnderRoot(root, `C:\a.zip`); err == nil {
			t.Fatalf("expected error for absolute path")
		}
		if _, err := SafeJoinUnderRoot(root, `\\server\\share\\a.zip`); err == nil {
			t.Fatalf("expected error for UNC absolute path")
		}
	} else {
		if _, err := SafeJoinUnderRoot(root, "/a.zip"); err == nil {
			t.Fatalf("expected error for absolute path")
		}
	}
}

func TestSafeMediaDir_RejectsSeparatorsAndTraversal(t *testing.T) {
	root := t.TempDir()

	bad := []string{
		"a/b",
		"a\\b",
		"..",
		".",
		"a..b",
		"a../b",
	}
	for _, name := range bad {
		if _, err := SafeMediaDir(root, name); err == nil {
			t.Fatalf("expected error for gameName=%q", name)
		}
	}
}

func TestSafeMediaDir_AllowsNormalName(t *testing.T) {
	root := t.TempDir()
	p, err := SafeMediaDir(root, "My Game")
	if err != nil {
		t.Fatalf("SafeMediaDir: %v", err)
	}
	want := filepath.Clean(filepath.Join(root, "media", "My Game"))
	if filepath.Clean(p) != want {
		t.Fatalf("path mismatch: got=%q want=%q", p, want)
	}
}
