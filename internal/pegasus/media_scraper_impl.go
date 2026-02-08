package pegasus

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/game_tool_box/internal/config"
	"github.com/game_tool_box/internal/pegasus/metadata"
	"github.com/game_tool_box/internal/pegasus/screenscraper"
)

var (
	ErrMediaScrapeNotConfigured = errors.New("screenscraper credentials not configured")
	ErrMediaScrapeNotFound      = errors.New("screenscraper game not found by hash")
)

type MediaScrapeOptions struct {
	Overwrite bool
	// MaxConcurrent controls parallel downloads.
	MaxConcurrent int
	// Timeout is the per-request timeout.
	Timeout time.Duration
}

// ScrapeSelectedMedia scrapes/downloads boxFront/logo/video for selected games into <rootDir>/media/<GameName>/
// by matching ScreenScraper entries using the ROM file hash (MD5).
func ScrapeSelectedMedia(rootDir string, games []GameViewModel) (MediaScrapeResult, error) {
	var res MediaScrapeResult
	rootDir = strings.TrimSpace(rootDir)
	if rootDir == "" {
		return res, errors.New("rootDir is empty")
	}

	c, _ := config.Load()
	devID := strings.TrimSpace(c.ScreenScraperDevID)
	devPass := strings.TrimSpace(c.ScreenScraperDevPassword)
	if devID == "" || devPass == "" {
		return res, ErrMediaScrapeNotConfigured
	}

	opt := MediaScrapeOptions{
		Overwrite:     c.MediaScrapeOverwrite,
		MaxConcurrent: 4,
		Timeout:       30 * time.Second,
	}

	cli := screenscraper.Client{
		DevID:       devID,
		DevPassword: devPass,
		User:        strings.TrimSpace(c.ScreenScraperUser),
		Password:    strings.TrimSpace(c.ScreenScraperPassword),
		SoftName:    "GameToolBox",
		HTTPClient:  &http.Client{Timeout: opt.Timeout},
	}

	return scrapeSelectedMediaWithClient(context.Background(), opt, cli, rootDir, games)
}

func scrapeSelectedMediaWithClient(ctx context.Context, opt MediaScrapeOptions, cli screenscraper.Client, rootDir string, games []GameViewModel) (MediaScrapeResult, error) {
	var res MediaScrapeResult
	if opt.MaxConcurrent <= 0 {
		opt.MaxConcurrent = 4
	}
	if opt.Timeout <= 0 {
		opt.Timeout = 30 * time.Second
	}
	client := &http.Client{Timeout: opt.Timeout}

	type job struct {
		gameName string
		url      string
		dst      string
		maxBytes int64
	}

	jobs := make([]job, 0)
	// Build jobs by querying ScreenScraper per selected game.
	for _, g := range games {
		if !g.Selected {
			continue
		}
		name := strings.TrimSpace(g.GameName)
		if name == "" {
			res.Failed++
			res.Errors = append(res.Errors, errors.New("gameName is empty"))
			continue
		}
		romRel := strings.TrimSpace(g.FileName)
		if romRel == "" {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Errorf("%s rom path is empty", name))
			continue
		}
		romAbs, err := SafeJoinUnderRoot(rootDir, romRel)
		if err != nil {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Errorf("%s refuse to access rom %q: %w", name, romRel, err))
			continue
		}

		hash, err := ComputeROMMD5(romAbs)
		if err != nil {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Errorf("%s compute md5: %w", name, err))
			continue
		}

		info, err := cli.GetGameByRomHash(ctx, "md5", hash)
		if err != nil {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Errorf("%s screenscraper query: %w", name, err))
			continue
		}
		if info == nil || info.Jeu == nil {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Errorf("%s: %w", name, ErrMediaScrapeNotFound))
			continue
		}

		// Pick URLs.
		var boxURL, logoURL, videoURL string
		for _, m := range info.Jeu.Medias.Media {
			t := strings.ToLower(strings.TrimSpace(m.Type))
			src := strings.TrimSpace(m.URL)
			if src == "" {
				continue
			}
			switch t {
			case "box-2d", "box2d", "box-2d-us", "box-2d-eu":
				if boxURL == "" {
					boxURL = src
				}
			case "wheel", "logo":
				if logoURL == "" {
					logoURL = src
				}
			case "video":
				if videoURL == "" {
					videoURL = src
				}
			}
		}

		mediaDir, err := SafeMediaDir(rootDir, name)
		if err != nil {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Errorf("%s invalid media dir: %w", name, err))
			continue
		}
		if err := os.MkdirAll(mediaDir, 0o755); err != nil {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Errorf("mkdir %s: %w", mediaDir, err))
			continue
		}

		if boxURL != "" {
			jobs = append(jobs, job{gameName: name, url: boxURL, dst: filepath.Join(mediaDir, "boxFront.png"), maxBytes: 20 << 20})
		}
		if logoURL != "" {
			jobs = append(jobs, job{gameName: name, url: logoURL, dst: filepath.Join(mediaDir, "logo.png"), maxBytes: 20 << 20})
		}
		if videoURL != "" {
			jobs = append(jobs, job{gameName: name, url: videoURL, dst: filepath.Join(mediaDir, "video.mp4"), maxBytes: 500 << 20})
		}
	}

	sem := make(chan struct{}, opt.MaxConcurrent)
	var wg sync.WaitGroup
	var mu sync.Mutex

	recordErr := func(e error) {
		mu.Lock()
		defer mu.Unlock()
		res.Errors = append(res.Errors, e)
		res.Failed++
	}

	incCreated := func() {
		mu.Lock()
		res.Created++
		mu.Unlock()
	}
	incSkipped := func() {
		mu.Lock()
		res.Skipped++
		mu.Unlock()
	}

	for _, j := range jobs {
		j := j
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				recordErr(ctx.Err())
				return
			}

			if !opt.Overwrite {
				if st, err := os.Stat(j.dst); err == nil && !st.IsDir() {
					incSkipped()
					return
				}
			}

			if err := downloadToFile(ctx, client, j.url, j.dst, j.maxBytes); err != nil {
				recordErr(fmt.Errorf("%s download %s: %w", j.gameName, j.url, err))
				return
			}
			incCreated()
		}()
	}

	wg.Wait()
	if len(res.Errors) > 0 {
		return res, res.Errors[0]
	}
	return res, nil
}

func downloadToFile(ctx context.Context, client *http.Client, u string, dst string, maxBytes int64) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http status %s", resp.Status)
	}

	dir := filepath.Dir(dst)
	base := filepath.Base(dst)
	h := sha256.Sum256([]byte(u))
	tmp := filepath.Join(dir, fmt.Sprintf(".%s.%x.tmp", base, h[:6]))

	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(tmp)
	}()

	var r io.Reader = resp.Body
	if maxBytes > 0 {
		r = io.LimitReader(resp.Body, maxBytes)
	}
	if _, err := io.Copy(f, r); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	return metadata.ReplaceFileAtomic(dst, tmp)
}
