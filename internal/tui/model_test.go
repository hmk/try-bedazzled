package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hmk/try-bedazzled/internal/dirs"
	"github.com/hmk/try-bedazzled/internal/theme"
)

var (
	now     = time.Date(2026, 4, 11, 12, 0, 0, 0, time.UTC)
	entries = []dirs.Entry{
		{Name: "2026-04-11-redis", Path: "/tries/2026-04-11-redis", Mtime: now.Add(-1 * time.Hour)},
		{Name: "2026-04-10-postgres", Path: "/tries/2026-04-10-postgres", Mtime: now.Add(-24 * time.Hour)},
		{Name: "2026-04-09-api-test", Path: "/tries/2026-04-09-api-test", Mtime: now.Add(-48 * time.Hour)},
		{Name: "2026-04-08-go-tui", Path: "/tries/2026-04-08-go-tui", Mtime: now.Add(-72 * time.Hour)},
	}
)

func newTestModel(filter string) Model {
	return New("/tries", entries, filter, theme.Default(), theme.Config{})
}

func sendKeys(m Model, keys ...string) Model {
	for _, k := range keys {
		var msg tea.KeyMsg
		switch k {
		case "up":
			msg = tea.KeyMsg{Type: tea.KeyUp}
		case "down":
			msg = tea.KeyMsg{Type: tea.KeyDown}
		case "enter":
			msg = tea.KeyMsg{Type: tea.KeyEnter}
		case "esc":
			msg = tea.KeyMsg{Type: tea.KeyEscape}
		case "backspace":
			msg = tea.KeyMsg{Type: tea.KeyBackspace}
		case "ctrl+d":
			msg = tea.KeyMsg{Type: tea.KeyCtrlD}
		case "ctrl+r":
			msg = tea.KeyMsg{Type: tea.KeyCtrlR}
		default:
			// Single character
			if len(k) == 1 {
				msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
			}
		}
		result, _ := m.Update(msg)
		m = result.(Model)
	}
	return m
}

// --- Initial state ---

func TestNewModelShowsAllEntries(t *testing.T) {
	m := newTestModel("")
	vis := m.visibleItems()
	if len(vis) != 4 {
		t.Errorf("expected 4 visible items, got %d", len(vis))
	}
}

func TestNewModelWithInitialFilter(t *testing.T) {
	m := newTestModel("redis")
	vis := m.visibleItems()
	// Should show redis match + "Create new" option
	if len(vis) < 1 {
		t.Error("expected at least 1 visible item with filter 'redis'")
	}
	// First visible should be the redis entry
	if vis[0] < 0 || vis[0] >= len(m.items) {
		t.Fatal("invalid visible index")
	}
	if m.items[vis[0]].entry.Name != "2026-04-11-redis" {
		t.Errorf("first item should be redis, got %s", m.items[vis[0]].entry.Name)
	}
}

func TestNewModelCursorStartsAtZero(t *testing.T) {
	m := newTestModel("")
	if m.cursor != 0 {
		t.Errorf("cursor should start at 0, got %d", m.cursor)
	}
}

// --- Navigation ---

func TestNavigateDown(t *testing.T) {
	m := newTestModel("")
	m = sendKeys(m, "down")
	if m.cursor != 1 {
		t.Errorf("cursor should be 1, got %d", m.cursor)
	}
}

func TestNavigateUp(t *testing.T) {
	m := newTestModel("")
	m = sendKeys(m, "down", "down", "up")
	if m.cursor != 1 {
		t.Errorf("cursor should be 1 after down-down-up, got %d", m.cursor)
	}
}

func TestNavigateUpAtTopStays(t *testing.T) {
	m := newTestModel("")
	m = sendKeys(m, "up")
	if m.cursor != 0 {
		t.Errorf("cursor should stay at 0, got %d", m.cursor)
	}
}

