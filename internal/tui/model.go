// Package tui implements the interactive directory selector using Bubble Tea.
package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	ModeConfirmDelete
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
		case ModeConfirmDelete:
			return m.updateConfirmDelete(msg)
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
	// Check if any items are marked for deletion → show confirmation
	var deleteNames []string
	for _, it := range m.items {
		if it.markedDelete {
			deleteNames = append(deleteNames, it.entry.Name)
		}
	}
	if len(deleteNames) > 0 {
		m.mode = ModeConfirmDelete
		return m, nil
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

func (m Model) updateConfirmDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		var deleteNames []string
		for _, it := range m.items {
			if it.markedDelete {
				deleteNames = append(deleteNames, it.entry.Name)
			}
		}
		m.result = Result{
			Action:      ActionDelete,
			DeleteNames: deleteNames,
		}
		m.done = true
		return m, tea.Quit

	case "n", "esc", "ctrl+c":
		m.mode = ModeSelect
	}
	return m, nil
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

	switch m.mode {
	case ModeConfirmDelete:
		return m.viewConfirmDelete()
	case ModeRename:
		b.WriteString(m.viewSearchBar())
		return m.viewRename(&b)
	}

	// Search bar
	b.WriteString(m.viewSearchBar())

	// Items
	visible := m.visibleItems()

	// Empty state
	if len(m.allEntries) == 0 && m.filter == "" {
		sym := m.styles.Symbols.Created
		if sym == "" {
			sym = "+"
		}
		b.WriteString("\n")
		b.WriteString(m.styles.Success.Render(fmt.Sprintf("  %s No tries yet!", sym)))
		b.WriteString("\n")
		b.WriteString(m.styles.Dim.Render("  Type a name and press enter to create your first."))
		b.WriteString("\n")
	} else if len(visible) == 0 {
		b.WriteString(m.styles.Dim.Render("  No matches"))
		b.WriteString("\n")
	} else {
		for row, idx := range visible {
			isSelected := row == m.cursor

			if idx == -1 {
				// "Create new" virtual entry
				slug, _ := dirs.NormalizeDirName(m.filter)
				name := dirs.FormatName(time.Now(), slug)
				sym := m.styles.Symbols.Created
				if sym == "" {
					sym = "+"
				}
				if isSelected {
					cursor := m.styles.Cursor.Render(m.styles.Symbols.Cursor)
					b.WriteString(m.styles.Success.Render(fmt.Sprintf("%s %s Create new: %s", cursor, sym, name)))
				} else {
					b.WriteString(m.styles.Success.Render(fmt.Sprintf("  %s Create new: %s", sym, name)))
				}
				b.WriteString("\n")
				continue
			}

			it := m.items[idx]
			line := m.renderItem(it, isSelected)
			b.WriteString(line)
			b.WriteString("\n")
		}

		// Scroll indicator
		b.WriteString(m.viewScrollIndicator(visible))

		// File tree preview of highlighted entry
		b.WriteString(m.viewPreview(visible))
	}

	// Status bar
	b.WriteString(m.viewStatusBar())

	return b.String()
}

// viewSearchBar renders the search/filter bar based on the theme's search_style.
func (m Model) viewSearchBar() string {
	// Count matches
	matchCount := 0
	for _, it := range m.items {
		if m.filter == "" || it.matched {
			matchCount++
		}
	}

	// Build filter text with ghost autocomplete
	filterText := m.filter
	ghostText := ""
	if m.filter != "" {
		ghostText = m.getGhostCompletion()
	}

	cursor := m.styles.Symbols.Cursor
	if m.filter == "" {
		cursor = " "
	}

	var content string
	if m.filter == "" {
		content = m.styles.Dim.Render("  Type to filter...")
	} else {
		content = m.styles.Cursor.Render(cursor) + " " + filterText
		if ghostText != "" {
			content += m.styles.Ghost.Render(ghostText)
		}
	}

	countLabel := ""
	if m.filter != "" {
		countLabel = m.styles.Dim.Render(fmt.Sprintf(" %d matches", matchCount))
	}

	switch m.theme.Layout.SearchStyle {
	case "bordered":
		// Bordered box with rounded corners
		inner := content
		if countLabel != "" {
			inner += "  " + countLabel
		}
		return m.styles.SearchBox.Render(inner) + "\n"

	case "underline":
		line := content
		if countLabel != "" {
			line += "  " + countLabel
		}
		return line + "\n" + m.styles.Dim.Render(" ─────────────────────────────────────") + "\n"

	default: // "minimal"
		line := content
		if countLabel != "" {
			line += "  " + countLabel
		}
		return " " + line + "\n"
	}
}

