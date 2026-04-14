package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func sendPickerKeys(m ThemePickerModel, keys ...string) ThemePickerModel {
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
		default:
			if len(k) == 1 {
				msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
			}
		}
		result, _ := m.Update(msg)
		m = result.(ThemePickerModel)
	}
	return m
}

func TestThemePickerInitialState(t *testing.T) {
	m := NewThemePicker()
	if len(m.themes) < 4 {
		t.Fatalf("expected at least 4 themes, got %d", len(m.themes))
	}
	if m.cursor != 0 {
		t.Errorf("cursor should start at 0, got %d", m.cursor)
	}
	if m.done {
		t.Error("should not be done initially")
	}
}

func TestThemePickerNavigation(t *testing.T) {
	m := NewThemePicker()
	m = sendPickerKeys(m, "down")
	if m.cursor != 1 {
		t.Errorf("cursor should be 1 after down, got %d", m.cursor)
	}
	m = sendPickerKeys(m, "down", "up")
	if m.cursor != 1 {
		t.Errorf("cursor should be 1 after down-up, got %d", m.cursor)
	}
}

func TestThemePickerNavigationVim(t *testing.T) {
	m := NewThemePicker()
	m = sendPickerKeys(m, "j")
	if m.cursor != 1 {
		t.Errorf("j should move down, cursor = %d", m.cursor)
	}
	m = sendPickerKeys(m, "k")
	if m.cursor != 0 {
		t.Errorf("k should move up, cursor = %d", m.cursor)
	}
}

func TestThemePickerUpAtTopStays(t *testing.T) {
	m := NewThemePicker()
	m = sendPickerKeys(m, "up")
	if m.cursor != 0 {
		t.Errorf("cursor should stay at 0, got %d", m.cursor)
	}
}

func TestThemePickerDownAtBottomStays(t *testing.T) {
	m := NewThemePicker()
	for i := 0; i < 20; i++ {
		m = sendPickerKeys(m, "down")
	}
	if m.cursor != len(m.themes)-1 {
		t.Errorf("cursor should be at last theme, got %d", m.cursor)
	}
}

func TestThemePickerSelectFirst(t *testing.T) {
	m := NewThemePicker()
	m = sendPickerKeys(m, "enter")
	if !m.Done() {
		t.Fatal("should be done after enter")
	}
	r := m.GetResult()
	if !r.Selected {
		t.Error("should have selected a theme")
	}
	if r.Name == "" {
		t.Error("theme name should not be empty")
	}
	if r.Name != m.themes[0] {
		t.Errorf("should select first theme %q, got %q", m.themes[0], r.Name)
	}
}

func TestThemePickerSelectSecond(t *testing.T) {
	m := NewThemePicker()
	m = sendPickerKeys(m, "down", "enter")
	if !m.Done() {
		t.Fatal("should be done")
	}
	r := m.GetResult()
	if r.Name != m.themes[1] {
		t.Errorf("should select second theme %q, got %q", m.themes[1], r.Name)
	}
}

func TestThemePickerCancel(t *testing.T) {
	m := NewThemePicker()
	m = sendPickerKeys(m, "esc")
	if !m.Done() {
		t.Fatal("should be done after esc")
	}
	r := m.GetResult()
	if r.Selected {
		t.Error("should not have selected anything on cancel")
	}
}

func TestThemePickerCancelQ(t *testing.T) {
	m := NewThemePicker()
	m = sendPickerKeys(m, "q")
	if !m.Done() {
		t.Fatal("should be done after q")
	}
	r := m.GetResult()
	if r.Selected {
		t.Error("q should cancel")
	}
}

func TestThemePickerViewContainsThemeNames(t *testing.T) {
	m := NewThemePicker()
	view := m.View()

	if !strings.Contains(view, "default") {
		t.Error("view should show 'default' theme")
	}
	if !strings.Contains(view, "catppuccin") {
		t.Error("view should show 'catppuccin' theme")
	}
}

func TestThemePickerViewContainsPreview(t *testing.T) {
	m := NewThemePicker()
	view := m.View()

	// Preview should show fake entries
	if !strings.Contains(view, "redis") {
		t.Error("preview should contain 'redis' entry")
	}
	if !strings.Contains(view, "go-api") {
		t.Error("preview should contain 'go-api' entry")
	}
}

func TestThemePickerViewContainsHints(t *testing.T) {
	m := NewThemePicker()
	view := m.View()

	if !strings.Contains(view, "enter") {
		t.Error("view should contain key hints")
	}
	if !strings.Contains(view, "esc") {
		t.Error("view should contain esc hint")
	}
}

func TestThemePickerDoneViewEmpty(t *testing.T) {
	m := NewThemePicker()
	m = sendPickerKeys(m, "enter")
	if m.View() != "" {
		t.Error("done view should be empty")
	}
}
