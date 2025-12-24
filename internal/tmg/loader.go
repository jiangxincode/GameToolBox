package tmg

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadGamesFromRootDir reads <rootDir>/metadata.pegasus.txt and returns the parsed games.
//
// It follows the Java implementation strictly:
//   - Recognize: game:, file:, sort-by:, developer:, description:
//   - Each "game:" starts a new record.
//   - Media files are resolved from <rootDir>/media/<gameName>/{logo.png,boxFront.png,video.mp4} if the directory exists.
func LoadGamesFromRootDir(rootDir string) ([]GameModel, error) {
	if strings.TrimSpace(rootDir) == "" {
		return nil, errors.New("root dir is empty")
	}
	metadataFile := filepath.Join(rootDir, "metadata.pegasus.txt")
	f, err := os.Open(metadataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("元数据文件不存在: %s", metadataFile)
		}
		return nil, err
	}
	defer f.Close()

	mediaDir := filepath.Join(rootDir, "media")
	mediaDirInfo, err := os.Stat(mediaDir)
	mediaExists := err == nil && mediaDirInfo.IsDir()

	scanner := bufio.NewScanner(f)
	// allow relatively long description lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 2*1024*1024)

	var games []GameModel
	var current *GameModel
	gameID := 1

	flush := func() {
		if current == nil {
			return
		}
		games = append(games, *current)
		current = nil
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		switch {
		case strings.HasPrefix(line, "game:"):
			flush()
			g := GameModel{ID: gameID, GameName: strings.TrimSpace(line[len("game:"):]), Selected: false}
			gameID++
			current = &g
		case strings.HasPrefix(line, "file:"):
			if current != nil {
				current.FileName = strings.TrimSpace(line[len("file:"):])
			}
		case strings.HasPrefix(line, "sort-by:"):
			if current != nil {
				current.SortBy = strings.TrimSpace(line[len("sort-by:"):])
			}
		case strings.HasPrefix(line, "developer:"):
			if current != nil {
				current.Developer = strings.TrimSpace(line[len("developer:"):])
			}
		case strings.HasPrefix(line, "description:"):
			if current != nil {
				current.Description = strings.TrimSpace(line[len("description:"):])
			}
		}

		if current != nil && mediaExists {
			specialMediaDir := filepath.Join(mediaDir, current.GameName)
			if info, err := os.Stat(specialMediaDir); err == nil && info.IsDir() {
				current.LogoImagePath = filepath.Join(specialMediaDir, "logo.png")
				current.BoxFrontImagePath = filepath.Join(specialMediaDir, "boxFront.png")
				current.VideoFilePath = filepath.Join(specialMediaDir, "video.mp4")
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	flush()
	return games, nil
}
