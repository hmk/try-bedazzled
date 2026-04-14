package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// keyMatch checks if a key message matches any of the given key bindings.
func keyMatch(msg tea.KeyMsg, keys ...string) bool {
	for _, k := range keys {
		if msg.String() == k {
			return true
		}
	}
	return false
}
