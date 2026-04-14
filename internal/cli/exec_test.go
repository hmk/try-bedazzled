package cli

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// --- isCloneURL ---

func TestIsCloneURL(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"https://github.com/tobi/try-cli.git", true},
		{"https://github.com/tobi/try-cli", true},
		{"http://github.com/tobi/try-cli.git", true},
		{"git@github.com:tobi/try-cli.git", true},
		{"git@gitlab.com:user/repo.git", true},
		{"redis", false},
		{"cd", false},
		{"clone", false},
		{"", false},
		{"./local-path", false},
		{"/absolute/path", false},
		{"https", false}, // prefix but not a URL
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isCloneURL(tt.input)
			if got != tt.want {
				t.Errorf("isCloneURL(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// --- parseKeyList ---

func TestParseKeyList(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		// Symbolic keys stay as-is
		{"ENTER", []string{"ENTER"}},
		{"ESCAPE", []string{"ESCAPE"}},
		{"UP", []string{"UP"}},
		{"DOWN", []string{"DOWN"}},
		{"BACKSPACE", []string{"BACKSPACE"}},
		{"TAB", []string{"TAB"}},
		{"SPACE", []string{"SPACE"}},

		// Literal text expands to individual characters
		{"redis", []string{"r", "e", "d", "i", "s"}},

		// Mixed: text then symbolic
		{"redis,ENTER", []string{"r", "e", "d", "i", "s", "ENTER"}},

		// Multiple symbolic
		{"UP,UP,DOWN,ENTER", []string{"UP", "UP", "DOWN", "ENTER"}},

		// Ctrl keys
		{"CTRL-D", []string{"CTRL-D"}},
		{"CTRL-R", []string{"CTRL-R"}},

		// Complex sequence
		{"redis,DOWN,DOWN,ENTER", []string{"r", "e", "d", "i", "s", "DOWN", "DOWN", "ENTER"}},

		// Whitespace handling
		{"redis , ENTER", []string{"r", "e", "d", "i", "s", "ENTER"}},

		// Empty parts ignored
		{"redis,,ENTER", []string{"r", "e", "d", "i", "s", "ENTER"}},

		// Empty string
		{"", nil},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseKeyList(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("parseKeyList(%q) = %v (len %d), want %v (len %d)",
					tt.input, got, len(got), tt.want, len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("parseKeyList(%q)[%d] = %q, want %q",
						tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

// --- parseKeyMsg ---

func TestParseKeyMsg(t *testing.T) {
	tests := []struct {
		input    string
		wantType tea.KeyType
	}{
		{"ENTER", tea.KeyEnter},
		{"ESCAPE", tea.KeyEscape},
		{"ESC", tea.KeyEscape},
		{"UP", tea.KeyUp},
		{"DOWN", tea.KeyDown},
		{"LEFT", tea.KeyLeft},
		{"RIGHT", tea.KeyRight},
		{"BACKSPACE", tea.KeyBackspace},
		{"TAB", tea.KeyTab},
		{"CTRL-C", tea.KeyCtrlC},
		{"CTRL-D", tea.KeyCtrlD},
		{"CTRL-R", tea.KeyCtrlR},
		{"CTRL-N", tea.KeyCtrlN},
		{"CTRL-P", tea.KeyCtrlP},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseKeyMsg(tt.input)
			if got.Type != tt.wantType {
				t.Errorf("parseKeyMsg(%q).Type = %v, want %v", tt.input, got.Type, tt.wantType)
			}
		})
	}
}

func TestParseKeyMsgSingleChar(t *testing.T) {
	msg := parseKeyMsg("a")
	if msg.Type != tea.KeyRunes {
		t.Errorf("expected KeyRunes, got %v", msg.Type)
	}
	if len(msg.Runes) != 1 || msg.Runes[0] != 'a' {
		t.Errorf("expected rune 'a', got %v", msg.Runes)
	}
}

func TestParseKeyMsgSpace(t *testing.T) {
	msg := parseKeyMsg("SPACE")
	if msg.Type != tea.KeyRunes {
		t.Errorf("expected KeyRunes for SPACE, got %v", msg.Type)
	}
	if len(msg.Runes) != 1 || msg.Runes[0] != ' ' {
		t.Errorf("expected space rune, got %v", msg.Runes)
	}
}

func TestParseKeyMsgCaseInsensitive(t *testing.T) {
	// Symbolic keys should work regardless of case
	enter := parseKeyMsg("enter")
	if enter.Type != tea.KeyEnter {
		t.Errorf("lowercase 'enter' should parse as KeyEnter, got %v", enter.Type)
	}

	ctrlD := parseKeyMsg("ctrl-d")
	if ctrlD.Type != tea.KeyCtrlD {
		t.Errorf("lowercase 'ctrl-d' should parse as KeyCtrlD, got %v", ctrlD.Type)
	}
}
