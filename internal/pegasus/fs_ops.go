package pegasus

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SafeJoinUnderRoot is the single entry point for translating a user-controlled
// relative path (typically from metadata "file:" or UI selection) into an
// absolute path under rootDir.
//
// It refuses absolute paths and any path that escapes rootDir after cleaning.
func SafeJoinUnderRoot(rootDir, rel string) (string, error) {
	return safePathUnderRoot(rootDir, rel)
}

// SafeMediaDir returns <rootDir>/media/<gameName> after validating that gameName
// is a single directory name (no separators, no traversal).
//
// This protects destructive ops like RemoveAll from being tricked into deleting
// outside the Pegasus root.
func SafeMediaDir(rootDir, gameName string) (string, error) {
	rootDir = strings.TrimSpace(rootDir)
	if rootDir == "" {
		return "", fmt.Errorf("root dir is empty")
	}
	name := strings.TrimSpace(gameName)
	if name == "" {
		return "", fmt.Errorf("game name is empty")
	}

	// Reject any path separators (both platforms) and traversal tokens.
	// filepath.Base("a/b") == "b" so we must check explicitly.
	if strings.ContainsAny(name, `/\\`) {
		return "", fmt.Errorf("game name contains path separator: %q", name)
	}
	if name == "." || name == ".." || strings.Contains(name, "..") {
		// Conservative: refuse any ".." to avoid surprises.
		return "", fmt.Errorf("game name contains traversal token: %q", name)
	}

	// Build via SafeJoinUnderRoot so the root containment check is centralized.
	return SafeJoinUnderRoot(rootDir, filepath.ToSlash(filepath.Join("media", name)))
}

// RemoveFileUnderRoot removes a file under rootDir specified by a relative path.
// It refuses unsafe paths.
func RemoveFileUnderRoot(rootDir, rel string) error {
	abs, err := SafeJoinUnderRoot(rootDir, rel)
	if err != nil {
		return err
	}
	if err := os.Remove(abs); err != nil {
		// Preserve os.IsNotExist semantics for callers.
		if os.IsNotExist(err) {
			return err
		}
		return fmt.Errorf("remove %s: %w", abs, err)
	}
	return nil
}

// RecreateDirUnderRoot removes (if exists) then creates a directory under rootDir.
// Useful for "clean output folder" flows.
func RecreateDirUnderRoot(rootDir, relDir string, perm os.FileMode) (string, error) {
	abs, err := SafeJoinUnderRoot(rootDir, relDir)
	if err != nil {
		return "", err
	}
	if err := os.RemoveAll(abs); err != nil {
		return "", fmt.Errorf("remove dir %s: %w", abs, err)
	}
	if err := os.MkdirAll(abs, perm); err != nil {
		return "", fmt.Errorf("mkdir %s: %w", abs, err)
	}
	return abs, nil
}
