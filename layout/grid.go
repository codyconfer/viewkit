package layout

import (
	"sort"
	"strings"

	"github.com/charmbracelet/x/ansi"

	"github.com/codyconfer/viewkit/theme"
)

// GridPos pins a pane to grid coordinates. Col/Row are zero-based; spans below
// 1 count as 1, and out-of-range values are clamped into the grid. Panes
// without a Pos cascade into the next free block instead.
type GridPos struct {
	Col     int
	Row     int
	ColSpan int
	RowSpan int
}

// Grid is a tiling Arranger: the frame is divided into Cols equal-width column
// tracks and panes are placed by GridPos (or auto-flowed), then composited to
// fill the frame exactly. Rows is a minimum row count; the grid grows past it
// as panes need. Panes tiled into grid tracks should render with the CellBox /
// CellTitledBox family, which stay inside their rect — see MinTrackWidth.
type Grid struct {
	Cols int
	Rows int
}

// SlimMinWidth is the narrowest a Slim pane will be squeezed to (in cells)
// before it stops shrinking.
const SlimMinWidth = 20

// MinTrackWidth is the narrowest column a tiling layout will create. It is the
// same floor Slim panes use, and it is what Frame.CellBox needs to stay inside
// the rect its layout hands it. It is not enough for Frame.TitledBox, which
// clamps its body up to theme.MinBodyWidth and therefore needs
// theme.MinBodyWidth+4 columns; panes that tile into narrow tracks want
// Frame.CellTitledBox, which clamps down instead.
const MinTrackWidth = SlimMinWidth

type gridCell struct {
	x, y, w, h int
}

type rect struct {
	x, y, w, h int
}

// Arrange implements Arranger. Column count is reduced with FitCols so no
// track drops below MinTrackWidth (pinned panes then reflow via auto-flow);
// Slim panes give up width to their row-mates; and the result is composited to
// exactly f.Width x f.Height, truncating pane output that overruns its rect.
// A frame without a height (f.Height < 1) falls back to SingleColumn.
func (g Grid) Arrange(f Frame, tier Tier, panes []Pane, focusedName string) string {
	height := f.Height
	if height < 1 {
		return SingleColumn{}.Arrange(f, tier, panes, focusedName)
	}
	visible := make([]Pane, 0, len(panes))
	for _, p := range panes {
		if tier >= p.MinTier {
			visible = append(visible, p)
		}
	}
	if len(visible) == 0 {
		return strings.Repeat("\n", height-1)
	}

	width := f.Width
	if width < 1 {
		width = theme.BodyWidth
	}
	fitted := g
	fitted.Cols = FitCols(g.Cols, width)
	slots := gridSlots(visible)
	if fitted.Cols < g.Cols {
		slots = autoFlow(slots, fitted.Cols)
	}
	cells, cols, rows := fitted.place(slots)
	rects := make([]rect, len(visible))
	slim := make([]bool, len(visible))
	for i, p := range visible {
		rects[i] = pixelRect(cells[i], cols, rows, width, height)
		slim[i] = p.Slim
	}
	applySlim(cells, slim, rects)

	blocks := make([]string, len(visible))
	for i, p := range visible {
		r := rects[i]
		pf := Frame{Width: r.w, Height: r.h, Focused: p.Focused(focusedName)}
		blocks[i] = p.Render(pf)
	}
	return composite(width, height, rects, blocks)
}

// FitCols reduces a requested column count until every column is at least
// MinTrackWidth wide, never dropping below a single column. A width below one
// cannot hold even a single track, so it also yields a single column.
func FitCols(cols, width int) int {
	if cols < 1 {
		cols = 1
	}
	if width < 1 {
		return 1
	}
	return min(cols, max(width/MinTrackWidth, 1))
}

// gridSlot is the placement intent for one pane. A pinned slot keeps the
// coordinates its caller declared; an unpinned slot cascades into the next free
// block while still honouring the spans it declared.
type gridSlot struct {
	pos    GridPos
	pinned bool
}

