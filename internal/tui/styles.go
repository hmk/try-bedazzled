package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/hmk/try-bedazzled/internal/theme"
)

// Styles holds all Lip Gloss styles derived from a Theme.
type Styles struct {
	// Entry rendering
	Cursor      lipgloss.Style
	Selected    lipgloss.Style
	SelectedRow lipgloss.Style // Background highlight for full row
	Normal      lipgloss.Style
	Dim         lipgloss.Style
	Match       lipgloss.Style
	Danger      lipgloss.Style
	Success     lipgloss.Style
	TimeText    lipgloss.Style // Relative time column
	Ghost       lipgloss.Style // Ghost autocomplete text

	// Chrome
	SearchBox  lipgloss.Style // Bordered search bar
	SearchBar  lipgloss.Style // Inner search content
	StatusBar  lipgloss.Style
	StatusKey  lipgloss.Style // Key name in status bar
	ConfirmBox lipgloss.Style // Delete confirmation dialog
	Title      lipgloss.Style
	ScrollHint lipgloss.Style // Scroll indicator

	// Symbols from the theme.
	Symbols theme.Symbols
}

// NewStyles creates Lip Gloss styles from a theme.
func NewStyles(t theme.Theme) Styles {
	s := Styles{
		Symbols: t.Symbols,
	}

	// Accent-based styles
	if t.Colors.Accent != "" {
		accentColor := lipgloss.Color(t.Colors.Accent)
		s.Cursor = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
		s.Selected = lipgloss.NewStyle().Foreground(accentColor)
		s.StatusKey = lipgloss.NewStyle().Foreground(accentColor).Bold(true)

		if t.Layout.SelectedBg {
			// Derive a subtle dark background from the accent
			s.SelectedRow = lipgloss.NewStyle().Background(lipgloss.Color("#1a1a2e"))
		} else {
			s.SelectedRow = lipgloss.NewStyle()
		}

		s.SearchBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accentColor).
			Padding(0, 1)
	} else {
		s.Cursor = lipgloss.NewStyle().Bold(true)
		s.Selected = lipgloss.NewStyle()
		s.SelectedRow = lipgloss.NewStyle()
		s.StatusKey = lipgloss.NewStyle().Bold(true)
		s.SearchBox = lipgloss.NewStyle().Padding(0, 1)
	}

	if t.Colors.Text != "" {
		s.Normal = lipgloss.NewStyle().Foreground(lipgloss.Color(t.Colors.Text))
	} else {
		s.Normal = lipgloss.NewStyle()
	}

	if t.Colors.Dim != "" {
		dimColor := lipgloss.Color(t.Colors.Dim)
		s.Dim = lipgloss.NewStyle().Foreground(dimColor)
		s.TimeText = lipgloss.NewStyle().Foreground(dimColor).Italic(true)
		s.Ghost = lipgloss.NewStyle().Foreground(dimColor)
		s.ScrollHint = lipgloss.NewStyle().Foreground(dimColor)
	} else {
		s.Dim = lipgloss.NewStyle().Faint(true)
		s.TimeText = lipgloss.NewStyle().Faint(true).Italic(true)
		s.Ghost = lipgloss.NewStyle().Faint(true)
		s.ScrollHint = lipgloss.NewStyle().Faint(true)
	}

	if t.Colors.Match != "" {
		s.Match = lipgloss.NewStyle().Foreground(lipgloss.Color(t.Colors.Match)).Bold(true)
	} else {
		s.Match = lipgloss.NewStyle().Bold(true)
	}

	if t.Colors.Danger != "" {
		dangerColor := lipgloss.Color(t.Colors.Danger)
		s.Danger = lipgloss.NewStyle().Foreground(dangerColor)
		s.ConfirmBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(dangerColor).
			Padding(1, 2)
	} else {
		s.Danger = lipgloss.NewStyle().Strikethrough(true)
		s.ConfirmBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2)
	}

	if t.Colors.Success != "" {
		s.Success = lipgloss.NewStyle().Foreground(lipgloss.Color(t.Colors.Success))
	} else {
		s.Success = lipgloss.NewStyle().Italic(true)
	}

	s.SearchBar = lipgloss.NewStyle().Padding(0, 1)

	s.StatusBar = lipgloss.NewStyle().
		Faint(true).
		MarginTop(1)

	s.Title = lipgloss.NewStyle().Bold(true).MarginBottom(1)

	return s
}
