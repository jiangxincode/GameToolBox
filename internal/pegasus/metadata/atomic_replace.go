package metadata

import (
	"io"
	"os"
)

// ReplaceFileAtomic replaces dstPath with tmpPath in a Windows-friendly way.
//
// Contract:
//   - tmpPath must already exist and be fully written/closed by the caller.
//   - On success, tmpPath will be moved into dstPath (or its contents copied).
//   - On failure, it tries to keep dstPath unchanged when possible.
//
// Strategy:
//  1. Try os.Rename(tmp, dst)
//  2. If that fails (commonly due to Windows file locking), try os.Remove(dst) then os.Rename(tmp, dst)
//  3. Last resort: copy tmp contents into dst (truncate/create), then remove tmp
func ReplaceFileAtomic(dstPath, tmpPath string) error {
	// Try rename first.
	if err := os.Rename(tmpPath, dstPath); err == nil {
		return nil
	}

	// If the destination exists/locked (common on Windows), try best-effort replace.
	if rmErr := os.Remove(dstPath); rmErr == nil || os.IsNotExist(rmErr) {
		if err := os.Rename(tmpPath, dstPath); err == nil {
			return nil
		}
	}

	// Last resort: copy tmp content into destination, then remove tmp.
	in, err := os.Open(tmpPath)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	dst, err := os.OpenFile(dstPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer func() { _ = dst.Close() }()

	if _, err := io.Copy(dst, in); err != nil {
		return err
	}

	// Best-effort cleanup.
	_ = dst.Close()
	_ = in.Close()
	_ = os.Remove(tmpPath)

	return nil
}
