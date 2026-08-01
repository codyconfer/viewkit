package panels

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/theme"
)

func fnum(f float64) string { return fmt.Sprintf("%.0f", f) }

var ansiRe = regexp.MustCompile("\x1b\\[[0-9;]*m")

func stripANSI(s string) string { return ansiRe.ReplaceAllString(s, "") }

func TestPlotEmpty(t *testing.T) {
	t.Parallel()
	if got := candlePlot(theme.Default(), nil, 10, 5, fnum); got != nil {
		t.Fatalf("candlePlot(nil) = %v, want nil", got)
	}
	if got := candlePlot(theme.Default(), []OHLC{{Open: 1, High: 2, Low: 0, Close: 1}}, 0, 5, fnum); got != nil {
		t.Fatalf("candlePlot(width 0) = %v, want nil", got)
	}
}

func TestPlotShape(t *testing.T) {
	t.Parallel()
	candles := []OHLC{
		{Open: 1, High: 5, Low: 1, Close: 4},
		{Open: 4, High: 6, Low: 3, Close: 5},
	}
	const height = 4
	lines := candlePlot(theme.Default(), candles, 10, height, fnum)
	if len(lines) != height+1 {
		t.Fatalf("candlePlot returned %d lines, want %d (height + axis)", len(lines), height+1)
	}
	if !strings.Contains(lines[0], "6") {
		t.Errorf("top row missing high label 6: %q", lines[0])
	}
	if !strings.Contains(lines[height-1], "1") {
		t.Errorf("bottom row missing low label 1: %q", lines[height-1])
	}
	if !strings.Contains(lines[len(lines)-1], "└") {
		t.Errorf("last line missing x-axis rule: %q", lines[len(lines)-1])
	}
}

func TestPlotTrimsToWidth(t *testing.T) {
	t.Parallel()
	candles := make([]OHLC, 50)
	for i := range candles {
		candles[i] = OHLC{Open: 1, High: 2, Low: 0, Close: 1}
	}
	lines := candlePlot(theme.Default(), candles, 8, 3, fnum)
	if len(lines) != 4 {
		t.Fatalf("candlePlot returned %d lines, want 4", len(lines))
	}
}

func TestCandleWrapsPlotWithFooter(t *testing.T) {
	t.Parallel()
	candles := []OHLC{{Open: 1, High: 5, Low: 1, Close: 4}}
	out := Candle(layout.DefaultFrame(), "EGG PRICE", candles, 10, 4, fnum, "now: 4")
	if !strings.Contains(out, "EGG PRICE") {
		t.Errorf("chart missing title:\n%s", out)
	}
	if !strings.Contains(out, "now: 4") {
		t.Errorf("chart missing footer:\n%s", out)
	}
}

func TestCandleRowsGeometry(t *testing.T) {
	t.Parallel()
	up := OHLC{Open: 2, High: 5, Low: 1, Close: 4}
	rows := candleRows(theme.Default(), []OHLC{up}, 1, 4, 1, 5)
	if len(rows) != 4 {
		t.Fatalf("got %d rows, want 4", len(rows))
	}
	for i, r := range rows {
		if w := len([]rune(stripANSI(r))); w != 1 {
			t.Fatalf("row %d width=%d, want 1", i, w)
		}
	}
	joined := stripANSI(strings.Join(rows, "\n"))
	if !strings.ContainsAny(joined, "█▀▄") {
		t.Error("expected a candle body glyph")
	}
	if !strings.Contains(joined, "│") {
		t.Error("a candle spanning a high/low range should draw a wick")
	}
}

func TestCandleDirectionOnlyChangesColor(t *testing.T) {
	t.Parallel()
	up := candleRows(theme.Default(), []OHLC{{Open: 2, High: 5, Low: 1, Close: 4}}, 1, 4, 1, 5)
	down := candleRows(theme.Default(), []OHLC{{Open: 4, High: 5, Low: 1, Close: 2}}, 1, 4, 1, 5)
	if stripANSI(strings.Join(up, "\n")) != stripANSI(strings.Join(down, "\n")) {
		t.Error("up and down candles with mirrored open/close should share geometry")
	}
}

func TestCandleRowsLeftPadsShortSeries(t *testing.T) {
	t.Parallel()
	rows := candleRows(theme.Default(), []OHLC{{2, 3, 2, 3}, {3, 4, 3, 4}}, 6, 3, 2, 4)
	for r, row := range rows {
		runes := []rune(stripANSI(row))
		for i := range 4 {
			if runes[i] != ' ' {
				t.Fatalf("row %d col %d: expected padding, got %q", r, i, runes[i])
			}
		}
	}
}
