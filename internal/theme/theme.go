// Package theme handles loading, resolving, and providing theme configuration
// for try-bedazzled's TUI rendering.
//
// Resolution order (first match wins):
//  1. NO_COLOR env or --no-colors flag → zero-color theme
//  2. TRY_THEME env var → load by name
//  3. config.toml theme key → load by name
//  4. Built-in "default" theme
package theme

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

//go:embed builtin/*.toml
var embeddedThemes embed.FS

// Colors holds the color palette for the theme. Values are hex strings (e.g. "#7C3AED").
// An empty string means "use terminal default" (transparent).
type Colors struct {
	Accent  string `toml:"accent"`
	Dim     string `toml:"dim"`
	Text    string `toml:"text"`
	Match   string `toml:"match"`
	Danger  string `toml:"danger"`
	Success string `toml:"success"`
	Bg      string `toml:"bg"`
}

// Symbols holds the unicode/ASCII glyphs used in the TUI.
type Symbols struct {
	Cursor  string `toml:"cursor"`
	Created string `toml:"created"`
	Deleted string `toml:"deleted"`
	Folder  string `toml:"folder"`
}

// Layout controls TUI layout options.
type Layout struct {
	ShowDatePrefix bool `toml:"show_date_prefix"`
	MaxVisible     int  `toml:"max_visible"`
}

// Theme is the full theme configuration.
type Theme struct {
	Colors  Colors  `toml:"colors"`
	Symbols Symbols `toml:"symbols"`
	Layout  Layout  `toml:"layout"`
}

// Default returns the built-in default theme.
func Default() Theme {
	t, _ := LoadBuiltin("default")
	return t
}

// NoColor returns a zero-color theme for NO_COLOR / --no-colors mode.
// All color values are empty (terminal default), symbols are ASCII-safe.
func NoColor() Theme {
	return Theme{
		Colors: Colors{},
		Symbols: Symbols{
			Cursor:  ">",
			Created: "+",
			Deleted: "x",
			Folder:  "",
		},
		Layout: Layout{
			ShowDatePrefix: true,
			MaxVisible:     12,
		},
	}
}

// BuiltinNames returns the names of all embedded built-in themes.
func BuiltinNames() []string {
	entries, err := embeddedThemes.ReadDir("builtin")
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		// Strip .toml extension
		if len(name) > 5 && name[len(name)-5:] == ".toml" {
			names = append(names, name[:len(name)-5])
		}
	}
	return names
}

// LoadBuiltin loads a built-in theme by name (e.g., "default", "catppuccin").
func LoadBuiltin(name string) (Theme, error) {
	data, err := embeddedThemes.ReadFile("builtin/" + name + ".toml")
	if err != nil {
		return Theme{}, fmt.Errorf("built-in theme %q not found: %w", name, err)
	}
	return parse(data)
}

// LoadFile loads a theme from a TOML file path.
func LoadFile(path string) (Theme, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Theme{}, fmt.Errorf("reading theme file %q: %w", path, err)
	}
	return parse(data)
}

func parse(data []byte) (Theme, error) {
	var t Theme
	if err := toml.Unmarshal(data, &t); err != nil {
		return Theme{}, fmt.Errorf("parsing theme: %w", err)
	}
	// Apply defaults for zero values
	if t.Layout.MaxVisible == 0 {
		t.Layout.MaxVisible = 12
	}
	return t, nil
}

// Resolve determines the active theme based on the resolution order:
//  1. noColors flag (from --no-colors)
//  2. NO_COLOR env or TERM=dumb
//  3. TRY_THEME env var
//  4. configTheme (from config.toml)
//  5. Custom theme dir (~/.config/try/themes/)
//  6. Built-in "default"
func Resolve(noColors bool, configTheme string) Theme {
	// 1. Explicit --no-colors flag
	if noColors {
		return NoColor()
	}

	// 2. NO_COLOR env (https://no-color.org/) or TERM=dumb
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return NoColor()
	}
	if os.Getenv("TERM") == "dumb" {
		return NoColor()
	}

	// 3. TRY_THEME env var
	if envTheme := os.Getenv("TRY_THEME"); envTheme != "" {
		return loadByName(envTheme)
	}

	// 4. Config file theme
	if configTheme != "" {
		return loadByName(configTheme)
	}

	// 5. Built-in default
	return Default()
}

// loadByName tries to load a theme: first as a built-in, then from the user's
// custom themes directory (~/.config/try/themes/<name>.toml).
func loadByName(name string) Theme {
	// Try built-in first
	if t, err := LoadBuiltin(name); err == nil {
		return t
	}

	// Try user custom themes dir
	configDir := userConfigDir()
	if configDir != "" {
		path := filepath.Join(configDir, "try", "themes", name+".toml")
		if t, err := LoadFile(path); err == nil {
			return t
		}
	}

	// Fallback to default
	return Default()
}

// userConfigDir returns XDG_CONFIG_HOME or ~/.config.
func userConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return xdg
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config")
}
