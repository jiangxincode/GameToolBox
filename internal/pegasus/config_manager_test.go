package pegasus

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/game_tool_box/internal/pegasus/metadata"
)

func TestLoadGamesFromRomFiles_SkipsMetadataAndMedia(t *testing.T) {
	root := t.TempDir()

	if err := os.WriteFile(filepath.Join(root, "metadata.pegasus.txt"), []byte("game: X\nfile: x.zip\n"), 0o644); err != nil {
		t.Fatalf("write metadata: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "media", "X"), 0o755); err != nil {
		t.Fatalf("mkdir media: %v", err)
	}
	// If media is not skipped, these would be returned — make sure they are ignored.
	if err := os.WriteFile(filepath.Join(root, "media", "X", "boxFront.png"), []byte("img"), 0o644); err != nil {
		t.Fatalf("write media file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "media", "media_should_not_be_scanned.zip"), []byte("no"), 0o644); err != nil {
		t.Fatalf("write media rom-like file: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(root, "roms"), 0o755); err != nil {
		t.Fatalf("mkdir roms: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "roms", "a.zip"), []byte("a"), 0o644); err != nil {
		t.Fatalf("write rom: %v", err)
	}

	games, err := LoadGamesFromRomFiles(root)
	if err != nil {
		t.Fatalf("LoadGamesFromRomFiles: %v", err)
	}
	if len(games) != 1 {
		t.Fatalf("expected 1 rom file, got %d (%v)", len(games), games)
	}
	if games[0].FileName != "roms/a.zip" {
		t.Fatalf("expected relative filename roms/a.zip, got %q", games[0].FileName)
	}
}

func TestDiffConfigAgainstRomFiles_MissingExtraDuplicate(t *testing.T) {
	root := t.TempDir()

	// Create ROM files: A.zip and B.zip
	if err := os.MkdirAll(filepath.Join(root, "roms"), 0o755); err != nil {
		t.Fatalf("mkdir roms: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "roms", "A.zip"), []byte("a"), 0o644); err != nil {
		t.Fatalf("write A.zip: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "roms", "B.zip"), []byte("b"), 0o644); err != nil {
		t.Fatalf("write B.zip: %v", err)
	}

	// metadata contains:
	// - A.zip twice (duplicate)
	// - C.zip (extra, missing on disk)
	meta := strings.Join([]string{
		"game: A",
		"file: roms/A.zip",
		"",
		"game: A2",
		"file: roms/A.zip",
		"",
		"game: C",
		"file: roms/C.zip",
		"",
	}, "\n")
	if err := os.WriteFile(filepath.Join(root, "metadata.pegasus.txt"), []byte(meta), 0o644); err != nil {
		t.Fatalf("write metadata: %v", err)
	}

	diff, err := DiffConfigAgainstRomFiles(root)
	if err != nil {
		t.Fatalf("DiffConfigAgainstRomFiles: %v", err)
	}

	// Missing: B.zip
	if len(diff.MissingInConfig) != 1 {
		t.Fatalf("expected MissingInConfig=1, got %d (%v)", len(diff.MissingInConfig), diff.MissingInConfig)
	}
	if diff.MissingInConfig[0].FileName != "roms/B.zip" {
		t.Fatalf("expected missing roms/B.zip, got %q", diff.MissingInConfig[0].FileName)
	}

	// Extra: C.zip
	if len(diff.ExtraInConfig) != 1 {
		t.Fatalf("expected ExtraInConfig=1, got %d (%v)", len(diff.ExtraInConfig), diff.ExtraInConfig)
	}
	if metadata.NormalizeFileKey(diff.ExtraInConfig[0].FileName) != "roms/c.zip" {
		t.Fatalf("expected extra roms/C.zip, got %q", diff.ExtraInConfig[0].FileName)
	}

	// Duplicate: roms/a.zip
	foundDup := false
	for _, d := range diff.DuplicateInConfig {
		if d == "roms/a.zip" {
			foundDup = true
			break
		}
	}
	if !foundDup {
		t.Fatalf("expected duplicate contains roms/a.zip, got %v", diff.DuplicateInConfig)
	}
}

func TestAppendMissingGamesToMetadata_AndRemoveGamesFromMetadata(t *testing.T) {
	root := t.TempDir()

	// Start with empty metadata and one ROM.
	if err := os.MkdirAll(filepath.Join(root, "roms"), 0o755); err != nil {
		t.Fatalf("mkdir roms: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "roms", "A.zip"), []byte("a"), 0o644); err != nil {
		t.Fatalf("write A.zip: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "metadata.pegasus.txt"), []byte(""), 0o644); err != nil {
		t.Fatalf("write metadata: %v", err)
	}

	// Append missing
	missing := []GameModel{{GameName: "A", FileName: "roms/A.zip"}}
	if err := AppendMissingGamesToMetadata(root, missing); err != nil {
		t.Fatalf("AppendMissingGamesToMetadata: %v", err)
	}

	b, err := os.ReadFile(filepath.Join(root, "metadata.pegasus.txt"))
	if err != nil {
		t.Fatalf("read metadata: %v", err)
	}
	out := string(b)
	if !strings.Contains(out, "game:A") || !strings.Contains(out, "file:roms/A.zip") {
		t.Fatalf("expected metadata contains appended block, got:\n%s", out)
	}
	if !strings.Contains(out, "developer:A") || !strings.Contains(out, "description:A") {
		t.Fatalf("expected developer/description equal game name, got:\n%s", out)
	}

	// Remove it back
	removed, err := RemoveGamesFromMetadata(root, []string{"roms/A.zip"})
	if err != nil {
		t.Fatalf("RemoveGamesFromMetadata: %v", err)
	}
	if removed != 1 {
		t.Fatalf("expected removed=1 got %d", removed)
	}

	b2, err := os.ReadFile(filepath.Join(root, "metadata.pegasus.txt"))
	if err != nil {
		t.Fatalf("read metadata: %v", err)
	}
	out2 := string(b2)
	if strings.Contains(out2, "file:roms/A.zip") {
		t.Fatalf("expected record removed, got:\n%s", out2)
	}
}

func TestAppendMissingGamesToMetadata_SortByIncrementsAndKeepsDeveloperDescription(t *testing.T) {
	root := t.TempDir()

	meta := strings.Join([]string{
		"game:A",
		"file:roms/A.zip",
		"sort-by:010",
		"developer:DevA",
		"description:DescA",
		"",
		"collection:Favorites",
		"sort-by:999",
		"",
	}, "\n")
	if err := os.MkdirAll(filepath.Join(root, "roms"), 0o755); err != nil {
		t.Fatalf("mkdir roms: %v", err)
	}
	_ = os.WriteFile(filepath.Join(root, "roms", "A.zip"), []byte("a"), 0o644)
	_ = os.WriteFile(filepath.Join(root, "roms", "Z.zip"), []byte("z"), 0o644)
	_ = os.WriteFile(filepath.Join(root, "roms", "B.zip"), []byte("b"), 0o644)

	if err := os.WriteFile(filepath.Join(root, "metadata.pegasus.txt"), []byte(meta), 0o644); err != nil {
		t.Fatalf("write metadata: %v", err)
	}

	missing := []GameModel{{GameName: "B", FileName: "roms/B.zip"}}
	if err := AppendMissingGamesToMetadata(root, missing); err != nil {
		t.Fatalf("AppendMissingGamesToMetadata: %v", err)
	}

	b, err := os.ReadFile(filepath.Join(root, "metadata.pegasus.txt"))
	if err != nil {
		t.Fatalf("read metadata: %v", err)
	}
	out := string(b)

	if !strings.Contains(out, "file:roms/B.zip") {
		t.Fatalf("expected metadata contains B record, got:\n%s", out)
	}
	// With max sort-by 010 (width=3), next should be 011.
	if !strings.Contains(out, "sort-by:011") {
		t.Fatalf("expected sort-by to be zero-padded (011), got:\n%s", out)
	}

	bBlockIdx := strings.LastIndex(out, "file:roms/B.zip")
	if bBlockIdx < 0 {
		t.Fatalf("cannot locate B block")
	}
	bTail := out[bBlockIdx:]
	if !strings.Contains(bTail, "developer:B") {
		t.Fatalf("expected developer equals game name, got:\n%s", bTail)
	}
	if !strings.Contains(bTail, "description:B") {
		t.Fatalf("expected description equals game name, got:\n%s", bTail)
	}
}

func TestAppendMissingGamesToMetadata_BlankLineFormatting(t *testing.T) {
	root := t.TempDir()
	metaPath := filepath.Join(root, "metadata.pegasus.txt")

	// Existing file ends WITHOUT a blank line (only one trailing \n)
	initial := strings.Join([]string{
		"game:A",
		"file:roms/A.zip",
		"sort-by:001",
		"developer:A",
		"description:A",
		"", // <== blank line between games
		"game:B",
		"file:roms/B.zip",
		"sort-by:002",
		"developer:B",
		"description:B",
	}, "\n") + "\n" // only one newline at end => NOT a blank line

	if err := os.MkdirAll(filepath.Join(root, "roms"), 0o755); err != nil {
		t.Fatalf("mkdir roms: %v", err)
	}
	_ = os.WriteFile(filepath.Join(root, "roms", "C.zip"), []byte("c"), 0o644)

	if err := os.WriteFile(metaPath, []byte(initial), 0o644); err != nil {
		t.Fatalf("write metadata: %v", err)
	}

	missing := []GameModel{
		{GameName: "C", FileName: "roms/C.zip"},
		{GameName: "D", FileName: "roms/D.zip"},
	}
	if err := AppendMissingGamesToMetadata(root, missing); err != nil {
		t.Fatalf("AppendMissingGamesToMetadata: %v", err)
	}

	b, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("read metadata: %v", err)
	}
	out := string(b)

	// Must end with exactly one blank line => ends with \n\n (or \r\n\r\n). We write with \n.
	if !strings.HasSuffix(out, "\n\n") {
		t.Fatalf("expected file ends with a blank line (\\n\\n), got tail=%q", tail(out, 20))
	}

	// Ensure between config blocks there is exactly one blank line.
	// We'll check the appended C and D blocks are separated by exactly one blank line.
	if !strings.Contains(out, "description:C\n\ngame:D\n") {
		t.Fatalf("expected exactly one blank line between appended games, got:\n%s", out)
	}

	// Ensure exactly one blank line between existing last game (B) and first appended (C)
	if !strings.Contains(out, "description:B\n\ngame:C\n") {
		t.Fatalf("expected exactly one blank line between existing and appended games, got:\n%s", out)
	}
}

func tail(s string, n int) string {
	if n <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[len(r)-n:])
}

func TestGenerateSelectedConfig_RewritesMetadataWithSelectedGames(t *testing.T) {
	root := t.TempDir()

	games := []GameModel{
		{Selected: true, GameName: "A", FileName: "roms/A.zip"},
		{Selected: false, GameName: "B", FileName: "roms/B.zip"},
		{Selected: true, GameName: "C", FileName: "roms/C.zip", Developer: "DevC", Description: "DescC"},
	}

	res := GenerateSelectedConfig(root, games)
	if len(res.Errors) > 0 {
		t.Fatalf("expected no errors, got %v", res.Errors[0])
	}
	if res.Written != 2 {
		t.Fatalf("expected Written=2 got %d (res=%+v)", res.Written, res)
	}

	b, err := os.ReadFile(filepath.Join(root, "metadata.pegasus.txt"))
	if err != nil {
		t.Fatalf("read metadata: %v", err)
	}
	out := string(b)

	// Should contain A and C, and not contain B.
	if !strings.Contains(out, "game:A") || !strings.Contains(out, "file:roms/A.zip") {
		t.Fatalf("expected A block, got:\n%s", out)
	}
	if strings.Contains(out, "game:B") || strings.Contains(out, "file:roms/B.zip") {
		t.Fatalf("expected B not written, got:\n%s", out)
	}
	if !strings.Contains(out, "game:C") || !strings.Contains(out, "file:roms/C.zip") {
		t.Fatalf("expected C block, got:\n%s", out)
	}

	// Default dev/desc for A should be game name.
	if !strings.Contains(out, "developer:A") || !strings.Contains(out, "description:A") {
		t.Fatalf("expected developer/description defaults for A, got:\n%s", out)
	}
	// Respect provided dev/desc for C.
	if !strings.Contains(out, "developer:DevC") || !strings.Contains(out, "description:DescC") {
		t.Fatalf("expected developer/description for C, got:\n%s", out)
	}

	// sort-by should be zero-padded.
	if !strings.Contains(out, "sort-by:001") || !strings.Contains(out, "sort-by:002") {
		t.Fatalf("expected sort-by entries 001/002, got:\n%s", out)
	}
}
