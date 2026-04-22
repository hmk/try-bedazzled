package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hmk/try-bedazzled/internal/theme"
)

// settingsRow identifies one row in the settings screen.
type settingsRow int

const (
	rowTheme settingsRow = iota
	rowDisplayMode
	rowPreview
	rowEmojis
	rowCustomIcons
)

// settingsMode is the sub-mode the screen is in.
type settingsMode int

const (
	modeList      settingsMode = iota // browsing the list of settings
	modeThemePick                     // enter on "theme" — browse+preview themes
	modeIconInput                     // enter on "custom icons" — text input modal
)

// SettingsResult is what the caller reads after the model quits.
type SettingsResult struct {
	Saved  bool         // true if user pressed Esc (which saves); false on Ctrl-C
	Config theme.Config // current state
}

// SettingsModel is the Bubble Tea model for the settings screen.
type SettingsModel struct {
	cfg      theme.Config
	themes   []string
	previews map[string]theme.Theme

	cursor   int // which settingsRow is selected
	mode     settingsMode
	themeIdx int // index into m.themes when in modeThemePick

	// icon input modal state
	iconInput string

	width  int
	height int

	done   bool
	result SettingsResult
}

// NewSettings builds a settings model from the current config.
func NewSettings(cfg theme.Config) SettingsModel {
	names := theme.BuiltinNames()
	previews := make(map[string]theme.Theme, len(names))
	for _, n := range names {
		if t, err := theme.LoadBuiltin(n); err == nil {
			previews[n] = t
		}
	}
	active := cfg.Theme
	if active == "" {
		active = theme.DefaultThemeName
	}
	startIdx := 0
	for i, n := range names {
		if n == active {
			startIdx = i
			break
		}
	}

	return SettingsModel{
		cfg:      cfg,
		themes:   names,
		previews: previews,
		themeIdx: startIdx,
	}
}

// Init implements tea.Model.
func (m SettingsModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m SettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch m.mode {
		case modeList:
			return m.updateList(msg)
		case modeThemePick:
			return m.updateThemePick(msg)
		case modeIconInput:
			return m.updateIconInput(msg)
		}
	}
	return m, nil
}

// rowCount returns the total number of rows in the list view.
func (m SettingsModel) rowCount() int { return 5 }

func (m SettingsModel) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		// Abandon without saving.
		m.result = SettingsResult{Saved: false, Config: m.cfg}
		m.done = true
		return m, tea.Quit

	case "esc":
		// Save current config and quit.
		m.result = SettingsResult{Saved: true, Config: m.cfg}
		m.done = true
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < m.rowCount()-1 {
			m.cursor++
		}

	case "enter", " ":
		return m.activate()
	}
	return m, nil
}

// activate reacts to Enter on the highlighted row.
func (m SettingsModel) activate() (tea.Model, tea.Cmd) {
	switch settingsRow(m.cursor) {
	case rowTheme:
		m.mode = modeThemePick

	case rowDisplayMode:
		if m.cfg.GetDisplayMode() == "inline" {
			m.cfg.DisplayMode = "fullscreen"
		} else {
			m.cfg.DisplayMode = "inline"
		}

	case rowPreview:
		cur := m.cfg.GetPreviewEnabled()
		next := !cur
		m.cfg.PreviewEnabled = &next

	case rowEmojis:
		cur := m.cfg.GetShowEmojis(true)
		next := !cur
		m.cfg.ShowEmojis = &next

	case rowCustomIcons:
		m.mode = modeIconInput
		m.iconInput = ""
	}
	return m, nil
}

func (m SettingsModel) updateThemePick(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Cancel theme choice — keep old selection.
		m.mode = modeList
	case "enter":
		if m.themeIdx >= 0 && m.themeIdx < len(m.themes) {
			m.cfg.Theme = m.themes[m.themeIdx]
		}
		m.mode = modeList
	case "up", "k":
		if m.themeIdx > 0 {
			m.themeIdx--
		}
	case "down", "j":
		if m.themeIdx < len(m.themes)-1 {
			m.themeIdx++
		}
	}
	return m, nil
}

func (m SettingsModel) updateIconInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = modeList
		m.iconInput = ""
	case "enter":
		parts := strings.SplitN(m.iconInput, "=", 2)
		if len(parts) == 2 {
			word := strings.TrimSpace(strings.ToLower(parts[0]))
			emoji := strings.TrimSpace(parts[1])
			if word != "" && emoji != "" {
				if m.cfg.CustomIcons == nil {
					m.cfg.CustomIcons = make(map[string]string)
				}
				m.cfg.CustomIcons[word] = emoji
			}
		}
		m.mode = modeList
		m.iconInput = ""
	case "backspace":
		if len(m.iconInput) > 0 {
			m.iconInput = m.iconInput[:len(m.iconInput)-1]
		}
	default:
		s := msg.String()
		if len(s) == 1 && s[0] >= 32 && s[0] < 127 {
			m.iconInput += s
		}
	}
	return m, nil
}

// --- View ---

