package dirs

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestScanEmpty(t *testing.T) {
	dir := t.TempDir()
	result, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(result.Entries))
	}
	if result.BasePath != dir {
		t.Errorf("BasePath = %q, want %q", result.BasePath, dir)
	}
}

func TestScanFindsDirectories(t *testing.T) {
	dir := t.TempDir()
	mkdirs(t, dir, "2026-04-11-redis", "2026-04-10-postgres", "2026-04-09-api")

	result, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(result.Entries))
	}
}

func TestScanSkipsHiddenDirs(t *testing.T) {
	dir := t.TempDir()
	mkdirs(t, dir, "2026-04-11-redis", ".hidden", ".git")

	result, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Entries) != 1 {
		t.Errorf("expected 1 entry (hidden dirs skipped), got %d", len(result.Entries))
	}
	if result.Entries[0].Name != "2026-04-11-redis" {
		t.Errorf("expected redis, got %q", result.Entries[0].Name)
	}
}

func TestScanSkipsFiles(t *testing.T) {
	dir := t.TempDir()
	mkdirs(t, dir, "2026-04-11-redis")
	// Create a regular file
	os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("hello"), 0644)

	result, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Entries) != 1 {
		t.Errorf("expected 1 entry (files skipped), got %d", len(result.Entries))
	}
}

func TestScanSortsByMtimeDescending(t *testing.T) {
	dir := t.TempDir()
	names := []string{"2026-04-09-oldest", "2026-04-10-middle", "2026-04-11-newest"}
	mkdirs(t, dir, names...)

	// Set mtimes explicitly: oldest first, newest last
	base := time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)
	for i, name := range names {
		p := filepath.Join(dir, name)
		mtime := base.Add(time.Duration(i) * 24 * time.Hour)
		os.Chtimes(p, mtime, mtime)
	}

	result, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Entries) != 3 {
		t.Fatalf("expected 3, got %d", len(result.Entries))
	}

	// Should be newest first
	if result.Entries[0].Name != "2026-04-11-newest" {
		t.Errorf("first entry should be newest, got %q", result.Entries[0].Name)
	}
	if result.Entries[2].Name != "2026-04-09-oldest" {
		t.Errorf("last entry should be oldest, got %q", result.Entries[2].Name)
	}
}

func TestScanEntryPaths(t *testing.T) {
	dir := t.TempDir()
	mkdirs(t, dir, "2026-04-11-redis")

	result, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}

	expected := filepath.Join(dir, "2026-04-11-redis")
	if result.Entries[0].Path != expected {
		t.Errorf("Path = %q, want %q", result.Entries[0].Path, expected)
	}
}

func TestScanNonExistentDir(t *testing.T) {
	_, err := Scan("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("expected error for non-existent directory")
	}
}

func TestScanManyEntries(t *testing.T) {
	dir := t.TempDir()
	names := make([]string, 100)
	for i := range names {
		names[i] = FormatName(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i), "test")
	}
	mkdirs(t, dir, names...)

	result, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Entries) != 100 {
		t.Errorf("expected 100 entries, got %d", len(result.Entries))
	}
}

// --- helpers ---

func mkdirs(t *testing.T, base string, names ...string) {
	t.Helper()
	for _, name := range names {
		err := os.Mkdir(filepath.Join(base, name), 0755)
		if err != nil {
			t.Fatal(err)
		}
	}
}
