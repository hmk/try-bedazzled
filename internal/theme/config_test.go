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
