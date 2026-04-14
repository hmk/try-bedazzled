// Package cli defines all Cobra commands for try-bedazzled.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags.
var Version = "dev"

var noColors bool

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "try",
		Short: "Manage ephemeral try directories with style",
		Long: `try-bedazzled — a beautiful, themeable directory manager.

Use try to quickly create, navigate, and manage scratch directories
for experiments, spikes, and one-off projects.

Setup: eval "$(try init [path])" in your shell config.`,
		Version: Version,
		// Default behavior: if args look like a query, run the selector
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			// Treat args as a query for the selector
			// This is handled by exec mode in the shell wrapper
			return fmt.Errorf("run 'try init' first to set up the shell wrapper")
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().BoolVar(&noColors, "no-colors", false, "Disable colors and use ASCII symbols")

	root.AddCommand(
		newInitCmd(),
		newExecCmd(),
		newCloneCmd(),
		newWorktreeCmd(),
		newThemeCmd(),
	)

	return root
}

// Execute runs the root command.
func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
