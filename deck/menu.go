package deck

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/theme"
)

// MenuItem is one row in a Menu.
type MenuItem struct {
	Label string
	Desc  string
	Icon  string
	Hue   int
	Do    func(h *Host) tea.Cmd
}

// Menu is a simple navigable list View.
type Menu struct {
	title  string
	items  []MenuItem
	cursor int
	ctx    [][2]string
}

// NewMenu builds a Menu view.
func NewMenu(title string, ctx [][2]string, items ...MenuItem) *Menu {
	return &Menu{title: title, items: items, ctx: ctx}
}

func (m *Menu) Title() string        { return m.title }
func (m *Menu) Init() tea.Cmd        { return nil }
func (m *Menu) Context() [][2]string { return m.ctx }
func (m *Menu) Hints() [][2]string {
	return [][2]string{{"↑/↓", "move"}, {"enter", "open"}}
}

func (m *Menu) Update(h *Host, msg tea.Msg) tea.Cmd {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}
	sc := keys.Cur()
	km := keys.NewMap(
		sc.Binding(keys.Up),
		sc.Binding(keys.Down),
		sc.Binding(keys.Confirm),
		sc.Binding(keys.Cancel),
		sc.Binding(keys.Quit),
	)
	act, ok := km.Action(key.String())
	if !ok {
		return nil
	}
	switch act {
	case keys.Up:
		if m.cursor > 0 {
			m.cursor--
		}
	case keys.Down:
		if m.cursor < len(m.items)-1 {
			m.cursor++
		}
	case keys.Confirm:
		if len(m.items) > 0 && m.items[m.cursor].Do != nil {
			return m.items[m.cursor].Do(h)
		}
	case keys.Cancel:
		return h.Pop()
	}
	return nil
}

func (m *Menu) Body(width, _ int) string {
	th := theme.Cur()
	f := layout.ScreenFrame(width)
	var lines []string
	for i, it := range m.items {
		cursor := "  "
		label := th.Val.Render(it.Label)
		if i == m.cursor {
			cursor = th.Key.Render("▸ ")
			label = th.Key.Render(it.Label)
		}
		row := cursor
		if it.Icon != "" {
			row += theme.Icon(it.Icon, it.Hue)
		}
		row += label
		if it.Desc != "" {
			row = f.Spread(row, th.Dim.Render(it.Desc))
		}
		lines = append(lines, row)
	}
	return f.TitledBox(strings.ToUpper(m.title), lines...)
}

// Message is a dismissible text View.
type Message struct {
	title string
	body  string
	ctx   [][2]string
}

// NewMessage builds a Message view.
func NewMessage(title, body string, ctx [][2]string) *Message {
	return &Message{title: title, body: body, ctx: ctx}
}

func (m *Message) Title() string        { return m.title }
func (m *Message) Init() tea.Cmd        { return nil }
func (m *Message) Context() [][2]string { return m.ctx }
func (m *Message) Hints() [][2]string   { return nil }

func (m *Message) Update(h *Host, msg tea.Msg) tea.Cmd {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}
	sc := keys.Cur()
	km := keys.NewMap(sc.Binding(keys.Cancel), sc.Binding(keys.Confirm), sc.Binding(keys.Quit))
	if act, ok := km.Action(key.String()); ok && (act == keys.Cancel || act == keys.Confirm) {
		return h.Pop()
	}
	return nil
}

func (m *Message) Body(width, _ int) string {
	f := layout.ScreenFrame(width)
	return f.TitledBox(strings.ToUpper(m.title), strings.Split(m.body, "\n")...)
}

// ScrollContent is a lazy-loaded scrollable text View.
type ScrollContent struct {
	title  string
	load   func() string
	hints  [][2]string
	ctx    [][2]string
	body   string
	loaded bool
	offset int
}

// NewScrollContent builds a content view that loads asynchronously.
func NewScrollContent(title string, ctx, hints [][2]string, load func() string) *ScrollContent {
	return &ScrollContent{title: title, load: load, ctx: ctx, hints: hints}
}

type scrollLoadedMsg struct{ body string }

func (c *ScrollContent) Title() string        { return c.title }
func (c *ScrollContent) Context() [][2]string { return c.ctx }
func (c *ScrollContent) Hints() [][2]string {
	return append([][2]string{{"↑/↓", "scroll"}}, c.hints...)
}

func (c *ScrollContent) Init() tea.Cmd {
	return func() tea.Msg { return scrollLoadedMsg{body: c.load()} }
}

func (c *ScrollContent) Update(h *Host, msg tea.Msg) tea.Cmd {
	switch m := msg.(type) {
	case scrollLoadedMsg:
		c.body, c.loaded = m.body, true
		return nil
	case tea.KeyMsg:
		sc := keys.Cur()
		km := keys.NewMap(
			sc.Binding(keys.Up),
			sc.Binding(keys.Down),
			sc.Binding(keys.Cancel),
			sc.Binding(keys.PageUp),
			sc.Binding(keys.PageDown),
		)
		act, ok := km.Action(m.String())
		if !ok {
			return nil
		}
		switch act {
		case keys.Up:
			if c.offset > 0 {
				c.offset--
			}
		case keys.Down:
			c.offset++
		case keys.PageUp:
			c.offset = max(c.offset-10, 0)
		case keys.PageDown:
			c.offset += 10
		case keys.Cancel:
			return h.Pop()
		}
	}
	return nil
}

func (c *ScrollContent) Body(width, height int) string {
	if !c.loaded {
		return theme.Cur().Dim.Render("loading…")
	}
	lines := strings.Split(c.body, "\n")
	if c.offset >= len(lines) {
		c.offset = max(len(lines)-1, 0)
	}
	end := min(c.offset+max(height, 1), len(lines))
	return strings.Join(lines[c.offset:end], "\n")
}
