package cli

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hmk/try-bedazzled/internal/dirs"
	"github.com/hmk/try-bedazzled/internal/shell"
	"github.com/hmk/try-bedazzled/internal/theme"
	"github.com/hmk/try-bedazzled/internal/tui"
	"github.com/spf13/cobra"
)

func newExecCmd() *cobra.Command {
	var (
		triesPath string
		andKeys   string
		andExit   bool
	)

	cmd := &cobra.Command{
		Use:    "exec",
		Short:  "Internal: run selector and output shell script",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if triesPath == "" {
				return fmt.Errorf("--path is required")
			}

			// Ensure tries directory exists
			if err := os.MkdirAll(triesPath, 0755); err != nil {
				return fmt.Errorf("creating tries directory: %w", err)
			}

			// Check if first arg is a URL → clone shorthand
			if len(args) > 0 {
				first := args[0]
				if isCloneURL(first) {
					return execClone(first, args[1:], triesPath)
				}

				// Check for subcommands passed through shell wrapper
				switch first {
				case "cd":
					args = args[1:]
				case "clone":
					if len(args) < 2 {
						return fmt.Errorf("usage: try clone <url> [name]")
					}
					return execClone(args[1], args[2:], triesPath)
				case "worktree":
					if len(args) < 2 {
						return fmt.Errorf("usage: try worktree <name>")
					}
					return execWorktree(args[1], triesPath)
				case ".":
					if len(args) < 2 {
						return fmt.Errorf("usage: try . <name>")
					}
					return execWorktree(args[1], triesPath)
				case "init":
					// Shouldn't reach here but handle gracefully
					return fmt.Errorf("try init should be called directly, not through exec")
				case "theme":
					// Run theme picker directly — it's not a selector query
					return newThemeCmd().RunE(cmd, args[1:])
				case "--help", "-h", "help":
					return cmd.Root().Help()
				case "--version", "-v":
					fmt.Printf("try version %s\n", Version)
					return nil
				}
			}

			// Interactive selector
			initialFilter := ""
			if len(args) > 0 {
				initialFilter = strings.Join(args, " ")
			}

			return runSelector(triesPath, initialFilter, andKeys, andExit)
		},
	}

	cmd.Flags().StringVar(&triesPath, "path", "", "Path to tries directory")
	cmd.Flags().StringVar(&andKeys, "and-keys", "", "Inject keypresses for testing (comma-separated)")
	cmd.Flags().BoolVar(&andExit, "and-exit", false, "Exit after one render (for testing)")

	return cmd
}

func runSelector(triesPath, initialFilter, andKeys string, andExit bool) error {
	cleanup := setupTTYRenderer()
	defer cleanup()

	// Scan directories
	scanResult, err := dirs.Scan(triesPath)
	if err != nil {
		return fmt.Errorf("scanning %s: %w", triesPath, err)
	}

	// Resolve theme — read from config file if available
	cfg, _ := theme.LoadConfig()
	t := theme.Resolve(noColors, cfg.Theme)

	// Create model with config for custom icons, preview pref, etc.
	m := tui.New(triesPath, scanResult.Entries, initialFilter, t, cfg)

	// Wire preview toggle to persist state to config
	m.SetPreviewToggleCallback(func(enabled bool) {
		_ = theme.SetPreviewEnabled(enabled)
	})

	// If test mode with key injection, send keys programmatically
	if andKeys != "" {
		return runWithInjectedKeys(m, andKeys, triesPath)
	}

	// Run Bubble Tea program, rendering to /dev/tty (stderr)
	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	fm := finalModel.(tui.Model)
	return handleResult(fm.GetResult(), triesPath)
}

func runWithInjectedKeys(m tui.Model, keys string, triesPath string) error {
	// Parse comma-separated key symbols
	keyList := parseKeyList(keys)

	var model tea.Model = m
	var lastView string
	for _, k := range keyList {
		// Capture frame before processing key (for test verification)
		lastView = model.(tui.Model).View()
		msg := parseKeyMsg(k)
		var cmd tea.Cmd
		model, cmd = model.Update(msg)
		_ = cmd
		if model.(tui.Model).Done() {
			break
		}
	}

	fm := model.(tui.Model)
	// Print last visible frame to stderr for test verification
	fmt.Fprint(os.Stderr, lastView)

	return handleResult(fm.GetResult(), triesPath)
}

func handleResult(r tui.Result, triesPath string) error {
	var script string

	switch r.Action {
	case tui.ActionCD:
		script = shell.BuildCDScript(r.Path)
	case tui.ActionMkdir:
		script = shell.BuildMkdirScript(r.Path)
	case tui.ActionDelete:
		cwd, _ := os.Getwd()
		script = shell.BuildDeleteScript(triesPath, r.DeleteNames, cwd)
	case tui.ActionRename:
		script = shell.BuildRenameScript(triesPath, r.RenameOld, r.RenameNew)
	case tui.ActionCancel:
		fmt.Fprintln(os.Stderr, "Cancelled.")
		return nil
	default:
		return nil
	}

	if script == "" {
		return fmt.Errorf("failed to generate script")
	}

	fmt.Print(script)
	return nil
}

func isCloneURL(s string) bool {
	return strings.HasPrefix(s, "https://") ||
		strings.HasPrefix(s, "http://") ||
		strings.HasPrefix(s, "git@")
}

// parseKeyList splits a comma-separated key list: "redis,ENTER" → ["r","e","d","i","s","ENTER"]
func parseKeyList(s string) []string {
	parts := strings.Split(s, ",")
	var keys []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		// Check if it's a symbolic key name
		upper := strings.ToUpper(p)
		switch upper {
		case "ENTER", "ESCAPE", "UP", "DOWN", "LEFT", "RIGHT",
			"BACKSPACE", "TAB", "SPACE":
			keys = append(keys, p)
		default:
			// Check for CTRL-X pattern
			if strings.HasPrefix(upper, "CTRL-") {
				keys = append(keys, p)
			} else {
				// It's literal text — expand to individual characters
				for _, ch := range p {
					keys = append(keys, string(ch))
				}
			}
		}
	}
	return keys
}

// parseKeyMsg converts a key name to a tea.KeyMsg.
func parseKeyMsg(key string) tea.KeyMsg {
	upper := strings.ToUpper(key)
	switch upper {
	case "ENTER":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "ESCAPE", "ESC":
		return tea.KeyMsg{Type: tea.KeyEscape}
	case "UP":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "DOWN":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "LEFT":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "RIGHT":
		return tea.KeyMsg{Type: tea.KeyRight}
	case "BACKSPACE":
		return tea.KeyMsg{Type: tea.KeyBackspace}
	case "TAB":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "SPACE":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	default:
		if strings.HasPrefix(upper, "CTRL-") && len(upper) == 6 {
			ch := rune(upper[5])
			switch ch {
			case 'C':
				return tea.KeyMsg{Type: tea.KeyCtrlC}
			case 'D':
				return tea.KeyMsg{Type: tea.KeyCtrlD}
			case 'N':
				return tea.KeyMsg{Type: tea.KeyCtrlN}
			case 'P':
				return tea.KeyMsg{Type: tea.KeyCtrlP}
			case 'R':
				return tea.KeyMsg{Type: tea.KeyCtrlR}
			}
		}
		// Single character
		if len(key) == 1 {
			return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
		}
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	}
}
