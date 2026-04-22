package theme

import (
	"os"
	"path/filepath"
	"testing"
)

// --- BuiltinNames ---

func TestBuiltinNames(t *testing.T) {
	names := BuiltinNames()
	if len(names) < 4 {
		t.Fatalf("expected at least 4 built-in themes, got %d: %v", len(names), names)
	}

	expected := map[string]bool{
		"bedazzled":  false,
		"catppuccin": false,
		"dracula":    false,
		"minimal":    false,
		"rainbow":    false,
	}

	if names[0] != "bedazzled" {
		t.Errorf("bedazzled should be first in the list, got %q first (full list: %v)", names[0], names)
	}

	for _, name := range names {
		if _, ok := expected[name]; ok {
			expected[name] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("missing built-in theme: %s", name)
		}
	}
}

// --- LoadBuiltin ---

func TestLoadBuiltinDefault(t *testing.T) {
	theme, err := LoadBuiltin(DefaultThemeName)
	if err != nil {
		t.Fatal(err)
	}

	// Verify some key values from bedazzled.toml (catppuccin palette + rainbow)
	if theme.Colors.Accent != "#CBA6F7" {
		t.Errorf("accent = %q, want #CBA6F7", theme.Colors.Accent)
	}
	if theme.Colors.Match != "#F9E2AF" {
		t.Errorf("match = %q, want #F9E2AF", theme.Colors.Match)
	}
	if theme.Symbols.Cursor != "▸" {
		t.Errorf("cursor = %q, want ▸", theme.Symbols.Cursor)
	}
	if theme.Layout.MaxVisible != 12 {
		t.Errorf("max_visible = %d, want 12", theme.Layout.MaxVisible)
	}
	if !theme.Layout.ShowIcons {
		t.Error("show_icons should be true for bedazzled theme")
	}
	if theme.Layout.ShowDate != "right" {
		t.Errorf("show_date = %q, want right", theme.Layout.ShowDate)
	}
	if !theme.Layout.ShowTime {
		t.Error("show_time should be true for bedazzled theme")
	}
	if !theme.Layout.Rainbow {
		t.Error("rainbow should be true for bedazzled theme")
	}
	if theme.Layout.SearchStyle != "bordered" {
		t.Errorf("search_style = %q, want bordered", theme.Layout.SearchStyle)
	}
	if len(theme.Layout.Columns) != 3 {
		t.Errorf("expected 3 columns, got %d", len(theme.Layout.Columns))
	}
	if theme.Symbols.Folder != "📂" {
		t.Errorf("folder = %q, want 📂", theme.Symbols.Folder)
	}
}

func TestLoadBuiltinCatppuccin(t *testing.T) {
	theme, err := LoadBuiltin("catppuccin")
	if err != nil {
		t.Fatal(err)
	}
	if theme.Colors.Accent != "#CBA6F7" {
		t.Errorf("catppuccin accent = %q, want #CBA6F7", theme.Colors.Accent)
	}
}

func TestLoadBuiltinDracula(t *testing.T) {
	theme, err := LoadBuiltin("dracula")
	if err != nil {
		t.Fatal(err)
	}
	if theme.Colors.Accent != "#BD93F9" {
		t.Errorf("dracula accent = %q, want #BD93F9", theme.Colors.Accent)
	}
}

func TestLoadBuiltinMinimal(t *testing.T) {
	theme, err := LoadBuiltin("minimal")
	if err != nil {
		t.Fatal(err)
	}
	if theme.Symbols.Cursor != ">" {
		t.Errorf("minimal cursor = %q, want >", theme.Symbols.Cursor)
	}
	if theme.Symbols.Created != "+" {
		t.Errorf("minimal created = %q, want +", theme.Symbols.Created)
	}
	if theme.Layout.ShowIcons {
		t.Error("minimal should have show_icons = false")
	}
	if theme.Layout.SearchStyle != "minimal" {
		t.Errorf("minimal search_style = %q, want minimal", theme.Layout.SearchStyle)
	}
	if theme.Layout.ShowTime {
		t.Error("minimal should have show_time = false")
	}
}

