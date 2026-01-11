package pegasus

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/game_tool_box/internal/logging"
	"github.com/game_tool_box/internal/pegasus/metadata"
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

func RemoveSelectedGames(rootDir string, games []GameViewModel) GameRemoveResult {
	var res GameRemoveResult

	rootDir = strings.TrimSpace(rootDir)
	if rootDir == "" {
		res.Failed++
		res.Errors = append(res.Errors, fmt.Errorf("root dir is empty"))
		return res
	}

	// Build lookup of selected games by name.
	selectedByName := make(map[string]GameViewModel)
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
		mediaDir, mediaErr := SafeMediaDir(rootDir, name)
		if mediaErr != nil {
			logging.Errorf("pegasus.RemoveSelectedGames: refuse to delete media dir game=%q err=%v", name, mediaErr)
			res.Failed++
			res.Errors = append(res.Errors, fmt.Errorf("refuse to delete media dir for %q: %w", name, mediaErr))
		} else {
			if err := os.RemoveAll(mediaDir); err != nil {
				logging.Errorf("pegasus.RemoveSelectedGames: remove media dir failed dir=%s err=%v", mediaDir, err)
				res.Failed++
				res.Errors = append(res.Errors, fmt.Errorf("remove media dir %s: %w", mediaDir, err))
			}
		}

		// rom file
		romRel := strings.TrimSpace(g.FileName)
		if romRel == "" {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Errorf("game %q fileName is empty", name))
			continue
		}

		if err := RemoveFileUnderRoot(rootDir, romRel); err != nil {
			if os.IsNotExist(err) {
				// If missing, treat as skipped (already removed or never existed).
				res.Skipped++
				continue
			}
			logging.Errorf("pegasus.RemoveSelectedGames: remove rom failed rel=%q err=%v", romRel, err)
			res.Failed++
			res.Errors = append(res.Errors, fmt.Errorf("remove rom %q: %w", romRel, err))
			continue
		}
		res.Removed++
	}

	return res
}

func removeSelectedFromMetadata(metadataPath string, selectedByName map[string]GameViewModel) (removed int, err error) {
	doc, err := metadata.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	names := make(map[string]struct{}, len(selectedByName))
	for name := range selectedByName {
		n := strings.TrimSpace(name)
		if n == "" {
			continue
		}
		names[n] = struct{}{}
	}

	removed = doc.RemoveByGameNames(names)
	if removed == 0 {
		return 0, nil
	}

	if err := doc.WriteFileAtomic(metadataPath); err != nil {
		return removed, err
	}
	return removed, nil
}
