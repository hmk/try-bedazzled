package shell

import (
	"fmt"
	"strings"
)

// ScriptHeader is prepended to all eval-mode scripts so users who accidentally
// run the binary directly (instead of through the shell wrapper) see a hint.
const ScriptHeader = "# if you can read this, you didn't launch try from an alias. run try --help.\n"

// BuildCDScript generates a shell script that touches (updates mtime) and cd's
// into the given path.
func BuildCDScript(path string) string {
	ep := Escape(path)
	return fmt.Sprintf("%stouch %s && \\\n  cd %s && \\\n  printf '%%s\\n' %s\n",
		ScriptHeader, ep, ep, ep)
}

// BuildMkdirScript generates a shell script that creates a directory and cd's into it.
func BuildMkdirScript(path string) string {
	ep := Escape(path)
	return fmt.Sprintf("%smkdir -p %s && \\\n  cd %s && \\\n  printf '%%s\\n' %s\n",
		ScriptHeader, ep, ep, ep)
}

// BuildCloneScript generates a shell script that git clones a URL into a path
// and cd's into it.
func BuildCloneScript(url, path string) string {
	eu := Escape(url)
	ep := Escape(path)
	return fmt.Sprintf("%sgit clone %s %s && \\\n  cd %s && \\\n  printf '%%s\\n' %s\n",
		ScriptHeader, eu, ep, ep, ep)
}

// BuildWorktreeScript generates a shell script that creates a git worktree
// and cd's into it.
func BuildWorktreeScript(worktreePath string) string {
	ep := Escape(worktreePath)
	return fmt.Sprintf("%sgit worktree add %s && \\\n  cd %s && \\\n  printf '%%s\\n' %s\n",
		ScriptHeader, ep, ep, ep)
}

// BuildDeleteScript generates a shell script that deletes the named directories
// under basePath. Names must not contain path separators.
func BuildDeleteScript(basePath string, names []string, cwd string) string {
	// Security: reject names with path separators
	for _, name := range names {
		if strings.Contains(name, "/") {
			return ""
		}
	}

	var b strings.Builder
	b.WriteString(ScriptHeader)

	eb := Escape(basePath)
	fmt.Fprintf(&b, "cd %s && \\\n", eb)

	for _, name := range names {
		en := Escape(name)
		fmt.Fprintf(&b, "  [[ -d %s ]] && rm -rf %s && \\\n", en, en)
	}

	// Restore cwd or fall back to $HOME
	if cwd != "" {
		ec := Escape(cwd)
		fmt.Fprintf(&b, "  ( cd %s 2>/dev/null || cd \"$HOME\" )\n", ec)
	} else {
		b.WriteString("  cd \"$HOME\"\n")
	}

	return b.String()
}

// BuildRenameScript generates a shell script that renames a directory and cd's
// into the new name.
func BuildRenameScript(basePath, oldName, newName string) string {
	// Security: names must not contain path separators
	if strings.Contains(oldName, "/") || strings.Contains(newName, "/") {
		return ""
	}

	eb := Escape(basePath)
	eo := Escape(oldName)
	en := Escape(newName)
	newPath := basePath + "/" + newName
	enp := Escape(newPath)

	return fmt.Sprintf("%scd %s && \\\nmv %s %s && \\\n  cd %s && \\\n  printf '%%s\\n' %s\n",
		ScriptHeader, eb, eo, en, enp, enp)
}
