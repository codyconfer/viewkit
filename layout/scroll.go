package layout

import (
	"fmt"
	"strings"

	"github.com/codyconfer/viewkit/theme"
)

// ScrollState is a scroll offset (in lines) with clamped mutation helpers.
// Embed or hold it by value in a view and pass the current totals into Scroll
// and Reveal; the state itself does not know the content size.
type ScrollState struct {
	Offset int
}

// Scroll moves the offset by delta lines and clamps it to [0, total-rows],
// where total is the line count and rows the visible window height.
func (s *ScrollState) Scroll(delta, total, rows int) {
	s.Offset += delta
	s.clamp(total, rows)
}

// Reveal adjusts the offset just enough to bring line index into the visible
// window, then clamps like Scroll. Lines already visible leave the offset
// untouched.
func (s *ScrollState) Reveal(index, total, rows int) {
	if rows < 1 {
		rows = 1
	}
	if index < s.Offset {
		s.Offset = index
	} else if index >= s.Offset+rows {
		s.Offset = index - rows + 1
	}
	s.clamp(total, rows)
}

func (s *ScrollState) clamp(total, rows int) {
	max := total - rows
	if max < 0 {
		max = 0
	}
	if s.Offset > max {
		s.Offset = max
	}
	if s.Offset < 0 {
		s.Offset = 0
	}
}

func scrollWindow(lines []string, rows, offset int) (window []string, footer string, ok bool) {
	total := len(lines)
	if rows < 1 {
		rows = 1
	}
	if total <= rows {
		return lines, "", false
	}
	max := total - rows
	if offset > max {
		offset = max
	}
	if offset < 0 {
		offset = 0
	}
	end := offset + rows
	return lines[offset:end], fmt.Sprintf("↕ %d–%d of %d", offset+1, end, total), true
}

// ScrollPanel is Frame.ScrollPanel on the default-width frame.
func ScrollPanel(title string, lines []string, rows, offset int) string {
	return DefaultFrame().ScrollPanel(title, lines, rows, offset)
}

// ScrollPanel renders a Panel showing a rows-high window of lines starting at
// offset. When lines overflow the window a dim "↕ a–b of n" footer line is
// appended inside the panel; offsets are clamped, never wrapped.
func (f Frame) ScrollPanel(title string, lines []string, rows, offset int) string {
	return f.ScrollPanelWithPrefix(title, nil, lines, rows, offset)
}

// ScrollPanelWithPrefix is ScrollPanel with fixed prefix lines pinned above
// the scrolling window (they do not scroll and do not count against rows).
// With no lines at all, the panel shows just the prefix.
func (f Frame) ScrollPanelWithPrefix(title string, prefix, lines []string, rows, offset int) string {
	if len(lines) == 0 {
		return f.Panel(title, prefix...)
	}
	window, footer, ok := scrollWindow(lines, rows, offset)
	out := make([]string, 0, len(prefix)+len(window)+1)
	out = append(out, prefix...)
	out = append(out, window...)
	if ok {
		out = append(out, theme.Cur().Dim.Render(footer))
	}
	return f.Panel(title, out...)
}

// Viewport shows a scrolled window of body within a rows budget. A body that
// fits is returned whole; otherwise the last row (plus, at rows >= 3, a blank
// margin row) is spent on a "▲▼ pgup/pgdn · a–b of n" hint, so the content
// window is ViewportContentRows(rows) tall. Offsets are clamped; rows < 1
// returns "".
func Viewport(body string, rows, offset int) string {
	lines := strings.Split(body, "\n")
	if rows < 1 {
		return ""
	}
	total := len(lines)
	if total <= rows {
		return body
	}
	if rows == 1 {
		off := clampOffset(offset, total, 1)
		return viewportHint(off, off+1, total)
	}

	margin := 0
	if rows >= 3 {
		margin = 1
	}
	windowRows := rows - 1 - margin
	off := clampOffset(offset, total, windowRows)
	end := off + windowRows
	out := make([]string, 0, windowRows+1+margin)
	out = append(out, lines[off:end]...)
	if margin == 1 {
		out = append(out, "")
	}
	out = append(out, viewportHint(off, end, total))
	return strings.Join(out, "\n")
}

// ViewportContentRows reports how many content rows Viewport shows for a
// given budget once the hint line and optional margin row are paid for —
// use it to clamp offsets consistently with Viewport's own windowing.
func ViewportContentRows(rows int) int {
	if rows < 2 {
		return 0
	}
	margin := 0
	if rows >= 3 {
		margin = 1
	}
	return rows - 1 - margin
}

func clampOffset(offset, total, rows int) int {
	max := total - rows
	if max < 0 {
		max = 0
	}
	if offset > max {
		offset = max
	}
	if offset < 0 {
		offset = 0
	}
	return offset
}

func viewportHint(offset, end, total int) string {
	up, down := "  ", "  "
	if offset > 0 {
		up = "▲ "
	}
	if end < total {
		down = "▼ "
	}
	return theme.Cur().Dim.Render(fmt.Sprintf("%s%s pgup/pgdn  ·  %d–%d of %d", up, down, offset+1, end, total))
}
