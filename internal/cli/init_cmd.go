package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hmk/try-bedazzled/internal/shell"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var triesPath string

	cmd := &cobra.Command{
		Use:   "init [path]",
		Short: "Output shell wrapper function for eval",
		Long: `Generates a shell function that wraps try for your shell.

Add this to your shell config (.bashrc, .zshrc, or config.fish):

  eval "$(try init)"
  eval "$(try init ~/my/custom/tries)"`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				triesPath = args[0]
			}

			if triesPath == "" {
				home, err := os.UserHomeDir()
				if err != nil {
					return fmt.Errorf("cannot determine home directory: %w", err)
				}
				triesPath = filepath.Join(home, "src", "tries")
			}

			// Expand ~ in path
			if len(triesPath) > 0 && triesPath[0] == '~' {
				home, _ := os.UserHomeDir()
				triesPath = filepath.Join(home, triesPath[1:])
			}

			// Get path to this executable
			selfPath, err := os.Executable()
			if err != nil {
				selfPath = "try" // fallback
			}
			selfPath, _ = filepath.EvalSymlinks(selfPath)

			shellEnv := os.Getenv("SHELL")
			shellType := shell.DetectShell(shellEnv)

			fmt.Print(shell.GenerateInit(shellType, selfPath, triesPath))
			return nil
		},
	}

	cmd.Flags().StringVar(&triesPath, "path", "", "Path to tries directory (default: ~/src/tries)")

	return cmd
}
