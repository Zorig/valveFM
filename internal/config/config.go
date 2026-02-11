package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// AppConfig holds application-level configuration.
type AppConfig struct {
	Theme string `json:"theme"`
}

// LoadConfig reads the app config from ~/.config/valvefm/config.json.
// Returns a zero-value AppConfig if the file does not exist.
func LoadConfig() AppConfig {
	path, err := configPath()
	if err != nil {
		return AppConfig{}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return AppConfig{}
	}
	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return AppConfig{}
	}
	return cfg
}

// SaveTheme persists the theme slug to the config file,
// preserving any other fields that may exist.
func SaveTheme(slug string) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	// Load existing config to preserve other fields.
	var raw map[string]interface{}
	data, err := os.ReadFile(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		raw = make(map[string]interface{})
	} else {
		if err := json.Unmarshal(data, &raw); err != nil {
			raw = make(map[string]interface{})
		}
	}

	raw["theme"] = slug

	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o644)
}

func configPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "valvefm", "config.json"), nil
}
