package pegasus

import (
	"path/filepath"

	"github.com/game_tool_box/internal/pegasus/metadata"
)

// GameModel is a UI/business-friendly wrapper around metadata.Game.
//
// It embeds metadata.Game to reuse the canonical Pegasus game fields, and adds
// view state (Selected/ID) needed by the app.
type GameModel struct {
	metadata.Game

	Selected bool
	ID       int
}

func (g GameModel) MediaDir(rootDir string) string {
	return filepath.Join(rootDir, "media", g.GameName)
}

func (g GameModel) LogoImagePath(rootDir string) string {
	return filepath.Join(g.MediaDir(rootDir), "logo.png")
}

func (g GameModel) BoxFrontImagePath(rootDir string) string {
	return filepath.Join(g.MediaDir(rootDir), "boxFront.png")
}

func (g GameModel) VideoFilePath(rootDir string) string {
	return filepath.Join(g.MediaDir(rootDir), "video.mp4")
}