// getGhostCompletion returns the ghost autocomplete text (like fish shell).
func (m Model) getGhostCompletion() string {
	if m.filter == "" || len(m.items) == 0 {
		return ""
	}

	filterLower := strings.ToLower(m.filter)

	// Find the best matching slug that starts with the filter
	for _, it := range m.items {
		if !it.matched {
			continue
		}
		// Extract slug from entry name
		slug := it.entry.Name
		if _, s, ok := dirs.ParseDatePrefix(it.entry.Name); ok {
			slug = s
		}
		slugLower := strings.ToLower(slug)
		if strings.HasPrefix(slugLower, filterLower) && len(slug) > len(m.filter) {
			return slug[len(m.filter):]
		}
	}
	return ""
}

// viewScrollIndicator shows scroll hints when items exceed max_visible.
func (m Model) viewScrollIndicator(visible []int) string {
	// Count total matching items (excluding create-new)
	totalMatching := 0
	for _, it := range m.items {
		if m.filter == "" || it.matched {
			totalMatching++
		}
	}

	if totalMatching <= m.maxVisible {
		return ""
	}

	hidden := totalMatching - m.maxVisible
	return m.styles.ScrollHint.Render(fmt.Sprintf("  ↕ %d more", hidden)) + "\n"
}

// viewPreview renders a file tree preview panel for the currently
// highlighted directory entry. Returns "" when no real entry is
// highlighted (e.g. on the "Create new" virtual item).
func (m Model) viewPreview(visible []int) string {
	if m.cursor >= len(visible) {
		return ""
	}
	idx := visible[m.cursor]
	if idx == -1 {
		// "Create new" — no preview
		return ""
	}
	entry := m.items[idx].entry

	// Render tree (cap at 8 lines, depth 2)
	tree := RenderFileTree(entry.Path, 2, 8, m.styles)
	if tree == "" {
		tree = m.styles.Dim.Render("  (empty)")
	}

	// Icon + name for the title
	icon := ""
	if m.theme.Layout.ShowIcons {
		slug := entry.Name
		if _, s, ok := dirs.ParseDatePrefix(entry.Name); ok {
			slug = s
		}
		icon = theme.LookupIcon(slug, m.styles.Symbols.Folder)
		if icon != "" {
			icon += " "
		}
	}
	title := fmt.Sprintf(" %s%s ", icon, entry.Name)

	// Bordered box with title in the top border
	box := m.styles.ConfirmBox.
		UnsetBorderForeground().
		BorderForeground(lipgloss.Color("#555")).
		Padding(0, 1).
		Render(tree)

	return title + "\n" + box + "\n"
}

// viewStatusBar renders context-sensitive key hints.
func (m Model) viewStatusBar() string {
	hasDeletes := false
	for _, it := range m.items {
		if it.markedDelete {
			hasDeletes = true
			break
		}
	}

	type hint struct {
		key  string
		desc string
	}

	var hints []hint
	if hasDeletes {
		hints = append(hints, hint{"enter", "delete marked"})
	} else {
		hints = append(hints, hint{"enter", "select"})
	}
	hints = append(hints,
		hint{"ctrl-d", "delete"},
		hint{"ctrl-r", "rename"},
		hint{"esc", "quit"},
	)

	var parts []string
	for _, h := range hints {
		parts = append(parts, m.styles.StatusKey.Render(h.key)+" "+m.styles.Dim.Render(h.desc))
	}

	separator := m.styles.Dim.Render(" ─────────────────────────────────────────")
	bar := "  " + strings.Join(parts, m.styles.Dim.Render("  •  "))
	return separator + "\n" + bar
}

