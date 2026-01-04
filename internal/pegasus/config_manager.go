package pegasus

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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

	selected := make([]GameModel, 0, len(games))
	for _, g := range games {
		if !g.Selected {
			continue
		}
		if strings.TrimSpace(g.GameName) == "" {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Errorf("selected game has empty name"))
			continue
		}
		if strings.TrimSpace(g.FileName) == "" {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Errorf("game %q fileName is empty", g.GameName))
			continue
		}
		selected = append(selected, g)
	}

	if len(selected) == 0 {
		return res
	}

	metaPath := filepath.Join(rootDir, "metadata.pegasus.txt")
	if err := os.MkdirAll(filepath.Dir(metaPath), 0o755); err != nil {
		res.Failed++
		res.Errors = append(res.Errors, err)
		return res
	}

	f, err := os.OpenFile(metaPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		res.Failed++
		res.Errors = append(res.Errors, err)
		return res
	}
	defer func() { _ = f.Close() }()

	w := bufio.NewWriter(f)
	defer func() { _ = w.Flush() }()

	for i, g := range selected {
		name := strings.TrimSpace(g.GameName)
		file := strings.TrimSpace(g.FileName)
		dev := strings.TrimSpace(g.Developer)
		if dev == "" {
			dev = name
		}
		desc := strings.TrimSpace(g.Description)
		if desc == "" {
			desc = name
		}

		// Keep a blank line between records for readability.
		if i > 0 {
			_, _ = w.WriteString("\n")
		}
		_, _ = w.WriteString("game:" + name + "\n")
		_, _ = w.WriteString("file:" + file + "\n")
		_, _ = w.WriteString(fmt.Sprintf("sort-by:%03d\n", i+1))
		_, _ = w.WriteString("developer:" + dev + "\n")
		_, _ = w.WriteString("description:" + desc + "\n")

		res.Written++
	}

	// Ensure file ends with a blank line.
	_, _ = w.WriteString("\n")

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

		g := GameModel{ID: id, GameName: strings.TrimSuffix(name, filepath.Ext(name)), FileName: rel}
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
		k := normalizeFileKey(g.FileName)
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
		k := normalizeFileKey(g.FileName)
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

	maxSort := 0
	maxSortWidth := 0
	endsWithBlankLine := false

	// Parse existing metadata:
	//   - maxSort: only from game blocks (ignore collection blocks)
	//   - maxSortWidth: preserve the digit width of the max sort-by (e.g. 003 => width=3)
	//   - endsWithBlankLine: whether file currently ends with a blank line
	// Accept both "game:" and "game: " styles for compatibility.
	if b, err := os.ReadFile(metaPath); err == nil {
		// ends with a blank line iff it ends with two line breaks.
		endsWithBlankLine = strings.HasSuffix(string(b), "\n\n") || strings.HasSuffix(string(b), "\r\n\r\n")

		scanner := bufio.NewScanner(strings.NewReader(string(b)))
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 2*1024*1024)

		inGame := false
		inCollection := false

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			switch {
			case strings.HasPrefix(line, "game:"):
				inGame = true
				inCollection = false
			case strings.HasPrefix(line, "collection:"):
				inCollection = true
				inGame = false
			case strings.HasPrefix(line, "sort-by:"):
				if inGame && !inCollection {
					svRaw := strings.TrimSpace(line[len("sort-by:"):])
					sv := strings.TrimLeft(svRaw, "0")
					if sv == "" {
						sv = "0"
					}
					if n, err := strconv.Atoi(sv); err == nil {
						if n > maxSort {
							maxSort = n
							maxSortWidth = len(svRaw)
						} else if n == maxSort {
							if len(svRaw) > maxSortWidth {
								maxSortWidth = len(svRaw)
							}
						}
					}
				}
			}
		}
		_ = scanner.Err()
	}

	if maxSortWidth <= 0 {
		maxSortWidth = 1
	}

	f, err := os.OpenFile(metaPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	w := bufio.NewWriter(f)
	defer func() { _ = w.Flush() }()

	// Ensure there is exactly one blank line separating the existing content and the first appended game.
	// If the file currently doesn't end with a blank line, add exactly one newline.
	if !endsWithBlankLine {
		_, _ = w.WriteString("\n")
	}

	sortNo := maxSort
	for i, g := range missing {
		name := strings.TrimSpace(g.GameName)
		if name == "" {
			name = strings.TrimSuffix(filepath.Base(filepath.FromSlash(g.FileName)), filepath.Ext(g.FileName))
		}
		sortNo++
		sortStr := fmt.Sprintf("%0*d", maxSortWidth, sortNo)

		if i > 0 {
			// Exactly one blank line between appended game blocks.
			_, _ = w.WriteString("\n")
		}

		_, _ = w.WriteString("game:" + name + "\n")
		_, _ = w.WriteString("file:" + g.FileName + "\n")
		_, _ = w.WriteString("sort-by:" + sortStr + "\n")
		_, _ = w.WriteString("developer:" + name + "\n")
		_, _ = w.WriteString("description:" + name + "\n")
	}

	// Ensure file ends with a blank line.
	_, _ = w.WriteString("\n")

	return nil
}

