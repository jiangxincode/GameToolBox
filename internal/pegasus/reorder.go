package pegasus

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/game_tool_box/internal/pegasus/metadata"
)

type ReorderResult struct {
	Updated bool
	Count   int
}

// ReorderGamesInMetadata rewrites the order of game blocks in metadata.pegasus.txt
// according to the given ordered file list (relative to rootDir).
//
// Important: It preserves all non-game sections (e.g. collection blocks, comments,
// unknown raw lines) exactly as-is, and only reorders game blocks.
func ReorderGamesInMetadata(rootDir string, orderedFiles []string) (ReorderResult, error) {
	var res ReorderResult

	rootDir = strings.TrimSpace(rootDir)
	if rootDir == "" {
		return res, fmt.Errorf("root dir is empty")
	}

	metaPath := filepath.Join(rootDir, "metadata.pegasus.txt")
	doc, err := metadata.ReadFile(metaPath)
	if err != nil {
		return res, err
	}

	items := doc.Items()
	gameItems := make([]metadata.GameItem, 0)
	for _, it := range items {
		if gi, ok := it.(metadata.GameItem); ok {
			gameItems = append(gameItems, gi)
		}
	}
	if len(gameItems) == 0 {
		return res, nil
	}

	// Build normalized order list (unique, first-win).
	order := make([]string, 0, len(orderedFiles))
	seen := map[string]bool{}
	for _, f := range orderedFiles {
		k := metadata.NormalizeFileKey(f)
		if k == "" || seen[k] {
			continue
		}
		seen[k] = true
		order = append(order, k)
	}

	// Choose games in requested order, then append leftovers keeping original order.
	used := make([]bool, len(gameItems))
	sorted := make([]metadata.GameItem, 0, len(gameItems))
	for _, k := range order {
		for i, gi := range gameItems {
			if used[i] {
				continue
			}
			if metadata.NormalizeFileKey(gi.Game().FileName) == k {
				sorted = append(sorted, gi)
				used[i] = true
				break
			}
		}
	}
	for i, gi := range gameItems {
		if used[i] {
			continue
		}
		sorted = append(sorted, gi)
	}

	// Update sort-by in the *lines* of each game block but keep any other lines unchanged.
	fmtInfo := doc.InferSortByFormat()
	width := fmtInfo.Width
	if width <= 0 {
		width = 3
	}
	for i := range sorted {
		newSort := fmt.Sprintf("%0*d", width, i+1)
		g := sorted[i].Game()
		g.SortBy = newSort
		lines := rewriteSortByLine(sorted[i].Lines(), newSort)
		sorted[i] = metadata.NewGameItem(lines, g)
	}

	// Rebuild items: keep all RawItems in place; replace each GameItem in encounter order.
	out := make([]metadata.Item, 0, len(items))
	giIdx := 0
	for _, it := range items {
		if it.Kind() == metadata.ItemGame {
			out = append(out, sorted[giIdx])
			giIdx++
			continue
		}
		out = append(out, it)
	}
	if giIdx != len(sorted) {
		return res, fmt.Errorf("internal reorder mismatch: expected %d games, wrote %d", len(sorted), giIdx)
	}

	doc.SetItems(out)
	if err := doc.WriteFileAtomic(metaPath); err != nil {
		return res, err
	}

	res.Updated = true
	res.Count = len(sorted)
	return res, nil
}

func rewriteSortByLine(lines []string, newSort string) []string {
	out := make([]string, 0, len(lines)+1)
	seen := false
	for _, ln := range lines {
		trim := strings.TrimSpace(ln)
		idx := strings.IndexByte(trim, ':')
		if idx >= 0 {
			key := strings.ToLower(strings.TrimSpace(trim[:idx]))
			if key == "sort-by" {
				out = append(out, "sort-by:"+newSort)
				seen = true
				continue
			}
		}
		out = append(out, ln)
	}
	if !seen {
		// Insert it after file: if possible, otherwise append.
		insertAt := -1
		for i, ln := range out {
			trim := strings.TrimSpace(ln)
			idx := strings.IndexByte(trim, ':')
			if idx >= 0 {
				key := strings.ToLower(strings.TrimSpace(trim[:idx]))
				if key == "file" {
					insertAt = i + 1
					break
				}
			}
		}
		if insertAt < 0 || insertAt > len(out) {
			out = append(out, "sort-by:"+newSort)
		} else {
			out = append(out[:insertAt], append([]string{"sort-by:" + newSort}, out[insertAt:]...)...)
		}
	}
	return out
}
