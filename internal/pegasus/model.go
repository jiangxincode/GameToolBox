package pegasus

import (
	"path/filepath"

	"github.com/game_tool_box/internal/pegasus/metadata"
)

// GameViewModel is a UI/business-friendly wrapper around metadata.Game.
//
// It embeds metadata.Game to reuse the canonical Pegasus game fields, and adds
// view state (Selected/ID) needed by the app.
//
// Placement rationale:
//   - It lives in package pegasus because many pegasus services accept/return it.
//   - UI may depend on pegasus, but pegasus must not depend on UI packages.
//   - metadata stays pure (no UI state like Selected/ID).
type GameViewModel struct {
	metadata.Game

	Selected bool
	ID       int
}

func (g GameViewModel) MediaDir(rootDir string) string {
	return filepath.Join(rootDir, "media", g.GameName)
}

func (g GameViewModel) LogoImagePath(rootDir string) string {
	return filepath.Join(g.MediaDir(rootDir), "logo.png")
}

func (g GameViewModel) BoxFrontImagePath(rootDir string) string {
	return filepath.Join(g.MediaDir(rootDir), "boxFront.png")
}

func (g GameViewModel) VideoFilePath(rootDir string) string {
	return filepath.Join(g.MediaDir(rootDir), "video.mp4")
}
