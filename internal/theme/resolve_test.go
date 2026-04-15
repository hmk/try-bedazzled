package theme

import (
	"os"
	"path/filepath"
	"testing"
)

// TestEndToEndThemePersistence simulates the full flow:
// 1. User picks a theme via `try theme` → SetTheme writes config
// 2. User runs `try redis` → LoadConfig + Resolve reads it back
// 3. The resolved theme should match what was saved
func TestEndToEndThemePersistence(t *testing.T) {
	// Set up isolated config dir
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)

	// Step 1: Simulate `try theme` saving "dracula"
	if err := SetTheme("dracula"); err != nil {
		t.Fatalf("SetTheme failed: %v", err)
	}

	// Verify file was written
	path := filepath.Join(configDir, "try", "config.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("config file not written: %v", err)
	}
	t.Logf("Config file contents:\n%s", string(data))

	// Step 2: Simulate `try exec` reading config
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	t.Logf("LoadConfig returned theme=%q", cfg.Theme)

	if cfg.Theme != "dracula" {
		t.Errorf("LoadConfig theme = %q, want 'dracula'", cfg.Theme)
	}

	// Step 3: Resolve should return dracula
	resolved := Resolve(false, cfg.Theme)
	t.Logf("Resolved accent=%q (dracula should be #BD93F9)", resolved.Colors.Accent)

	if resolved.Colors.Accent != "#BD93F9" {
		t.Errorf("Resolved accent = %q, want #BD93F9 (dracula)", resolved.Colors.Accent)
	}

	// Verify it's NOT the default theme
	defaultTheme := Default()
	if resolved.Colors.Accent == defaultTheme.Colors.Accent {
		t.Error("Resolved theme is the same as default — config was not applied")
	}
}

// TestEndToEndThemeSwitch tests switching themes
func TestEndToEndThemeSwitch(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)

	// Set catppuccin
	SetTheme("catppuccin")
	cfg, _ := LoadConfig()
	r1 := Resolve(false, cfg.Theme)
	if r1.Colors.Accent != "#CBA6F7" {
		t.Errorf("after setting catppuccin, accent = %q, want #CBA6F7", r1.Colors.Accent)
	}

	// Switch to dracula
	SetTheme("dracula")
	cfg, _ = LoadConfig()
	r2 := Resolve(false, cfg.Theme)
	if r2.Colors.Accent != "#BD93F9" {
		t.Errorf("after switching to dracula, accent = %q, want #BD93F9", r2.Colors.Accent)
	}
}

// TestConfigPathActual prints the real config path (not mocked) for debugging
func TestConfigPathActual(t *testing.T) {
	// Don't override XDG — show what the binary actually uses
	t.Logf("ConfigPath() = %s", ConfigPath())
	t.Logf("XDG_CONFIG_HOME = %q", os.Getenv("XDG_CONFIG_HOME"))
	t.Logf("HOME = %q", os.Getenv("HOME"))
}
