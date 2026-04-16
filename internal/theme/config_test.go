package theme

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigPathUsesXDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/custom/config")
	path := ConfigPath()
	if path != "/custom/config/try/config.toml" {
		t.Errorf("ConfigPath with XDG = %q, want /custom/config/try/config.toml", path)
	}
}

func TestLoadConfigMissing(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	c, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if c.Theme != "" || c.TriesPath != "" {
		t.Error("missing config should return empty Config")
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)

	c := Config{
		TriesPath: "~/src/tries",
		Theme:     "dracula",
	}
	if err := SaveConfig(c); err != nil {
		t.Fatal(err)
	}

	// Verify file exists
	path := filepath.Join(configDir, "try", "config.toml")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	// Load it back
	loaded, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Theme != "dracula" {
		t.Errorf("theme = %q, want dracula", loaded.Theme)
	}
	if loaded.TriesPath != "~/src/tries" {
		t.Errorf("tries_path = %q, want ~/src/tries", loaded.TriesPath)
	}
}

func TestSetTheme(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)

	// Set theme on fresh config
	if err := SetTheme("catppuccin"); err != nil {
		t.Fatal(err)
	}

	c, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if c.Theme != "catppuccin" {
		t.Errorf("theme = %q, want catppuccin", c.Theme)
	}
}

func TestSetThemePreservesOtherSettings(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)

	// First save a config with tries_path
	if err := SaveConfig(Config{TriesPath: "/my/tries", Theme: "default"}); err != nil {
		t.Fatal(err)
	}

	// Now set theme — should preserve tries_path
	if err := SetTheme("dracula"); err != nil {
		t.Fatal(err)
	}

	c, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if c.Theme != "dracula" {
		t.Errorf("theme = %q, want dracula", c.Theme)
	}
	if c.TriesPath != "/my/tries" {
		t.Errorf("tries_path = %q, want /my/tries (should be preserved)", c.TriesPath)
	}
}

// --- New config fields ---

func TestGetPreviewEnabledDefault(t *testing.T) {
	c := Config{}
	if !c.GetPreviewEnabled() {
		t.Error("preview should default to true when unset")
	}
}

func TestGetPreviewEnabledExplicit(t *testing.T) {
	f := false
	c := Config{PreviewEnabled: &f}
	if c.GetPreviewEnabled() {
		t.Error("preview should be false when explicitly set to false")
	}

	tr := true
	c = Config{PreviewEnabled: &tr}
	if !c.GetPreviewEnabled() {
		t.Error("preview should be true when explicitly set to true")
	}
}

func TestGetShowEmojisFallsBackToTheme(t *testing.T) {
	c := Config{} // ShowEmojis unset
	if !c.GetShowEmojis(true) {
		t.Error("unset show_emojis should use theme default (true)")
	}
	if c.GetShowEmojis(false) {
		t.Error("unset show_emojis should use theme default (false)")
	}
}

func TestGetShowEmojisOverridesTheme(t *testing.T) {
	f := false
	c := Config{ShowEmojis: &f}
	if c.GetShowEmojis(true) {
		t.Error("explicit false should override theme default true")
	}
}

func TestGetDisplayMode(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "inline"},
		{"inline", "inline"},
		{"fullscreen", "fullscreen"},
		{"gibberish", "inline"},
	}
	for _, tt := range tests {
		c := Config{DisplayMode: tt.input}
		got := c.GetDisplayMode()
		if got != tt.want {
			t.Errorf("GetDisplayMode(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestGetInlineMinRows(t *testing.T) {
	tests := []struct {
		input int
		want  int
	}{
		{0, 15},
		{-5, 15},
		{5, 5},
		{30, 30},
	}
	for _, tt := range tests {
		c := Config{InlineMinRows: tt.input}
		got := c.GetInlineMinRows()
		if got != tt.want {
			t.Errorf("GetInlineMinRows(%d) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestSaveAndLoadNewFields(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)

	showEmojis := false
	previewEnabled := true
	c := Config{
		Theme:          "dracula",
		ShowEmojis:     &showEmojis,
		PreviewEnabled: &previewEnabled,
		DisplayMode:    "fullscreen",
		InlineMinRows:  20,
		CustomIcons:    map[string]string{"django": "🎭", "flask": "🧪"},
	}

	if err := SaveConfig(c); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}

	if loaded.ShowEmojis == nil || *loaded.ShowEmojis != false {
		t.Errorf("ShowEmojis round-trip failed")
	}
	if loaded.PreviewEnabled == nil || *loaded.PreviewEnabled != true {
		t.Errorf("PreviewEnabled round-trip failed")
	}
	if loaded.DisplayMode != "fullscreen" {
		t.Errorf("DisplayMode = %q", loaded.DisplayMode)
	}
	if loaded.InlineMinRows != 20 {
		t.Errorf("InlineMinRows = %d", loaded.InlineMinRows)
	}
	if loaded.CustomIcons["django"] != "🎭" {
		t.Errorf("CustomIcons[django] = %q", loaded.CustomIcons["django"])
	}
	if loaded.CustomIcons["flask"] != "🧪" {
		t.Errorf("CustomIcons[flask] = %q", loaded.CustomIcons["flask"])
	}
}

func TestSetPreviewEnabledPreservesOther(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)

	if err := SaveConfig(Config{Theme: "dracula"}); err != nil {
		t.Fatal(err)
	}

	if err := SetPreviewEnabled(false); err != nil {
		t.Fatal(err)
	}

	c, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if c.Theme != "dracula" {
		t.Errorf("theme lost: %q", c.Theme)
	}
	if c.PreviewEnabled == nil || *c.PreviewEnabled != false {
		t.Error("PreviewEnabled should be false")
	}
}
