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
// Checks each word in the slug (split by hyphens) against the registry.
// Returns the theme's default folder icon if no match is found.
func LookupIcon(slug string, fallback string) string {
	parts := strings.Split(strings.ToLower(slug), "-")
	for _, part := range parts {
		if icon, ok := IconRegistry[part]; ok {
			return icon
		}
	}
	return fallback
}
