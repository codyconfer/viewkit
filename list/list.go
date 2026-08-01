// Package list renders a cursor-and-scroll list over pre-rendered row blocks.
//
// Rows arrive as Items whose Block text is already styled and wrapped; the
// model owns only selection and viewport state. It is not a tea model — deck
// views and CLI panes drive it by calling Move/Scroll and printing View.
package list

import (
	"slices"
	"strings"

	"github.com/codyconfer/viewkit/theme"
)

// Item is one list row. Block is the pre-rendered (possibly multi-line) body;
// Key identifies the row to the caller; unselectable rows are skipped by the
// cursor. GapStem, when set, is repeated in the blank line drawn after the row
// so tree rules survive the gap.
type Item struct {
	Block      string
	Key        string
	Selectable bool
	GapStem    string
	// Payload carries the caller's domain value for the row, handed back by
	// Selected and selection callbacks so callers need no key→value side table.
	Payload any
}

// Model is the list state: rows plus cursor, scroll offset, size, and focus.
// Callers mutate it with SetItems/Move/Scroll and render it with View; it is
// not a Bubble Tea model and has no Update loop.
type Model struct {
	items   []Item
	cursor  int
	offset  int
	width   int
	height  int
	focused bool
	th      *theme.Theme
}

// New returns an empty model with no selection.
func New() Model { return Model{cursor: -1} }

// SetTheme pins the theme View renders with; unset, the process default applies.
func (m *Model) SetTheme(t theme.Theme) { m.th = &t }

// SetItems replaces the rows and resets cursor and scroll to the top.
func (m *Model) SetItems(items []Item) {
	m.items = items
	m.cursor = m.firstSelectable()
	m.offset = 0
}

// SetItemsKeepingCursor replaces the rows while trying to keep the current
// selection: first by Key, then by index, falling back to the first
// selectable row.
func (m *Model) SetItemsKeepingCursor(items []Item) {
	prevIdx, prevOffset := m.cursor, m.offset
	hadCursor := m.cursor >= 0
	prevKey := ""
	if it, ok := m.Selected(); ok {
		prevKey = it.Key
	}
	m.items = items
	if i := m.reacquire(prevKey, prevIdx, hadCursor); i >= 0 {
		m.cursor, m.offset = i, prevOffset
		m.clampOffset(m.totalLines())
		m.ensureVisible()
		return
	}
	m.cursor = m.firstSelectable()
	if !hadCursor && m.cursor < 0 {
		m.offset = prevOffset
		m.clampOffset(m.totalLines())
		return
	}
	m.offset = 0
}

func (m *Model) reacquire(key string, idx int, hadCursor bool) int {
	switch {
	case !hadCursor:
		return -1
	case key != "":
		return m.indexOfSelectable(key)
	case idx >= 0 && idx < len(m.items) && m.items[idx].Selectable:
		return idx
	}
	return -1
}

func (m *Model) indexOfSelectable(key string) int {
	for i, it := range m.items {
		if it.Selectable && it.Key == key {
			return i
		}
	}
	return -1
}

// SetSize sets the viewport dimensions in cells; height bounds View's output.
func (m *Model) SetSize(w, h int) { m.width, m.height = w, h }

// EnsureVisible scrolls just enough to bring the selected row into view.
func (m *Model) EnsureVisible() { m.ensureVisible() }

// Height returns the viewport height set by SetSize.
func (m *Model) Height() int { return m.height }

// SetFocused switches the cursor glyph between focused and dimmed styles.
func (m *Model) SetFocused(f bool) { m.focused = f }

// Selectable reports whether any row currently holds the cursor.
func (m *Model) Selectable() bool { return m.cursor >= 0 }

// Selected returns the row under the cursor, if any.
func (m *Model) Selected() (Item, bool) {
	if m.cursor < 0 || m.cursor >= len(m.items) {
		return Item{}, false
	}
	return m.items[m.cursor], true
}

func (m *Model) firstSelectable() int {
	for i, it := range m.items {
		if it.Selectable {
			return i
		}
	}
	return -1
}

func (m *Model) lastSelectable() int {
	for i, it := range slices.Backward(m.items) {
		if it.Selectable {
			return i
		}
	}
	return -1
}

