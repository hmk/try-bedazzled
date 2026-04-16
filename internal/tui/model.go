// Package tui implements the interactive directory selector using Bubble Tea.
package tui

import (
	"fmt"
	"math"
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
	ActionOpenSettings
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
	basePath    string
	styles      Styles
	theme       theme.Theme
	customIcons map[string]string // User-defined slug→icon overrides

	// State
	items       []item
	allEntries  []dirs.Entry
	filter      string
	cursor      int
	offset      int // Index of first visible item for scrolling
	mode        Mode
	renameInput string
	renameIdx   int // index of item being renamed

	// Preference state
	previewEnabled bool

	// Callbacks (for persisting preferences that change at runtime)
	onPreviewToggle func(bool)

	// Output
	result Result
	done   bool

	// Layout
	maxVisible int
	width      int
	height     int
}

// New creates a new selector Model.
// cfg provides user preferences (custom icons, preview enabled, etc.).
// Pass theme.Config{} for defaults.
func New(basePath string, entries []dirs.Entry, initialFilter string, t theme.Theme, cfg theme.Config) Model {
	m := Model{
		basePath:       basePath,
		allEntries:     entries,
		filter:         initialFilter,
		styles:         NewStyles(t),
		theme:          t,
		customIcons:    cfg.CustomIcons,
		previewEnabled: cfg.GetPreviewEnabled(),
		maxVisible:     t.Layout.MaxVisible,
	}
	if m.maxVisible == 0 {
		m.maxVisible = 12
	}
	m.items = buildItems(m.allEntries, m.filter)
	return m
}

