package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/hmk/try-bedazzled/internal/theme"
)

// Styles holds all Lip Gloss styles derived from a Theme.
type Styles struct {
	// Cursor is the style for the selection cursor symbol.
	Cursor lipgloss.Style

	// Selected is the style for the currently highlighted item row.
	Selected lipgloss.Style

	// Normal is the style for unselected item text.
	Normal lipgloss.Style

	// Dim is the style for secondary text (date prefixes, etc.).
	Dim lipgloss.Style

	// Match is the style for fuzzy-matched characters.
	Match lipgloss.Style

	// Danger is the style for items marked for deletion.
	Danger lipgloss.Style

	// Success is the style for the "Create new" entry.
	Success lipgloss.Style

	// SearchBar is the style for the search/filter input area.
	SearchBar lipgloss.Style

	// StatusBar is the style for the bottom status line.
	StatusBar lipgloss.Style

	// Title is the style for header text.
	Title lipgloss.Style

	// Symbols from the theme.
	Symbols theme.Symbols
}

// NewStyles creates Lip Gloss styles from a theme.
func NewStyles(t theme.Theme) Styles {
	s := Styles{
		Symbols: t.Symbols,
	}

	if t.Colors.Accent != "" {
		s.Cursor = lipgloss.NewStyle().Foreground(lipgloss.Color(t.Colors.Accent)).Bold(true)
		s.Selected = lipgloss.NewStyle().Foreground(lipgloss.Color(t.Colors.Accent))
	} else {
		s.Cursor = lipgloss.NewStyle().Bold(true)
		s.Selected = lipgloss.NewStyle()
	}

	if t.Colors.Text != "" {
		s.Normal = lipgloss.NewStyle().Foreground(lipgloss.Color(t.Colors.Text))
	} else {
		s.Normal = lipgloss.NewStyle()
	}

	if t.Colors.Dim != "" {
		s.Dim = lipgloss.NewStyle().Foreground(lipgloss.Color(t.Colors.Dim))
	} else {
		s.Dim = lipgloss.NewStyle().Faint(true)
	}

	if t.Colors.Match != "" {
		s.Match = lipgloss.NewStyle().Foreground(lipgloss.Color(t.Colors.Match)).Bold(true)
	} else {
		s.Match = lipgloss.NewStyle().Bold(true)
	}

	if t.Colors.Danger != "" {
		s.Danger = lipgloss.NewStyle().Foreground(lipgloss.Color(t.Colors.Danger))
	} else {
		s.Danger = lipgloss.NewStyle().Strikethrough(true)
	}

	if t.Colors.Success != "" {
		s.Success = lipgloss.NewStyle().Foreground(lipgloss.Color(t.Colors.Success))
	} else {
		s.Success = lipgloss.NewStyle().Italic(true)
	}

	s.SearchBar = lipgloss.NewStyle().
		Padding(0, 1).
		MarginBottom(1)

	s.StatusBar = lipgloss.NewStyle().
		Faint(true).
		MarginTop(1)

	s.Title = lipgloss.NewStyle().Bold(true).MarginBottom(1)

	return s
}
