package fuzzy

import (
	"math"
	"testing"
	"time"
)

// Fixed reference time for deterministic tests
var (
	now   = time.Date(2026, 4, 11, 12, 0, 0, 0, time.UTC)
	fresh = now.Add(-1 * time.Hour)    // 1 hour ago
	old   = now.Add(-720 * time.Hour)  // 30 days ago
	stale = now.Add(-2160 * time.Hour) // 90 days ago
)

func matchAt(text, query string, mtime time.Time) Result {
	return MatchAt(text, query, mtime, now)
}

// --- HasDatePrefix ---

func TestHasDatePrefix(t *testing.T) {
	tests := []struct {
		text string
		want bool
	}{
		{"2026-04-11-redis", true},
		{"2025-01-01-foo", true},
		{"2026-12-31-bar", true},
		{"2026-04-11-", true}, // just the prefix, empty slug
		{"no-date-here", false},
		{"202-04-11-short", false}, // year too short
		{"2026-4-11-nopad", false}, // month not zero-padded
		{"2026-04-1-nopad", false}, // day not zero-padded
		{"abcd-ef-gh-nope", false},
		{"", false},
		{"2026-04-11redis", false},  // missing trailing dash
		{"2026/04/11-slash", false}, // wrong separator
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			got := HasDatePrefix(tt.text)
			if got != tt.want {
				t.Errorf("HasDatePrefix(%q) = %v, want %v", tt.text, got, tt.want)
			}
		})
	}
}

// --- Empty query ---

func TestEmptyQueryReturnsRecencyOnly(t *testing.T) {
	r := matchAt("2026-04-11-redis", "", fresh)
	if !r.Matched {
		t.Fatal("empty query should always match")
	}
	if r.Score <= 0 {
		t.Fatal("empty query should have positive recency score")
	}
	if len(r.MatchPositions) != 0 {
		t.Fatal("empty query should have no match positions")
	}
}

func TestEmptyQueryRecencyFreshBeatsOld(t *testing.T) {
	freshResult := matchAt("2026-04-11-redis", "", fresh)
	oldResult := matchAt("2026-04-11-redis", "", old)
	if freshResult.Score <= oldResult.Score {
		t.Errorf("fresh (%f) should score higher than old (%f)", freshResult.Score, oldResult.Score)
	}
}

// --- No match ---

func TestNoMatchReturnsZero(t *testing.T) {
	r := matchAt("2026-04-11-redis", "zzzzz", fresh)
	if r.Matched {
		t.Fatal("should not match")
	}
	if r.Score != 0 {
		t.Errorf("unmatched score should be 0, got %f", r.Score)
	}
}

func TestPartialQueryNoMatch(t *testing.T) {
	// "redisxyz" — the "xyz" part won't match
	r := matchAt("2026-04-11-redis", "redisxyz", fresh)
	if r.Matched {
		t.Fatal("should not match partial query")
	}
	if r.Score != 0 {
		t.Errorf("score should be 0, got %f", r.Score)
	}
}

// --- Basic matching ---

func TestExactSubstringMatch(t *testing.T) {
	r := matchAt("2026-04-11-redis", "redis", fresh)
	if !r.Matched {
		t.Fatal("should match")
	}
	if r.Score <= 0 {
		t.Error("score should be positive")
	}
}

func TestCaseInsensitive(t *testing.T) {
	lower := matchAt("2026-04-11-redis", "redis", fresh)
	upper := matchAt("2026-04-11-redis", "REDIS", fresh)
	mixed := matchAt("2026-04-11-redis", "ReDiS", fresh)

	if !lower.Matched || !upper.Matched || !mixed.Matched {
		t.Fatal("all case variants should match")
	}
	// Scores should be identical since matching is case-insensitive
	if lower.Score != upper.Score || lower.Score != mixed.Score {
		t.Errorf("case variants should have equal scores: lower=%f upper=%f mixed=%f",
			lower.Score, upper.Score, mixed.Score)
	}
}

func TestSingleCharMatch(t *testing.T) {
	r := matchAt("2026-04-11-redis", "r", fresh)
	if !r.Matched {
		t.Fatal("single char should match")
	}
	if len(r.MatchPositions) != 1 {
		t.Errorf("expected 1 match position, got %d", len(r.MatchPositions))
	}
}

// --- Match positions ---

