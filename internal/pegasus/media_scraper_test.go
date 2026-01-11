package pegasus

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/game_tool_box/internal/pegasus/metadata"
	"github.com/game_tool_box/internal/pegasus/screenscraper"
)

func TestScrapeSelectedMediaWithClient_DownloadsAndSkips(t *testing.T) {
	png := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00}
	mp4 := []byte{0x00, 0x00, 0x00, 0x18, 'f', 't', 'y', 'p'}

	root := t.TempDir()
	romRel := filepath.ToSlash(filepath.Join("roms", "A.zip"))
	romAbs := filepath.Join(root, filepath.FromSlash(romRel))
	if err := os.MkdirAll(filepath.Dir(romAbs), 0o755); err != nil {
		t.Fatalf("mkdir rom dir: %v", err)
	}
	romContent := []byte("rom-content")
	if err := os.WriteFile(romAbs, romContent, 0o644); err != nil {
		t.Fatalf("write rom: %v", err)
	}
	h := md5.Sum(romContent)
	romMD5 := hex.EncodeToString(h[:])

	var base string
	// Fake ScreenScraper API + media files on same server.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api2/jeuInfos.php":
			q := r.URL.Query()
			if q.Get("romhashmd5") != romMD5 {
				http.Error(w, "bad hash", http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"jeu":{"id":"1","medias":{"media":[`+
				`{"type":"box-2d","url":"%s/media/A/boxFront.png"},`+
				`{"type":"wheel","url":"%s/media/A/logo.png"},`+
				`{"type":"video","url":"%s/media/A/video.mp4"}`+
				`]}}}`, base, base, base)
		case r.URL.Path == "/media/A/boxFront.png":
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(png)
		case r.URL.Path == "/media/A/logo.png":
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(png)
		case r.URL.Path == "/media/A/video.mp4":
			w.Header().Set("Content-Type", "video/mp4")
			_, _ = w.Write(mp4)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	base = srv.URL

	cli := screenscraper.Client{BaseURL: srv.URL + "/api2", DevID: "x", DevPassword: "y", SoftName: "GameToolBox", HTTPClient: &http.Client{Timeout: 2 * time.Second}}

	games := []GameViewModel{{Game: metadata.Game{GameName: "A", FileName: romRel}, Selected: true}}
	opt := MediaScrapeOptions{Overwrite: false, MaxConcurrent: 2, Timeout: 2 * time.Second}

	res, err := scrapeSelectedMediaWithClient(context.Background(), opt, cli, root, games)
	if err != nil {
		t.Fatalf("scrape: %v", err)
	}
	if res.Created != 3 {
		t.Fatalf("Created=%d want 3", res.Created)
	}
	if res.Skipped != 0 {
		t.Fatalf("Skipped=%d want 0", res.Skipped)
	}

	// Run again: should skip all (overwrite=false).
	res2, err := scrapeSelectedMediaWithClient(context.Background(), opt, cli, root, games)
	if err != nil {
		t.Fatalf("scrape2: %v", err)
	}
	if res2.Skipped != 3 {
		t.Fatalf("Skipped=%d want 3", res2.Skipped)
	}

	// Files exist.
	if _, err := os.Stat(filepath.Join(root, "media", "A", "boxFront.png")); err != nil {
		t.Fatalf("boxFront.png missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "media", "A", "logo.png")); err != nil {
		t.Fatalf("logo.png missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "media", "A", "video.mp4")); err != nil {
		t.Fatalf("video.mp4 missing: %v", err)
	}
}

func TestScrapeSelectedMediaWithClient_NotFound(t *testing.T) {
	root := t.TempDir()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api2/jeuInfos.php" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"jeu":null}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()
	cli := screenscraper.Client{BaseURL: srv.URL + "/api2", DevID: "x", DevPassword: "y", SoftName: "GameToolBox", HTTPClient: &http.Client{Timeout: 2 * time.Second}}

	// Need a real ROM file because code computes MD5.
	romRel := filepath.ToSlash(filepath.Join("roms", "A.zip"))
	romAbs := filepath.Join(root, filepath.FromSlash(romRel))
	_ = os.MkdirAll(filepath.Dir(romAbs), 0o755)
	_ = os.WriteFile(romAbs, []byte("rom"), 0o644)

	games := []GameViewModel{{Game: metadata.Game{GameName: "A", FileName: romRel}, Selected: true}}
	_, err := scrapeSelectedMediaWithClient(context.Background(), MediaScrapeOptions{Timeout: 2 * time.Second}, cli, root, games)
	if err == nil {
		t.Fatalf("expected error")
	}
}
