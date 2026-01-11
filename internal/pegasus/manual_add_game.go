package pegasus

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/game_tool_box/internal/pegasus/metadata"
)

type ManualAddGameRequest struct {
	// GameName is the display name in metadata. If empty, inferred from ROM file name.
	GameName string

	// Developer and Description are optional metadata fields.
	Developer   string
	Description string

	// SourceRomPath is an existing file selected by user.
	SourceRomPath string

	// Optional media files.
	BoxFrontPath string
	LogoPath     string
	VideoPath    string

	// If true, overwrite existing ROM/media destination files.
	Overwrite bool
}

type ManualAddGameResult struct {
	RomRelPath   string
	MediaDirName string

	MetadataUpdated bool
	RomCopied       bool
	MediaCopied     int
}

// ManualAddGame copies ROM/media into the Pegasus root directory and upserts a
// corresponding game block into metadata.pegasus.txt.
func ManualAddGame(rootDir string, req ManualAddGameRequest) (ManualAddGameResult, error) {
	var res ManualAddGameResult

	rootDir = strings.TrimSpace(rootDir)
	if rootDir == "" {
		return res, fmt.Errorf("root dir is empty")
	}

	if strings.TrimSpace(req.SourceRomPath) == "" {
		return res, fmt.Errorf("source rom path is empty")
	}
	if st, err := os.Stat(req.SourceRomPath); err != nil {
		return res, fmt.Errorf("stat rom: %w", err)
	} else if st.IsDir() {
		return res, fmt.Errorf("rom path is a directory: %s", req.SourceRomPath)
	}

	romRel := filepath.Base(req.SourceRomPath)
	romRel = filepath.ToSlash(strings.TrimPrefix(strings.TrimSpace(romRel), "./"))

	romDstAbs, err := SafeJoinUnderRoot(rootDir, romRel)
	if err != nil {
		return res, err
	}
	if err := copyFileWithDirs(req.SourceRomPath, romDstAbs, req.Overwrite); err != nil {
		return res, err
	}
	res.RomCopied = true
	res.RomRelPath = romRel

	gameName := strings.TrimSpace(req.GameName)
	if gameName == "" {
		base := filepath.Base(req.SourceRomPath)
		gameName = strings.TrimSuffix(base, filepath.Ext(base))
	}
	if strings.TrimSpace(gameName) == "" {
		return res, fmt.Errorf("game name is empty")
	}

	dev := strings.TrimSpace(req.Developer)
	desc := strings.TrimSpace(req.Description)

	// Media copy.
	if needsMedia(req) {
		mediaName, err := sanitizeMediaDirName(gameName)
		if err != nil {
			return res, err
		}
		res.MediaDirName = mediaName

		mediaDir, err := SafeMediaDir(rootDir, mediaName)
		if err != nil {
			return res, err
		}
		if err := os.MkdirAll(mediaDir, 0o755); err != nil {
			return res, fmt.Errorf("mkdir media dir: %w", err)
		}

		if p := strings.TrimSpace(req.BoxFrontPath); p != "" {
			if err := copyFileWithDirs(p, filepath.Join(mediaDir, "boxFront.png"), req.Overwrite); err != nil {
				return res, err
			}
			res.MediaCopied++
		}
		if p := strings.TrimSpace(req.LogoPath); p != "" {
			if err := copyFileWithDirs(p, filepath.Join(mediaDir, "logo.png"), req.Overwrite); err != nil {
				return res, err
			}
			res.MediaCopied++
		}
		if p := strings.TrimSpace(req.VideoPath); p != "" {
			if err := copyFileWithDirs(p, filepath.Join(mediaDir, "video.mp4"), req.Overwrite); err != nil {
				return res, err
			}
			res.MediaCopied++
		}
	} else {
		// No media selected; still keep consistent name for result.
		res.MediaDirName = gameName
	}

	// Metadata upsert.
	metaPath := filepath.Join(rootDir, "metadata.pegasus.txt")
	var doc *metadata.Document
	if existing, err := metadata.ReadFile(metaPath); err == nil {
		doc = existing
	} else {
		if !os.IsNotExist(err) {
			return res, err
		}
		doc = metadata.Parse("")
	}

	chg, _, err := doc.UpsertGameByFile(metadata.Game{GameName: gameName, FileName: romRel, Developer: dev, Description: desc}, metadata.UpsertOptions{})
	if err != nil {
		return res, err
	}
	if chg {
		if err := doc.WriteFileAtomic(metaPath); err != nil {
			return res, err
		}
		res.MetadataUpdated = true
	}

	return res, nil
}

func needsMedia(req ManualAddGameRequest) bool {
	return strings.TrimSpace(req.BoxFrontPath) != "" || strings.TrimSpace(req.LogoPath) != "" || strings.TrimSpace(req.VideoPath) != ""
}

func sanitizeMediaDirName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("media dir name is empty")
	}
	// Windows invalid filename characters.
	if strings.ContainsAny(name, "<>:\"/\\|?*") {
		return "", fmt.Errorf("game name contains invalid path characters: %q", name)
	}
	// Trailing dot/space is not allowed on Windows.
	name = strings.TrimRight(name, " .")
	if name == "" {
		return "", fmt.Errorf("game name becomes empty after trimming")
	}
	return name, nil
}

func copyFileWithDirs(src, dst string, overwrite bool) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	if st, err := os.Stat(dst); err == nil {
		if st.IsDir() {
			return fmt.Errorf("destination exists and is a directory: %s", dst)
		}
		if !overwrite {
			return fmt.Errorf("destination already exists: %s", dst)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat dst %s: %w", dst, err)
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(dst), err)
	}

	in, err := os.Open(src)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("source not found: %s", src)
		}
		return fmt.Errorf("open %s: %w", src, err)
	}
	defer func() { _ = in.Close() }()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy %s -> %s: %w", src, dst, err)
	}
	if err := out.Sync(); err != nil {
		return fmt.Errorf("sync %s: %w", dst, err)
	}
	return nil
}
