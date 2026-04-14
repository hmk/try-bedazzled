// Package tui implements the interactive directory selector using Bubble Tea.
package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hmk/try-bedazzled/internal/dirs"
	"github.com/hmk/try-bedazzled/internal/fuzzy"
	"github.com/hmk/try-bedazzled/internal/theme"
)

// ActionType represents the user's chosen action.
type ActionType int

const (
	ActionNone ActionType = iota
	ActionCD
	ActionMkdir
	ActionCancel
	ActionDelete
	ActionRename
)

// Result is the outcome of the TUI selector.
type Result struct {
	Action       ActionType
	Path         string   // For CD/Mkdir: the target path
	DeleteNames  []string // For Delete: names to remove
	RenameOld    string   // For Rename: original name
	RenameNew    string   // For Rename: new name
}

// Mode tracks the current UI state.
type Mode int

const (
	ModeSelect Mode = iota
	ModeRename
)

// item is a scored, renderable directory entry.
type item struct {
	entry          dirs.Entry
	score          float64
	matchPositions []int
	matched        bool
	markedDelete   bool
}

// Model is the Bubble Tea model for the selector TUI.
type Model struct {
	// Config
	basePath string
	styles   Styles
	theme    theme.Theme

	// State
	items       []item
	allEntries  []dirs.Entry
	filter      string
	cursor      int
	mode        Mode
	renameInput string
	renameIdx   int // index of item being renamed

	// Output
	result Result
	done   bool

	// Layout
	maxVisible int
	width      int
	height     int
}

// New creates a new selector Model.
func New(basePath string, entries []dirs.Entry, initialFilter string, t theme.Theme) Model {
	m := Model{
		basePath:   basePath,
		allEntries: entries,
		filter:     initialFilter,
		styles:     NewStyles(t),
		theme:      t,
		maxVisible: t.Layout.MaxVisible,
	}
	if m.maxVisible == 0 {
		m.maxVisible = 12
	}
	m.items = buildItems(m.allEntries, m.filter)
	return m
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case ModeSelect:
			return m.updateSelect(msg)
		case ModeRename:
			return m.updateRename(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m Model) updateSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case keyMatch(msg, "ctrl+c", "esc"):
		m.result = Result{Action: ActionCancel}
		m.done = true
		return m, tea.Quit

	case keyMatch(msg, "up", "ctrl+p"):
		if m.cursor > 0 {
			m.cursor--
		}

	case keyMatch(msg, "down", "ctrl+n"):
		maxIdx := m.visibleCount() - 1
		if m.cursor < maxIdx {
			m.cursor++
		}

	case keyMatch(msg, "enter"):
		return m.selectCurrent()

	case keyMatch(msg, "ctrl+d"):
		m = m.toggleDelete()

	case keyMatch(msg, "ctrl+r"):
		m = m.startRename()

	case keyMatch(msg, "backspace"):
		if len(m.filter) > 0 {
			m.filter = m.filter[:len(m.filter)-1]
			m.items = buildItems(m.allEntries, m.filter)
			m.cursor = clampCursor(m.cursor, m.visibleCount())
		}

	default:
		// Printable character → add to filter
		str := msg.String()
		if len(str) == 1 && str[0] >= 32 && str[0] < 127 {
			m.filter += str
			m.items = buildItems(m.allEntries, m.filter)
			m.cursor = clampCursor(m.cursor, m.visibleCount())
		}
	}

	return m, nil
}

func (m Model) updateRename(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case keyMatch(msg, "esc"):
		m.mode = ModeSelect
		m.renameInput = ""

	case keyMatch(msg, "enter"):
		if m.renameInput != "" {
			entry := m.items[m.renameIdx].entry
			// Preserve date prefix if present
			newName := m.renameInput
			if _, _, ok := dirs.ParseDatePrefix(entry.Name); ok {
				datePrefix := entry.Name[:dirs.DatePrefixLen]
				newName = datePrefix + m.renameInput
			}
			m.result = Result{
				Action:    ActionRename,
				RenameOld: entry.Name,
				RenameNew: newName,
			}
			m.done = true
			return m, tea.Quit
		}
		m.mode = ModeSelect

	case keyMatch(msg, "backspace"):
		if len(m.renameInput) > 0 {
			m.renameInput = m.renameInput[:len(m.renameInput)-1]
		}

	default:
		str := msg.String()
		if len(str) == 1 && str[0] >= 32 && str[0] < 127 {
			m.renameInput += str
		}
	}

	return m, nil
}

