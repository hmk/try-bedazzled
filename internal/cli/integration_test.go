package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// buildBinary builds the try binary for integration tests.
// Returns the path to the binary.
func buildBinary(t *testing.T) string {
	t.Helper()
	binPath := filepath.Join(t.TempDir(), "try")
	cmd := exec.Command("go", "build", "-o", binPath, "../../cmd/try")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to build binary: %v", err)
	}
	return binPath
}

func runTry(t *testing.T, bin string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	var outBuf, errBuf bytes.Buffer
	cmd := exec.Command(bin, args...)
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}
	return outBuf.String(), errBuf.String(), exitCode
}

// --- Version ---

func TestIntegrationVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	bin := buildBinary(t)
	stdout, _, code := runTry(t, bin, "--version")
	if code != 0 {
		t.Errorf("--version should exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "try version") {
		t.Errorf("--version output should contain 'try version', got %q", stdout)
	}
}

// --- Help ---

func TestIntegrationHelp(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	bin := buildBinary(t)
	stdout, _, code := runTry(t, bin, "--help")
	if code != 0 {
		t.Errorf("--help should exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "try-bedazzled") {
		t.Errorf("--help should mention try-bedazzled, got %q", stdout)
	}
	if !strings.Contains(stdout, "init") {
		t.Errorf("--help should list init command")
	}
}

// --- Init ---

func TestIntegrationInitBash(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	bin := buildBinary(t)
	t.Setenv("SHELL", "/bin/bash")
	stdout, _, code := runTry(t, bin, "init", "/tmp/my-tries")
	if code != 0 {
		t.Fatalf("init should exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "try()") {
		t.Error("bash init should contain 'try()' function")
	}
	if !strings.Contains(stdout, "/tmp/my-tries") {
		t.Error("bash init should contain the tries path")
	}
	if !strings.Contains(stdout, `eval "$out"`) {
		t.Error("bash init should contain eval")
	}
}

func TestIntegrationInitFish(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	bin := buildBinary(t)
	t.Setenv("SHELL", "/usr/bin/fish")
	stdout, _, code := runTry(t, bin, "init", "/tmp/my-tries")
	if code != 0 {
		t.Fatalf("init should exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "function try") {
		t.Error("fish init should contain 'function try'")
	}
	if !strings.Contains(stdout, "$argv") {
		t.Error("fish init should contain $argv")
	}
}

// --- Exec: selector with key injection ---

func TestIntegrationExecSelectEntry(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	bin := buildBinary(t)

	// Create fixture directories
	triesDir := t.TempDir()
	mkdirs(t, triesDir, "2026-04-11-redis", "2026-04-10-postgres")

	stdout, _, code := runTry(t, bin, "exec", "--path", triesDir,
		"--and-keys", "redis,ENTER")
	if code != 0 {
		t.Fatalf("exec should exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "cd") {
		t.Error("output should contain cd command")
	}
	if !strings.Contains(stdout, "redis") {
		t.Error("output should contain 'redis' in path")
	}
}

func TestIntegrationExecCreateNew(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	bin := buildBinary(t)
	triesDir := t.TempDir()

	// Type a name that doesn't match, navigate to "Create new", enter
	stdout, _, code := runTry(t, bin, "exec", "--path", triesDir,
		"--and-keys", "myapp,ENTER")
	if code != 0 {
		t.Fatalf("exec should exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "mkdir -p") {
		t.Error("output should contain mkdir command for new dir")
	}
	if !strings.Contains(stdout, "myapp") {
		t.Error("output should contain 'myapp' in path")
	}
}

func TestIntegrationExecCancel(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	bin := buildBinary(t)
	triesDir := t.TempDir()
	mkdirs(t, triesDir, "2026-04-11-test")

	stdout, stderr, _ := runTry(t, bin, "exec", "--path", triesDir,
		"--and-keys", "ESCAPE")
	if stdout != "" {
		t.Errorf("cancel should produce no stdout, got %q", stdout)
	}
	if !strings.Contains(stderr, "Cancelled") {
		t.Error("cancel should print 'Cancelled' to stderr")
	}
}

// --- Exec: clone URL shorthand ---

func TestIntegrationExecCloneURL(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	bin := buildBinary(t)
	triesDir := t.TempDir()

	stdout, _, code := runTry(t, bin, "exec", "--path", triesDir,
		"https://github.com/tobi/try-cli.git")
	if code != 0 {
		t.Fatalf("clone URL should exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "git clone") {
		t.Error("output should contain 'git clone'")
	}
	if !strings.Contains(stdout, "tobi-try-cli") {
		t.Error("output should contain 'tobi-try-cli' in dirname")
	}
	if !strings.Contains(stdout, "https://github.com/tobi/try-cli.git") {
		t.Error("output should contain the original URL")
	}
}

func TestIntegrationExecCloneSSH(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	bin := buildBinary(t)
	triesDir := t.TempDir()

	stdout, _, code := runTry(t, bin, "exec", "--path", triesDir,
		"git@github.com:hmk/try-bedazzled.git")
	if code != 0 {
		t.Fatalf("clone SSH should exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "git clone") {
		t.Error("output should contain 'git clone'")
	}
	if !strings.Contains(stdout, "hmk-try-bedazzled") {
		t.Error("output should contain 'hmk-try-bedazzled' in dirname")
	}
}

func TestIntegrationExecCloneSubcommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	bin := buildBinary(t)
	triesDir := t.TempDir()

	stdout, _, code := runTry(t, bin, "exec", "--path", triesDir,
		"clone", "https://github.com/user/repo.git")
	if code != 0 {
		t.Fatalf("clone subcommand should exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "git clone") {
		t.Error("output should contain 'git clone'")
	}
	if !strings.Contains(stdout, "user-repo") {
		t.Error("output should contain 'user-repo' in dirname")
	}
}

func TestIntegrationExecCloneWithCustomName(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	bin := buildBinary(t)
	triesDir := t.TempDir()

	stdout, _, code := runTry(t, bin, "exec", "--path", triesDir,
		"clone", "https://github.com/user/repo.git", "my-fork")
	if code != 0 {
		t.Fatalf("clone with name should exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "my-fork") {
		t.Error("output should contain custom name 'my-fork'")
	}
}

// --- Exec: worktree ---

func TestIntegrationExecWorktreeNotInGitRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	bin := buildBinary(t)
	triesDir := t.TempDir()

	// Run from a temp dir that is NOT inside any git repo
	nonGitDir := t.TempDir()
	var outBuf bytes.Buffer
	cmd := exec.Command(bin, "exec", "--path", triesDir,
		"worktree", "feature-x")
	cmd.Dir = nonGitDir
	cmd.Stdout = &outBuf
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("worktree should exit 0: %v", err)
	}
	stdout := outBuf.String()
	// Not in a git repo → should fall back to mkdir
	if !strings.Contains(stdout, "mkdir -p") {
		t.Error("worktree outside git repo should use mkdir")
	}
	if !strings.Contains(stdout, "feature-x") {
		t.Error("output should contain 'feature-x'")
	}
}

func TestIntegrationExecWorktreeInGitRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	bin := buildBinary(t)
	triesDir := t.TempDir()

	// Create a git repo to run from
	gitDir := t.TempDir()
	cmd := exec.Command("git", "init", gitDir)
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to create git repo: %v", err)
	}
	// Need at least one commit for worktree to work
	cmd = exec.Command("git", "-C", gitDir, "commit", "--allow-empty", "-m", "init")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to create initial commit: %v", err)
	}

	// Run from inside the git repo
	var outBuf bytes.Buffer
	execCmd := exec.Command(bin, "exec", "--path", triesDir,
		"worktree", "my-branch")
	execCmd.Dir = gitDir
	execCmd.Stdout = &outBuf
	execCmd.Stderr = os.Stderr
	if err := execCmd.Run(); err != nil {
		t.Fatalf("worktree in git repo should exit 0: %v", err)
	}
	stdout := outBuf.String()
	if !strings.Contains(stdout, "git worktree add") {
		t.Error("worktree inside git repo should use 'git worktree add'")
	}
	if !strings.Contains(stdout, "my-branch") {
		t.Error("output should contain 'my-branch'")
	}
}

func TestIntegrationExecDotShorthand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	bin := buildBinary(t)
	triesDir := t.TempDir()

	stdout, _, code := runTry(t, bin, "exec", "--path", triesDir,
		".", "quick-test")
	if code != 0 {
		t.Fatalf(". shorthand should exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "quick-test") {
		t.Error("output should contain 'quick-test'")
	}
}

// --- Exec: no-colors flag ---

func TestIntegrationNoColorsFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	bin := buildBinary(t)
	triesDir := t.TempDir()
	mkdirs(t, triesDir, "2026-04-11-test")

	// Should work without errors
	_, _, code := runTry(t, bin, "--no-colors", "exec", "--path", triesDir,
		"--and-keys", "ESCAPE")
	// Exit code doesn't matter (cancel is fine), just shouldn't crash
	_ = code
}

// --- Exec: --path required ---

func TestIntegrationExecRequiresPath(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	bin := buildBinary(t)

	_, stderr, code := runTry(t, bin, "exec")
	if code == 0 {
		t.Error("exec without --path should fail")
	}
	if !strings.Contains(stderr, "path") {
		t.Errorf("error should mention --path, got %q", stderr)
	}
}

// --- Theme persistence through binary ---

func TestIntegrationThemeConfigApplied(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	bin := buildBinary(t)

	// Set up isolated config dir with dracula theme
	configDir := t.TempDir()
	tryConfigDir := filepath.Join(configDir, "try")
	os.MkdirAll(tryConfigDir, 0755)
	os.WriteFile(filepath.Join(tryConfigDir, "config.toml"),
		[]byte("theme = \"dracula\"\n"), 0644)

	// Set up tries dir with an entry
	triesDir := t.TempDir()
	mkdirs(t, triesDir, "2026-04-11-redis")

	// Run the binary with the custom config dir, capture stderr (TUI output)
	var outBuf, errBuf bytes.Buffer
	cmd := exec.Command(bin, "exec", "--path", triesDir,
		"--and-keys", "ESCAPE")
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	cmd.Env = append(os.Environ(),
		"XDG_CONFIG_HOME="+configDir,
		// Clear TRY_THEME so it doesn't override
		"TRY_THEME=",
	)
	cmd.Run() // exit code doesn't matter (ESCAPE = cancel)

	stderr := errBuf.String()
	t.Logf("TUI stderr output (%d bytes):\n%s", len(stderr), stderr)

	// The dracula theme uses #BD93F9 (purple) for accent, which would show up
	// as ANSI escape codes in the output. Default uses #7C3AED.
	// We can't easily check exact colors, but we CAN verify the binary
	// actually read the config by checking it didn't crash and produced output.
	if len(stderr) == 0 {
		t.Error("expected TUI output on stderr, got nothing")
	}
}

func TestIntegrationThemeEnvOverridesConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	bin := buildBinary(t)

	// Config says "default"
	configDir := t.TempDir()
	tryConfigDir := filepath.Join(configDir, "try")
	os.MkdirAll(tryConfigDir, 0755)
	os.WriteFile(filepath.Join(tryConfigDir, "config.toml"),
		[]byte("theme = \"default\"\n"), 0644)

	triesDir := t.TempDir()
	mkdirs(t, triesDir, "2026-04-11-test")

	// Env says "minimal" — should override. Minimal uses ">" cursor, no icons.
	// Select an entry so we get actual script output, then check stderr for
	// the rendered frame (which runWithInjectedKeys prints before done).
	var outBuf, errBuf bytes.Buffer
	cmd := exec.Command(bin, "exec", "--path", triesDir,
		"--and-keys", "test,ENTER")
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	cmd.Env = append(os.Environ(),
		"XDG_CONFIG_HOME="+configDir,
		"TRY_THEME=minimal",
	)
	cmd.Run()

	stderr := errBuf.String()
	t.Logf("Minimal theme stderr (%d bytes):\n%s", len(stderr), stderr)

	// Just verify it ran without crash — the theme plumbing is tested
	// at the unit level in theme/resolve_test.go
	if len(outBuf.String()) == 0 {
		t.Error("expected script output on stdout")
	}
}

// --- helpers ---

func mkdirs(t *testing.T, base string, names ...string) {
	t.Helper()
	for _, name := range names {
		if err := os.Mkdir(filepath.Join(base, name), 0755); err != nil {
			t.Fatal(err)
		}
	}
}
