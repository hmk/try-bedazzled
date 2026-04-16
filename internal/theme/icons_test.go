package theme

import "testing"

func TestLookupIconMatchesKeyword(t *testing.T) {
	tests := []struct {
		slug string
		want string
	}{
		{"redis-cache", "🔴"},
		{"go-api", "🐹"},
		{"python-ml-pipeline", "🐍"}, // first match wins
		{"rust-cli", "🦀"},
		{"my-react-app", "⚛️"},
		{"postgres-cluster", "🐘"},
		{"docker-compose-test", "🐳"},
	}

	for _, tt := range tests {
		t.Run(tt.slug, func(t *testing.T) {
			got := LookupIcon(tt.slug, "📂", nil)
			if got != tt.want {
				t.Errorf("LookupIcon(%q) = %q, want %q", tt.slug, got, tt.want)
			}
		})
	}
}

func TestLookupIconFallback(t *testing.T) {
	got := LookupIcon("my-random-project", "📂", nil)
	if got != "📂" {
		t.Errorf("unrecognized slug should return fallback, got %q", got)
	}
}

func TestLookupIconEmptySlug(t *testing.T) {
	got := LookupIcon("", "📂", nil)
	if got != "📂" {
		t.Errorf("empty slug should return fallback, got %q", got)
	}
}

func TestLookupIconCaseInsensitive(t *testing.T) {
	got := LookupIcon("Redis-Cache", "📂", nil)
	if got != "🔴" {
		t.Errorf("should be case insensitive, got %q", got)
	}
}

// --- Custom icons ---

func TestLookupIconCustomBeatsBuiltin(t *testing.T) {
	// User overrides "redis" → custom icon
	custom := map[string]string{"redis": "❤️"}
	got := LookupIcon("redis-cache", "📂", custom)
	if got != "❤️" {
		t.Errorf("custom should beat built-in: got %q, want ❤️", got)
	}
}

func TestLookupIconCustomAddsNewKeyword(t *testing.T) {
	// "django" isn't in built-in registry
	custom := map[string]string{"django": "🎭"}
	got := LookupIcon("my-django-app", "📂", custom)
	if got != "🎭" {
		t.Errorf("custom keyword should match: got %q, want 🎭", got)
	}
}

func TestLookupIconCustomCaseInsensitive(t *testing.T) {
	custom := map[string]string{"Django": "🎭"}
	got := LookupIcon("my-django-app", "📂", custom)
	if got != "🎭" {
		t.Errorf("custom keys should be case-insensitive: got %q", got)
	}
}

func TestLookupIconCustomFallsBackToBuiltin(t *testing.T) {
	custom := map[string]string{"django": "🎭"}
	// No django in slug — should fall to built-in for "redis"
	got := LookupIcon("redis-cache", "📂", custom)
	if got != "🔴" {
		t.Errorf("should fall back to built-in: got %q, want 🔴", got)
	}
}

func TestLookupIconEmptyCustomMap(t *testing.T) {
	got := LookupIcon("go-api", "📂", map[string]string{})
	if got != "🐹" {
		t.Errorf("empty custom map should not break: got %q", got)
	}
}
