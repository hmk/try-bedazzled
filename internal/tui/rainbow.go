package tui

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// hsl2hex converts an HSL color (h in [0,360], s and l in [0,1]) to a #RRGGBB hex string.
// Algorithm from the standard CSS HSL-to-RGB conversion.
func hsl2hex(h, s, l float64) string {
	// Normalize h to [0, 360)
	h = math.Mod(h, 360.0)
	if h < 0 {
		h += 360.0
	}

	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60.0, 2)-1))
	m := l - c/2

	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	ri := int(math.Round((r + m) * 255))
	gi := int(math.Round((g + m) * 255))
	bi := int(math.Round((b + m) * 255))
	return fmt.Sprintf("#%02X%02X%02X", ri, gi, bi)
}

// rainbowHue returns the hex color at the given position along a rainbow arc.
// pos is the 0-based index; total is the length of the strip (minimum 1).
// The hue steps from startHue across a full 360° so consecutive runs look distinct.
func rainbowHue(pos, total int, startHue float64) string {
	if total <= 0 {
		total = 1
	}
	h := startHue + float64(pos)*(360.0/float64(total))
	return hsl2hex(h, 0.72, 0.62)
}

// rainbowText renders s with each rune in a different hue across a 360° arc.
// Optional bold styling matches the non-rainbow cursor glyph feel.
func rainbowText(s string, bold bool) string {
	runes := []rune(s)
	if len(runes) == 0 {
		return ""
	}
	var b strings.Builder
	for i, r := range runes {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(rainbowHue(i, len(runes), 0)))
		if bold {
			style = style.Bold(true)
		}
		b.WriteString(style.Render(string(r)))
	}
	return b.String()
}

// rainbowRule returns a horizontal rule of n box-drawing chars, each
// colored along the rainbow. A leading space aligns it with the usual
// " ─────…" layout.
func rainbowRule(n int) string {
	if n <= 0 {
		return ""
	}
	var b strings.Builder
	b.WriteByte(' ')
	for i := 0; i < n; i++ {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(rainbowHue(i, n, 0)))
		b.WriteString(style.Render("─"))
	}
	return b.String()
}
