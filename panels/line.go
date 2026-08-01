package panels

import (
	"strings"

	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/theme"
)

// Line renders a scatter-style line chart panel plotting series as one '•'
// per value, scaled between the series min and max. width and height are the
// plot size in terminal cells; width is capped to the frame body minus the
// axis gutter, and when the series is longer than width only the most recent
// values are shown (shorter series are right-aligned). fmtVal formats the
// max/min axis labels on the first and last rows. Any footer lines are
// appended inside the panel. An empty series or non-positive width/height
// yields a panel containing only the footer.
func Line(f layout.Frame, title string, series []float64, width, height int, fmtVal func(float64) string, footer ...string) string {
	if max := f.BodyWidth() - 7; max > 0 && width > max {
		width = max
	}
	lines := linePlot(series, width, height, fmtVal)
	lines = append(lines, footer...)
	return f.Panel(title, lines...)
}

func linePlot(series []float64, width, height int, fmtVal func(float64) string) []string {
	if len(series) == 0 || width < 1 || height < 1 {
		return nil
	}
	if len(series) > width {
		series = series[len(series)-width:]
	}

	lo, hi := seriesBounds(series)
	span := chartSpan(lo, hi)
	rowOf := func(v float64) int {
		return height - 1 - chartLevel(v, lo, span, height)
	}

	pad := width - len(series)
	grid := make([][]rune, height)
	for i := range grid {
		grid[i] = []rune(strings.Repeat(" ", width))
	}
	for x, v := range series {
		grid[rowOf(v)][pad+x] = '•'
	}

	out := make([]string, 0, height+1)
	for i, row := range grid {
		out = append(out, chartGutter(chartLabel(i, height-1, lo, hi, fmtVal))+theme.Cur().Can.Render(string(row)))
	}
	out = append(out, chartBaseline(width))
	return out
}