func TestNavigateDownAtBottomStays(t *testing.T) {
	m := newTestModel("")
	for i := 0; i < 20; i++ {
		m = sendKeys(m, "down")
	}
	max := m.visibleCount() - 1
	if m.cursor != max {
		t.Errorf("cursor should be at max %d, got %d", max, m.cursor)
	}
}

// --- Filtering ---

func TestFilterNarrowsList(t *testing.T) {
	m := newTestModel("")
	m = sendKeys(m, "r", "e", "d", "i", "s")

	vis := m.visibleItems()
	// Should show matched entries + "Create new"
	foundRedis := false
	for _, idx := range vis {
		if idx >= 0 && m.items[idx].entry.Name == "2026-04-11-redis" {
			foundRedis = true
		}
	}
	if !foundRedis {
		t.Error("redis should be visible after typing 'redis'")
	}
	if m.filter != "redis" {
		t.Errorf("filter should be 'redis', got %q", m.filter)
	}
}

func TestFilterNoMatch(t *testing.T) {
	m := newTestModel("")
	m = sendKeys(m, "z", "z", "z", "z")

	vis := m.visibleItems()
	// Should only have "Create new: YYYY-MM-DD-zzzz"
	if len(vis) != 1 || vis[0] != -1 {
		t.Errorf("expected only create-new option, got %v", vis)
	}
}

func TestBackspaceRemovesFilterChar(t *testing.T) {
	m := newTestModel("")
	m = sendKeys(m, "r", "e", "d")
	if m.filter != "red" {
		t.Fatalf("expected 'red', got %q", m.filter)
	}
	m = sendKeys(m, "backspace")
	if m.filter != "re" {
		t.Errorf("expected 're' after backspace, got %q", m.filter)
	}
}

func TestBackspaceOnEmptyFilterDoesNothing(t *testing.T) {
	m := newTestModel("")
	m = sendKeys(m, "backspace")
	if m.filter != "" {
		t.Errorf("filter should still be empty, got %q", m.filter)
	}
}

// --- Selection ---

func TestSelectFirstItem(t *testing.T) {
	m := newTestModel("")
	m = sendKeys(m, "enter")
	if !m.Done() {
		t.Fatal("should be done after enter")
	}
	r := m.GetResult()
	if r.Action != ActionCD {
		t.Errorf("action should be CD, got %d", r.Action)
	}
	if r.Path == "" {
		t.Error("path should not be empty")
	}
}

func TestSelectSecondItem(t *testing.T) {
	m := newTestModel("")
	m = sendKeys(m, "down", "enter")
	if !m.Done() {
		t.Fatal("should be done")
	}
	r := m.GetResult()
	if r.Action != ActionCD {
		t.Errorf("action should be CD, got %d", r.Action)
	}
}

func TestSelectCreateNew(t *testing.T) {
	m := newTestModel("")
	m = sendKeys(m, "m", "y", "a", "p", "p")

	// Navigate down to the "Create new" entry (last item)
	vis := m.visibleItems()
	createIdx := len(vis) - 1
	for i := 0; i < createIdx; i++ {
		m = sendKeys(m, "down")
	}
	m = sendKeys(m, "enter")

	if !m.Done() {
		t.Fatal("should be done")
	}
	r := m.GetResult()
	if r.Action != ActionMkdir {
		t.Errorf("action should be Mkdir, got %d", r.Action)
	}
	if !strings.Contains(r.Path, "myapp") {
		t.Errorf("path should contain 'myapp', got %q", r.Path)
	}
}

// --- Cancel ---

func TestEscCancels(t *testing.T) {
	m := newTestModel("")
	m = sendKeys(m, "esc")
	if !m.Done() {
		t.Fatal("should be done after esc")
	}
	r := m.GetResult()
	if r.Action != ActionCancel {
		t.Errorf("action should be Cancel, got %d", r.Action)
	}
}

// --- Delete marking ---