// viewConfirmDelete renders the delete confirmation dialog.
func (m Model) viewConfirmDelete() string {
	var names []string
	for _, it := range m.items {
		if it.markedDelete {
			names = append(names, it.entry.Name)
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Delete %d %s?\n\n",
		len(names), pluralize(len(names), "directory", "directories"))

	for _, name := range names {
		sym := m.styles.Symbols.Deleted
		if sym == "" {
			sym = "x"
		}
		fmt.Fprintf(&b, "  %s %s\n", m.styles.Danger.Render(sym), m.styles.Danger.Render(name))
	}

	b.WriteString("\n")
	b.WriteString(m.styles.StatusKey.Render("y") + m.styles.Dim.Render(" confirm") +
		m.styles.Dim.Render("  •  ") +
		m.styles.StatusKey.Render("n") + m.styles.Dim.Render("/") +
		m.styles.StatusKey.Render("esc") + m.styles.Dim.Render(" cancel"))

	return m.styles.ConfirmBox.Render(b.String())
}

func (m Model) viewRename(b *strings.Builder) string {
	entry := m.items[m.renameIdx].entry
	fmt.Fprintf(b, "  Rename: %s\n", m.styles.Dim.Render(entry.Name))
	fmt.Fprintf(b, "  New name: %s▏\n", m.renameInput)
	b.WriteString(m.styles.Dim.Render(" ─────────────────────────────────────────") + "\n")
	b.WriteString("  " + m.styles.StatusKey.Render("enter") + m.styles.Dim.Render(" confirm") +
		m.styles.Dim.Render("  •  ") +
		m.styles.StatusKey.Render("esc") + m.styles.Dim.Render(" cancel"))
	return b.String()
}

// renderItem renders a single directory entry with column-based layout.
func (m Model) renderItem(it item, isSelected bool) string {
	var b strings.Builder

	// Cursor or indent
	if isSelected {
		b.WriteString(m.styles.Cursor.Render(m.styles.Symbols.Cursor) + " ")
	} else {
		b.WriteString("  ")
	}

	name := it.entry.Name
	hasDate := fuzzy.HasDatePrefix(name)

	// Extract slug and date parts
	var datePart, slugPart string
	if hasDate {
		datePart = name[:10] // "2026-04-11"
		slugPart = name[dirs.DatePrefixLen:]
	} else {
		slugPart = name
	}

	// Build match set adjusted for slug position
	matchSet := make(map[int]bool)
	for _, pos := range it.matchPositions {
		matchSet[pos] = true
	}

	// Render columns based on theme config
	columns := m.theme.Layout.Columns
	if len(columns) == 0 {
		columns = []string{"icon", "name", "date", "time"}
	}

	for ci, col := range columns {
		if ci > 0 {
			b.WriteString(" ")
		}

		switch col {
		case "icon":
			if m.theme.Layout.ShowIcons {
				if it.markedDelete {
					sym := m.styles.Symbols.Deleted
					if sym == "" {
						sym = "x"
					}
					b.WriteString(m.styles.Danger.Render(sym))
				} else {
					// Look up content-aware icon from slug, fall back to theme default
					icon := theme.LookupIcon(slugPart, m.styles.Symbols.Folder)
					if icon != "" {
						b.WriteString(icon)
					}
				}
			}

		case "name":
			m.renderName(&b, it, isSelected, slugPart, hasDate, name, matchSet)

		case "date":
			if datePart != "" && m.theme.Layout.ShowDate != "hide" {
				b.WriteString(m.styles.Dim.Render(datePart))
			}

		case "time":
			if m.theme.Layout.ShowTime {
				relTime := dirs.FormatRelativeTime(it.entry.Mtime)
				b.WriteString(m.styles.TimeText.Render(relTime))
			}
		}
	}

	// Delete marker (if no icon column showed it)
	if it.markedDelete && !m.theme.Layout.ShowIcons {
		b.WriteString(" " + m.styles.Danger.Render(m.styles.Symbols.Deleted))
	}

	return b.String()
}

// renderName renders the name/slug portion with fuzzy highlights.
func (m Model) renderName(b *strings.Builder, it item, isSelected bool, slugPart string, hasDate bool, fullName string, matchSet map[int]bool) {
	showDate := m.theme.Layout.ShowDate

	if showDate == "inline" && hasDate {
		// Inline: render full name with date dimmed
		for i, ch := range fullName {
			char := string(ch)
			if it.markedDelete {
				b.WriteString(m.styles.Danger.Render(char))
			} else if matchSet[i] {
				b.WriteString(m.styles.Match.Render(char))
			} else if i < dirs.DatePrefixLen {
				b.WriteString(m.styles.Dim.Render(char))
			} else if isSelected {
				b.WriteString(m.styles.Selected.Render(char))
			} else {
				b.WriteString(m.styles.Normal.Render(char))
			}
		}
	} else {
		// Separated: render just the slug, highlight matches adjusted
		offset := 0
		if hasDate {
			offset = dirs.DatePrefixLen
		}

		for i, ch := range slugPart {
			char := string(ch)
			origIdx := i + offset
			if it.markedDelete {
				b.WriteString(m.styles.Danger.Render(char))
			} else if matchSet[origIdx] {
				b.WriteString(m.styles.Match.Render(char))
			} else if isSelected {
				b.WriteString(m.styles.Selected.Render(char))
			} else {
				b.WriteString(m.styles.Normal.Render(char))
			}
		}
	}
}

func pluralize(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}
