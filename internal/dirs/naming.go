// Package dirs handles try directory naming, scanning, and operations.
package dirs

import (
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode"
)

// DateFormat is the date prefix format used in try directory names.
const DateFormat = "2006-01-02"

// DatePrefixLen is the length of "YYYY-MM-DD-" (11 chars).
const DatePrefixLen = 11

// FormatName creates a dated directory name: YYYY-MM-DD-<slug>.
func FormatName(t time.Time, slug string) string {
	return t.Format(DateFormat) + "-" + slug
}

// ParseDatePrefix extracts the date and slug from a "YYYY-MM-DD-slug" name.
// Returns zero time and empty slug if the name doesn't have a valid date prefix.
func ParseDatePrefix(name string) (date time.Time, slug string, ok bool) {
	if len(name) < DatePrefixLen {
		return time.Time{}, "", false
	}
	dateStr := name[:10]
	if name[10] != '-' {
		return time.Time{}, "", false
	}
	t, err := time.Parse(DateFormat, dateStr)
	if err != nil {
		return time.Time{}, "", false
	}
	return t, name[DatePrefixLen:], true
}

// NormalizeDirName normalizes a directory name:
//   - Converts spaces to hyphens
//   - Collapses multiple consecutive hyphens to a single hyphen
//   - Strips leading/trailing hyphens and spaces
//   - Returns ("", false) if the name contains invalid characters
//
// Valid characters: [a-zA-Z0-9_.-] and spaces (which are converted to hyphens).
func NormalizeDirName(name string) (string, bool) {
	if name == "" {
		return "", false
	}

	// First pass: check for invalid characters
	for _, c := range name {
		if !isValidDirChar(c) && !unicode.IsSpace(c) {
			return "", false
		}
	}

	// Second pass: normalize
	var b strings.Builder
	lastWasHyphen := true // start true to strip leading hyphens/spaces
	for _, c := range name {
		if unicode.IsSpace(c) || c == '-' {
			if !lastWasHyphen {
				b.WriteByte('-')
				lastWasHyphen = true
			}
		} else {
			b.WriteRune(c)
			lastWasHyphen = false
		}
	}

	result := b.String()
	// Strip trailing hyphen
	result = strings.TrimRight(result, "-")

	if result == "" {
		return "", false
	}
	return result, true
}

// IsValidDirName checks if a name contains only valid directory name characters.
// Valid: [a-zA-Z0-9_.-] and spaces.
func IsValidDirName(name string) bool {
	if name == "" {
		return false
	}
	for _, c := range name {
		if !isValidDirChar(c) && !unicode.IsSpace(c) {
			return false
		}
	}
	return true
}

func isValidDirChar(c rune) bool {
	return unicode.IsLetter(c) || unicode.IsDigit(c) || c == '_' || c == '-' || c == '.'
}

// MakeCloneDirname generates a date-prefixed directory name from a clone URL.
// URL formats supported:
//   - https://github.com/user/repo.git → YYYY-MM-DD-user-repo
//   - git@github.com:user/repo.git     → YYYY-MM-DD-user-repo
//
// If customName is non-empty, it's used instead of parsing the URL.
func MakeCloneDirname(cloneURL string, customName string, t time.Time) string {
	datePrefix := t.Format(DateFormat) + "-"

	if customName != "" {
		return datePrefix + customName
	}

	user, repo := parseCloneURL(cloneURL)

	var slug string
	if user != "" {
		slug = user + "-" + repo
	} else {
		slug = repo
	}

	return datePrefix + slug
}

// parseCloneURL extracts user and repo from a git clone URL.
func parseCloneURL(rawURL string) (user, repo string) {
	// Handle SSH format: git@github.com:user/repo.git
	if strings.HasPrefix(rawURL, "git@") {
		colonIdx := strings.LastIndex(rawURL, ":")
		if colonIdx < 0 {
			return "", rawURL
		}
		path := rawURL[colonIdx+1:]
		return splitUserRepo(path)
	}

	// Handle HTTPS format
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", rawURL
	}

	path := strings.TrimPrefix(parsed.Path, "/")
	return splitUserRepo(path)
}

// splitUserRepo splits "user/repo.git" into ("user", "repo").
func splitUserRepo(path string) (user, repo string) {
	// Strip .git suffix
	path = strings.TrimSuffix(path, ".git")

	parts := strings.Split(path, "/")
	switch len(parts) {
	case 0:
		return "", ""
	case 1:
		return "", parts[0]
	default:
		// Take the last two segments (user/repo)
		return parts[len(parts)-2], parts[len(parts)-1]
	}
}

// FormatRelativeTime formats a time as a human-readable relative string.
func FormatRelativeTime(mtime time.Time) string {
	diff := time.Since(mtime)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	case diff < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(diff.Hours()/24))
	}
}
