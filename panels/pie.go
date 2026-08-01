package panels

import (
	"fmt"
	"strings"

	"github.com/codyconfer/viewkit/layout"
)

// Pie renders proportions as a panel with a single stacked horizontal bar —
// one block-glyph segment per positive Datum, colored from the theme's series
// palette — followed by a legend line per segment showing a color swatch, the
// label, the fmtNum-formatted value, and its rounded percentage of the total.
// barWidth is the bar length in terminal cells, clamped to [1, body width];
// rounding never overflows it because later segments are capped to the cells
// left. Zero, negative, and NaN/Inf values are skipped entirely (bar and
// legend), and if no value is positive the panel shows empty in the dim
// style.
func Pie(f layout.Frame, title string, data []Datum, barWidth int, fmtNum func(float64) string, empty string) string {
	total := 0.0
	for _, d := range data {
		if v := finite(d.Value); v > 0 {
			total += v
		}
	}
	th := f.Theme()
	if total = finite(total); total <= 0 {
		return f.Panel(title, th.Dim.Render(empty))
	}
	if barWidth < 1 {
		barWidth = 1
	}
	if barWidth > f.BodyWidth() {
		barWidth = f.BodyWidth()
	}

	var bar strings.Builder
	var legend []string
	filled := 0
	for i, d := range data {
		v := finite(d.Value)
		if v <= 0 {
			continue
		}
		frac := v / total
		sty := seriesAt(th, i)
		n := min(max(int(frac*float64(barWidth)+0.5), 0), barWidth-filled)
		filled += n
		bar.WriteString(sty.Render(strings.Repeat("█", n)))
		legend = append(legend, f.Spread(
			sty.Render("■ ")+th.Val.Render(d.Label),
			th.Dim.Render(fmt.Sprintf("%s  ·  %.0f%%", fmtNum(d.Value), frac*100))))
	}

	lines := append([]string{bar.String()}, legend...)
	return f.Panel(title, lines...)
}