// SetPreviewToggleCallback sets a function that's called whenever the user
// toggles the preview panel. Used to persist the state to config.
func (m *Model) SetPreviewToggleCallback(fn func(bool)) {
	m.onPreviewToggle = fn
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

	case keyMatch(msg, "up", "ctrl+k"):
		// Move cursor up; scroll viewport if at top
		if m.cursor > 0 {
			m.cursor--
		} else if m.offset > 0 {
			m.offset--
		}

	case keyMatch(msg, "down", "ctrl+j"):
		// Move cursor down; scroll viewport if at bottom
		visible := m.visibleItems()
		lastVisibleIdx := len(visible) - 1
		if m.cursor < lastVisibleIdx {
			m.cursor++
		} else {
			// Cursor is at bottom of viewport — try to scroll
			// Only real items scroll (the "Create new" virtual entry
			// stays pinned at the bottom of the viewport)
			total := m.totalMatchingCount()
			realVisibleCount := m.maxVisible
			if m.hasCreateNew() {
				realVisibleCount--
			}
			maxOffset := total - realVisibleCount
			if maxOffset < 0 {
				maxOffset = 0
			}
			if m.offset < maxOffset {
				m.offset++
			}
		}

	case keyMatch(msg, "enter"):
		return m.selectCurrent()

	case keyMatch(msg, "ctrl+d"):
		m = m.toggleDelete()

	case keyMatch(msg, "ctrl+r"):
		m = m.startRename()

	case keyMatch(msg, "ctrl+p"):
		m.previewEnabled = !m.previewEnabled
		if m.onPreviewToggle != nil {
			m.onPreviewToggle(m.previewEnabled)
		}

	case keyMatch(msg, "ctrl+g"):
		m.result = Result{Action: ActionOpenSettings}
		m.done = true
		return m, tea.Quit

	case keyMatch(msg, "backspace"):
		if len(m.filter) > 0 {
			m.filter = m.filter[:len(m.filter)-1]
			m.items = buildItems(m.allEntries, m.filter)
			m.offset = 0 // reset scroll when filter changes
			m.cursor = clampCursor(m.cursor, m.visibleCount())
		}

	default:
		// Printable character → add to filter
		str := msg.String()
		if len(str) == 1 && str[0] >= 32 && str[0] < 127 {
			m.filter += str
			m.items = buildItems(m.allEntries, m.filter)
			m.offset = 0 // reset scroll when filter changes
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

// allMatchingIndices returns indices into m.items for all matching items
// (ignoring viewport/offset). Used for scroll math.
func (m Model) allMatchingIndices() []int {
	var all []int
	for i, it := range m.items {
		if m.filter == "" || it.matched {
			all = append(all, i)
		}
	}
	return all
}

// hasCreateNew returns true when the filter produces a valid slug
// and thus a "Create new" virtual entry is shown.
func (m Model) hasCreateNew() bool {
	if m.filter == "" {
		return false
	}
	_, ok := dirs.NormalizeDirName(m.filter)
	return ok
}

// visibleItems returns indices into m.items for items displayed in the
// current viewport. Returns -1 as the index for the "Create new" virtual
// entry which always appears at the bottom of the visible list.
func (m Model) visibleItems() []int {
	all := m.allMatchingIndices()

	start := m.offset
	if start > len(all) {
		start = len(all)
	}
	if start < 0 {
		start = 0
	}

	end := start + m.maxVisible
	// Reserve a slot for "Create new" when it applies
	if m.hasCreateNew() && end > start {
		end--
	}
	if end > len(all) {
		end = len(all)
	}

	var indices []int
	indices = append(indices, all[start:end]...)

	if m.hasCreateNew() {
		indices = append(indices, -1)
	}

	return indices
}

// visibleCount returns the number of rows currently rendered in the viewport
// (including the "Create new" virtual row when applicable).
func (m Model) visibleCount() int {
	return len(m.visibleItems())
}

// totalMatchingCount returns the total number of matching real items
// (excludes "Create new").
func (m Model) totalMatchingCount() int {
	return len(m.allMatchingIndices())
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

// clampOffset ensures offset doesn't scroll past the end of the list.
func clampOffset(offset, totalMatching, maxVisible int) int {
	if offset < 0 {
		return 0
	}
	maxOffset := totalMatching - maxVisible
	if maxOffset < 0 {
		maxOffset = 0
	}
	if offset > maxOffset {
		return maxOffset
	}
	return offset
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

// rainbowCursor returns the themed cursor glyph, colored along a rainbow arc
// when the active theme opts in (bedazzled). rowIndex lets each visible row
// pick a different hue so the cursor shifts as the user scrolls.
func (m Model) rainbowCursor(rowIndex int) string {
	sym := m.styles.Symbols.Cursor
	if !m.theme.Layout.Rainbow {
		return m.styles.Cursor.Render(sym)
	}
	hue := math.Mod(float64(rowIndex)*47.0, 360.0)
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(hsl2hex(hue, 0.8, 0.6))).
		Bold(true).
		Render(sym)
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
				// "Create new" virtual entry — preview the content-aware icon
				// so the user sees which emoji the new directory will get.
				slug, _ := dirs.NormalizeDirName(m.filter)
				name := dirs.FormatName(time.Now(), slug)
				fallback := m.styles.Symbols.Created
				if fallback == "" {
					fallback = "+"
				}
				icon := theme.LookupIcon(slug, fallback, m.customIcons)
				if isSelected {
					cursor := m.styles.Cursor.Render(m.styles.Symbols.Cursor)
					b.WriteString(m.styles.Success.Render(fmt.Sprintf("%s %s Create new: %s", cursor, icon, name)))
				} else {
					b.WriteString(m.styles.Success.Render(fmt.Sprintf("  %s Create new: %s", icon, name)))
				}
				b.WriteString("\n")
				continue
			}

			it := m.items[idx]
			line := m.renderItem(it, isSelected, row)
			b.WriteString(line)
			b.WriteString("\n")
		}

		// Scroll indicator
		b.WriteString(m.viewScrollIndicator())

		// File tree preview of highlighted entry (if enabled)
		if m.previewEnabled {
			b.WriteString(m.viewPreview(visible))
		}
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

	// Rule width adapts to the terminal, with a sensible fallback.
	ruleChars := 41
	if m.width > 4 {
		ruleChars = m.width - 2
	}
	var rule string
	if m.theme.Layout.Rainbow {
		rule = rainbowRule(ruleChars)
	} else {
		rule = " " + m.styles.Dim.Render(strings.Repeat("─", ruleChars))
	}

	switch m.theme.Layout.SearchStyle {
	case "bordered":
		// Top-and-bottom rules only (no left/right border)
		line := " " + content
		if countLabel != "" {
			line += "  " + countLabel
		}
		return rule + "\n" + line + "\n" + rule + "\n"

	case "underline":
		line := content
		if countLabel != "" {
			line += "  " + countLabel
		}
		return line + "\n" + rule + "\n"

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
// Displays counts of items above and below the viewport.
func (m Model) viewScrollIndicator() string {
	total := m.totalMatchingCount()
	realVisible := m.maxVisible
	if m.hasCreateNew() {
		realVisible--
	}
	if total <= realVisible {
		return ""
	}

	above := m.offset
	below := total - m.offset - realVisible
	if below < 0 {
		below = 0
	}

	switch {
	case above > 0 && below > 0:
		return m.styles.ScrollHint.Render(
			fmt.Sprintf("  ↑ %d above · ↓ %d below", above, below)) + "\n"
	case above > 0:
		return m.styles.ScrollHint.Render(
			fmt.Sprintf("  ↑ %d above", above)) + "\n"
	case below > 0:
		return m.styles.ScrollHint.Render(
			fmt.Sprintf("  ↓ %d below", below)) + "\n"
	}
	return ""
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
		icon = theme.LookupIcon(slug, m.styles.Symbols.Folder, m.customIcons)
		if icon != "" {
			icon += " "
		}
	}
	title := fmt.Sprintf(" %s%s ", icon, entry.Name)

	// Preview box spans most of the terminal width — 2-col gutter + 2 for
	// the box borders. Falls back to a reasonable default pre-WindowSizeMsg.
	boxWidth := m.width - 4
	if boxWidth < 40 {
		boxWidth = 40
	}

	box := m.styles.ConfirmBox.
		UnsetBorderForeground().
		BorderForeground(lipgloss.Color("#555")).
		Padding(0, 1).
		Width(boxWidth).
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
	previewLabel := "preview on"
	if !m.previewEnabled {
		previewLabel = "preview off"
	}
	hints = append(hints,
		hint{"ctrl-d", "delete"},
		hint{"ctrl-r", "rename"},
		hint{"ctrl-p", previewLabel},
		hint{"ctrl-g", "settings"},
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
// Left-side columns (icon, date, name) render first; the "time" (and optional
// fuzzy score) are right-aligned to the terminal width — matching tobi/try-cli's
// `"3h ago, 18.5"` tail format.
func (m Model) renderItem(it item, isSelected bool, rowIndex int) string {
	var left strings.Builder

	// Cursor or indent
	if isSelected {
		left.WriteString(m.rainbowCursor(rowIndex) + " ")
	} else {
		left.WriteString("  ")
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

	// Render columns based on theme config (left side only — no "time" here)
	columns := m.theme.Layout.Columns
	if len(columns) == 0 {
		columns = []string{"icon", "date", "name"}
	}

	wroteAny := false
	for _, col := range columns {
		if col == "time" {
			continue // time renders right-aligned below
		}
		if wroteAny {
			left.WriteString(" ")
		}
		wroteAny = true

		switch col {
		case "icon":
			if m.theme.Layout.ShowIcons {
				if it.markedDelete {
					sym := m.styles.Symbols.Deleted
					if sym == "" {
						sym = "x"
					}
					left.WriteString(m.styles.Danger.Render(sym))
				} else {
					icon := theme.LookupIcon(slugPart, m.styles.Symbols.Folder, m.customIcons)
					if icon != "" {
						left.WriteString(icon)
					}
				}
			}

		case "name":
			m.renderName(&left, it, isSelected, slugPart, hasDate, name, matchSet)

		case "date":
			if datePart != "" && m.theme.Layout.ShowDate != "hide" {
				left.WriteString(m.styles.Dim.Render(datePart))
			}
		}
	}

	// Delete marker (if no icon column showed it)
	if it.markedDelete && !m.theme.Layout.ShowIcons {
		left.WriteString(" " + m.styles.Danger.Render(m.styles.Symbols.Deleted))
	}

	leftStr := left.String()

	// Right-aligned tail: "3h ago" [, "12.4"]
	var tail string
	if m.theme.Layout.ShowTime {
		rel := dirs.FormatRelativeTime(it.entry.Mtime)
		if m.theme.Layout.ShowScore {
			tail = fmt.Sprintf("%s, %.1f", rel, it.score)
		} else {
			tail = rel
		}
	} else if m.theme.Layout.ShowScore {
		tail = fmt.Sprintf("%.1f", it.score)
	}

	if tail == "" {
		return leftStr
	}

	// Pad the middle with spaces so the tail hits the right edge.
	width := m.width
	if width <= 0 {
		width = 80 // sensible fallback when we haven't gotten a WindowSizeMsg yet
	}
	leftWidth := lipgloss.Width(leftStr)
	tailWidth := lipgloss.Width(tail)
	gap := width - leftWidth - tailWidth - 1 // 1-col right margin
	if gap < 2 {
		gap = 2 // never collide
	}

	return leftStr + strings.Repeat(" ", gap) + m.styles.TimeText.Render(tail)
}

// matchChar renders a single matched character — rainbow-hued when the theme
// opts in (each character gets a position-derived hue), otherwise the theme's
// Match style.
func (m Model) matchChar(ch string, matchPos int) string {
	if !m.theme.Layout.Rainbow {
		return m.styles.Match.Render(ch)
	}
	// Step the hue by a prime-ish offset so consecutive matches don't blur.
	hue := math.Mod(float64(matchPos)*61.0, 360.0)
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(hsl2hex(hue, 0.85, 0.62))).
		Bold(true).
		Render(ch)
}

// renderName renders the name/slug portion with fuzzy highlights.
func (m Model) renderName(b *strings.Builder, it item, isSelected bool, slugPart string, hasDate bool, fullName string, matchSet map[int]bool) {
	showDate := m.theme.Layout.ShowDate

	if showDate == "inline" && hasDate {
		// Inline: render full name with date dimmed
		matchIdx := 0
		for i, ch := range fullName {
			char := string(ch)
			if it.markedDelete {
				b.WriteString(m.styles.Danger.Render(char))
			} else if matchSet[i] {
				b.WriteString(m.matchChar(char, matchIdx))
				matchIdx++
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

		matchIdx := 0
		for i, ch := range slugPart {
			char := string(ch)
			origIdx := i + offset
			if it.markedDelete {
				b.WriteString(m.styles.Danger.Render(char))
			} else if matchSet[origIdx] {
				b.WriteString(m.matchChar(char, matchIdx))
				matchIdx++
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
