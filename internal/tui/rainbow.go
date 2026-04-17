package tui

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	lgv2 "charm.land/lipgloss/v2"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"
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

// rainbowStops are the six rainbow anchor colors for row-background gradients.
var rainbowStops = []color.Color{
	lipgloss.Color("#FF2D95"), // hot pink
	lipgloss.Color("#FF6B00"), // orange
	lipgloss.Color("#FACC15"), // yellow
	lipgloss.Color("#10B981"), // green
	lipgloss.Color("#3B82F6"), // blue
	lipgloss.Color("#8B5CF6"), // purple
}

// gradientRowBG wraps a rendered row in a left→right CIELAB rainbow
// background (via lipgloss/v2 Blend1D), padding to `width` columns and
// overriding the foreground to bold white for legibility. Any existing
// ANSI styles in `rendered` are stripped first — on selected rows the
// gradient carries the emphasis by itself.
func gradientRowBG(rendered string, width int) string {
	if width < 1 {
		width = 1
	}
	plain := ansi.Strip(rendered)
	// Trim trailing whitespace since we'll repad to exact width.
	plain = strings.TrimRight(plain, " ")
	plainWidth := runewidth.StringWidth(plain)
	if plainWidth < width {
		plain += strings.Repeat(" ", width-plainWidth)
	}

	colors := lgv2.Blend1D(width, rainbowStops...)

	var b strings.Builder
	col := 0
	for _, r := range plain {
		if col >= width {
			break
		}
		w := runewidth.RuneWidth(r)
		if w == 0 {
			b.WriteRune(r) // combining mark; don't advance the color index
			continue
		}
		idx := col
		if idx >= len(colors) {
			idx = len(colors) - 1
		}
		bg := hexOf(colors[idx])
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color(bg)).
			Bold(true).
			Render(string(r)))
		col += w
	}
	return b.String()
}

// hexOf converts a color.Color to a #RRGGBB string.
func hexOf(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02X%02X%02X", r>>8, g>>8, b>>8)
}

// solidRowBG wraps a row in a single-color background, padding to `width`
// columns with bold text at `fgHex`. Strips any existing ANSI so the
// coloring is uniform — the BG carries the "selected row" signal by itself.
func solidRowBG(rendered string, bgHex, fgHex string, width int) string {
	if width < 1 {
		width = 1
	}
	plain := ansi.Strip(rendered)
	plain = strings.TrimRight(plain, " ")
	plainWidth := runewidth.StringWidth(plain)
	if plainWidth < width {
		plain += strings.Repeat(" ", width-plainWidth)
	}
	style := lipgloss.NewStyle().
		Background(lipgloss.Color(bgHex)).
		Bold(true)
	if fgHex != "" {
		style = style.Foreground(lipgloss.Color(fgHex))
	}
	return style.Render(plain)
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
