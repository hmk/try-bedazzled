package shell

import (
	"strings"
	"testing"
)

func TestBuildCDScript(t *testing.T) {
	script := BuildCDScript("/home/user/src/tries/2026-04-11-redis")

	// Should contain header
	if !strings.HasPrefix(script, ScriptHeader) {
		t.Error("missing script header")
	}

	// Should touch, cd, and printf
	if !strings.Contains(script, "touch '/home/user/src/tries/2026-04-11-redis'") {
		t.Error("missing touch command")
	}
	if !strings.Contains(script, "cd '/home/user/src/tries/2026-04-11-redis'") {
		t.Error("missing cd command")
	}
	if !strings.Contains(script, "printf '%s\\n' '/home/user/src/tries/2026-04-11-redis'") {
		t.Error("missing printf command")
	}
}

func TestBuildMkdirScript(t *testing.T) {
	script := BuildMkdirScript("/home/user/src/tries/2026-04-11-new")

	if !strings.Contains(script, "mkdir -p '/home/user/src/tries/2026-04-11-new'") {
		t.Error("missing mkdir command")
	}
	if !strings.Contains(script, "cd '/home/user/src/tries/2026-04-11-new'") {
		t.Error("missing cd command")
	}
}

func TestBuildCloneScript(t *testing.T) {
	script := BuildCloneScript("https://github.com/tobi/try-cli.git", "/home/user/src/tries/2026-04-11-tobi-try-cli")

	if !strings.Contains(script, "git clone 'https://github.com/tobi/try-cli.git'") {
		t.Error("missing git clone with URL")
	}
	if !strings.Contains(script, "cd '/home/user/src/tries/2026-04-11-tobi-try-cli'") {
		t.Error("missing cd command")
	}
}

func TestBuildWorktreeScript(t *testing.T) {
	script := BuildWorktreeScript("/home/user/src/tries/2026-04-11-feature")

	if !strings.Contains(script, "git worktree add '/home/user/src/tries/2026-04-11-feature'") {
		t.Error("missing git worktree command")
	}
	if !strings.Contains(script, "cd '/home/user/src/tries/2026-04-11-feature'") {
		t.Error("missing cd command")
	}
}

func TestBuildDeleteScript(t *testing.T) {
	script := BuildDeleteScript("/home/user/src/tries",
		[]string{"2026-04-11-redis", "2026-04-10-old"},
		"/home/user/projects")

	if !strings.Contains(script, "cd '/home/user/src/tries'") {
		t.Error("missing cd to base")
	}
	if !strings.Contains(script, "rm -rf '2026-04-11-redis'") {
		t.Error("missing delete of redis")
	}
	if !strings.Contains(script, "rm -rf '2026-04-10-old'") {
		t.Error("missing delete of old")
	}
	if !strings.Contains(script, "cd '/home/user/projects'") {
		t.Error("missing cwd restoration")
	}
}

func TestBuildDeleteScriptRejectsPathSeparators(t *testing.T) {
	script := BuildDeleteScript("/base", []string{"../etc/passwd"}, "/cwd")
	if script != "" {
		t.Error("should return empty script for names with path separators")
	}
}

func TestBuildDeleteScriptNoCwd(t *testing.T) {
	script := BuildDeleteScript("/base", []string{"dir"}, "")
	if !strings.Contains(script, `cd "$HOME"`) {
		t.Error("should fall back to $HOME when cwd is empty")
	}
}

func TestBuildRenameScript(t *testing.T) {
	script := BuildRenameScript("/home/user/src/tries", "2026-04-11-old", "2026-04-11-new")

	if !strings.Contains(script, "mv '2026-04-11-old' '2026-04-11-new'") {
		t.Error("missing mv command")
	}
	if !strings.Contains(script, "cd '/home/user/src/tries/2026-04-11-new'") {
		t.Error("missing cd to new path")
	}
}

func TestBuildRenameScriptRejectsPathSeparators(t *testing.T) {
	if s := BuildRenameScript("/base", "../evil", "good"); s != "" {
		t.Error("should reject old name with path separator")
	}
	if s := BuildRenameScript("/base", "good", "../evil"); s != "" {
		t.Error("should reject new name with path separator")
	}
}

func TestScriptsEscapeQuotesInPaths(t *testing.T) {
	// Path with a single quote — should be properly escaped
	script := BuildCDScript("/home/user/it's a test")
	// The full escaped path should be: '/home/user/it'"'"'s a test'
	escaped := Escape("/home/user/it's a test")
	if !strings.Contains(script, "touch "+escaped) {
		t.Errorf("single quotes in path not escaped properly: %s", script)
	}
}
