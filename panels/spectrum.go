package panels

import (
	"strings"

	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/theme"
)

var vBlocks = [8]string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

// SpectrumOpts tunes Spectrum. Peaks holds optional peak-hold marks in the
// same 0..1 scale as the levels, drawn as a dim ▔ above each bar (extra
// entries beyond the levels are ignored). BarWide is each bar's width and
// BarGap the space between bars, both in terminal cells.
type SpectrumOpts struct {
	Peaks   []float64
	BarGap  int
	BarWide int
}

// Spectrum renders an audio-analyzer-style panel of vertical bars, one per
// level, cycling through the theme's series palette. Levels are fractions in
// 0..1 of the plot height (values outside that range, NaN, and Inf are
// clamped/zeroed) drawn with eighth-block glyphs for sub-cell resolution.
// height is the plot height in terminal cells; bands that don't fit the frame
// body width are dropped from the end. Only opts[0] is consulted; the default
// is 1-cell bars with a 1-cell gap, and BarWide/BarGap are clamped to at
// least 1/0. With no levels or height < 1 the panel shows empty in the dim
// style.
func Spectrum(f layout.Frame, title string, levels []float64, height int, empty string, opts ...SpectrumOpts) string {
	o := SpectrumOpts{BarWide: 1, BarGap: 1}
	if len(opts) > 0 {
		o = opts[0]
		if o.BarWide < 1 {
			o.BarWide = 1
		}
		if o.BarGap < 0 {
			o.BarGap = 0
		}
	}
	if len(levels) == 0 || height < 1 {
		return f.Panel(title, theme.Cur().Dim.Render(empty))
	}

	span := o.BarWide + o.BarGap
	maxBands := (f.BodyWidth() + o.BarGap) / span
	if maxBands < 1 {
		maxBands = 1
	}
	if len(levels) > maxBands {
		levels = levels[:maxBands]
	}
	peaks := o.Peaks
	if len(peaks) > len(levels) {
		peaks = peaks[:len(levels)]
	}

	series := theme.Cur().Series
	dim := theme.Cur().Dim
	rows := make([]string, height)
	for row := range height {
		cell := height - 1 - row
		var b strings.Builder
		for i, lvl := range levels {
			if i > 0 {
				b.WriteString(strings.Repeat(" ", o.BarGap))
			}
			glyph := spectrumCell(lvl, cell, height)
			switch {
			case glyph != "":
				b.WriteString(seriesAt(series, i).Render(strings.Repeat(glyph, o.BarWide)))
			case i < len(peaks) && peakCell(peaks[i], height) == cell:
				b.WriteString(dim.Render(strings.Repeat("▔", o.BarWide)))
			default:
				b.WriteString(strings.Repeat(" ", o.BarWide))
			}
		}
		rows[row] = b.String()
	}
	return f.Panel(title, rows...)
}

func spectrumCell(level float64, cell, height int) string {
	level = finite(level)
	if level < 0 {
		level = 0
	}
	if level > 1 {
		level = 1
	}
	eighths := int(level*float64(height)*8 + 0.5)
	n := eighths - cell*8
	switch {
	case n <= 0:
		return ""
	case n >= 8:
		return "█"
	default:
		return vBlocks[n-1]
	}
}

func peakCell(peak float64, height int) int {
	peak = finite(peak)
	if peak < 0 {
		peak = 0
	}
	if peak > 1 {
		peak = 1
	}
	eighths := int(peak*float64(height)*8 + 0.5)
	if eighths <= 0 {
		return -1
	}
	return (eighths - 1) / 8
}
