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

	parsed := doc.Games()
	games := make([]GameModel, 0, len(parsed))
	for i, g := range parsed {
		gm := GameModel{
			Game:     g,
			ID:       i + 1,
			Selected: false,
		}
		games = append(games, gm)
	}

	return games, nil
}
