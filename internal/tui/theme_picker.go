package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hmk/try-bedazzled/internal/dirs"
	"github.com/hmk/try-bedazzled/internal/fuzzy"
	"github.com/hmk/try-bedazzled/internal/theme"
)

// ThemePickerResult is the outcome of the theme picker.
type ThemePickerResult struct {
	Selected bool   // true if user picked a theme
	Name     string // theme name
}

// ThemePickerModel is the Bubble Tea model for the theme picker.
type ThemePickerModel struct {
	themes    []string // theme names
	cursor    int
	done      bool
	result    ThemePickerResult
	previews  map[string]theme.Theme // cached loaded themes
	width     int
	height    int
}

// NewThemePicker creates a new theme picker model.
func NewThemePicker() ThemePickerModel {
	names := theme.BuiltinNames()
	previews := make(map[string]theme.Theme, len(names))
	for _, name := range names {
		t, err := theme.LoadBuiltin(name)
		if err == nil {
			previews[name] = t
		}
	}

	return ThemePickerModel{
		themes:   names,
		previews: previews,
	}
}

func (m ThemePickerModel) Init() tea.Cmd {
	return nil
}

func (m ThemePickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			m.result = ThemePickerResult{Selected: false}
			m.done = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.themes)-1 {
				m.cursor++
			}

		case "enter":
			m.result = ThemePickerResult{
				Selected: true,
				Name:     m.themes[m.cursor],
			}
			m.done = true
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m ThemePickerModel) View() string {
	if m.done {
		return ""
	}

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		MarginBottom(1).
		Render("  Pick a theme:")

	// Build left pane: theme list
	var left strings.Builder
	for i, name := range m.themes {
		if i == m.cursor {
			cursor := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7C3AED")).Render("▸")
			label := lipgloss.NewStyle().Bold(true).Render(name)
			left.WriteString(fmt.Sprintf("  %s %s\n", cursor, label))
		} else {
			left.WriteString(fmt.Sprintf("    %s\n", name))
		}
	}

	// Build right pane: live preview of the selected theme
	selectedName := m.themes[m.cursor]
	selectedTheme := m.previews[selectedName]
	preview := renderPreview(selectedTheme)

	// Layout: side by side
	leftPane := lipgloss.NewStyle().
		Width(20).
		Render(left.String())

	rightPane := lipgloss.NewStyle().
		Width(40).
		BorderLeft(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#555")).
		PaddingLeft(2).
		Render(preview)

	panes := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	// Status bar
	status := lipgloss.NewStyle().
		Faint(true).
		MarginTop(1).
		Render("  enter: apply  •  ↑/↓: browse  •  esc: cancel")

	return title + "\n" + panes + "\n" + status
}

// Done returns true if the user has made a choice.
func (m ThemePickerModel) Done() bool {
	return m.done
}

// GetResult returns the picker result.
func (m ThemePickerModel) GetResult() ThemePickerResult {
	return m.result
}

// renderPreview generates a fake try-list preview using the given theme.
func renderPreview(t theme.Theme) string {
	styles := NewStyles(t)

	// Fake entries for preview
	fakeEntries := []struct {
		name   string
		query  string
		mtime  time.Duration
		cursor bool
	}{
		{"2026-04-14-redis-cache", "redis", 1 * time.Hour, true},
		{"2026-04-13-go-api", "", 24 * time.Hour, false},
		{"2026-04-12-react-app", "", 48 * time.Hour, false},
		{"2026-04-10-ml-pipeline", "", 96 * time.Hour, false},
	}

	now := time.Now()
	var b strings.Builder

	// Search bar
	b.WriteString(styles.Cursor.Render(styles.Symbols.Cursor))
	b.WriteString(" redis\n")

	for _, fe := range fakeEntries {
		// Cursor or indent
		if fe.cursor {
			b.WriteString(styles.Cursor.Render(styles.Symbols.Cursor) + " ")
		} else {
			b.WriteString("  ")
		}

		// Render name with fuzzy highlights
		mtime := now.Add(-fe.mtime)
		name := fe.name

		if fe.query != "" {
			r := fuzzy.MatchAt(name, fe.query, mtime, now)
			matchSet := make(map[int]bool)
			for _, pos := range r.MatchPositions {
				matchSet[pos] = true
			}

			hasDate := fuzzy.HasDatePrefix(name)
			for i, ch := range name {
				char := string(ch)
				if matchSet[i] {
					b.WriteString(styles.Match.Render(char))
				} else if hasDate && i < dirs.DatePrefixLen {
					b.WriteString(styles.Dim.Render(char))
				} else if fe.cursor {
					b.WriteString(styles.Selected.Render(char))
				} else {
					b.WriteString(styles.Normal.Render(char))
				}
			}
		} else {
			hasDate := fuzzy.HasDatePrefix(name)
			for i, ch := range name {
				char := string(ch)
				if hasDate && i < dirs.DatePrefixLen {
					b.WriteString(styles.Dim.Render(char))
				} else {
					b.WriteString(styles.Normal.Render(char))
				}
			}
		}

		b.WriteString("\n")
	}

	// Fake "Create new" line
	b.WriteString("  ")
	b.WriteString(styles.Success.Render(fmt.Sprintf("%s Create new: 2026-04-14-redis", styles.Symbols.Created)))

	return b.String()
}
