// Package fuzzy implements the scoring algorithm from tobi/try-cli.
//
// The algorithm scores directory names against a query string using:
//   - Character match bonuses (case-insensitive)
//   - Word boundary bonuses
//   - Proximity bonuses for consecutive/near matches
//   - Density and length multipliers
//   - Date prefix bonus for YYYY-MM-DD- formatted names
//   - Recency bonus based on last access time
package fuzzy

import (
	"math"
	"strings"
	"time"
	"unicode"
)

// Result holds the outcome of a fuzzy match.
type Result struct {
	// Score is the total match score. Zero means no match (query not fully matched).
	Score float64

	// MatchPositions tracks which character indices in the text were matched.
	// Used for highlight rendering.
	MatchPositions []int

	// Matched is true if all query characters were found in the text.
	Matched bool
}

// HasDatePrefix returns true if the text starts with a YYYY-MM-DD- pattern.
func HasDatePrefix(text string) bool {
	if len(text) < 11 {
		return false
	}
	for _, i := range []int{0, 1, 2, 3} {
		if !isDigit(text[i]) {
			return false
		}
	}
	if text[4] != '-' {
		return false
	}
	for _, i := range []int{5, 6} {
		if !isDigit(text[i]) {
			return false
		}
	}
	if text[7] != '-' {
		return false
	}
	for _, i := range []int{8, 9} {
		if !isDigit(text[i]) {
			return false
		}
	}
	return text[10] == '-'
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

// Match scores a text against a query, using mtime for recency bonus.
// If query is empty, only recency scoring is applied.
func Match(text string, query string, mtime time.Time) Result {
	now := time.Now()
	hoursSinceAccess := now.Sub(mtime).Hours()
	if hoursSinceAccess < 0 {
		hoursSinceAccess = 0
	}
	recencyBonus := 3.0 / math.Sqrt(hoursSinceAccess+1)

	// Empty query: recency-only scoring
	if query == "" {
		return Result{
			Score:   recencyBonus,
			Matched: true,
		}
	}

	textLower := strings.ToLower(text)
	queryLower := strings.ToLower(query)

	queryLen := len(queryLower)
	textLen := len(text)
	queryIdx := 0
	lastPos := -1
	var matchPositions []int
	var fuzzyScore float64

	for i := 0; i < len(textLower) && queryIdx < queryLen; i++ {
		if textLower[i] == queryLower[queryIdx] {
			// Match found
			fuzzyScore += 1.0

			// Word boundary bonus
			if i == 0 || !isAlnum(rune(textLower[i-1])) {
				fuzzyScore += 1.0
			}

			// Proximity bonus
			if lastPos >= 0 {
				gap := i - lastPos - 1
				fuzzyScore += 2.0 / math.Sqrt(float64(gap)+1)
			}

			lastPos = i
			queryIdx++
			matchPositions = append(matchPositions, i)
		}
	}

	// If we didn't match the full query, score is 0
	if queryIdx < queryLen {
		return Result{
			Score:          0,
			MatchPositions: matchPositions,
			Matched:        false,
		}
	}

	// Density bonus multiplier
	if lastPos >= 0 {
		fuzzyScore *= float64(queryLen) / float64(lastPos+1)
	}

	// Length penalty multiplier
	fuzzyScore *= 10.0 / float64(textLen+10)

	// Date prefix bonus (applied after multipliers)
	var dateBonus float64
	if HasDatePrefix(text) {
		dateBonus = 2.0
	}

	totalScore := fuzzyScore + dateBonus + recencyBonus

	return Result{
		Score:          totalScore,
		MatchPositions: matchPositions,
		Matched:        true,
	}
}

// MatchAt scores a text against a query at a specific reference time.
// This is useful for deterministic testing where you control "now".
func MatchAt(text string, query string, mtime time.Time, now time.Time) Result {
	hoursSinceAccess := now.Sub(mtime).Hours()
	if hoursSinceAccess < 0 {
		hoursSinceAccess = 0
	}
	recencyBonus := 3.0 / math.Sqrt(hoursSinceAccess+1)

	if query == "" {
		return Result{
			Score:   recencyBonus,
			Matched: true,
		}
	}

	textLower := strings.ToLower(text)
	queryLower := strings.ToLower(query)

	queryLen := len(queryLower)
	textLen := len(text)
	queryIdx := 0
	lastPos := -1
	var matchPositions []int
	var fuzzyScore float64

	for i := 0; i < len(textLower) && queryIdx < queryLen; i++ {
		if textLower[i] == queryLower[queryIdx] {
			fuzzyScore += 1.0

			if i == 0 || !isAlnum(rune(textLower[i-1])) {
				fuzzyScore += 1.0
			}

			if lastPos >= 0 {
				gap := i - lastPos - 1
				fuzzyScore += 2.0 / math.Sqrt(float64(gap)+1)
			}

			lastPos = i
			queryIdx++
			matchPositions = append(matchPositions, i)
		}
	}

	if queryIdx < queryLen {
		return Result{
			Score:          0,
			MatchPositions: matchPositions,
			Matched:        false,
		}
	}

	if lastPos >= 0 {
		fuzzyScore *= float64(queryLen) / float64(lastPos+1)
	}

	fuzzyScore *= 10.0 / float64(textLen+10)

	var dateBonus float64
	if HasDatePrefix(text) {
		dateBonus = 2.0
	}

	totalScore := fuzzyScore + dateBonus + recencyBonus

	return Result{
		Score:          totalScore,
		MatchPositions: matchPositions,
		Matched:        true,
	}
}

func isAlnum(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}
