package pegasus

import "path/filepath"

// GameModel mirrors the fields used by the Java version.
// It represents one game entry parsed from metadata.pegasus.txt.
//
// Note: Media files are optional and will be resolved if a media/<gameName>/ directory exists.
type GameModel struct {
	Selected    bool
	ID          int
	GameName    string
	FileName    string
	SortBy      string
	Developer   string
	Description string

	LogoImagePath     string
	BoxFrontImagePath string
	VideoFilePath     string
}

func (g GameModel) MediaDir(rootDir string) string {
	return filepath.Join(rootDir, "media", g.GameName)
}
