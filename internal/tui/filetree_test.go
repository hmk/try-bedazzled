package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hmk/try-bedazzled/internal/theme"
)

func TestRenderFileTreeEmpty(t *testing.T) {
	dir := t.TempDir()
	styles := NewStyles(theme.Default())

	result := RenderFileTree(dir, 2, 8, styles)
	if result != "" {
		t.Errorf("empty dir should return empty tree, got %q", result)
	}
}

func TestRenderFileTreeFiles(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "README.md"), []byte("hi"))
	mustWriteFile(t, filepath.Join(dir, "go.mod"), []byte("module x"))

	styles := NewStyles(theme.Default())
	result := RenderFileTree(dir, 2, 8, styles)

	if !strings.Contains(result, "README.md") {
		t.Error("tree should contain README.md")
	}
	if !strings.Contains(result, "go.mod") {
		t.Error("tree should contain go.mod")
	}
}

func TestRenderFileTreeNestedDir(t *testing.T) {
	dir := t.TempDir()
	mustMkdir(t, filepath.Join(dir, "src"))
	mustWriteFile(t, filepath.Join(dir, "src", "main.go"), []byte(""))
	mustWriteFile(t, filepath.Join(dir, "README.md"), []byte(""))

	styles := NewStyles(theme.Default())
	result := RenderFileTree(dir, 2, 8, styles)

	// Should show the dir with trailing /
	if !strings.Contains(result, "src/") {
		t.Error("tree should show 'src/' directory")
	}
	// Should descend and show main.go
	if !strings.Contains(result, "main.go") {
		t.Error("tree should descend and show 'main.go'")
	}
}

func TestRenderFileTreeSkipsHidden(t *testing.T) {
	dir := t.TempDir()
	mustMkdir(t, filepath.Join(dir, ".git"))
	mustWriteFile(t, filepath.Join(dir, ".DS_Store"), []byte(""))
	mustWriteFile(t, filepath.Join(dir, "visible.txt"), []byte(""))

	styles := NewStyles(theme.Default())
	result := RenderFileTree(dir, 2, 8, styles)

	if strings.Contains(result, ".git") {
		t.Error("tree should skip .git")
	}
	if strings.Contains(result, ".DS_Store") {
		t.Error("tree should skip .DS_Store")
	}
	if !strings.Contains(result, "visible.txt") {
		t.Error("tree should show visible.txt")
	}
}

func TestRenderFileTreeDirsBeforeFiles(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "zebra.txt"), []byte(""))
	mustMkdir(t, filepath.Join(dir, "alpha"))

	styles := NewStyles(theme.Default())
	result := RenderFileTree(dir, 1, 8, styles)

	// Even though 'zebra.txt' comes alphabetically after 'alpha',
	// dirs should be listed first.
	alphaIdx := strings.Index(result, "alpha")
	zebraIdx := strings.Index(result, "zebra.txt")
	if alphaIdx < 0 || zebraIdx < 0 {
		t.Fatalf("both entries should appear in tree: %q", result)
	}
	if alphaIdx > zebraIdx {
		t.Error("directories should be listed before files")
	}
}

func TestRenderFileTreeMaxLines(t *testing.T) {
	dir := t.TempDir()
	// Create many files
	for i := 0; i < 20; i++ {
		mustWriteFile(t, filepath.Join(dir, "file"+string(rune('a'+i))+".txt"), []byte(""))
	}

	styles := NewStyles(theme.Default())
	result := RenderFileTree(dir, 2, 5, styles)

	// Should show truncation indicator
	if !strings.Contains(result, "...") {
		t.Error("tree should show '...' truncation when exceeding maxLines")
	}
}

func TestRenderFileTreeMaxDepth(t *testing.T) {
	dir := t.TempDir()
	// Create nested structure: dir/a/b/c/deep.txt
	if err := os.MkdirAll(filepath.Join(dir, "a", "b", "c"), 0755); err != nil {
		t.Fatal(err)
	}
	mustWriteFile(t, filepath.Join(dir, "a", "b", "c", "deep.txt"), []byte(""))

	styles := NewStyles(theme.Default())
	result := RenderFileTree(dir, 2, 20, styles)

	// With maxDepth=2, we should see "a/" and "b/" but not "c/" or "deep.txt"
	if !strings.Contains(result, "a/") {
		t.Error("tree should show 'a/' at depth 0")
	}
	if strings.Contains(result, "deep.txt") {
		t.Error("tree should NOT show 'deep.txt' at depth 3 when maxDepth=2")
	}
}

func TestRenderFileTreeNonExistent(t *testing.T) {
	styles := NewStyles(theme.Default())
	result := RenderFileTree("/path/does/not/exist/anywhere", 2, 8, styles)
	// Should not panic, should return empty or just a dim message
	if len(result) > 100 {
		t.Errorf("non-existent path should return short output, got %d chars", len(result))
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatal(err)
	}
}

func mustWriteFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
}
