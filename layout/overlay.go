package layout

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// OverlayPos anchors an overlay as fractions of the free space between the
// foreground and background blocks: 0 is the left/top edge, 1 the right/bottom,
// 0.5 centered.
type OverlayPos struct {
	XFrac float64
	YFrac float64
}

// Center anchors an overlay in the middle of the background.
var Center = OverlayPos{XFrac: 0.5, YFrac: 0.5}

// Overlay draws fg on top of bg at the given anchor (Center when omitted).
// The covered strip of each background line is replaced — ANSI-aware, with the
// foreground padded to its block width so background text cannot bleed
// through. A foreground taller than the background extends the result
// downward rather than being clipped.
func Overlay(bg, fg string, pos ...OverlayPos) string {
	p := Center
	if len(pos) > 0 {
		p = pos[0]
	}

	bgLines := strings.Split(bg, "\n")
	fgLines := strings.Split(fg, "\n")

	bgW := blockWidth(bgLines)
	fgW := blockWidth(fgLines)
	bgH := len(bgLines)
	fgH := len(fgLines)

	x := anchor(p.XFrac, bgW-fgW)
	y := anchor(p.YFrac, bgH-fgH)

	totalRows := bgH
	if y+fgH > totalRows {
		totalRows = y + fgH
	}

	out := make([]string, totalRows)
	for r := 0; r < totalRows; r++ {
		var bgLine string
		if r < len(bgLines) {
			bgLine = bgLines[r]
		}
		if r < y || r >= y+fgH {
			out[r] = bgLine
			continue
		}
		left := padTo(ansi.Truncate(bgLine, x, ""), x)
		mid := padTo(fgLines[r-y], fgW)
		right := ansi.TruncateLeft(bgLine, x+fgW, "")
		out[r] = left + mid + right
	}
	return strings.Join(out, "\n")
}

func blockWidth(lines []string) int {
	w := 0
	for _, l := range lines {
		if lw := ansi.StringWidth(l); lw > w {
			w = lw
		}
	}
	return w
}

func anchor(frac float64, span int) int {
	if span < 0 {
		span = 0
	}
	v := int(frac*float64(span) + 0.5)
	if v < 0 {
		return 0
	}
	if v > span {
		return span
	}
	return v
}
