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
	// OnSelect runs when the row is confirmed.
	OnSelect func(m *Model) tea.Cmd
}

// Menu is a simple navigable list View.
type Menu struct {
	title  string
	items  []MenuItem
	cursor int
	ctx    []keys.Hint
}

// NewMenu builds a Menu view.
func NewMenu(title string, ctx []keys.Hint, items ...MenuItem) *Menu {
	return &Menu{title: title, items: items, ctx: ctx}
}

// Title implements View.
func (m *Menu) Title() string { return m.title }

// Init implements View; a Menu needs no startup command.
func (m *Menu) Init() tea.Cmd { return nil }

// Context implements View.
func (m *Menu) Context() []keys.Hint { return m.ctx }

// Hints implements View.
func (m *Menu) Hints() []keys.Hint {
	km := navMap()
	return []keys.Hint{
		km.HintLabeled(keys.Up, "move"),
		km.HintLabeled(keys.Confirm, "open"),
	}
}

// Update implements View: Up/Down move the cursor, Confirm runs the row's
// OnSelect, Cancel pops.
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
		if len(m.items) > 0 && m.items[m.cursor].OnSelect != nil {
			return m.items[m.cursor].OnSelect(h)
		}
	case keys.Cancel:
		return h.Pop()
	}
	return nil
}

const menuBoxChrome = 2

// Body implements View, rendering the rows in a titled box.
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
	ctx   []keys.Hint
}

// NewError builds a Message that renders err in the theme's failure style —
// the standard error screen, so hosts stop hand-rolling Cant.Render(err).
func NewError(title string, err error, ctx []keys.Hint) *Message {
	return NewMessage(title, theme.Cur().Cant.Render(err.Error()), ctx)
}

// NewMessage builds a Message view.
func NewMessage(title, body string, ctx []keys.Hint) *Message {
	return &Message{title: title, body: body, ctx: ctx}
}

// Title implements View.
func (m *Message) Title() string { return m.title }

// Init implements View; a Message needs no startup command.
func (m *Message) Init() tea.Cmd { return nil }

// Context implements View.
func (m *Message) Context() []keys.Hint { return m.ctx }

// Hints implements View; a Message adds no legend entries of its own.
func (m *Message) Hints() []keys.Hint { return nil }

// Update implements View: Cancel or Confirm dismisses the message.
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

// Body implements View, rendering the text in a titled box.
func (m *Message) Body(width, _ int) string {
	f := layout.ScreenFrame(width)
	return f.TitledBox(strings.ToUpper(m.title), strings.Split(m.body, "\n")...)
}