func gridSlots(panes []Pane) []gridSlot {
	slots := make([]gridSlot, len(panes))
	for i, p := range panes {
		if p.Pos != nil {
			slots[i] = gridSlot{pos: *p.Pos, pinned: true}
		}
	}
	return slots
}

// autoFlow drops explicit coordinates so panes reflow into the narrowed column
// count instead of colliding on clamped ones. Spans survive, clamped to the
// columns that are left, so a declared full-width header stays full width.
func autoFlow(slots []gridSlot, cols int) []gridSlot {
	out := make([]gridSlot, len(slots))
	copy(out, slots)
	for i := range out {
		if out[i].pinned {
			out[i] = unpin(out[i], cols, len(out))
		}
	}
	return out
}

// unpin keeps the spans and drops the coordinates. A cascading pane searches for
// a free block, so its row span is also capped at the pane count: no span can
// usefully reach deeper than there are panes to stack, and an unbounded one would
// make that search walk (and occupy) cells forever.
func unpin(s gridSlot, cols, panes int) gridSlot {
	return gridSlot{pos: GridPos{
		ColSpan: min(s.pos.ColSpan, cols),
		RowSpan: min(s.pos.RowSpan, max(panes, 1)),
	}}
}

// place assigns every pane a cell. Pinned slots whose cells would overlap an
// already-claimed cell are unpinned and cascade instead: two panes declaring the
// same coordinates both get drawn, because dropping one silently is the one
// outcome a layout must never produce.
func (g Grid) place(slots []gridSlot) (cells []gridCell, cols, rows int) {
	cols = g.Cols
	if cols < 1 {
		cols = 1
	}
	cells = make([]gridCell, len(slots))
	occupied := map[[2]int]bool{}
	for i := range slots {
		if !slots[i].pinned {
			continue
		}
		c := cellFor(slots[i].pos, cols)
		if !freeBlock(occupied, c) {
			slots[i] = unpin(slots[i], cols, len(slots))
			continue
		}
		cells[i] = c
		occupy(occupied, c)
	}
	for i := range slots {
		if slots[i].pinned {
			continue
		}
		c := cellFor(slots[i].pos, cols)
		c.x, c.y = nextFreeBlock(occupied, cols, c.w, c.h)
		cells[i] = c
		occupy(occupied, c)
	}
	for i := range slots {
		if slots[i].pinned || slots[i].pos.ColSpan >= 1 || cells[i].w >= cols {
			continue
		}
		if soleInRows(cells, i) {
			cells[i].x = 0
			cells[i].w = cols
		}
	}
	rows = g.Rows
	for _, c := range cells {
		if c.y+c.h > rows {
			rows = c.y + c.h
		}
	}
	if rows < 1 {
		rows = 1
	}
	return cells, cols, rows
}

func soleInRows(cells []gridCell, i int) bool {
	for j := range cells {
		if j == i {
			continue
		}
		if cells[j].y < cells[i].y+cells[i].h && cells[i].y < cells[j].y+cells[j].h {
			return false
		}
	}
	return true
}

func applySlim(cells []gridCell, slim []bool, rects []rect) {
	byRow := map[int][]int{}
	for i := range cells {
		byRow[cells[i].y] = append(byRow[cells[i].y], i)
	}
	for _, idxs := range byRow {
		var freed, stdBase, totalW int
		hasSlim := false
		for _, i := range idxs {
			totalW += rects[i].w
			if slim[i] {
				hasSlim = true
				freed += rects[i].w - slimWidth(rects[i].w)
			} else {
				stdBase += rects[i].w
			}
		}
		if !hasSlim || stdBase == 0 {
			continue
		}

		sort.SliceStable(idxs, func(a, b int) bool { return rects[idxs[a]].x < rects[idxs[b]].x })
		rightEdge := rects[idxs[0]].x + totalW
		x := rects[idxs[0]].x
		for n, i := range idxs {
			w := slimWidth(rects[i].w)
			if !slim[i] {
				w = rects[i].w + freed*rects[i].w/stdBase
			}
			if n == len(idxs)-1 {
				w = rightEdge - x
			}
			rects[i].x = x
			rects[i].w = w
			x += w
		}
	}
}

