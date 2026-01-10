package pegasus

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/game_tool_box/internal/pegasus/metadata"
)

// ConfigManager is the backend for the "Pegasus Config Manager" feature.
// For now it intentionally mirrors the ROM manager behavior, but lives in its
// own file so future changes won't affect other generators.
//
// Contract:
//   - Input: rootDir (Pegasus root dir), games list with Selected flag
//   - Output: GenerateResult with Created/Skipped/Failed and Errors
//
// NOTE: This is intentionally a thin wrapper around the existing generator.
// We'll diverge the logic later.

type ConfigDiff struct {
	MissingInConfig   []GameModel // present in ROM dir but not in metadata
	ExtraInConfig     []GameModel // present in metadata but ROM file missing
	DuplicateInConfig []string    // duplicated file: entries (normalized)
}

type ConfigGenerateResult struct {
	Written int
	Skipped int
	Failed  int

	Errors []error
}

// GenerateSelectedConfig rewrites <rootDir>/metadata.pegasus.txt using selected games.
//
// Contract:
//   - Writes a standard game block for each selected game.
//   - Uses existing GameModel fields where available; falls back to GameName for developer/description.
//   - sort-by will be assigned incrementally starting at 1.
func GenerateSelectedConfig(rootDir string, games []GameModel) ConfigGenerateResult {
	var res ConfigGenerateResult

	rootDir = strings.TrimSpace(rootDir)
	if rootDir == "" {
		res.Failed++
		res.Errors = append(res.Errors, fmt.Errorf("root dir is empty"))
		return res
	}

	selected := make([]metadata.Game, 0, len(games))
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
		file := strings.TrimSpace(g.FileName)
		if file == "" {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Errorf("game %q fileName is empty", g.GameName))
			continue
		}

		dev := strings.TrimSpace(g.Developer)
		desc := strings.TrimSpace(g.Description)
		selected = append(selected, metadata.Game{GameName: name, FileName: file, Developer: dev, Description: desc})
	}

	if len(selected) == 0 {
		return res
	}

	metaPath := filepath.Join(rootDir, "metadata.pegasus.txt")

	doc := metadata.New()
	doc.SetGames(selected)

	if err := doc.WriteFileAtomic(metaPath); err != nil {
		res.Failed++
		res.Errors = append(res.Errors, err)
		return res
	}

	res.Written = len(selected)
	return res
}

// LoadGamesFromRomFiles builds games from ROM files under rootDir.
// It walks the directory and returns all files (excluding metadata.pegasus.txt and media/**).
// FileName is stored as a relative path to rootDir using forward slashes.
func LoadGamesFromRomFiles(rootDir string) ([]GameModel, error) {
	rootDir = strings.TrimSpace(rootDir)
	if rootDir == "" {
		return nil, fmt.Errorf("root dir is empty")
	}

	var games []GameModel
	id := 1
	walkErr := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// skip media directory
			if strings.EqualFold(d.Name(), "media") {
				return filepath.SkipDir
			}
			return nil
		}
		name := d.Name()
		if strings.EqualFold(name, "metadata.pegasus.txt") {
			return nil
		}

		rel, err := filepath.Rel(rootDir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		g := GameModel{ID: id, Game: metadata.Game{GameName: strings.TrimSuffix(name, filepath.Ext(name)), FileName: rel}}
		id++
		games = append(games, g)
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}
	return games, nil
}

func DiffConfigAgainstRomFiles(rootDir string) (ConfigDiff, error) {
	cfgGames, err := LoadGamesFromRootDir(rootDir)
	if err != nil {
		return ConfigDiff{}, err
	}
	romGames, err := LoadGamesFromRomFiles(rootDir)
	if err != nil {
		return ConfigDiff{}, err
	}

	cfgByFile := map[string]GameModel{}
	duplicates := map[string]int{}
	for _, g := range cfgGames {
		k := metadata.NormalizeFileKey(g.FileName)
		if k == "" {
			continue
		}
		duplicates[k]++
		// Keep the first for reporting.
		if _, ok := cfgByFile[k]; !ok {
			cfgByFile[k] = g
		}
	}

	romByFile := map[string]GameModel{}
	for _, g := range romGames {
		k := metadata.NormalizeFileKey(g.FileName)
		if k == "" {
			continue
		}
		romByFile[k] = g
	}

	var missing []GameModel
	for k, g := range romByFile {
		if _, ok := cfgByFile[k]; !ok {
			missing = append(missing, g)
		}
	}

	var extra []GameModel
	for k, g := range cfgByFile {
		// Extra means in metadata but ROM file doesn't exist.
		if _, ok := romByFile[k]; ok {
			continue
		}
		// Directly stat to handle absolute paths in metadata.
		full := g.FileName
		if !filepath.IsAbs(full) {
			full = filepath.Join(rootDir, filepath.FromSlash(full))
		}
		if _, err := os.Stat(full); err == nil {
			continue
		}
		extra = append(extra, g)
	}

	var dup []string
	for k, n := range duplicates {
		if n > 1 {
			dup = append(dup, k)
		}
	}
	// stable output
	sort.Slice(missing, func(i, j int) bool { return missing[i].FileName < missing[j].FileName })
	sort.Slice(extra, func(i, j int) bool { return extra[i].FileName < extra[j].FileName })
	sort.Strings(dup)

	return ConfigDiff{MissingInConfig: missing, ExtraInConfig: extra, DuplicateInConfig: dup}, nil
}

// AppendMissingGamesToMetadata appends missing games as new game blocks to metadata.pegasus.txt.
func AppendMissingGamesToMetadata(rootDir string, missing []GameModel) error {
	if len(missing) == 0 {
		return nil
	}

	metaPath := filepath.Join(rootDir, "metadata.pegasus.txt")

	var doc *metadata.Document
	if existing, err := metadata.ReadFile(metaPath); err == nil {
		doc = existing
	} else {
		if !os.IsNotExist(err) {
			return err
		}
		// file missing => treat as empty
		doc = metadata.Parse("")
	}

	mg := make([]metadata.Game, 0, len(missing))
	for _, g := range missing {
		mg = append(mg, metadata.Game{
			GameName: g.GameName,
			FileName: g.FileName,
			// Developer/Description intentionally left empty to match old defaulting behavior
			// SortBy is auto-generated based on existing max.
		})
	}

	if err := doc.AppendGames(mg); err != nil {
		return err
	}
	return doc.WriteFileAtomic(metaPath)
}

// RemoveGamesFromMetadata removes entries whose normalized file key matches any in filesToRemove.
func RemoveGamesFromMetadata(rootDir string, filesToRemove []string) (removed int, err error) {
	meta := filepath.Join(rootDir, "metadata.pegasus.txt")

	removeSet := map[string]struct{}{}
	for _, f := range filesToRemove {
		k := metadata.NormalizeFileKey(f)
		if k != "" {
			removeSet[k] = struct{}{}
		}
	}
	if len(removeSet) == 0 {
		return 0, nil
	}

	doc, err := metadata.ReadFile(meta)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	removed = doc.RemoveByFileKeys(removeSet)
	if removed == 0 {
		return 0, nil
	}

	if err := doc.WriteFileAtomic(meta); err != nil {
		return removed, err
	}
	return removed, nil
}
