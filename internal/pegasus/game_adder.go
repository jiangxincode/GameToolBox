package pegasus

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/game_tool_box/internal/pegasus/metadata"
)

// GameAddResult tracks game addition results
type GameAddResult struct {
	Added   int
	Skipped int
	Failed  int
	Errors  []error
}

// AddSingleGame adds a single game to the metadata.pegasus.txt file.
// It checks if the game already exists and skips if it does.
//
// Parameters:
//   - rootDir: the Pegasus root directory
//   - game: the game to add
//
// Returns:
//   - GameAddResult with statistics and any errors
func AddSingleGame(rootDir string, game GameModel) GameAddResult {
	return AddGames(rootDir, []GameModel{game})
}

// AddGames adds multiple games to the metadata.pegasus.txt file.
// It checks for duplicates and skips games that already exist.
//
// Parameters:
//   - rootDir: the Pegasus root directory
//   - games: the games to add
//
// Returns:
//   - GameAddResult with statistics and any errors
func AddGames(rootDir string, games []GameModel) GameAddResult {
	var res GameAddResult

	rootDir = strings.TrimSpace(rootDir)
	if rootDir == "" {
		res.Failed++
		res.Errors = append(res.Errors, fmt.Errorf("root dir is empty"))
		return res
	}

	if len(games) == 0 {
		return res
	}

	metaPath := filepath.Join(rootDir, "metadata.pegasus.txt")

	// Load existing metadata or create new document
	var doc *metadata.Document
	if existing, err := metadata.ReadFile(metaPath); err == nil {
		doc = existing
	} else {
		if !os.IsNotExist(err) {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Errorf("failed to read metadata: %w", err))
			return res
		}
		// file missing => treat as empty
		doc = metadata.Parse("")
	}

	// Get existing games to check for duplicates
	existingGames := doc.Games()
	existingFileKeys := make(map[string]bool)
	existingGameNames := make(map[string]bool)
	for _, g := range existingGames {
		fileKey := metadata.NormalizeFileKey(g.FileName)
		if fileKey != "" {
			existingFileKeys[fileKey] = true
		}
		gameName := strings.TrimSpace(g.GameName)
		if gameName != "" {
			existingGameNames[strings.ToLower(gameName)] = true
		}
	}

	// Filter out duplicates and prepare games to add
	toAdd := make([]metadata.Game, 0, len(games))
	for _, g := range games {
		name := strings.TrimSpace(g.GameName)
		if name == "" {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Errorf("game has empty name"))
			continue
		}

		file := strings.TrimSpace(g.FileName)
		if file == "" {
			res.Failed++
			res.Errors = append(res.Errors, fmt.Errorf("game %q fileName is empty", name))
			continue
		}

		// Check for duplicate by file key
		fileKey := metadata.NormalizeFileKey(file)
		if existingFileKeys[fileKey] {
			res.Skipped++
			continue
		}

		// Check for duplicate by game name (case-insensitive)
		if existingGameNames[strings.ToLower(name)] {
			res.Skipped++
			continue
		}

		dev := strings.TrimSpace(g.Developer)
		desc := strings.TrimSpace(g.Description)
		toAdd = append(toAdd, metadata.Game{
			GameName:    name,
			FileName:    file,
			Developer:   dev,
			Description: desc,
		})
	}

	if len(toAdd) == 0 {
		// All games were either skipped or failed
		return res
	}

	// Append the new games
	if err := doc.AppendGames(toAdd); err != nil {
		res.Failed++
		res.Errors = append(res.Errors, fmt.Errorf("failed to append games: %w", err))
		return res
	}

	// Write the updated metadata
	if err := doc.WriteFileAtomic(metaPath); err != nil {
		res.Failed++
		res.Errors = append(res.Errors, fmt.Errorf("failed to write metadata: %w", err))
		return res
	}

	res.Added = len(toAdd)
	return res
}