func TestCtrlDTogglesDeleteMark(t *testing.T) {
	m := newTestModel("")
	vis := m.visibleItems()
	firstIdx := vis[0]

	m = sendKeys(m, "ctrl+d")
	if !m.items[firstIdx].markedDelete {
		t.Error("item should be marked for delete")
	}

	m = sendKeys(m, "ctrl+d")
	if m.items[firstIdx].markedDelete {
		t.Error("item should be unmarked after second ctrl+d")
	}
}

func TestDeleteMarkedItemsGoesToConfirm(t *testing.T) {
	m := newTestModel("")
	m = sendKeys(m, "ctrl+d")         // mark first
	m = sendKeys(m, "down", "ctrl+d") // mark second
	m = sendKeys(m, "enter")

	// Should be in confirm mode, not done yet
	if m.Done() {
		t.Fatal("should not be done — should be in confirm mode")
	}
	if m.mode != ModeConfirmDelete {
		t.Errorf("should be in ModeConfirmDelete, got %d", m.mode)
	}
}

func TestDeleteConfirmYes(t *testing.T) {
	m := newTestModel("")
	m = sendKeys(m, "ctrl+d", "down", "ctrl+d", "enter") // mark 2, enter → confirm
	m = sendKeys(m, "y")                                 // confirm

	if !m.Done() {
		t.Fatal("should be done after confirm")
	}
	r := m.GetResult()
	if r.Action != ActionDelete {
		t.Errorf("action should be Delete, got %d", r.Action)
	}
	if len(r.DeleteNames) != 2 {
		t.Errorf("expected 2 delete names, got %d", len(r.DeleteNames))
	}
}

func TestDeleteConfirmNo(t *testing.T) {
	m := newTestModel("")
	m = sendKeys(m, "ctrl+d", "enter") // mark 1, enter → confirm
	m = sendKeys(m, "n")               // cancel

	if m.Done() {
		t.Error("should not be done after cancel")
	}
	if m.mode != ModeSelect {
		t.Errorf("should be back in ModeSelect, got %d", m.mode)
	}
}

func TestDeleteConfirmEsc(t *testing.T) {
	m := newTestModel("")
	m = sendKeys(m, "ctrl+d", "enter") // mark 1, enter → confirm
	m = sendKeys(m, "esc")             // cancel

	if m.Done() {
		t.Error("should not be done after esc cancel")
	}
	if m.mode != ModeSelect {
		t.Errorf("should be back in ModeSelect, got %d", m.mode)
	}
}

// --- Rename ---

func TestCtrlRStartsRename(t *testing.T) {
	m := newTestModel("")
	m = sendKeys(m, "ctrl+r")
	if m.mode != ModeRename {
		t.Error("should be in rename mode")
	}
	// Rename input should be pre-filled with slug (without date prefix)
	if m.renameInput == "" {
		t.Error("rename input should be pre-filled")
	}
}

func TestRenameEscCancels(t *testing.T) {
	m := newTestModel("")
	m = sendKeys(m, "ctrl+r")
	m = sendKeys(m, "esc")
	if m.mode != ModeSelect {
		t.Error("should be back in select mode after esc")
	}
	if m.Done() {
		t.Error("should not be done — rename was cancelled")
	}
}

func TestRenameEnterConfirms(t *testing.T) {
	m := newTestModel("")
	m = sendKeys(m, "ctrl+r")

	// Clear and type new name
	for range m.renameInput {
		m = sendKeys(m, "backspace")
	}
	m = sendKeys(m, "n", "e", "w", "enter")

	if !m.Done() {
		t.Fatal("should be done after rename enter")
	}
	r := m.GetResult()
	if r.Action != ActionRename {
		t.Errorf("action should be Rename, got %d", r.Action)
	}
	if r.RenameOld == "" {
		t.Error("old name should not be empty")
	}
	if !strings.Contains(r.RenameNew, "new") {
		t.Errorf("new name should contain 'new', got %q", r.RenameNew)
	}
}

// --- View rendering ---

func TestViewContainsEntryNames(t *testing.T) {
	m := newTestModel("")
	view := m.View()

	if !strings.Contains(view, "redis") {
		t.Error("view should contain 'redis'")
	}
	if !strings.Contains(view, "postgres") {
		t.Error("view should contain 'postgres'")
	}
}

