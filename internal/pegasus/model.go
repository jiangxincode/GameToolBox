package pegasus

import "path/filepath"

// GameModel mirrors the fields used by the Java version.
// It represents one game entry parsed from metadata.pegasus.txt.
//
// Note: Media files are optional. Their paths can be derived from rootDir + GameName.
type GameModel struct {
	Selected    bool
	ID          int
	GameName    string
	FileName    string
	SortBy      string
	Developer   string
	Description string
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
