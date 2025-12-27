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
}

func configDir() (string, error) {
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
	dir, err := configDir()
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
