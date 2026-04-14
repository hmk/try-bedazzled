package theme

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config represents the user's try configuration file.
type Config struct {
	TriesPath string `toml:"tries_path,omitempty"`
	Theme     string `toml:"theme,omitempty"`
}

// ConfigPath returns the path to the config file.
func ConfigPath() string {
	return filepath.Join(userConfigDir(), "try", "config.toml")
}

// LoadConfig reads the config file, returning an empty Config if it doesn't exist.
func LoadConfig() (Config, error) {
	path := ConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, nil
		}
		return Config{}, err
	}
	var c Config
	if err := toml.Unmarshal(data, &c); err != nil {
		return Config{}, fmt.Errorf("parsing config: %w", err)
	}
	return c, nil
}

// SaveConfig writes the config file, creating directories as needed.
func SaveConfig(c Config) error {
	path := ConfigPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating config file: %w", err)
	}
	defer f.Close()

	enc := toml.NewEncoder(f)
	if err := enc.Encode(c); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// SetTheme updates only the theme field in the config, preserving other settings.
func SetTheme(name string) error {
	c, err := LoadConfig()
	if err != nil {
		// If config is corrupt, start fresh
		c = Config{}
	}
	c.Theme = name
	return SaveConfig(c)
}
