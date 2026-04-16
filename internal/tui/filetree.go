package tui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RenderFileTree returns a tree-style listing of the given directory,
// limited to maxDepth levels and maxLines total lines.
// Directories are sorted first, then files, alphabetically.
// Hidden entries (starting with ".") are skipped.
// Renders tree connectors with the dim style from the passed Styles.
func RenderFileTree(path string, maxDepth int, maxLines int, styles Styles) string {
	if maxDepth <= 0 {
		maxDepth = 2
	}
	if maxLines <= 0 {
		maxLines = 8
	}

	var lines []string
	truncated := renderTreeLevel(path, "", 0, maxDepth, &lines, maxLines, styles)

	if truncated {
		lines = append(lines, styles.Dim.Render("  ..."))
	}

	return strings.Join(lines, "\n")
}

// renderTreeLevel recursively walks dir, appending tree-style lines.
// Returns true if truncation occurred.
func renderTreeLevel(path string, prefix string, depth int, maxDepth int, lines *[]string, maxLines int, styles Styles) bool {
	if depth >= maxDepth {
		return false
	}
	if len(*lines) >= maxLines {
		return true
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}

	// Filter and sort: dirs first, then files, alpha within each
	var dirs, files []os.DirEntry
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		if e.IsDir() {
			dirs = append(dirs, e)
		} else {
			files = append(files, e)
		}
	}
	sort.Slice(dirs, func(i, j int) bool { return dirs[i].Name() < dirs[j].Name() })
	sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })

	visible := append(dirs, files...)

	for i, e := range visible {
		if len(*lines) >= maxLines {
			return true
		}
		isLast := i == len(visible)-1

		var connector string
		var nextPrefix string
		if isLast {
			connector = "└── "
			nextPrefix = prefix + "    "
		} else {
			connector = "├── "
			nextPrefix = prefix + "│   "
		}

		name := e.Name()
		if e.IsDir() {
			name += "/"
		}

		line := styles.Dim.Render(prefix+connector) + styles.Normal.Render(name)
		*lines = append(*lines, line)

		if e.IsDir() {
			if renderTreeLevel(filepath.Join(path, e.Name()), nextPrefix, depth+1, maxDepth, lines, maxLines, styles) {
				return true
			}
		}
	}

	return false
}
