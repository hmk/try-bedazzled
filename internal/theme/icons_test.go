package theme

import "testing"

func TestLookupIconMatchesKeyword(t *testing.T) {
	tests := []struct {
		slug string
		want string
	}{
		{"redis-cache", "🔴"},
		{"go-api", "🐹"},
		{"python-ml-pipeline", "🐍"},    // first match wins
		{"rust-cli", "🦀"},
		{"my-react-app", "⚛️"},
		{"postgres-cluster", "🐘"},
		{"docker-compose-test", "🐳"},
	}

	for _, tt := range tests {
		t.Run(tt.slug, func(t *testing.T) {
			got := LookupIcon(tt.slug, "📂")
			if got != tt.want {
				t.Errorf("LookupIcon(%q) = %q, want %q", tt.slug, got, tt.want)
			}
		})
	}
}

func TestLookupIconFallback(t *testing.T) {
	got := LookupIcon("my-random-project", "📂")
	if got != "📂" {
		t.Errorf("unrecognized slug should return fallback, got %q", got)
	}
}

func TestLookupIconEmptySlug(t *testing.T) {
	got := LookupIcon("", "📂")
	if got != "📂" {
		t.Errorf("empty slug should return fallback, got %q", got)
	}
}

func TestLookupIconCaseInsensitive(t *testing.T) {
	got := LookupIcon("Redis-Cache", "📂")
	if got != "🔴" {
		t.Errorf("should be case insensitive, got %q", got)
	}
}
