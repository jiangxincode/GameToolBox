package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDir_UsesDotGameToolBoxUnderHomeOverride(t *testing.T) {
	tmp := t.TempDir()
	if err := os.Setenv("GAMETOOLBOX_HOME", tmp); err != nil {
		t.Fatalf("Setenv error: %v", err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("GAMETOOLBOX_HOME") })

	dir, err := configDir()
	if err != nil {
		t.Fatalf("configDir() error: %v", err)
	}

	expected := filepath.Join(tmp, ".gametoolbox")
	if dir != expected {
		t.Fatalf("configDir()=%q want %q", dir, expected)
	}

	p, err := configPath()
	if err != nil {
		t.Fatalf("configPath() error: %v", err)
	}
	expectedP := filepath.Join(expected, "config.json")
	if p != expectedP {
		t.Fatalf("configPath()=%q want %q", p, expectedP)
	}
}

func TestSaveLoad_RoundTrip(t *testing.T) {
	tmp := t.TempDir()
	if err := os.Setenv("GAMETOOLBOX_HOME", tmp); err != nil {
		t.Fatalf("Setenv error: %v", err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("GAMETOOLBOX_HOME") })

	in := Config{Lang: "en", Theme: ""}
	if err := Save(in); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	out, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if out.Lang != in.Lang {
		t.Fatalf("Lang=%q want %q", out.Lang, in.Lang)
	}
}