// RemoveGamesFromMetadata removes entries whose normalized file key matches any in filesToRemove.
func RemoveGamesFromMetadata(rootDir string, filesToRemove []string) (removed int, err error) {
	meta := filepath.Join(rootDir, "metadata.pegasus.txt")

	removeSet := map[string]struct{}{}
	for _, f := range filesToRemove {
		k := normalizeFileKey(f)
		if k != "" {
			removeSet[k] = struct{}{}
		}
	}
	if len(removeSet) == 0 {
		return 0, nil
	}

	in, err := os.Open(meta)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	defer func() { _ = in.Close() }()

	tmp := meta + ".tmp"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = out.Close()
			_ = os.Remove(tmp)
		}
	}()

	scanner := bufio.NewScanner(in)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 2*1024*1024)

	w := bufio.NewWriter(out)

	skip := false
	blockFileKey := ""
	blockLines := make([]string, 0, 16)
	flushBlock := func() error {
		if len(blockLines) == 0 {
			return nil
		}
		if blockFileKey != "" {
			if _, ok := removeSet[blockFileKey]; ok {
				removed++
				blockLines = blockLines[:0]
				blockFileKey = ""
				skip = false
				return nil
			}
		}
		for _, ln := range blockLines {
			if _, werr := w.WriteString(ln + "\n"); werr != nil {
				return werr
			}
		}
		blockLines = blockLines[:0]
		blockFileKey = ""
		skip = false
		return nil
	}

	for scanner.Scan() {
		line := scanner.Text()
		trim := strings.TrimSpace(line)

		if strings.HasPrefix(trim, "game:") {
			if err := flushBlock(); err != nil {
				return removed, err
			}
			blockLines = append(blockLines, line)
			continue
		}
		if len(blockLines) == 0 {
			// outside any game block, preserve
			if _, werr := w.WriteString(line + "\n"); werr != nil {
				return removed, werr
			}
			continue
		}

		if strings.HasPrefix(trim, "file:") {
			fileVal := strings.TrimSpace(trim[len("file:"):])
			blockFileKey = normalizeFileKey(fileVal)
			if _, ok := removeSet[blockFileKey]; ok {
				skip = true
			}
		}

		if !skip {
			blockLines = append(blockLines, line)
		}
	}
	if scanErr := scanner.Err(); scanErr != nil {
		return removed, scanErr
	}
	if err := flushBlock(); err != nil {
		return removed, err
	}

	if flushErr := w.Flush(); flushErr != nil {
		return removed, flushErr
	}
	if closeErr := out.Close(); closeErr != nil {
		return removed, closeErr
	}

	if repErr := metadata.ReplaceFileAtomic(meta, tmp); repErr != nil {
		return removed, repErr
	}

	return removed, nil
}

func normalizeFileKey(fileName string) string {
	f := strings.TrimSpace(fileName)
	if f == "" {
		return ""
	}
	f = filepath.ToSlash(f)
	f = strings.TrimPrefix(f, "./")
	return strings.ToLower(f)
}