func TestMatchPositionsCorrect(t *testing.T) {
	// "redis" in "2026-04-11-redis" starts at index 11
	r := matchAt("2026-04-11-redis", "redis", fresh)
	if !r.Matched {
		t.Fatal("should match")
	}
	expected := []int{11, 12, 13, 14, 15}
	if len(r.MatchPositions) != len(expected) {
		t.Fatalf("expected %d positions, got %d", len(expected), len(r.MatchPositions))
	}
	for i, pos := range r.MatchPositions {
		if pos != expected[i] {
			t.Errorf("position[%d] = %d, want %d", i, pos, expected[i])
		}
	}
}

func TestFuzzyMatchPositions(t *testing.T) {
	// "rd" in "redis" — matches r(0) then d is at index 2 in "redis"
	r := matchAt("redis", "rd", fresh)
	if !r.Matched {
		t.Fatal("should match")
	}
	if len(r.MatchPositions) != 2 {
		t.Fatalf("expected 2 positions, got %d", len(r.MatchPositions))
	}
	// r at 0, d at 2 (skipping e)
	if r.MatchPositions[0] != 0 || r.MatchPositions[1] != 2 {
		t.Errorf("positions = %v, want [0, 2]", r.MatchPositions)
	}
}

// --- Scoring properties ---

func TestConsecutiveMatchScoresHigherThanSpread(t *testing.T) {
	// "redis" is a consecutive match in "2026-04-11-redis"
	// "rds" is a spread match in "2026-04-11-redis" (r, d, s with gaps)
	consecutive := matchAt("2026-04-11-redis", "redis", fresh)
	spread := matchAt("2026-04-11-redis", "rds", fresh)

	// Consecutive should benefit from proximity bonus
	// But spread has fewer chars → need to compare per-char or just verify consecutive is higher
	// Actually the algorithm doesn't normalize by query length, so "redis" (5 chars) should
	// score higher than "rds" (3 chars) due to more match bonuses.
	if !consecutive.Matched || !spread.Matched {
		t.Fatal("both should match")
	}
	if consecutive.Score <= spread.Score {
		t.Errorf("consecutive (%f) should score higher than spread (%f)", consecutive.Score, spread.Score)
	}
}

func TestWordBoundaryBonus(t *testing.T) {
	// "a" at word boundary (after hyphen) vs "a" in middle of word
	// "2026-04-11-api" — "a" matches at position 11 (after '-', word boundary)
	// "2026-04-11-bar" — "a" matches at position 12 (middle of "bar", NOT a boundary)
	boundary := matchAt("2026-04-11-api", "a", fresh)
	middle := matchAt("2026-04-11-bar", "a", fresh)

	if !boundary.Matched || !middle.Matched {
		t.Fatal("both should match")
	}
	// The one at word boundary should score higher
	// Note: "a" in "api" is at pos 11 (after '-'), which IS a boundary.
	// "a" in "bar" is at pos 12 (after 'b'), which is NOT a boundary.
	if boundary.Score <= middle.Score {
		t.Errorf("boundary match (%f) should score higher than middle match (%f)",
			boundary.Score, middle.Score)
	}
}

func TestRecencyBonus(t *testing.T) {
	freshR := matchAt("redis", "redis", fresh)
	oldR := matchAt("redis", "redis", old)
	staleR := matchAt("redis", "redis", stale)

	if !freshR.Matched || !oldR.Matched || !staleR.Matched {
		t.Fatal("all should match")
	}
	if freshR.Score <= oldR.Score {
		t.Errorf("fresh (%f) should score higher than old (%f)", freshR.Score, oldR.Score)
	}
	if oldR.Score <= staleR.Score {
		t.Errorf("old (%f) should score higher than stale (%f)", oldR.Score, staleR.Score)
	}
}

func TestDatePrefixBonus(t *testing.T) {
	// Compare two names of the same length, one with and one without a date prefix.
	// This isolates the +2.0 date bonus from length penalty differences.
	withDate := matchAt("2026-04-11-redis", "redis", fresh)
	withoutDate := matchAt("some-prefix-redis", "redis", fresh)

	if !withDate.Matched || !withoutDate.Matched {
		t.Fatal("both should match")
	}
	// Same length text, same query position — date prefix adds +2.0 bonus
	if withDate.Score <= withoutDate.Score {
		t.Errorf("date-prefixed (%f) should score higher than non-dated same-length (%f)",
			withDate.Score, withoutDate.Score)
	}
}

func TestLengthPenalty(t *testing.T) {
	short := matchAt("redis", "r", fresh)
	long := matchAt("redis-cluster-proxy-sidecar", "r", fresh)

	if !short.Matched || !long.Matched {
		t.Fatal("both should match")
	}
	if short.Score <= long.Score {
		t.Errorf("shorter text (%f) should score higher than longer text (%f)",
			short.Score, long.Score)
	}
}