func TestViewContainsStatusBar(t *testing.T) {
	m := newTestModel("")
	view := m.View()

	if !strings.Contains(view, "enter") {
		t.Error("view should contain key hints")
	}
	if !strings.Contains(view, "ctrl-d") {
		t.Error("view should contain ctrl-d hint")
	}
}

func TestViewShowsCreateNewWhenFiltered(t *testing.T) {
	m := newTestModel("")
	m = sendKeys(m, "m", "y", "a", "p", "p")
	view := m.View()

	if !strings.Contains(view, "Create new") {
		t.Error("view should show 'Create new' when filter is active")
	}
}

func TestViewShowsNoMatchesMessage(t *testing.T) {
	// Use a filter that matches nothing AND is invalid for dir name
	m := newTestModel("")
	m = sendKeys(m, "!")

	view := m.View()
	if !strings.Contains(view, "No matches") {
		t.Error("view should show 'No matches' when nothing matches and no create option")
	}
}

func TestViewDoneReturnsEmpty(t *testing.T) {
	m := newTestModel("")
	m = sendKeys(m, "esc")
	view := m.View()
	if view != "" {
		t.Errorf("view should be empty when done, got %q", view)
	}
}

func TestViewRenameMode(t *testing.T) {
	m := newTestModel("")
	m = sendKeys(m, "ctrl+r")
	view := m.View()

	if !strings.Contains(view, "Rename") {
		t.Error("rename mode view should contain 'Rename'")
	}
	if !strings.Contains(view, "New name") {
		t.Error("rename mode view should contain 'New name'")
	}
}

func TestViewConfirmDeleteMode(t *testing.T) {
	m := newTestModel("")
	m = sendKeys(m, "ctrl+d", "enter") // mark + enter → confirm
	view := m.View()

	if !strings.Contains(view, "Delete") {
		t.Error("confirm view should contain 'Delete'")
	}
	if !strings.Contains(view, "confirm") {
		t.Error("confirm view should contain 'confirm' hint")
	}
	if !strings.Contains(view, "cancel") {
		t.Error("confirm view should contain 'cancel' hint")
	}
}

func TestViewGhostAutocomplete(t *testing.T) {
	m := newTestModel("")
	m = sendKeys(m, "r", "e", "d")
	view := m.View()

	// Ghost should show remaining chars of "redis" after "red"
	// The ghost text "is" should appear somewhere (may be styled)
	_ = view // Ghost rendering is visual — hard to assert exact content
	// Just verify the view doesn't crash and contains our filter
	if !strings.Contains(view, "red") {
		t.Error("view should contain filter text 'red'")
	}
}

func TestViewEmptyState(t *testing.T) {
	m := New("/tries", nil, "", theme.Default(), theme.Config{})
	view := m.View()

	if !strings.Contains(view, "No tries yet") {
		t.Error("empty state should show 'No tries yet'")
	}
}

func TestViewSearchBarPlaceholder(t *testing.T) {
	m := newTestModel("")
	view := m.View()

	if !strings.Contains(view, "Type to filter") {
		t.Error("empty filter should show placeholder")
	}
}

func TestViewMatchCount(t *testing.T) {
	m := newTestModel("")
	m = sendKeys(m, "r", "e", "d", "i", "s")
	view := m.View()

	if !strings.Contains(view, "match") {
		t.Error("filtered view should show match count")
	}
}

// --- Cursor clamping ---

func TestCursorClampsAfterFilter(t *testing.T) {
	m := newTestModel("")
	// Move cursor to bottom
	for i := 0; i < 10; i++ {
		m = sendKeys(m, "down")
	}
	// Type filter that reduces visible items
	m = sendKeys(m, "r", "e", "d", "i", "s")
	// Cursor should be clamped
	if m.cursor >= m.visibleCount() {
		t.Errorf("cursor %d should be < visible count %d", m.cursor, m.visibleCount())
	}
}