func slimWidth(w int) int {
	if half := w / 2; half >= SlimMinWidth {
		return half
	}
	if w < SlimMinWidth {
		return w
	}
	return SlimMinWidth
}

func cellFor(p GridPos, cols int) gridCell {
	w := p.ColSpan
	if w < 1 {
		w = 1
	}
	h := p.RowSpan
	if h < 1 {
		h = 1
	}
	x := p.Col
	if x < 0 {
		x = 0
	}
	if x > cols-1 {
		x = cols - 1
	}
	if x+w > cols {
		w = cols - x
	}
	if w < 1 {
		w = 1
	}
	y := p.Row
	if y < 0 {
		y = 0
	}
	return gridCell{x: x, y: y, w: w, h: h}
}

func occupy(occ map[[2]int]bool, c gridCell) {
	for dy := 0; dy < c.h; dy++ {
		for dx := 0; dx < c.w; dx++ {
			occ[[2]int{c.x + dx, c.y + dy}] = true
		}
	}
}

func freeBlock(occ map[[2]int]bool, c gridCell) bool {
	for dy := 0; dy < c.h; dy++ {
		for dx := 0; dx < c.w; dx++ {
			if occ[[2]int{c.x + dx, c.y + dy}] {
				return false
			}
		}
	}
	return true
}

func nextFreeBlock(occ map[[2]int]bool, cols, w, h int) (int, int) {
	w = min(max(w, 1), cols)
	h = max(h, 1)
	for y := 0; ; y++ {
		for x := 0; x+w <= cols; x++ {
			if freeBlock(occ, gridCell{x: x, y: y, w: w, h: h}) {
				return x, y
			}
		}
	}
}

func pixelRect(c gridCell, cols, rows, width, height int) rect {
	x := c.x * width / cols
	xEnd := (c.x + c.w) * width / cols
	y := c.y * height / rows
	yEnd := (c.y + c.h) * height / rows
	return rect{x: x, y: y, w: xEnd - x, h: yEnd - y}
}

type segment struct {
	x    int
	text string
}

func composite(width, height int, rects []rect, blocks []string) string {
	rowSegs := make([][]segment, height)
	for i, r := range rects {
		lines := fitLines(blocks[i], r.w, r.h)
		for dy := 0; dy < r.h; dy++ {
			ry := r.y + dy
			if ry < 0 || ry >= height {
				continue
			}
			rowSegs[ry] = append(rowSegs[ry], segment{x: r.x, text: lines[dy]})
		}
	}
	out := make([]string, height)
	for y := 0; y < height; y++ {
		segs := rowSegs[y]
		sort.SliceStable(segs, func(a, b int) bool { return segs[a].x < segs[b].x })
		var b strings.Builder
		cursor := 0
		for _, s := range segs {
			if s.x > cursor {
				b.WriteString(strings.Repeat(" ", s.x-cursor))
				cursor = s.x
			}
			b.WriteString(s.text)
			cursor += ansi.StringWidth(s.text)
		}
		out[y] = padTo(b.String(), width)
	}
	return strings.Join(out, "\n")
}

// FitBlock forces block to exactly w x h cells: each line is ANSI-aware
// truncated or space-padded to w, and rows are dropped or blank-padded to h.
func FitBlock(block string, w, h int) string {
	return strings.Join(fitLines(block, w, h), "\n")
}

func fitLines(block string, w, h int) []string {
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}
	raw := strings.Split(block, "\n")
	out := make([]string, h)
	blank := strings.Repeat(" ", w)
	for i := 0; i < h; i++ {
		if i < len(raw) {
			out[i] = padTo(raw[i], w)
		} else {
			out[i] = blank
		}
	}
	return out
}

func padTo(line string, w int) string {
	if w < 0 {
		w = 0
	}
	lw := ansi.StringWidth(line)
	switch {
	case lw > w:
		return ansi.Truncate(line, w, "")
	case lw < w:
		return line + strings.Repeat(" ", w-lw)
	default:
		return line
	}
}
