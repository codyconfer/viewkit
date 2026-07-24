package list

import (
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

func (m *Model) SetSize(w, h int) { m.width, m.height = w, h }

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
	m.clampOffset(len(m.render()))
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

func (m *Model) ensureVisible() {
	if m.height <= 0 || m.cursor < 0 {
		return
	}
	start := m.itemStart(m.cursor)
	end := start + len(strings.Split(m.items[m.cursor].Block, "\n")) - 1
	if start < m.offset {
		m.offset = start
	}
	if end >= m.offset+m.height {
		m.offset = end - m.height + 1
	}
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