func TestLayoutDefaultsFillIn(t *testing.T) {
	// A theme with only max_visible set should get defaults for everything else
	dir := t.TempDir()
	path := filepath.Join(dir, "sparse.toml")
	if err := os.WriteFile(path, []byte(`
[layout]
max_visible = 8
`), 0644); err != nil {
		t.Fatal(err)
	}

	theme, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if theme.Layout.MaxVisible != 8 {
		t.Errorf("max_visible = %d, want 8", theme.Layout.MaxVisible)
	}
	if theme.Layout.ShowDate != "right" {
		t.Errorf("show_date should default to 'right', got %q", theme.Layout.ShowDate)
	}
	if theme.Layout.SearchStyle != "bordered" {
		t.Errorf("search_style should default to 'bordered', got %q", theme.Layout.SearchStyle)
	}
	if len(theme.Layout.Columns) != 4 {
		t.Errorf("columns should default to 4 entries, got %d", len(theme.Layout.Columns))
	}
}

func TestLoadBuiltinNotFound(t *testing.T) {
	_, err := LoadBuiltin("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent theme")
	}
}

// --- LoadFile ---

func TestLoadFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "custom.toml")
	content := `
[colors]
accent = "#FF00FF"
dim = "#333333"

[symbols]
cursor = "→"

[layout]
max_visible = 20
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	theme, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if theme.Colors.Accent != "#FF00FF" {
		t.Errorf("accent = %q, want #FF00FF", theme.Colors.Accent)
	}
	if theme.Symbols.Cursor != "→" {
		t.Errorf("cursor = %q, want →", theme.Symbols.Cursor)
	}
	if theme.Layout.MaxVisible != 20 {
		t.Errorf("max_visible = %d, want 20", theme.Layout.MaxVisible)
	}
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := LoadFile("/nonexistent/path/theme.toml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadFileInvalidTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.toml")
	if err := os.WriteFile(path, []byte("this is not valid {{toml"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadFile(path)
	if err == nil {
		t.Error("expected error for invalid TOML")
	}
}

// --- Default ---

func TestDefault(t *testing.T) {
	theme := Default()
	if theme.Colors.Accent == "" {
		t.Error("default theme should have non-empty accent color")
	}
	if theme.Symbols.Cursor == "" {
		t.Error("default theme should have non-empty cursor symbol")
	}
}

// --- NoColor ---

func TestNoColor(t *testing.T) {
	theme := NoColor()

	// All colors should be empty
	if theme.Colors.Accent != "" {
		t.Errorf("NoColor accent should be empty, got %q", theme.Colors.Accent)
	}
	if theme.Colors.Dim != "" {
		t.Errorf("NoColor dim should be empty, got %q", theme.Colors.Dim)
	}
	if theme.Colors.Text != "" {
		t.Errorf("NoColor text should be empty, got %q", theme.Colors.Text)
	}
	if theme.Colors.Match != "" {
		t.Errorf("NoColor match should be empty, got %q", theme.Colors.Match)
	}

	// Symbols should be ASCII
	if theme.Symbols.Cursor != ">" {
		t.Errorf("NoColor cursor = %q, want >", theme.Symbols.Cursor)
	}
	if theme.Symbols.Created != "+" {
		t.Errorf("NoColor created = %q, want +", theme.Symbols.Created)
	}

	// Layout defaults
	if theme.Layout.MaxVisible != 12 {
		t.Errorf("NoColor max_visible = %d, want 12", theme.Layout.MaxVisible)
	}
	if theme.Layout.ShowIcons {
		t.Error("NoColor should have show_icons = false")
	}
	if theme.Layout.SearchStyle != "minimal" {
		t.Errorf("NoColor search_style = %q, want minimal", theme.Layout.SearchStyle)
	}
}

// --- Resolve ---

func TestResolveNoColorsFlag(t *testing.T) {
	theme := Resolve(true, "catppuccin")
	// noColors flag should override everything, even an explicit config theme
	if theme.Colors.Accent != "" {
		t.Error("noColors flag should produce zero-color theme")
	}
	if theme.Symbols.Cursor != ">" {
		t.Error("noColors flag should produce ASCII symbols")
	}
}

func TestResolveNO_COLOREnv(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	t.Setenv("TRY_THEME", "catppuccin") // should be ignored

	theme := Resolve(false, "")
	if theme.Colors.Accent != "" {
		t.Error("NO_COLOR env should produce zero-color theme")
	}
}

func TestResolveNO_COLOREmpty(t *testing.T) {
	// NO_COLOR spec says presence of the var is enough, even if empty
	t.Setenv("NO_COLOR", "")

	theme := Resolve(false, "")
	if theme.Colors.Accent != "" {
		t.Error("NO_COLOR='' should still produce zero-color theme")
	}
}

func TestResolveTERMDumb(t *testing.T) {
	t.Setenv("TERM", "dumb")

	theme := Resolve(false, "")
	if theme.Colors.Accent != "" {
		t.Error("TERM=dumb should produce zero-color theme")
	}
}

func TestResolveTRY_THEMEEnv(t *testing.T) {
	t.Setenv("TRY_THEME", "dracula")

	theme := Resolve(false, "")
	if theme.Colors.Accent != "#BD93F9" {
		t.Errorf("TRY_THEME=dracula should load dracula, got accent=%q", theme.Colors.Accent)
	}
}

func TestResolveTRY_THEMEOverridesConfig(t *testing.T) {
	t.Setenv("TRY_THEME", "dracula")

	// Config says catppuccin, env says dracula — env wins
	theme := Resolve(false, "catppuccin")
	if theme.Colors.Accent != "#BD93F9" {
		t.Errorf("TRY_THEME should override config, got accent=%q", theme.Colors.Accent)
	}
}

func TestResolveConfigTheme(t *testing.T) {
	theme := Resolve(false, "catppuccin")
	if theme.Colors.Accent != "#CBA6F7" {
		t.Errorf("config theme catppuccin should load, got accent=%q", theme.Colors.Accent)
	}
}

func TestResolveDefaultFallback(t *testing.T) {
	theme := Resolve(false, "")
	if theme.Colors.Accent != "#CBA6F7" {
		t.Errorf("default fallback should load bedazzled theme, got accent=%q", theme.Colors.Accent)
	}
}

func TestResolveUnknownThemeFallsBackToDefault(t *testing.T) {
	theme := Resolve(false, "nonexistent-theme-xyz")
	if theme.Colors.Accent != "#CBA6F7" {
		t.Errorf("unknown theme should fall back to bedazzled, got accent=%q", theme.Colors.Accent)
	}
}

// --- Custom user themes dir ---

func TestResolveCustomThemeFromDir(t *testing.T) {
	// Set up a fake XDG_CONFIG_HOME with a custom theme
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)

	themesDir := filepath.Join(configDir, "try", "themes")
	if err := os.MkdirAll(themesDir, 0755); err != nil {
		t.Fatal(err)
	}

	customTheme := `
