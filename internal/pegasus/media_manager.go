package pegasus

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// OssHandheldGenerateResult is the result for OSS handheld media generation.
// It is intentionally separated from GenerateResult to avoid coupling with the original generator.
type OssHandheldGenerateResult struct {
	Created int
	Skipped int
	Failed  int

	Errors []error
}

// GenerateSelectedFilesForOssHandheld generates files for the OSS handheld media converter feature.
//
// Current behavior:
//   - Recreate <rootDir>/images (if exists: delete then mkdir)
//   - For each selected game:
//     copy <rootDir>/media/<GameName>/boxFront.png -> <rootDir>/images/<GameName>.png
func GenerateSelectedFilesForOssHandheld(rootDir string, games []GameViewModel) OssHandheldGenerateResult {
	var res OssHandheldGenerateResult

	rootDir = strings.TrimSpace(rootDir)
	if rootDir == "" {
		res.Failed++
		res.Errors = append(res.Errors, fmt.Errorf("rootDir is empty"))
		return res
	}

	imagesDir, err := RecreateDirUnderRoot(rootDir, "images", 0o755)
	if err != nil {
		res.Failed++
		res.Errors = append(res.Errors, err)
		return res
	}

	for _, g := range games {
		if !g.Selected {
			continue
		}
		name := strings.TrimSpace(g.GameName)
		if name == "" {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Errorf("gameName is empty"))
			continue
		}

		mediaDir, mediaErr := SafeMediaDir(rootDir, name)
		if mediaErr != nil {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Errorf("refuse to access media for %q: %w", name, mediaErr))
			continue
		}

		src := filepath.Join(mediaDir, "boxFront.png")
		dst := filepath.Join(imagesDir, name+".png")

		if err := copyFile(src, dst); err != nil {
			res.Failed++
			res.Errors = append(res.Errors, err)
			continue
		}
		res.Created++
	}
	return res
}

// MediaFolderGenerateResult is the result for generating empty media folders.
// It creates <rootDir>/media/<GameName>/ for each selected game.
type MediaFolderGenerateResult struct {
	Created int
	Skipped int
	Failed  int

	Errors []error
}

// GenerateEmptyMediaFolders creates empty media folders for selected games.
//
// Behavior:
//   - Ensure <rootDir>/media exists.
//   - For each selected game:
//     mkdir -p <rootDir>/media/<GameName>
//     If the directory already exists, it is counted as Skipped.
func GenerateEmptyMediaFolders(rootDir string, games []GameViewModel) MediaFolderGenerateResult {
	var res MediaFolderGenerateResult

	rootDir = strings.TrimSpace(rootDir)
	if rootDir == "" {
		res.Failed++
		res.Errors = append(res.Errors, fmt.Errorf("rootDir is empty"))
		return res
	}

	mediaDir := filepath.Join(rootDir, "media")
	if err := os.MkdirAll(mediaDir, 0o755); err != nil {
		res.Failed++
		res.Errors = append(res.Errors, fmt.Errorf("mkdir media dir %s: %w", mediaDir, err))
		return res
	}

	for _, g := range games {
		if !g.Selected {
			continue
		}
		name := strings.TrimSpace(g.GameName)
		if name == "" {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Errorf("gameName is empty"))
			continue
		}

		target := filepath.Join(mediaDir, name)
		if st, err := os.Stat(target); err == nil {
			if st.IsDir() {
				res.Skipped++
				continue
			}
			res.Failed++
			res.Errors = append(res.Errors, fmt.Errorf("target exists but is not a directory: %s", target))
			continue
		} else if !os.IsNotExist(err) {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Errorf("stat %s: %w", target, err))
			continue
		}

		if err := os.MkdirAll(target, 0o755); err != nil {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Errorf("mkdir %s: %w", target, err))
			continue
		}
		res.Created++
	}

	return res
}

func copyFile(src, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

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
	defer func() {
		_ = out.Close()
	}()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy %s -> %s: %w", src, dst, err)
	}
	if err := out.Sync(); err != nil {
		return fmt.Errorf("sync %s: %w", dst, err)
	}
	return nil
}
