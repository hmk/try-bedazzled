package dirs

import (
	"testing"
	"time"
)

var refTime = time.Date(2026, 4, 11, 12, 0, 0, 0, time.UTC)

// --- FormatName ---

func TestFormatName(t *testing.T) {
	tests := []struct {
		slug string
		want string
	}{
		{"redis", "2026-04-11-redis"},
		{"go-tui", "2026-04-11-go-tui"},
		{"my_project.v2", "2026-04-11-my_project.v2"},
		{"a", "2026-04-11-a"},
	}

	for _, tt := range tests {
		t.Run(tt.slug, func(t *testing.T) {
			got := FormatName(refTime, tt.slug)
			if got != tt.want {
				t.Errorf("FormatName(%q) = %q, want %q", tt.slug, got, tt.want)
			}
		})
	}
}

func TestFormatNameDifferentDates(t *testing.T) {
	jan1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	got := FormatName(jan1, "test")
	if got != "2025-01-01-test" {
		t.Errorf("got %q, want %q", got, "2025-01-01-test")
	}
}

// --- ParseDatePrefix ---

func TestParseDatePrefix(t *testing.T) {
	tests := []struct {
		name     string
		wantDate string
		wantSlug string
		wantOK   bool
	}{
		{"2026-04-11-redis", "2026-04-11", "redis", true},
		{"2025-01-01-foo", "2025-01-01", "foo", true},
		{"2026-12-31-bar-baz", "2026-12-31", "bar-baz", true},
		{"2026-04-11-", "2026-04-11", "", true}, // empty slug is valid parse
		{"no-date-here", "", "", false},
		{"short", "", "", false},
		{"2026-04-11redis", "", "", false},  // missing dash after date
		{"abcd-ef-gh-nope", "", "", false},  // non-numeric date
		{"2026-13-01-bad", "", "", false},   // invalid month
		{"2026-04-32-bad", "", "", false},   // invalid day
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, slug, ok := ParseDatePrefix(tt.name)
			if ok != tt.wantOK {
				t.Fatalf("ParseDatePrefix(%q) ok = %v, want %v", tt.name, ok, tt.wantOK)
			}
			if !ok {
				return
			}
			gotDate := date.Format(DateFormat)
			if gotDate != tt.wantDate {
				t.Errorf("date = %q, want %q", gotDate, tt.wantDate)
			}
			if slug != tt.wantSlug {
				t.Errorf("slug = %q, want %q", slug, tt.wantSlug)
			}
		})
	}
}

// --- NormalizeDirName ---

func TestNormalizeDirName(t *testing.T) {
	tests := []struct {
		input string
		want  string
		ok    bool
	}{
		// Basic valid names
		{"redis", "redis", true},
		{"my-project", "my-project", true},
		{"my_project", "my_project", true},
		{"project.v2", "project.v2", true},

		// Space to hyphen conversion
		{"my project", "my-project", true},
		{"hello world test", "hello-world-test", true},

		// Collapse multiple hyphens
		{"my---project", "my-project", true},
		{"a - - b", "a-b", true},

		// Strip leading/trailing hyphens and spaces
		{"-redis", "redis", true},
		{"redis-", "redis", true},
		{"-redis-", "redis", true},
		{"--redis--", "redis", true},
		{" redis ", "redis", true},
		{" -redis- ", "redis", true},

		// Mixed spaces and hyphens
		{" my  project ", "my-project", true},
		{"  hello  --  world  ", "hello-world", true},

		// Invalid characters
		{"my/project", "", false},
		{"my@project", "", false},
		{"my!project", "", false},
		{"hello$world", "", false},
		{"path\\to", "", false},

		// Edge cases
		{"", "", false},
		{"-", "", false},
		{"---", "", false},
		{"   ", "", false},
		{" - - - ", "", false},

		// Valid with dots and underscores
		{"v1.2.3", "v1.2.3", true},
		{"my_app_v2", "my_app_v2", true},
		{"test.something.here", "test.something.here", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := NormalizeDirName(tt.input)
			if ok != tt.ok {
				t.Fatalf("NormalizeDirName(%q) ok = %v, want %v", tt.input, ok, tt.ok)
			}
			if got != tt.want {
				t.Errorf("NormalizeDirName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// --- IsValidDirName ---

func TestIsValidDirName(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"redis", true},
		{"my-project", true},
		{"my_project", true},
		{"v1.2.3", true},
		{"hello world", true}, // spaces are valid (normalized to hyphens)
		{"my/project", false},
		{"my@project", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidDirName(tt.name)
			if got != tt.want {
				t.Errorf("IsValidDirName(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

// --- MakeCloneDirname ---

func TestMakeCloneDirname(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		customName string
		want       string
	}{
		{
			name: "https with .git",
			url:  "https://github.com/tobi/try-cli.git",
			want: "2026-04-11-tobi-try-cli",
		},
		{
			name: "https without .git",
			url:  "https://github.com/tobi/try-cli",
			want: "2026-04-11-tobi-try-cli",
		},
		{
			name: "ssh format",
			url:  "git@github.com:tobi/try-cli.git",
			want: "2026-04-11-tobi-try-cli",
		},
		{
			name: "ssh without .git",
			url:  "git@github.com:tobi/try-cli",
			want: "2026-04-11-tobi-try-cli",
		},
		{
			name:       "custom name overrides URL parsing",
			url:        "https://github.com/tobi/try-cli.git",
			customName: "my-fork",
			want:       "2026-04-11-my-fork",
		},
		{
			name: "gitlab url",
			url:  "https://gitlab.com/user/project.git",
			want: "2026-04-11-user-project",
		},
		{
			name: "nested github path",
			url:  "https://github.com/org/suborg/repo.git",
			want: "2026-04-11-suborg-repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MakeCloneDirname(tt.url, tt.customName, refTime)
			if got != tt.want {
				t.Errorf("MakeCloneDirname(%q, %q) = %q, want %q",
					tt.url, tt.customName, got, tt.want)
			}
		})
	}
}

// --- FormatRelativeTime ---

func TestFormatRelativeTime(t *testing.T) {
	// These tests are approximate since they use time.Since internally.
	// We test the boundaries.
	now := time.Now()

	tests := []struct {
		name  string
		mtime time.Time
		want  string
	}{
		{"just now", now.Add(-30 * time.Second), "just now"},
		{"minutes ago", now.Add(-5 * time.Minute), "5m ago"},
		{"hours ago", now.Add(-3 * time.Hour), "3h ago"},
		{"days ago", now.Add(-7 * 24 * time.Hour), "7d ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatRelativeTime(tt.mtime)
			if got != tt.want {
				t.Errorf("FormatRelativeTime() = %q, want %q", got, tt.want)
			}
		})
	}
}

// --- Round-trip ---

func TestFormatParseRoundTrip(t *testing.T) {
	slug := "redis-test"
	name := FormatName(refTime, slug)

	date, parsedSlug, ok := ParseDatePrefix(name)
	if !ok {
		t.Fatal("round-trip parse failed")
	}
	if date.Format(DateFormat) != refTime.Format(DateFormat) {
		t.Errorf("date mismatch: %v vs %v", date, refTime)
	}
	if parsedSlug != slug {
		t.Errorf("slug mismatch: %q vs %q", parsedSlug, slug)
	}
}
