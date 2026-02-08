package pegasus

import (
	"fmt"
	"path/filepath"
	"strings"
)

// safePathUnderRoot joins rootDir and rel (a path from metadata/file lists) and
// returns a cleaned absolute candidate path.
//
// Safety rules:
//   - rel must be a relative path (absolute paths are refused)
//   - after cleaning, the resulting path must be under rootDir
//   - rel can use forward slashes; it will be converted with filepath.FromSlash
func safePathUnderRoot(rootDir, rel string) (string, error) {
	rootDir = strings.TrimSpace(rootDir)
	if rootDir == "" {
		return "", fmt.Errorf("root dir is empty")
	}
	rel = strings.TrimSpace(rel)
	if rel == "" {
		return "", fmt.Errorf("path is empty")
	}
	if filepath.IsAbs(rel) {
		return "", fmt.Errorf("absolute path is not allowed: %q", rel)
	}

	rootClean := filepath.Clean(rootDir)
	candidate := filepath.Join(rootClean, filepath.FromSlash(rel))
	candidate = filepath.Clean(candidate)

	// Ensure candidate is within rootDir to avoid path traversal.
	relToRoot, err := filepath.Rel(rootClean, candidate)
	if err != nil {
		return "", err
	}
	if relToRoot == "." || strings.HasPrefix(relToRoot, "..") {
		return "", fmt.Errorf("refuse to access outside root: %q", rel)
	}

	return candidate, nil
}
