package pegasus

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/game_tool_box/internal/pegasus/metadata"
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

	doc, err := metadata.ReadFile(metadataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("元数据文件不存在: %s", metadataFile)
		}
		return nil, err
	}

	mediaDir := filepath.Join(rootDir, "media")
	mediaDirInfo, err := os.Stat(mediaDir)
	mediaExists := err == nil && mediaDirInfo.IsDir()

	parsed := doc.Games()
	games := make([]GameModel, 0, len(parsed))
	for i, g := range parsed {
		gm := GameModel{
			ID:          i + 1,
			Selected:    false,
			GameName:    g.GameName,
			FileName:    g.FileName,
			SortBy:      g.SortBy,
			Developer:   g.Developer,
			Description: g.Description,
		}

		if mediaExists {
			specialMediaDir := filepath.Join(mediaDir, gm.GameName)
			if info, err := os.Stat(specialMediaDir); err == nil && info.IsDir() {
				gm.LogoImagePath = filepath.Join(specialMediaDir, "logo.png")
				gm.BoxFrontImagePath = filepath.Join(specialMediaDir, "boxFront.png")
				gm.VideoFilePath = filepath.Join(specialMediaDir, "video.mp4")
			}
		}

		games = append(games, gm)
	}

	return games, nil
}
