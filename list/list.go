package list

import (
	"slices"
	"strings"

	"github.com/codyconfer/viewkit/theme"
)

type Item struct {
	Block      string
	Key        string
	Selectable bool
	GapStem    string
}

type Model struct {
	items   []Item
	cursor  int
	offset  int
	width   int
	height  int
	focused bool
}

func New() Model { return Model{cursor: -1} }

func (m *Model) SetItems(items []Item) {
	m.items = items
	m.cursor = m.firstSelectable()
	m.offset = 0
}

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

func (m *Model) SetSize(w, h int) { m.width, m.height = w, h }

func (m *Model) EnsureVisible() { m.ensureVisible() }

func (m *Model) Height() int { return m.height }

func (m *Model) SetFocused(f bool) { m.focused = f }

func (m *Model) Selectable() bool { return m.cursor >= 0 }

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

func (m *Model) Scroll(delta int) {
	m.offset += delta
	m.clampOffset(m.totalLines())
	m.syncCursor()
}

func (m *Model) render() []string {
	th := theme.Cur()
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

func (m *Model) View() string {
	lines := m.render()
	if m.height <= 0 || len(lines) <= m.height {
		return strings.Join(lines, "\n")
	}
	m.clampOffset(len(lines))
	end := min(m.offset+m.height, len(lines))
	return strings.Join(lines[m.offset:end], "\n")
}