[colors]
accent = "#CUSTOM1"

[symbols]
cursor = "~"
`
	if err := os.WriteFile(filepath.Join(themesDir, "mytheme.toml"), []byte(customTheme), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("TRY_THEME", "mytheme")
	theme := Resolve(false, "")

	if theme.Colors.Accent != "#CUSTOM1" {
		t.Errorf("custom theme should load, got accent=%q", theme.Colors.Accent)
	}
	if theme.Symbols.Cursor != "~" {
		t.Errorf("custom theme cursor = %q, want ~", theme.Symbols.Cursor)
	}
}

// --- Partial themes ---

func TestPartialThemeGetsDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "partial.toml")
	// Only set accent color, nothing else
	if err := os.WriteFile(path, []byte(`
[colors]
accent = "#123456"
`), 0644); err != nil {
		t.Fatal(err)
	}

	theme, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if theme.Colors.Accent != "#123456" {
		t.Errorf("accent = %q, want #123456", theme.Colors.Accent)
	}
	// max_visible should get the default of 12
	if theme.Layout.MaxVisible != 12 {
		t.Errorf("max_visible should default to 12, got %d", theme.Layout.MaxVisible)
	}
}

// --- Resolution priority order ---

func TestResolutionPriorityOrder(t *testing.T) {
	// Set everything: noColors flag beats all
	t.Setenv("NO_COLOR", "1")
	t.Setenv("TRY_THEME", "dracula")

	// noColors=true should win over NO_COLOR env, TRY_THEME, and config
	theme := Resolve(true, "catppuccin")
	if theme.Colors.Accent != "" {
		t.Error("noColors flag should have highest priority")
	}
}
