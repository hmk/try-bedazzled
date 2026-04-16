package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
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
	table       table.Model
	tableStyles table.Styles
	themes      []string // theme names in order
	done        bool
	result      ThemePickerResult
	previews    map[string]theme.Theme
	width       int
	height      int
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

	// Build table
	columns := []table.Column{
		{Title: "Theme", Width: 14},
	}

	rows := make([]table.Row, len(names))
	for i, name := range names {
		rows[i] = table.Row{name}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(len(names)+1),
	)

	// Style the table
	s := table.DefaultStyles()
	s.Header = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED")).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#555"))
	s.Selected = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED"))
	s.Cell = lipgloss.NewStyle().
		PaddingLeft(1)
	t.SetStyles(s)

	return ThemePickerModel{
		table:       t,
		tableStyles: s,
		themes:      names,
		previews:    previews,
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

		case "enter":
			cursor := m.table.Cursor()
			if cursor >= 0 && cursor < len(m.themes) {
				m.result = ThemePickerResult{
					Selected: true,
					Name:     m.themes[cursor],
				}
			}
			m.done = true
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	// Delegate to table for navigation
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)

	// Update table selected style to match the previewed theme's accent
	cursor := m.table.Cursor()
	if cursor >= 0 && cursor < len(m.themes) {
		t := m.previews[m.themes[cursor]]
		if t.Colors.Accent != "" {
			m.tableStyles.Selected = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(t.Colors.Accent))
			m.table.SetStyles(m.tableStyles)
		}
	}

	return m, cmd
}

func (m ThemePickerModel) View() string {
	if m.done {
		return ""
	}

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Render("  Pick a theme:")

	// Top separator
	sep := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#555")).
		Render(" ─────────────────────────────────────────────────────────────")

	// Left pane: Bubbles table
	leftPane := lipgloss.NewStyle().
		Width(18).
		Render(m.table.View())

	// Right pane: live preview
	cursor := m.table.Cursor()
	var preview string
	if cursor >= 0 && cursor < len(m.themes) {
		selectedTheme := m.previews[m.themes[cursor]]
		preview = RenderPreview(selectedTheme)
	}

	rightPane := lipgloss.NewStyle().
		Width(48).
		BorderLeft(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#555")).
		PaddingLeft(1).
		Render(preview)

	panes := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	// Status bar
	status := lipgloss.NewStyle().
		Faint(true).
		Render("  enter apply  •  ↑/↓ browse  •  esc cancel")

	return title + "\n" + sep + "\n" + panes + "\n" + sep + "\n" + status
}

// Done returns true if the user has made a choice.
func (m ThemePickerModel) Done() bool {
	return m.done
}

// GetResult returns the picker result.
func (m ThemePickerModel) GetResult() ThemePickerResult {
	return m.result
}

// RenderPreview generates a fake try-list preview using the given theme.
// Used by both the standalone theme picker and the in-settings theme submode.
func RenderPreview(t theme.Theme) string {
	styles := NewStyles(t)

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

	// Search bar preview (simple, with top/bottom lines)
	searchLine := styles.Dim.Render("─────────────────────────────────────")
	b.WriteString(" " + searchLine + "\n")
	b.WriteString(" " + styles.Cursor.Render(styles.Symbols.Cursor) + " redis")
	b.WriteString(styles.Dim.Render("                    3 matches"))
	b.WriteString("\n")
	b.WriteString(" " + searchLine + "\n")

	for _, fe := range fakeEntries {
		if fe.cursor {
			b.WriteString(" " + styles.Cursor.Render(styles.Symbols.Cursor) + " ")
		} else {
			b.WriteString("   ")
		}

		// Icon
		name := fe.name
		slug := name
		if _, s, ok := dirs.ParseDatePrefix(name); ok {
			slug = s
		}
		icon := theme.LookupIcon(slug, styles.Symbols.Folder, nil)
		if icon != "" && t.Layout.ShowIcons {
			b.WriteString(icon + " ")
		}

		// Name with fuzzy highlights
		mtime := now.Add(-fe.mtime)
		if fe.query != "" {
			r := fuzzy.MatchAt(name, fe.query, mtime, now)
			matchSet := make(map[int]bool)
			for _, pos := range r.MatchPositions {
				matchSet[pos] = true
			}

			hasDate := fuzzy.HasDatePrefix(name)
			offset := 0
			if hasDate {
				offset = dirs.DatePrefixLen
			}
			for i, ch := range slug {
				char := string(ch)
				if matchSet[i+offset] {
					b.WriteString(styles.Match.Render(char))
				} else if fe.cursor {
					b.WriteString(styles.Selected.Render(char))
				} else {
					b.WriteString(styles.Normal.Render(char))
				}
			}
		} else {
			b.WriteString(styles.Normal.Render(slug))
		}

		// Date + time
		if t.Layout.ShowDate != "hide" {
			date := ""
			if _, _, ok := dirs.ParseDatePrefix(name); ok {
				date = name[:10]
			}
			if date != "" {
				b.WriteString("  " + styles.Dim.Render(date))
			}
		}
		if t.Layout.ShowTime {
			b.WriteString("  " + styles.TimeText.Render(dirs.FormatRelativeTime(now.Add(-fe.mtime))))
		}

		b.WriteString("\n")
	}

	// Create new line
	sym := styles.Symbols.Created
	if sym == "" {
		sym = "+"
	}
	b.WriteString("   " + styles.Success.Render(fmt.Sprintf("%s Create new: 2026-04-14-redis", sym)))

	return b.String()
}
