package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/hmk/try-bedazzled/internal/theme"
	"github.com/spf13/cobra"
)

func newSettingsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "settings",
		Short: "Configure try preferences interactively",
		RunE: func(cmd *cobra.Command, args []string) error {
			cleanup := setupTTYRenderer()
			defer cleanup()
			return runSettings()
		},
	}
}

func runSettings() error {
	cfg, _ := theme.LoadConfig()

	// --- Form field values ---
	selectedTheme := cfg.Theme
	if selectedTheme == "" {
		selectedTheme = "default"
	}

	displayMode := cfg.GetDisplayMode()
	previewEnabled := cfg.GetPreviewEnabled()
	showEmojis := cfg.GetShowEmojis(true)
	inlineMinRows := cfg.GetInlineMinRows()

	// Build theme options (built-ins + any custom files)
	builtinNames := theme.BuiltinNames()
	themeOptions := make([]huh.Option[string], 0, len(builtinNames))
	for _, name := range builtinNames {
		themeOptions = append(themeOptions, huh.NewOption(name, name))
	}

	// Custom icon editor state: slice of "slug=emoji" strings
	iconPairs := iconMapToSlice(cfg.CustomIcons)
	newPair := ""

	form := huh.NewForm(
		// Page 1: Display
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Theme").
				Description("Color scheme for the TUI").
				Options(themeOptions...).
				Value(&selectedTheme),

			huh.NewSelect[string]().
				Title("Display mode").
				Description("How the selector fits in your terminal").
				Options(
					huh.NewOption("Inline (below current line)", "inline"),
					huh.NewOption("Fullscreen (alt screen)", "fullscreen"),
				).
				Value(&displayMode),

			huh.NewConfirm().
				Title("Show preview panel").
				Description("Press Ctrl-P at any time to toggle").
				Value(&previewEnabled),

			huh.NewConfirm().
				Title("Show emoji icons").
				Description("Folder icons in the list (requires font support)").
				Value(&showEmojis),
		),

		// Page 2: Custom icons
		huh.NewGroup(
			huh.NewNote().
				Title("Custom icons").
				Description(formatIconPairs(iconPairs)),

			huh.NewInput().
				Title("Add icon mapping").
				Description("Format: word=emoji  (e.g. django=🎭, rust=🦀)").
				Placeholder("word=emoji").
				Value(&newPair),
		),
	)

	if err := form.Run(); err != nil {
		// User quit with Ctrl-C / Esc — that's fine, don't error
		fmt.Fprintln(os.Stderr, "Settings not saved.")
		return nil
	}

	// Apply new pair if provided
	if newPair != "" {
		parts := strings.SplitN(newPair, "=", 2)
		if len(parts) == 2 {
			word := strings.TrimSpace(strings.ToLower(parts[0]))
			emoji := strings.TrimSpace(parts[1])
			if word != "" && emoji != "" {
				if cfg.CustomIcons == nil {
					cfg.CustomIcons = make(map[string]string)
				}
				cfg.CustomIcons[word] = emoji
			}
		}
	}

	// Write back
	cfg.Theme = selectedTheme
	cfg.DisplayMode = displayMode
	cfg.PreviewEnabled = &previewEnabled
	cfg.ShowEmojis = &showEmojis
	cfg.InlineMinRows = inlineMinRows

	if err := theme.SaveConfig(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Fprintln(os.Stderr, "✓ Settings saved.")
	return nil
}

// iconMapToSlice converts the custom icon map to a display slice.
func iconMapToSlice(m map[string]string) []string {
	if len(m) == 0 {
		return nil
	}
	result := make([]string, 0, len(m))
	for k, v := range m {
		result = append(result, k+"="+v)
	}
	return result
}

// formatIconPairs formats existing icon pairs for display in the Note widget.
func formatIconPairs(pairs []string) string {
	if len(pairs) == 0 {
		return "No custom icons yet."
	}
	return strings.Join(pairs, "\n")
}
