package cli

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

// setupTTYRenderer configures Lip Gloss to detect color support from /dev/tty
// instead of stdout. This is necessary because in exec mode, stdout is captured
// by the shell wrapper (pipe), which makes Lip Gloss think there's no terminal.
// Returns a cleanup function to close the tty file.
func setupTTYRenderer() func() {
	ttyFile, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		return func() {} // no-op cleanup
	}
	renderer := lipgloss.NewRenderer(ttyFile)
	lipgloss.SetDefaultRenderer(renderer)
	return func() { ttyFile.Close() }
}
