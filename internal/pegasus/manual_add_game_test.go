package pegasus

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManualAddGame_CopiesFilesAndUpsertsMetadata(t *testing.T) {
	root := t.TempDir()
	srcDir := t.TempDir()

	srcRom := filepath.Join(srcDir, "MyGame.zip")
	if err := os.WriteFile(srcRom, []byte("rom"), 0o644); err != nil {
		t.Fatalf("write rom: %v", err)
	}
	cover := filepath.Join(srcDir, "cover.png")
	if err := os.WriteFile(cover, []byte("cover"), 0o644); err != nil {
		t.Fatalf("write cover: %v", err)
	}
	logo := filepath.Join(srcDir, "logo.png")
	if err := os.WriteFile(logo, []byte("logo"), 0o644); err != nil {
		t.Fatalf("write logo: %v", err)
	}
	video := filepath.Join(srcDir, "video.mp4")
	if err := os.WriteFile(video, []byte("video"), 0o644); err != nil {
		t.Fatalf("write video: %v", err)
	}

	res, err := ManualAddGame(root, ManualAddGameRequest{
		GameName:      "My Game",
		Developer:     "My Dev",
		Description:   "My Desc",
		SourceRomPath: srcRom,
		BoxFrontPath:  cover,
		LogoPath:      logo,
		VideoPath:     video,
		Overwrite:     true,
	})
	if err != nil {
		t.Fatalf("ManualAddGame: %v", err)
	}
	if res.RomRelPath != "MyGame.zip" {
		t.Fatalf("unexpected RomRelPath: %s", res.RomRelPath)
	}

	if _, err := os.Stat(filepath.Join(root, "MyGame.zip")); err != nil {
		t.Fatalf("expected rom copied: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "media", "My Game", "boxFront.png")); err != nil {
		t.Fatalf("expected cover copied: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "media", "My Game", "logo.png")); err != nil {
		t.Fatalf("expected logo copied: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "media", "My Game", "video.mp4")); err != nil {
		t.Fatalf("expected video copied: %v", err)
	}

	b, err := os.ReadFile(filepath.Join(root, "metadata.pegasus.txt"))
	if err != nil {
		t.Fatalf("read metadata: %v", err)
	}
	out := string(b)
	if !strings.Contains(out, "game:My Game") || !strings.Contains(out, "file:MyGame.zip") {
		t.Fatalf("unexpected metadata:\n%s", out)
	}
	if !strings.Contains(out, "developer:My Dev") {
		t.Fatalf("expected developer in metadata, got:\n%s", out)
	}
	if !strings.Contains(out, "description:My Desc") {
		t.Fatalf("expected description in metadata, got:\n%s", out)
	}
}
