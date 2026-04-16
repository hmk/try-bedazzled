package theme

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Default values for config fields when not set.
const (
	DefaultDisplayMode   = "inline"
	DefaultInlineMinRows = 15
)

// Config represents the user's try configuration file.
//
// Pointer fields (ShowEmojis, PreviewEnabled) use *bool so we can distinguish
// "unset" (use default) from "explicitly false". This matters for theme overrides.
type Config struct {
	TriesPath      string            `toml:"tries_path,omitempty"`
	Theme          string            `toml:"theme,omitempty"`
	ShowEmojis     *bool             `toml:"show_emojis,omitempty"`     // nil = use theme default
	PreviewEnabled *bool             `toml:"preview_enabled,omitempty"` // nil = default true
	DisplayMode    string            `toml:"display_mode,omitempty"`    // "fullscreen" | "inline"
	InlineMinRows  int               `toml:"inline_min_rows,omitempty"` // default 15
	CustomIcons    map[string]string `toml:"custom_icons,omitempty"`    // slug-word → emoji
}

// GetPreviewEnabled returns whether the preview panel should be shown.
// Defaults to true when unset.
func (c Config) GetPreviewEnabled() bool {
	if c.PreviewEnabled == nil {
		return true
	}
	return *c.PreviewEnabled
}

// GetShowEmojis returns whether emoji icons should be shown.
// Falls back to the theme's default when the config value is unset.
func (c Config) GetShowEmojis(themeDefault bool) bool {
	if c.ShowEmojis == nil {
		return themeDefault
	}
	return *c.ShowEmojis
}

// GetDisplayMode returns "fullscreen" or "inline".
// Defaults to "inline" when unset or invalid.
func (c Config) GetDisplayMode() string {
	switch c.DisplayMode {
	case "fullscreen", "inline":
		return c.DisplayMode
	default:
		return DefaultDisplayMode
	}
}

// GetInlineMinRows returns the minimum number of rows the TUI should occupy
// in inline mode. Defaults to 15 when unset or non-positive.
func (c Config) GetInlineMinRows() int {
	if c.InlineMinRows <= 0 {
		return DefaultInlineMinRows
	}
	return c.InlineMinRows
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

// SetPreviewEnabled updates only the preview_enabled field in the config.
func SetPreviewEnabled(enabled bool) error {
	c, err := LoadConfig()
	if err != nil {
		c = Config{}
	}
	c.PreviewEnabled = &enabled
	return SaveConfig(c)
}
