package metadata

import (
	"strings"
	"testing"
)

func TestDocument_UpsertGameByFile_AppendWhenMissing(t *testing.T) {
	doc := Parse("")
	changed, found, err := doc.UpsertGameByFile(Game{GameName: "Test", FileName: "roms/Test.zip"}, UpsertOptions{})
	if err != nil {
		t.Fatalf("UpsertGameByFile: %v", err)
	}
	if !changed {
		t.Fatalf("expected changed=true")
	}
	if found {
		t.Fatalf("expected found=false")
	}
	out := doc.Render()
	if !strings.Contains(out, "game:Test") || !strings.Contains(out, "file:roms/Test.zip") {
		t.Fatalf("unexpected render:\n%s", out)
	}
}

func TestDocument_UpsertGameByFile_UpdateWhenExists_PreserveUnknownLines(t *testing.T) {
	in := strings.Join([]string{
		"game: Old Title",
		"file: ROMS/TEST.ZIP",
		"custom-key: keepme",
		"description: old",
		"",
	}, "\n")
	doc := Parse(in)
	changed, found, err := doc.UpsertGameByFile(Game{GameName: "New Title", FileName: "roms/test.zip", Developer: "Dev"}, UpsertOptions{})
	if err != nil {
		t.Fatalf("UpsertGameByFile: %v", err)
	}
	if !changed {
		t.Fatalf("expected changed=true")
	}
	if !found {
		t.Fatalf("expected found=true")
	}
	out := doc.Render()
	if !strings.Contains(out, "game:New Title") {
		t.Fatalf("expected title updated, got:\n%s", out)
	}
	if !strings.Contains(out, "custom-key: keepme") {
		t.Fatalf("expected unknown line preserved, got:\n%s", out)
	}
	if !strings.Contains(out, "developer:Dev") {
		t.Fatalf("expected developer set, got:\n%s", out)
	}
}
