package cli

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/hmk/try-bedazzled/internal/dirs"
	"github.com/hmk/try-bedazzled/internal/shell"
	"github.com/spf13/cobra"
)

func newCloneCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clone <url> [name]",
		Short: "Clone a git repo into a dated try directory",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("clone must be run through the shell wrapper. Run 'try init' first")
		},
	}
}

func execClone(url string, args []string, triesPath string) error {
	var customName string
	if len(args) > 0 {
		customName = args[0]
	}

	dirName := dirs.MakeCloneDirname(url, customName, time.Now())
	fullPath := filepath.Join(triesPath, dirName)

	script := shell.BuildCloneScript(url, fullPath)
	fmt.Print(script)
	return nil
}