func (m Model) selectCurrent() (tea.Model, tea.Cmd) {
	// Check if any items are marked for deletion
	var deleteNames []string
	for _, it := range m.items {
		if it.markedDelete {
			deleteNames = append(deleteNames, it.entry.Name)
		}
	}
	if len(deleteNames) > 0 {
		m.result = Result{
			Action:      ActionDelete,
			DeleteNames: deleteNames,
		}
		m.done = true
		return m, tea.Quit
	}

	visible := m.visibleItems()
	if m.cursor >= len(visible) {
		return m, nil
	}

	idx := visible[m.cursor]

	// "Create new" is a virtual item at the end when filter is active
	if idx == -1 {
		// Create new directory
		slug, ok := dirs.NormalizeDirName(m.filter)
		if !ok {
			return m, nil
		}
		name := dirs.FormatName(time.Now(), slug)
		path := m.basePath + "/" + name
		m.result = Result{Action: ActionMkdir, Path: path}
		m.done = true
		return m, tea.Quit
	}

	it := m.items[idx]
	m.result = Result{Action: ActionCD, Path: it.entry.Path}
	m.done = true
	return m, tea.Quit
}

func (m Model) toggleDelete() Model {
	visible := m.visibleItems()
	if m.cursor >= len(visible) {
		return m
	}
	idx := visible[m.cursor]
	if idx == -1 {
		return m // can't delete the "create new" option
	}
	m.items[idx].markedDelete = !m.items[idx].markedDelete
	return m
}

func (m Model) startRename() Model {
	visible := m.visibleItems()
	if m.cursor >= len(visible) {
		return m
	}
	idx := visible[m.cursor]
	if idx == -1 {
		return m
	}
	m.mode = ModeRename
	m.renameIdx = idx
	// Pre-fill with slug (strip date prefix if present)
	entry := m.items[idx].entry
	if _, slug, ok := dirs.ParseDatePrefix(entry.Name); ok {
		m.renameInput = slug
	} else {
		m.renameInput = entry.Name
	}
	return m
}

// buildItems rescores and sorts entries against the given filter.
func buildItems(allEntries []dirs.Entry, filter string) []item {
	now := time.Now()
	items := make([]item, len(allEntries))

	for i, entry := range allEntries {
		r := fuzzy.MatchAt(entry.Name, filter, entry.Mtime, now)
		items[i] = item{
			entry:          entry,
			score:          r.Score,
			matchPositions: r.MatchPositions,
			matched:        r.Matched,
		}
	}

	// Sort by score descending
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].score > items[j].score
	})

	return items
}

// visibleItems returns indices into m.items for items to display.
// Returns -1 as the index for the "Create new" virtual entry.
func (m Model) visibleItems() []int {
	var indices []int

	for i, it := range m.items {
		if m.filter == "" || it.matched {
			indices = append(indices, i)
		}
		if len(indices) >= m.maxVisible {
			break
		}
	}

	// Add "Create new" option if filter is active and produces a valid name
	if m.filter != "" {
		if _, ok := dirs.NormalizeDirName(m.filter); ok {
			indices = append(indices, -1) // -1 = create new
		}
	}

	return indices
}

func (m Model) visibleCount() int {
	return len(m.visibleItems())
}

// clampCursor ensures cursor doesn't exceed the visible item count.
func clampCursor(cursor, visibleCount int) int {
	max := visibleCount - 1
	if max < 0 {
		max = 0
	}
	if cursor > max {
		return max
	}
	return cursor
}

// Done returns true if the user has made a selection.
func (m Model) Done() bool {
	return m.done
}

