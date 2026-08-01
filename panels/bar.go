package panels

import (
	"strings"

	"github.com/charmbracelet/x/ansi"

	"github.com/codyconfer/viewkit/layout"
)

// Datum is one labeled value for the Bar, BarScroll, and Pie charts.
type Datum struct {
	Label string
	Value float64
}

// Bar renders a horizontal bar chart panel: one row per Datum with the label
// on the left, a bar of block glyphs scaled against the largest absolute
// value, and the fmtNum-formatted value right-aligned. Negative values get the
// "can't" (negative) style; NaN/Inf values are treated as 0. width is the
// maximum bar length in terminal cells and is shrunk to fit the frame body
// after reserving room for the widest label and formatted value. With no data
// the panel shows empty in the dim style instead.
func Bar(f layout.Frame, title string, data []Datum, width int, fmtNum func(float64) string, empty string) string {
	lines, ok := barLines(f, data, width, fmtNum)
	if !ok {
		return f.Panel(title, f.Theme().Dim.Render(empty))
	}
	return f.Panel(title, lines...)
}

// BarScroll is Bar in a scroll panel: only visible rows are shown starting at
// row offset, with a scroll-position footer when the data overflows. Bars are
// still scaled against the whole dataset, not just the visible window. With no
// data the panel shows empty in the dim style.
func BarScroll(f layout.Frame, title string, data []Datum, width int, fmtNum func(float64) string, empty string, visible, offset int) string {
	lines, ok := barLines(f, data, width, fmtNum)
	if !ok {
		return f.Panel(title, f.Theme().Dim.Render(empty))
	}
	return f.ScrollPanel(title, lines, visible, offset)
}

func barLines(f layout.Frame, data []Datum, width int, fmtNum func(float64) string) ([]string, bool) {
	if len(data) == 0 {
		return nil, false
	}
	if width < 1 {
		width = 1
	}
	peak, labelW, valueW := 0.0, 0, 0
	for _, d := range data {
		if a := absf(finite(d.Value)); a > peak {
			peak = a
		}
		if w := ansi.StringWidth(d.Label); w > labelW {
			labelW = w
		}
		if w := ansi.StringWidth(fmtNum(finite(d.Value))); w > valueW {
			valueW = w
		}
	}
	if peak = finite(peak); peak == 0 {
		peak = 1
	}
	available := f.BodyWidth() - labelW - valueW - 2
	if available < 1 {
		available = 1
	}
	if width > available {
		width = available
	}
	th := f.Theme()
	lines := make([]string, len(data))
	for i, d := range data {
		v := finite(d.Value)
		n := min(max(int(absf(v)/peak*float64(width)+0.5), 0), width)
		sty := th.Can
		if v < 0 {
			sty = th.Cant
		}
		label := th.Dim.Render(padRight(d.Label, labelW))
		bar := sty.Render(strings.Repeat("█", n))
		lines[i] = f.Spread(label+" "+bar, sty.Render(fmtNum(v)))
	}
	return lines, true
}

func absf(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func padRight(s string, w int) string {
	if gap := w - ansi.StringWidth(s); gap > 0 {
		return s + strings.Repeat(" ", gap)
	}
	return s
}
