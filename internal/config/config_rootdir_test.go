package config

import (
	"os"
	"testing"
)

func TestSaveLoad_RootDirPersisted(t *testing.T) {
	tmp := t.TempDir()
	if err := os.Setenv("GAMETOOLBOX_HOME", tmp); err != nil {
		t.Fatalf("Setenv error: %v", err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("GAMETOOLBOX_HOME") })

	in := Config{Lang: "en", RootDir: "C:/demo/root"}
	if err := Save(in); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	out, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if out.RootDir != in.RootDir {
		t.Fatalf("RootDir=%q want %q", out.RootDir, in.RootDir)
	}
}
