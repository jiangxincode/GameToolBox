package pegasus

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type GenerateResult struct {
	Created int
	Skipped int
	Failed  int

	Errors []error
}

type DeleteResult struct {
	Deleted int
	Skipped int
	Failed  int

	Errors []error
}

// GenerateSelectedFiles creates empty files for all selected games under rootDir.
// It matches the Java behavior:
//   - create <rootDir>/<fileName> if it doesn't exist
//   - if exists: skip
func GenerateSelectedFiles(rootDir string, games []GameViewModel) GenerateResult {
	var res GenerateResult
	for _, g := range games {
		if !g.Selected {
			continue
		}
		if g.FileName == "" {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Errorf("game %q fileName is empty", g.GameName))
			continue
		}
		target := filepath.Join(rootDir, filepath.FromSlash(g.FileName))
		if _, err := os.Stat(target); err == nil {
			res.Skipped++
			continue
		} else if !os.IsNotExist(err) {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Errorf("stat %s: %w", target, err))
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Errorf("mkdir %s: %w", filepath.Dir(target), err))
			continue
		}

		f, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err != nil {
			// If a race created it, treat as skipped to mirror "exists" behavior.
			if os.IsExist(err) {
				res.Skipped++
				continue
			}
			res.Failed++
			res.Errors = append(res.Errors, fmt.Errorf("create %s: %w", target, err))
			continue
		}
		_ = f.Close()
		res.Created++
	}
	return res
}

// DeleteRomsNotInConfig deletes ROM files that exist on disk but are missing in metadata.pegasus.txt.
// It is the inverse of "ExtraInConfig":
//   - MissingInConfig: present in ROM dir, absent in config => should be deleted.
//
// Safety:
//   - Only deletes paths that are under rootDir after filepath.Clean.
//   - Skips non-existent files.
func DeleteRomsNotInConfig(rootDir string) DeleteResult {
	var res DeleteResult

	diff, err := DiffConfigAgainstRomFiles(rootDir)
	if err != nil {
		res.Failed++
		res.Errors = append(res.Errors, err)
		return res
	}

	rootClean := filepath.Clean(rootDir)

	for _, g := range diff.MissingInConfig {
		if strings.TrimSpace(g.FileName) == "" {
			res.Skipped++
			continue
		}

		candidate := filepath.Join(rootClean, filepath.FromSlash(g.FileName))
		candidate = filepath.Clean(candidate)

		// Ensure candidate is within rootDir to avoid path traversal.
		rel, relErr := filepath.Rel(rootClean, candidate)
		if relErr != nil || rel == "." || strings.HasPrefix(rel, "..") {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Errorf("refuse to delete outside root: %q", g.FileName))
			continue
		}

		if _, statErr := os.Stat(candidate); statErr != nil {
			if os.IsNotExist(statErr) {
				res.Skipped++
				continue
			}
			res.Failed++
			res.Errors = append(res.Errors, fmt.Errorf("stat %s: %w", candidate, statErr))
			continue
		}

		if rmErr := os.Remove(candidate); rmErr != nil {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Errorf("remove %s: %w", candidate, rmErr))
			continue
		}
		res.Deleted++
	}

	return res
}
