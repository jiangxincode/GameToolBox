package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Lang string `json:"lang"`
	// Theme reserved for future use.
	Theme string `json:"theme,omitempty"`
	// RootDir remembers last selected root directory for Pegasus G generator.
	RootDir string `json:"rootDir,omitempty"`

	// MediaScrapeBaseURL is an optional base URL used by the media scraper.
	//
	// If set, the scraper will try to download (per selected game):
	//   <baseURL>/<url.PathEscape(GameName)>/boxFront.png
	//   <baseURL>/<url.PathEscape(GameName)>/logo.png
	//   <baseURL>/<url.PathEscape(GameName)>/video.mp4
	//
	// Example: https://example.com/pegasus/media
	MediaScrapeBaseURL string `json:"mediaScrapeBaseURL,omitempty"`

	// MediaScrapeOverwrite controls whether existing local media files will be overwritten.
	// Default false: only download missing files.
	MediaScrapeOverwrite bool `json:"mediaScrapeOverwrite,omitempty"`

	// --- ScreenScraper credentials ---
	// ScreenScraperDevID / ScreenScraperDevPassword are required by ScreenScraper API.
	ScreenScraperDevID       string `json:"screenScraperDevID,omitempty"`
	ScreenScraperDevPassword string `json:"screenScraperDevPassword,omitempty"`

	// Optional user account for higher rate limits / additional access.
	ScreenScraperUser     string `json:"screenScraperUser,omitempty"`
	ScreenScraperPassword string `json:"screenScraperPassword,omitempty"`

	// --- ROM Converter settings ---
	// RomConverterSourceDir remembers last selected source directory for ROM conversion.
	RomConverterSourceDir string `json:"romConverterSourceDir,omitempty"`
	// RomConverterTargetDir remembers last selected target directory for ROM conversion.
	RomConverterTargetDir string `json:"romConverterTargetDir,omitempty"`
}

// Dir returns the application data directory used to store config/log files.
func Dir() (string, error) {
	// Allow overriding for testing or portable setups.
	if base := os.Getenv("GAMETOOLBOX_HOME"); base != "" {
		return filepath.Join(base, ".gametoolbox"), nil
	}

	base, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, ".gametoolbox"), nil
}

func configPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func Load() (Config, error) {
	p, err := configPath()
	if err != nil {
		return Config{}, err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, nil
		}
		return Config{}, err
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		// If config is corrupted, don't block app startup.
		return Config{}, nil
	}
	return c, nil
}

func Save(c Config) error {
	p, err := configPath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, b, 0o644)
}