func (m SettingsModel) View() string {
	if m.done {
		return ""
	}

	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	bold := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF2D95"))
	key := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF2D95"))

	title := bold.Render("  Settings")
	sep := dim.Render(" ─────────────────────────────────────────────────────────────────────")

	rows := m.rows()
	left := m.viewList(rows)
	right := m.viewRight(rows)

	leftPane := lipgloss.NewStyle().Width(38).Render(left)
	rightPane := lipgloss.NewStyle().
		BorderLeft(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#555")).
		PaddingLeft(1).
		Render(right)

	panes := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	var status string
	switch m.mode {
	case modeList:
		status = "  " + key.Render("enter") + dim.Render(" edit") +
			dim.Render("  •  ") +
			key.Render("esc") + dim.Render(" save & close") +
			dim.Render("  •  ") +
			key.Render("ctrl-c") + dim.Render(" abandon")
	case modeThemePick:
		status = "  " + key.Render("↑/↓") + dim.Render(" browse") +
			dim.Render("  •  ") +
			key.Render("enter") + dim.Render(" apply") +
			dim.Render("  •  ") +
			key.Render("esc") + dim.Render(" cancel")
	case modeIconInput:
		status = "  " + key.Render("enter") + dim.Render(" add") +
			dim.Render("  •  ") +
			key.Render("esc") + dim.Render(" cancel")
	}

	return title + "\n" + sep + "\n" + panes + "\n" + sep + "\n" + status
}

type settingRow struct {
	label string
	value string
}

func (m SettingsModel) rows() []settingRow {
	themeLabel := m.cfg.Theme
	if themeLabel == "" {
		themeLabel = theme.DefaultThemeName + " (default)"
	}

	previewStr := "on"
	if !m.cfg.GetPreviewEnabled() {
		previewStr = "off"
	}
	emojiStr := "on"
	if !m.cfg.GetShowEmojis(true) {
		emojiStr = "off"
	}
	iconStr := fmt.Sprintf("%d defined", len(m.cfg.CustomIcons))

	return []settingRow{
		{"Theme", themeLabel},
		{"Display mode", m.cfg.GetDisplayMode()},
		{"Preview panel", previewStr},
		{"Emoji icons", emojiStr},
		{"Custom icons", iconStr},
	}
}

func (m SettingsModel) viewList(rows []settingRow) string {
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	cursor := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF2D95"))
	label := lipgloss.NewStyle().Foreground(lipgloss.Color("#F5F3FF"))

	var b strings.Builder
	for i, r := range rows {
		if i == m.cursor && m.mode == modeList {
			b.WriteString(cursor.Render("▸ "))
			b.WriteString(label.Render(padRight(r.label, 16)))
			b.WriteString(cursor.Render(r.value))
		} else {
			b.WriteString("  ")
			b.WriteString(label.Render(padRight(r.label, 16)))
			b.WriteString(dim.Render(r.value))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (m SettingsModel) viewRight(rows []settingRow) string {
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	bold := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F5F3FF"))

	switch m.mode {
	case modeThemePick:
		return m.viewThemePick()

	case modeIconInput:
		var b strings.Builder
		b.WriteString(bold.Render("Add custom icon") + "\n\n")
		b.WriteString(dim.Render("Format: word=emoji (e.g. django=🎭)") + "\n\n")
		b.WriteString("  " + m.iconInput + "▏\n\n")
		if len(m.cfg.CustomIcons) > 0 {
			b.WriteString(dim.Render("Current mappings:") + "\n")
			for word, emoji := range m.cfg.CustomIcons {
				_, _ = fmt.Fprintf(&b, "  %s = %s\n", word, emoji)
			}
		}
		return b.String()
	}

	// Default: context help for the selected row.
	switch settingsRow(m.cursor) {
	case rowTheme:
		active := m.cfg.Theme
		if active == "" {
			active = theme.DefaultThemeName
		}
		if t, ok := m.previews[active]; ok {
			return bold.Render("Preview: "+active) + "\n\n" + RenderPreview(t)
		}
		return dim.Render("(no preview available)")

	case rowDisplayMode:
		return bold.Render("Display mode") + "\n\n" +
			dim.Render("inline     — renders below the current prompt (default)\n") +
			dim.Render("fullscreen — takes over the screen (alt screen mode)\n\n") +
			dim.Render("Enter to cycle.")

	case rowPreview:
		return bold.Render("Preview panel") + "\n\n" +
			dim.Render("Shows a file tree of the highlighted directory.\n") +
			dim.Render("Toggle any time with Ctrl-P.\n\n") +
			dim.Render("Enter to toggle.")

	case rowEmojis:
		return bold.Render("Emoji icons") + "\n\n" +
			dim.Render("Content-aware folder icons (🐹 Go, 🦀 Rust, etc.).\n") +
			dim.Render("Turn off for terminals without emoji fonts.\n\n") +
			dim.Render("Enter to toggle.")

	case rowCustomIcons:
		return bold.Render("Custom icons") + "\n\n" +
			dim.Render("Add your own slug-word → emoji mappings.\n") +
			dim.Render("Example: django=🎭 makes '2026-04-16-django-app' show 🎭.\n\n") +
			dim.Render("Enter to add a new mapping.")
	}

	_ = rows
	return ""
}

func (m SettingsModel) viewThemePick() string {
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	bold := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F5F3FF"))
	active := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF2D95"))

	var b strings.Builder
	b.WriteString(bold.Render("Pick a theme") + "\n\n")
	for i, name := range m.themes {
		if i == m.themeIdx {
			b.WriteString(active.Render("  ▸ "+name) + "\n")
		} else {
			b.WriteString(dim.Render("    "+name) + "\n")
		}
	}

	if m.themeIdx >= 0 && m.themeIdx < len(m.themes) {
		t := m.previews[m.themes[m.themeIdx]]
		b.WriteString("\n" + RenderPreview(t))
	}
	return b.String()
}

// Done returns true after the user has finished editing.
func (m SettingsModel) Done() bool { return m.done }

// GetResult returns the final result.
func (m SettingsModel) GetResult() SettingsResult { return m.result }

func padRight(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}
