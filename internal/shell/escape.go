// Package shell handles shell script generation, escaping, and init wrapper output.
package shell

import "strings"

// Escape returns a shell-safe single-quoted string.
// Internal single quotes are escaped as '"'"' (close, escaped quote, open).
// This matches Python's shlex.quote() and the C version's shell_escape().
func Escape(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}
