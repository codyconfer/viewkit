package deck

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/theme"
	"github.com/codyconfer/viewkit/ui"
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
func (m *Menu) Context(scope *ui.Scope) []keys.Hint { return m.ctx }

// Hints implements View.
func (m *Menu) Hints(scope *ui.Scope) []keys.Hint {
	km := navMapFor(schemeOf(scope))
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
	act, ok := navMapFor(schemeOf(h.UI())).Action(key.String())
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
func (m *Menu) Body(f layout.Frame) string {
	th := f.Theme()
	f = f.Screen()
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
			row += th.Icon(it.Icon, it.Hue)
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
	title   string
	body    string
	errText string
	ctx     []keys.Hint
}

// NewError builds a Message that renders err in the theme's failure style —
// the standard error screen, so hosts stop hand-rolling Cant.Render(err).
// Styling is deferred to Body so the render-time scope applies.
func NewError(title string, err error, ctx []keys.Hint) *Message {
	return &Message{title: title, errText: err.Error(), ctx: ctx}
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
func (m *Message) Context(scope *ui.Scope) []keys.Hint { return m.ctx }

// Hints implements View; a Message adds no legend entries of its own.
func (m *Message) Hints(scope *ui.Scope) []keys.Hint { return nil }

// Update implements View: Cancel or Confirm dismisses the message.
func (m *Message) Update(h *Model, msg tea.Msg) tea.Cmd {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}
	sc := schemeOf(h.UI())
	km := keys.NewMap(sc.Binding(keys.Cancel), sc.Binding(keys.Confirm), sc.Binding(keys.Quit))
	if act, ok := km.Action(key.String()); ok && (act == keys.Cancel || act == keys.Confirm) {
		return h.Pop()
	}
	return nil
}

// Body implements View, rendering the text in a titled box.
func (m *Message) Body(f layout.Frame) string {
	f = f.Screen()
	body := m.body
	if m.errText != "" {
		body = f.Theme().Cant.Render(m.errText)
	}
	return f.TitledBox(strings.ToUpper(m.title), strings.Split(body, "\n")...)
}
