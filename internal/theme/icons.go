package theme

import "strings"

// IconRegistry maps semantic keywords found in directory slugs to emoji icons.
// Used to auto-detect icons based on directory name content.
// The theme's Symbols.Folder is used as the fallback when no keyword matches.
var IconRegistry = map[string]string{
	// Languages
	"go":         "🐹",
	"golang":     "🐹",
	"rust":       "🦀",
	"python":     "🐍",
	"ruby":       "💎",
	"node":       "💚",
	"typescript": "💙",
	"js":         "💛",
	"java":       "☕",
	"swift":      "🐦",
	"kotlin":     "🟣",
	"elixir":     "🧪",
	"zig":        "⚡",
	"c":          "⚙️",
	"cpp":        "⚙️",

	// Tools & services
	"redis":      "🔴",
	"postgres":   "🐘",
	"mysql":      "🐬",
	"docker":     "🐳",
	"k8s":        "☸️",
	"kubernetes": "☸️",
	"terraform":  "🏗️",
	"nginx":      "🌐",
	"react":      "⚛️",
	"vue":        "💚",
	"svelte":     "🔥",

	// Concepts
	"api":        "🔌",
	"cli":        "💻",
	"tui":        "🖥️",
	"ml":         "🧠",
	"ai":         "🤖",
	"auth":       "🔐",
	"test":       "🧪",
	"benchmark":  "📊",
	"docs":       "📝",
	"config":     "⚙️",
	"data":       "📦",
	"web":        "🌐",
	"bot":        "🤖",
	"game":       "🎮",
}

// LookupIcon returns the best icon for a directory slug.
// Checks each hyphen-separated word in the slug, with priority:
//  1. user custom map (passed in)
//  2. built-in IconRegistry
//  3. fallback (typically the theme's Symbols.Folder)
//
// custom may be nil. Keys in both maps are lower-cased for comparison.
func LookupIcon(slug string, fallback string, custom map[string]string) string {
	parts := strings.Split(strings.ToLower(slug), "-")

	// Pre-lowercase custom map keys once for case-insensitive lookup
	var customLower map[string]string
	if len(custom) > 0 {
		customLower = make(map[string]string, len(custom))
		for k, v := range custom {
			customLower[strings.ToLower(k)] = v
		}
	}

	for _, part := range parts {
		if icon, ok := customLower[part]; ok {
			return icon
		}
		if icon, ok := IconRegistry[part]; ok {
			return icon
		}
	}
	return fallback
}