// Move steps the cursor by delta over selectable rows, scrolling as needed;
// with no selectable rows it scrolls the viewport instead.
func (m *Model) Move(delta int) {
	if m.cursor < 0 {
		m.Scroll(delta)
		return
	}
	i := m.cursor
	for {
		i += delta
		if i < 0 || i >= len(m.items) {
			return
		}
		if m.items[i].Selectable {
			m.cursor = i
			m.ensureVisible()
			return
		}
	}
}

// Scroll shifts the viewport by delta lines and drags the cursor along so it
// stays on a visible row.
func (m *Model) Scroll(delta int) {
	m.offset += delta
	m.clampOffset(m.totalLines())
	m.syncCursor()
}

func (m *Model) render() []string {
	th := theme.Default()
	if m.th != nil {
		th = *m.th
	}
	var out []string
	for i, it := range m.items {
		if i > 0 {
			for range theme.ListItemGapY {
				out = append(out, gapLine(m.items[i-1]))
			}
		}
		selected := i == m.cursor
		for j, bl := range strings.Split(it.Block, "\n") {
			prefix := "  "
			switch {
			case !it.Selectable:
				prefix = ""
			case selected && j == 0 && m.focused:
				prefix = th.Key.Render("▸ ")
			case selected && j == 0:
				prefix = th.Dim.Render("▸ ")
			}
			out = append(out, prefix+bl)
		}
	}
	return out
}

func gapLine(prev Item) string {
	if prev.GapStem == "" {
		return ""
	}
	return "  " + prev.GapStem
}

func (m *Model) itemStart(idx int) int {
	line := 0
	for i, it := range m.items {
		if i == idx {
			return line
		}
		line += len(strings.Split(it.Block, "\n")) + theme.ListItemGapY
	}
	return line
}

func (m *Model) totalLines() int { return len(m.render()) }

func (m *Model) itemSpan(idx int) (int, int) {
	start := m.itemStart(idx)
	return start, start + len(strings.Split(m.items[idx].Block, "\n")) - 1
}

func (m *Model) ensureVisible() {
	if m.height <= 0 || m.cursor < 0 {
		return
	}
	start, end := m.itemSpan(m.cursor)
	lo, hi := start, end
	if m.cursor == m.firstSelectable() {
		lo = 0
	}
	if m.cursor == m.lastSelectable() {
		hi = m.totalLines() - 1
	}
	if hi >= m.offset+m.height {
		m.offset = hi - m.height + 1
	}
	if lo < m.offset {
		m.offset = lo
	}
	if end >= m.offset+m.height {
		m.offset = end - m.height + 1
	}
	if start < m.offset {
		m.offset = start
	}
}

func (m *Model) syncCursor() {
	if m.cursor < 0 || m.height <= 0 {
		return
	}
	if start, end := m.itemSpan(m.cursor); start >= m.offset && end < m.offset+m.height {
		return
	}
	best := m.nearestSelectable(true)
	if best < 0 {
		best = m.nearestSelectable(false)
	}
	if best >= 0 {
		m.cursor = best
	}
}

func (m *Model) nearestSelectable(whole bool) int {
	best, bestDist, line := -1, 0, 0
	for i, it := range m.items {
		start := line
		end := start + len(strings.Split(it.Block, "\n")) - 1
		line = end + 1 + theme.ListItemGapY
		if !it.Selectable {
			continue
		}
		var fits bool
		if whole {
			fits = start >= m.offset && end < m.offset+m.height
		} else {
			fits = end >= m.offset && start < m.offset+m.height
		}
		if !fits {
			continue
		}
		if d := max(i-m.cursor, m.cursor-i); best < 0 || d < bestDist {
			best, bestDist = i, d
		}
	}
	return best
}

func (m *Model) clampOffset(total int) {
	maxOff := max(total-m.height, 0)
	if m.offset > maxOff {
		m.offset = maxOff
	}
	if m.offset < 0 {
		m.offset = 0
	}
}

// View renders the visible window of rows as a plain string.
func (m *Model) View() string {
	lines := m.render()
	if m.height <= 0 || len(lines) <= m.height {
		return strings.Join(lines, "\n")
	}
	m.clampOffset(len(lines))
	end := min(m.offset+m.height, len(lines))
	return strings.Join(lines[m.offset:end], "\n")
}
