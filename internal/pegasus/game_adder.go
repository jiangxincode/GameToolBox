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
// If media files are specified in the GameModel, they will be copied to the media directory.
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
	gamesToProcess := make([]GameModel, 0, len(games))
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
		gamesToProcess = append(gamesToProcess, g)
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

	// Copy media files if specified
	for _, g := range gamesToProcess {
		if err := copyMediaFiles(rootDir, g); err != nil {
			// Log error but don't fail the operation since metadata is already written
			res.Errors = append(res.Errors, fmt.Errorf("failed to copy media for game %q: %w", g.GameName, err))
		}
	}

	res.Added = len(toAdd)
	return res
}

// copyMediaFiles copies media files (logo, boxFront, video) to the game's media directory.
// It creates the media directory if it doesn't exist.
func copyMediaFiles(rootDir string, game GameModel) error {
	// Check if any media files are specified
	hasMedia := strings.TrimSpace(game.LogoImagePath) != "" ||
		strings.TrimSpace(game.BoxFrontImagePath) != "" ||
		strings.TrimSpace(game.VideoFilePath) != ""

	if !hasMedia {
		return nil // No media files to copy
	}

	// Create media directory for this game
	mediaDir := filepath.Join(rootDir, "media", strings.TrimSpace(game.GameName))
	if err := os.MkdirAll(mediaDir, 0o755); err != nil {
		return fmt.Errorf("create media directory: %w", err)
	}

	// Copy logo file
	if logoPath := strings.TrimSpace(game.LogoImagePath); logoPath != "" {
		dstPath := filepath.Join(mediaDir, "logo.png")
		if err := copyFileIfExists(logoPath, dstPath); err != nil {
			return fmt.Errorf("copy logo: %w", err)
		}
	}

	// Copy boxFront file
	if boxFrontPath := strings.TrimSpace(game.BoxFrontImagePath); boxFrontPath != "" {
		dstPath := filepath.Join(mediaDir, "boxFront.png")
		if err := copyFileIfExists(boxFrontPath, dstPath); err != nil {
			return fmt.Errorf("copy boxFront: %w", err)
		}
	}

	// Copy video file
	if videoPath := strings.TrimSpace(game.VideoFilePath); videoPath != "" {
		dstPath := filepath.Join(mediaDir, "video.mp4")
		if err := copyFileIfExists(videoPath, dstPath); err != nil {
			return fmt.Errorf("copy video: %w", err)
		}
	}

	return nil
}

// copyFileIfExists copies a file from src to dst. Returns an error if src doesn't exist.
func copyFileIfExists(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source file %s: %w", src, err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create destination file %s: %w", dst, err)
	}
	defer dstFile.Close()

	if _, err := dstFile.ReadFrom(srcFile); err != nil {
		return fmt.Errorf("copy from %s to %s: %w", src, dst, err)
	}

	return dstFile.Sync()
}
