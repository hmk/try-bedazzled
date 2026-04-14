package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hmk/try-bedazzled/internal/dirs"
	"github.com/hmk/try-bedazzled/internal/shell"
	"github.com/spf13/cobra"
)

func newWorktreeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "worktree <name>",
		Short: "Create a git worktree in a dated try directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("worktree must be run through the shell wrapper. Run 'try init' first")
		},
	}
}

func execWorktree(name string, triesPath string) error {
	datePrefix := time.Now().Format(dirs.DateFormat)
	dirName := datePrefix + "-" + name
	fullPath := filepath.Join(triesPath, dirName)

	// Check if we're in a git repo
	if isInGitRepo() {
		script := shell.BuildWorktreeScript(fullPath)
		fmt.Print(script)
	} else {
		script := shell.BuildMkdirScript(fullPath)
		fmt.Print(script)
	}
	return nil
}

func isInGitRepo() bool {
	dir, err := os.Getwd()
	if err != nil {
		return false
	}

	for {
		gitPath := filepath.Join(dir, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return false
		}
		dir = parent
	}
}