// --- Space-as-dash fuzzy query ---

func TestFilterSpaceMatchesDash(t *testing.T) {
	dashed := []dirs.Entry{
		{Name: "2026-04-14-redis-cache", Path: "/t/2026-04-14-redis-cache", Mtime: now.Add(-1 * time.Hour)},
		{Name: "2026-04-13-go-api", Path: "/t/2026-04-13-go-api", Mtime: now.Add(-24 * time.Hour)},
	}
	m := New("/t", dashed, "redis cache", theme.Default(), theme.Config{})

	var matched []string
	for _, it := range m.items {
		if it.matched {
			matched = append(matched, it.entry.Name)
		}
	}
	if len(matched) == 0 || matched[0] != "2026-04-14-redis-cache" {
		t.Errorf("expected `redis cache` to match redis-cache entry, got %v", matched)
	}
}

func TestFilterSpaceMatchesDashSameAsDashed(t *testing.T) {
	dashed := []dirs.Entry{
		{Name: "2026-04-14-redis-cache", Path: "/t/2026-04-14-redis-cache", Mtime: now.Add(-1 * time.Hour)},
	}
	spaced := New("/t", dashed, "redis cache", theme.Default(), theme.Config{})
	hyph := New("/t", dashed, "redis-cache", theme.Default(), theme.Config{})
	// Recency uses wall-clock time.Now(), so scores can differ by a tiny epsilon.
	diff := spaced.items[0].score - hyph.items[0].score
	if diff < -1e-3 || diff > 1e-3 {
		t.Errorf("space and dash queries should score (nearly) identically: space=%v dash=%v",
			spaced.items[0].score, hyph.items[0].score)
	}
}

func TestGhostCompletionAcceptsSpace(t *testing.T) {
	dashed := []dirs.Entry{
		{Name: "2026-04-14-redis-cache", Path: "/t/2026-04-14-redis-cache", Mtime: now.Add(-1 * time.Hour)},
	}
	m := New("/t", dashed, "redis ca", theme.Default(), theme.Config{})
	if got := m.getGhostCompletion(); got != "che" {
		t.Errorf("ghost completion for `redis ca` = %q, want \"che\"", got)
	}
}

// --- Adaptive list height (fullscreen) ---

func TestEffectiveMaxVisibleInlineUsesThemeCap(t *testing.T) {
	m := newTestModel("")
	m.height = 80
	if got := m.effectiveMaxVisible(); got != m.maxVisible {
		t.Errorf("inline effectiveMaxVisible = %d, want %d", got, m.maxVisible)
	}
}

func TestEffectiveMaxVisibleFullscreenGrowsWithHeight(t *testing.T) {
	m := New("/t", entries, "", theme.Default(), theme.Config{DisplayMode: "fullscreen"})
	m.previewEnabled = false

	m.height = 20
	small := m.effectiveMaxVisible()
	m.height = 60
	big := m.effectiveMaxVisible()

	if big <= small {
		t.Errorf("fullscreen effectiveMaxVisible should grow with height: small=%d big=%d", small, big)
	}
}

func TestEffectiveMaxVisibleShrinksWhenPreviewOn(t *testing.T) {
	m := New("/t", entries, "", theme.Default(), theme.Config{DisplayMode: "fullscreen"})
	m.height = 60

	m.previewEnabled = false
	off := m.effectiveMaxVisible()
	m.previewEnabled = true
	on := m.effectiveMaxVisible()

	if off <= on {
		t.Errorf("fullscreen list should be larger with preview off: off=%d on=%d", off, on)
	}
}

func TestEffectiveMaxVisibleFloorsAtFive(t *testing.T) {
	m := New("/t", entries, "", theme.Default(), theme.Config{DisplayMode: "fullscreen"})
	m.previewEnabled = true
	m.height = 5
	if got := m.effectiveMaxVisible(); got < 5 {
		t.Errorf("effectiveMaxVisible floor should be >= 5, got %d", got)
	}
}
