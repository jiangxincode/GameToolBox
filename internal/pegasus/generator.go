package pegasus

import (
	"fmt"
	"os"
	"path/filepath"
)

type GenerateResult struct {
	Created int
	Skipped int
	Failed  int

	Errors []error
}

// GenerateSelectedFiles creates empty files for all selected games under rootDir.
// It matches the Java behavior:
//   - create <rootDir>/<fileName> if it doesn't exist
//   - if exists: skip
func GenerateSelectedFiles(rootDir string, games []GameModel) GenerateResult {
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
		target := filepath.Join(rootDir, g.FileName)
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
