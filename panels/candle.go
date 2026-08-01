package panels

import (
	"strings"

	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/theme"
)

// OHLC is one candlestick's open, high, low, and close prices.
type OHLC struct{ Open, High, Low, Close float64 }

// Candle renders a candlestick chart panel at one column per candle, using
// half-block glyphs for two price levels per row and │ for wicks. Falling
// candles (Close < Open) use the negative style, others the positive one.
// width and height are the plot size in terminal cells; width is capped to
// the frame body minus the axis gutter, and when there are more candles than
// columns only the most recent ones are shown (fewer candles are right-
// aligned). fmtVal formats the high/low axis labels on the first and last
// rows. Any footer lines are appended inside the panel. Empty candles or a
// non-positive width/height yields a panel containing only the footer.
func Candle(f layout.Frame, title string, candles []OHLC, width, height int, fmtVal func(float64) string, footer ...string) string {
	if max := f.BodyWidth() - 7; max > 0 && width > max {
		width = max
	}
	lines := candlePlot(f.Theme(), candles, width, height, fmtVal)
	lines = append(lines, footer...)
	return f.Panel(title, lines...)
}

func candlePlot(th theme.Theme, candles []OHLC, width, height int, fmtVal func(float64) string) []string {
	if len(candles) == 0 || width < 1 || height < 1 {
		return nil
	}
	if len(candles) > width {
		candles = candles[len(candles)-width:]
	}

	lo, hi := candles[0].Low, candles[0].High
	for _, c := range candles {
		if c.Low < lo {
			lo = c.Low
		}
		if c.High > hi {
			hi = c.High
		}
	}

	body := candleRows(th, candles, width, height, lo, hi)
	lines := make([]string, 0, height+1)
	for i, row := range body {
		lines = append(lines, chartGutter(th, chartLabel(i, height-1, lo, hi, fmtVal))+row)
	}
	lines = append(lines, chartBaseline(th, width))
	return lines
}

func candleRows(th theme.Theme, candles []OHLC, width, height int, lo, hi float64) []string {
	span := chartSpan(lo, hi)
	levels := height * 2
	lvl := func(p float64) int {
		return chartLevel(p, lo, span, levels)
	}

	pad := max(width-len(candles), 0)
	rows := make([]string, height)
	for row := range height {
		cell := height - 1 - row
		upper, lower := 2*cell+1, 2*cell
		var b strings.Builder
		if pad > 0 {
			b.WriteString(strings.Repeat(" ", pad))
		}
		for _, c := range candles {
			hiL, loL := lvl(c.High), lvl(c.Low)
			top, bot := lvl(c.Open), lvl(c.Close)
			if bot > top {
				top, bot = bot, top
			}
			sty := th.Can
			if c.Close < c.Open {
				sty = th.Cant
			}
			bodyU := upper >= bot && upper <= top
			bodyL := lower >= bot && lower <= top
			switch {
			case bodyU && bodyL:
				b.WriteString(sty.Render("█"))
			case bodyU:
				b.WriteString(sty.Render("▀"))
			case bodyL:
				b.WriteString(sty.Render("▄"))
			case (upper >= loL && upper <= hiL) || (lower >= loL && lower <= hiL):
				b.WriteString(sty.Render("│"))
			default:
				b.WriteByte(' ')
			}
		}
		rows[row] = b.String()
	}
	return rows
}
