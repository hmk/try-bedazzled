package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hmk/try-bedazzled/internal/theme"
	"github.com/hmk/try-bedazzled/internal/tui"
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

	m := tui.NewSettings(cfg)
	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("settings TUI: %w", err)
	}

	sm := finalModel.(tui.SettingsModel)
	res := sm.GetResult()
	if !res.Saved {
		fmt.Fprintln(os.Stderr, "Settings not saved.")
		return nil
	}

	if err := theme.SaveConfig(res.Config); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}
	fmt.Fprintln(os.Stderr, "✓ Settings saved.")
	return nil
}
