package shell

import "fmt"

// ShellType represents the user's shell.
type ShellType int

const (
	Bash ShellType = iota
	Zsh
	Fish
)

// DetectShell returns the shell type based on the SHELL environment variable value.
func DetectShell(shellEnv string) ShellType {
	// Check if it contains "fish" anywhere (e.g., /usr/bin/fish, /opt/homebrew/bin/fish)
	for i := 0; i <= len(shellEnv)-4; i++ {
		if shellEnv[i:i+4] == "fish" {
			return Fish
		}
	}
	return Bash // Bash and Zsh use the same syntax
}

// GenerateInit outputs the shell wrapper function that users eval in their shell config.
// selfPath is the absolute path to the try binary.
// triesPath is the default tries directory.
func GenerateInit(shell ShellType, selfPath, triesPath string) string {
	es := Escape(selfPath)
	et := Escape(triesPath)

	switch shell {
	case Fish:
		return fmt.Sprintf(`function try
  set -l out (%s exec --path %s $argv 2>/dev/tty)
  or begin; echo $out; return $status; end
  eval $out
end
`, es, et)

	default: // Bash / Zsh
		return fmt.Sprintf(`try() {
  local out
  out=$(%s exec --path %s "$@" 2>/dev/tty) || {
    echo "$out"
    return $?
  }
  eval "$out"
}
`, es, et)
	}
}
