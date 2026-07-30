package layout

import "github.com/codyconfer/viewkit/theme"

// CellBox renders a bordered box that fills exactly f.Width columns: the
// border and padding cost 4 columns, so the body is f.Width-4. Unlike
// NewFrame, a cell narrower than theme.MinBodyWidth+4 clamps the body *down*
// rather than up, so the box never spills past the rect its layout gave it and
// never loses its right border to the compositor.
func (f Frame) CellBox(title string, lines ...string) string {
	return f.fitCell(f.panelAt(f.cellBody(), title, lines...))
}

// CellPanel is CellBox with a scrolling body sized to f.Height.
func (f Frame) CellPanel(title string, lines []string, offset int) string {
	rows := f.Height - 3
	if rows < 1 {
		rows = 1
	}
	body := lines
	if window, footer, ok := scrollWindow(lines, rows, offset); ok {
		body = append(append(make([]string, 0, len(window)+1), window...), theme.Cur().Dim.Render(footer))
	}
	return f.fitCell(f.panelAt(f.cellBody(), title, body...))
}

// cellBody is the body width of a box that must fit inside f.Width. Frames
// with no usable width keep NewFrame's fallback to the default body width.
func (f Frame) cellBody() int {
	if f.Width-4 < 1 {
		return NewFrame(f.Width - 4).Width
	}
	return f.Width - 4
}

func (f Frame) fitCell(box string) string {
	switch {
	case f.Height > 0:
		return FitBlock(box, f.Width, f.Height)
	case f.Width > 0:
		return FitBlock(box, f.Width, CountLines(box))
	default:
		return box
	}
}