// GetResult returns the selection result.
func (m Model) GetResult() Result {
	return m.result
}

// View implements tea.Model.
func (m Model) View() string {
	if m.done {
		return ""
	}

	var b strings.Builder

	// Search bar
	searchPrefix := "  "
	if m.filter != "" {
		searchPrefix = m.styles.Cursor.Render(m.styles.Symbols.Cursor) + " "
	}
	b.WriteString(m.styles.SearchBar.Render(searchPrefix + m.filter))
	b.WriteString("\n")

	if m.mode == ModeRename {
		return m.viewRename(&b)
	}

	// Items
	visible := m.visibleItems()
	if len(visible) == 0 {
		b.WriteString(m.styles.Dim.Render("  No matches"))
		b.WriteString("\n")
	}

	for row, idx := range visible {
		isSelected := row == m.cursor

		if idx == -1 {
			// "Create new" virtual entry
			slug, _ := dirs.NormalizeDirName(m.filter)
			name := dirs.FormatName(time.Now(), slug)
			label := fmt.Sprintf("  Create new: %s", name)
			if isSelected {
				cursor := m.styles.Cursor.Render(m.styles.Symbols.Cursor)
				label = fmt.Sprintf("%s Create new: %s", cursor, name)
				b.WriteString(m.styles.Success.Render(label))
			} else {
				b.WriteString(m.styles.Success.Render(label))
			}
			b.WriteString("\n")
			continue
		}

		it := m.items[idx]
		line := m.renderItem(it, isSelected)
		b.WriteString(line)
		b.WriteString("\n")
	}

	// Status bar
	hasDeletes := false
	for _, it := range m.items {
		if it.markedDelete {
			hasDeletes = true
			break
		}
	}

	var hints []string
	if hasDeletes {
		hints = append(hints, "enter: delete marked")
	} else {
		hints = append(hints, "enter: select")
	}
	hints = append(hints, "ctrl-d: mark delete", "ctrl-r: rename", "esc: cancel")
	b.WriteString(m.styles.StatusBar.Render("  " + strings.Join(hints, "  •  ")))

	return b.String()
}

func (m Model) viewRename(b *strings.Builder) string {
	entry := m.items[m.renameIdx].entry
	b.WriteString(fmt.Sprintf("  Rename: %s\n", m.styles.Dim.Render(entry.Name)))
	b.WriteString(fmt.Sprintf("  New name: %s▏\n", m.renameInput))
	b.WriteString(m.styles.StatusBar.Render("  enter: confirm  •  esc: cancel"))
	return b.String()
}

// renderItem renders a single directory entry with fuzzy highlights.
func (m Model) renderItem(it item, isSelected bool) string {
	var b strings.Builder

	// Cursor or indent
	if isSelected {
		b.WriteString(m.styles.Cursor.Render(m.styles.Symbols.Cursor) + " ")
	} else {
		b.WriteString("  ")
	}

	name := it.entry.Name
	matchSet := make(map[int]bool)
	for _, pos := range it.matchPositions {
		matchSet[pos] = true
	}

	hasDate := fuzzy.HasDatePrefix(name)
	showDate := m.theme.Layout.ShowDatePrefix

	for i, ch := range name {
		char := string(ch)

		if it.markedDelete {
			b.WriteString(m.styles.Danger.Render(char))
			continue
		}

		if matchSet[i] {
			// Matched character
			if isSelected {
				b.WriteString(m.styles.Match.Render(char))
			} else {
				b.WriteString(m.styles.Match.Render(char))
			}
		} else if hasDate && showDate && i < dirs.DatePrefixLen {
			// Date prefix — dimmed
			b.WriteString(m.styles.Dim.Render(char))
		} else if isSelected {
			b.WriteString(m.styles.Selected.Render(char))
		} else {
			b.WriteString(m.styles.Normal.Render(char))
		}
	}

	// Delete marker
	if it.markedDelete {
		b.WriteString(" " + m.styles.Danger.Render(m.styles.Symbols.Deleted))
	}

	return b.String()
}
