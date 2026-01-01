package pegasus

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/game_tool_box/internal/logging"
)

// GameRemover is the backend for the "Pegasus Game Remover" feature.
//
// Behavior:
//   - Remove selected games from <rootDir>/metadata.pegasus.txt (remove full blocks starting at "game:" until next "game:" or EOF)
//   - Remove <rootDir>/media/<gameName>/ directory (if exists)
//   - Remove ROM file referenced by "file:" (relative to rootDir; if absolute, use as-is)
//
// Contract:
//   - Input: rootDir, games with Selected flag
//   - Output: GameRemoveResult with Removed/Skipped/Failed and collected Errors

type GameRemoveResult struct {
	Removed int
	Skipped int
	Failed  int

	Errors []error
}

func RemoveSelectedGames(rootDir string, games []GameModel) GameRemoveResult {
	var res GameRemoveResult

	rootDir = strings.TrimSpace(rootDir)
	if rootDir == "" {
		res.Failed++
		res.Errors = append(res.Errors, fmt.Errorf("root dir is empty"))
		return res
	}

	// Build lookup of selected games by name.
	selectedByName := make(map[string]GameModel)
	selectedCount := 0
	for _, g := range games {
		if !g.Selected {
			continue
		}
		name := strings.TrimSpace(g.GameName)
		if name == "" {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Errorf("selected game has empty name"))
			continue
		}
		selectedByName[name] = g
		selectedCount++
	}
	if selectedCount == 0 {
		return res
	}

	// 1) Rewrite metadata.pegasus.txt
	metadataPath := filepath.Join(rootDir, "metadata.pegasus.txt")
	removedFromMetadata, err := removeSelectedFromMetadata(metadataPath, selectedByName)
	if err != nil {
		logging.Errorf("pegasus.RemoveSelectedGames: removeSelectedFromMetadata failed path=%s err=%v", metadataPath, err)
		res.Failed++
		res.Errors = append(res.Errors, err)
	} else {
		res.Removed += removedFromMetadata
		// If a selected game wasn't found in metadata, count as skipped for that aspect.
		res.Skipped += selectedCount - removedFromMetadata
	}

	// 2) Delete media dirs and ROM files
	for name, g := range selectedByName {
		// media/<gameName>
		mediaDir := filepath.Join(rootDir, "media", name)
		if err := os.RemoveAll(mediaDir); err != nil {
			logging.Errorf("pegasus.RemoveSelectedGames: remove media dir failed dir=%s err=%v", mediaDir, err)
			res.Failed++
			res.Errors = append(res.Errors, fmt.Errorf("remove media dir %s: %w", mediaDir, err))
		}

		// rom file
		rom := strings.TrimSpace(g.FileName)
		if rom == "" {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Errorf("game %q fileName is empty", name))
			continue
		}
		romPath := rom
		if !filepath.IsAbs(romPath) {
			romPath = filepath.Join(rootDir, filepath.FromSlash(romPath))
		}
		if err := os.Remove(romPath); err != nil {
			if os.IsNotExist(err) {
				// If missing, treat as skipped (already removed or never existed).
				res.Skipped++
				continue
			}
			logging.Errorf("pegasus.RemoveSelectedGames: remove rom failed path=%s err=%v", romPath, err)
			res.Failed++
			res.Errors = append(res.Errors, fmt.Errorf("remove rom %s: %w", romPath, err))
			continue
		}
		res.Removed++
	}

	return res
}

func removeSelectedFromMetadata(metadataPath string, selectedByName map[string]GameModel) (removed int, err error) {
	f, err := os.Open(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Nothing to do
			return 0, nil
		}
		return 0, err
	}
	defer func() { _ = f.Close() }()

	tmp := metadataPath + ".tmp"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return 0, err
	}
	// Ensure out is closed on all error paths; on success we'll close explicitly before rename.
	defer func() {
		if err != nil {
			_ = out.Close()
			_ = os.Remove(tmp)
		}
	}()

	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 2*1024*1024)

	writer := bufio.NewWriter(out)

	skipBlock := false

	for scanner.Scan() {
		line := scanner.Text()
		trim := strings.TrimSpace(line)

		if strings.HasPrefix(trim, "game:") {
			currentGame := strings.TrimSpace(trim[len("game:"):])
			if _, ok := selectedByName[currentGame]; ok {
				skipBlock = true
				removed++
				continue // don't write this line
			}
			skipBlock = false
		}

		if skipBlock {
			continue
		}

		if _, werr := writer.WriteString(line + "\n"); werr != nil {
			err = werr
			return removed, err
		}
	}

	if scanErr := scanner.Err(); scanErr != nil {
		err = scanErr
		return removed, err
	}

	if flushErr := writer.Flush(); flushErr != nil {
		err = flushErr
		return removed, err
	}
	if closeErr := out.Close(); closeErr != nil {
		err = closeErr
		return removed, err
	}

	// Try rename first.
	if renameErr := os.Rename(tmp, metadataPath); renameErr == nil {
		return removed, nil
	}

	// If the destination exists/locked (common on Windows), try best-effort replace.
	if rmErr := os.Remove(metadataPath); rmErr == nil || os.IsNotExist(rmErr) {
		if renameErr2 := os.Rename(tmp, metadataPath); renameErr2 == nil {
			return removed, nil
		}
	}

	// Last resort: copy temp content into destination, then remove temp.
	in, inErr := os.Open(tmp)
	if inErr != nil {
		err = inErr
		return removed, err
	}
	defer func() { _ = in.Close() }()

	// Truncate/write destination
	dst, dstErr := os.OpenFile(metadataPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if dstErr != nil {
		err = dstErr
		return removed, err
	}
	defer func() { _ = dst.Close() }()

	if _, cErr := io.Copy(dst, in); cErr != nil {
		err = cErr
		return removed, err
	}
	_ = dst.Close()
	_ = in.Close()
	_ = os.Remove(tmp)

	return removed, nil
}
