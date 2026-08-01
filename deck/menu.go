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
	Do    func(m *Model) tea.Cmd
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
	km := navMap()
	return [][2]string{
		km.HintLabeled(keys.Up, "move"),
		km.HintLabeled(keys.Confirm, "open"),
	}
}

func (m *Menu) Update(h *Model, msg tea.Msg) tea.Cmd {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}
	act, ok := navMap().Action(key.String())
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

const menuBoxChrome = 2

func (m *Menu) Body(width, height int) string {
	th := theme.Cur()
	f := layout.ScreenFrame(width)
	anyIcon := false
	for _, it := range m.items {
		if it.Icon != "" {
			anyIcon = true
			break
		}
	}
	rows := make([]string, 0, len(m.items))
	for i, it := range m.items {
		cursor := "  "
		label := th.Val.Render(it.Label)
		if i == m.cursor {
			cursor = th.Key.Render("▸ ")
			label = th.Key.Render(it.Label)
		}
		row := cursor
		switch {
		case it.Icon != "":
			row += theme.Icon(it.Icon, it.Hue)
		case anyIcon:
			row += theme.IconBlank()
		}
		row += label
		if it.Desc != "" {
			row = f.Spread(row, th.Dim.Render(it.Desc))
		}
		rows = append(rows, row)
	}
	return f.TitledBox(strings.ToUpper(m.title), rows...)
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

func (m *Message) Update(h *Model, msg tea.Msg) tea.Cmd {
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

type scrollLoadedMsg struct {
	own  *ScrollContent
	body string
}

func (m scrollLoadedMsg) recipient() View { return m.own }

func (c *ScrollContent) Title() string        { return c.title }
func (c *ScrollContent) Context() [][2]string { return c.ctx }
func (c *ScrollContent) Hints() [][2]string {
	km := navMap()
	return append([][2]string{
		km.HintLabeled(keys.Up, "scroll"),
		km.HintLabeled(keys.PageUp, "page"),
	}, c.hints...)
}

func (c *ScrollContent) Init() tea.Cmd {
	return func() tea.Msg { return scrollLoadedMsg{own: c, body: c.load()} }
}

func (c *ScrollContent) Update(h *Model, msg tea.Msg) tea.Cmd {
	switch m := msg.(type) {
	case scrollLoadedMsg:
		c.body, c.loaded = m.body, true
		return nil
	case tea.KeyMsg:
		act, ok := navMap().Action(m.String())
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
