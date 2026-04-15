package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hmk/try-bedazzled/internal/theme"
	"github.com/hmk/try-bedazzled/internal/tui"
	"github.com/spf13/cobra"
)

func newThemeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "theme",
		Short: "Pick a theme with live preview",
		Long: `Launch an interactive theme picker with a live preview panel.

Browse built-in themes, see how they look in real-time, and press
Enter to apply. Your choice is saved to ~/.config/try/config.toml.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cleanup := setupTTYRenderer()
			defer cleanup()

			m := tui.NewThemePicker()

			p := tea.NewProgram(m, tea.WithOutput(os.Stderr))
			finalModel, err := p.Run()
			if err != nil {
				return fmt.Errorf("theme picker error: %w", err)
			}

			fm := finalModel.(tui.ThemePickerModel)
			result := fm.GetResult()

			if !result.Selected {
				fmt.Fprintln(os.Stderr, "No theme selected.")
				return nil
			}

			if err := theme.SetTheme(result.Name); err != nil {
				return fmt.Errorf("saving theme: %w", err)
			}

			fmt.Fprintf(os.Stderr, "Theme set to %q. Saved to %s\n",
				result.Name, theme.ConfigPath())
			return nil
		},
	}
}
