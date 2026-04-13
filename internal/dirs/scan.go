package dirs

import (
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Entry represents a single try directory.
type Entry struct {
	// Name is the directory name (e.g., "2026-04-11-redis").
	Name string

	// Path is the full absolute path to the directory.
	Path string

	// Mtime is the last modification time (used for recency scoring).
	Mtime time.Time
}

// ScanResult holds the result of scanning a tries directory.
type ScanResult struct {
	Entries []Entry
	BasePath string
}

// Scan reads all directories in basePath, skipping hidden dirs (starting with '.').
// Results are sorted by mtime descending (most recently accessed first).
func Scan(basePath string) (ScanResult, error) {
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return ScanResult{BasePath: basePath}, err
	}

	var result []Entry
	for _, e := range entries {
		// Skip hidden directories
		if e.Name()[0] == '.' {
			continue
		}
		// Skip non-directories
		if !e.IsDir() {
			continue
		}

		fullPath := filepath.Join(basePath, e.Name())
		info, err := e.Info()
		if err != nil {
			continue // skip entries we can't stat
		}

		result = append(result, Entry{
			Name:  e.Name(),
			Path:  fullPath,
			Mtime: info.ModTime(),
		})
	}

	// Sort by mtime descending (most recent first)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Mtime.After(result[j].Mtime)
	})

	return ScanResult{
		Entries:  result,
		BasePath: basePath,
	}, nil
}