// --- Ranking order ---

func TestRankingExactMatchBeatsFuzzy(t *testing.T) {
	entries := []struct {
		name  string
		query string
	}{
		{"2026-04-11-redis", "redis"},
		{"2026-04-11-redis-cluster", "redis"},
	}

	scores := make([]float64, len(entries))
	for i, e := range entries {
		r := matchAt(e.name, e.query, fresh)
		scores[i] = r.Score
	}

	// "redis" in shorter name should rank higher (length penalty)
	if scores[0] <= scores[1] {
		t.Errorf("shorter exact match (%f) should rank above longer (%f)", scores[0], scores[1])
	}
}

func TestRankingWithMultipleCandidates(t *testing.T) {
	candidates := []string{
		"2026-04-11-redis",
		"2026-04-10-redis-cluster",
		"2026-04-09-go-redis-client",
		"2026-04-08-something-else",
	}

	type scored struct {
		name  string
		score float64
	}

	var results []scored
	for _, c := range candidates {
		r := matchAt(c, "redis", fresh)
		if r.Matched {
			results = append(results, scored{c, r.Score})
		}
	}

	// "something-else" should not match "redis"
	for _, r := range results {
		if r.name == "2026-04-08-something-else" {
			t.Error("something-else should not match redis")
		}
	}

	// Should have 3 matches
	if len(results) != 3 {
		t.Fatalf("expected 3 matches, got %d", len(results))
	}

	// Verify they're all scored > 0
	for _, r := range results {
		if r.score <= 0 {
			t.Errorf("%s should have positive score, got %f", r.name, r.score)
		}
	}
}

// --- Recency formula ---

func TestRecencyFormulaValues(t *testing.T) {
	// Verify the formula: 3.0 / sqrt(hours + 1)
	tests := []struct {
		hours float64
		want  float64
	}{
		{0, 3.0}, // just accessed
		{1, 3.0 / math.Sqrt(2)},
		{8, 1.0},  // 3.0 / sqrt(9) = 1.0
		{24, 0.6}, // 3.0 / sqrt(25) = 0.6
		{720, 3.0 / math.Sqrt(721)},
	}

	for _, tt := range tests {
		mtime := now.Add(-time.Duration(tt.hours) * time.Hour)
		r := matchAt("something", "", mtime)
		// Allow small floating point tolerance
		if math.Abs(r.Score-tt.want) > 0.001 {
			t.Errorf("recency at %.0f hours: got %f, want %f", tt.hours, r.Score, tt.want)
		}
	}
}

// --- Edge cases ---

func TestEmptyText(t *testing.T) {
	r := matchAt("", "redis", fresh)
	if r.Matched {
		t.Error("empty text should not match non-empty query")
	}
}

func TestEmptyTextEmptyQuery(t *testing.T) {
	r := matchAt("", "", fresh)
	if !r.Matched {
		t.Error("empty text + empty query should match")
	}
}

func TestQueryLongerThanText(t *testing.T) {
	r := matchAt("ab", "abcdef", fresh)
	if r.Matched {
		t.Error("should not match when query is longer than text")
	}
}

func TestSpecialCharsInName(t *testing.T) {
	r := matchAt("2026-04-11-my_project.v2", "project", fresh)
	if !r.Matched {
		t.Error("should match through special chars")
	}
}

func TestUnicodeInName(t *testing.T) {
	// Basic ASCII matching should still work even with unicode in name
	r := matchAt("2026-04-11-café-app", "cafe", fresh)
	// 'é' != 'e', so this might not fully match — depends on normalization
	// Our implementation is byte-level like the C version, so é (2 bytes) won't match 'e'
	// This is expected behavior matching the C version
	if r.Matched {
		// If it matched, that's fine too — just verify score is positive
		if r.Score <= 0 {
			t.Error("matched but score is not positive")
		}
	}
}

func TestDensityBonus(t *testing.T) {
	// "abc" consecutive vs scattered — consecutive should score higher
	// due to density bonus: queryLen / (lastPos + 1)
	dense := matchAt("abcdef", "abc", fresh)         // lastPos=2, density=3/3=1.0
	sparse := matchAt("a---b-------c", "abc", fresh) // lastPos=12, density=3/13≈0.23

	if !dense.Matched || !sparse.Matched {
		t.Fatal("both should match")
	}
	if dense.Score <= sparse.Score {
		t.Errorf("dense (%f) should score higher than sparse (%f)", dense.Score, sparse.Score)
	}
}
