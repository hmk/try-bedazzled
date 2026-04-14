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
	return New("/tries", entries, filter, theme.Default())
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

func TestDeleteMarkedItemsOnEnter(t *testing.T) {
	m := newTestModel("")
	m = sendKeys(m, "ctrl+d") // mark first
	m = sendKeys(m, "down", "ctrl+d") // mark second
	m = sendKeys(m, "enter")

	if !m.Done() {
		t.Fatal("should be done")
	}
	r := m.GetResult()
	if r.Action != ActionDelete {
		t.Errorf("action should be Delete, got %d", r.Action)
	}
	if len(r.DeleteNames) != 2 {
		t.Errorf("expected 2 delete names, got %d", len(r.DeleteNames))
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
