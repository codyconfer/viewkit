package deck

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/viewkit/forms"
	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/theme"
)

// Confirm runs a yes/no tea prompt and returns whether the user confirmed.
func Confirm(title, message, yesLabel, noLabel string) (bool, error) {
	m := &confirmModel{c: forms.Confirm{Title: title, Message: message, YesLabel: yesLabel, NoLabel: noLabel}}
	out, err := tea.NewProgram(m).Run()
	if err != nil {
		return false, err
	}
	fm := out.(*confirmModel)
	return fm.confirmed && fm.c.Yes, nil
}

type confirmModel struct {
	c         forms.Confirm
	confirmed bool
}

func (m *confirmModel) Init() tea.Cmd { return nil }

func (m *confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	sc := keys.Cur()
	km := keys.NewMap(
		sc.Binding(keys.Left),
		sc.Binding(keys.Right),
		sc.Binding(keys.Confirm),
		sc.Binding(keys.Cancel),
		sc.Binding(keys.Quit),
		keys.Binding{Keys: []string{"y", "Y"}, Action: "confirm.yes", Glyph: "y", Label: "yes"},
		keys.Binding{Keys: []string{"n", "N"}, Action: "confirm.no", Glyph: "n", Label: "no"},
	)
	act, ok := km.Action(key.String())
	if !ok {
		return m, nil
	}
	switch act {
	case keys.Left:
		m.c.Handle(keys.Left)
	case keys.Right:
		m.c.Handle(keys.Right)
	case "confirm.yes":
		m.c.Yes, m.confirmed = true, true
		return m, tea.Quit
	case "confirm.no", keys.Cancel, keys.Quit:
		m.confirmed = false
		return m, tea.Quit
	case keys.Confirm:
		m.confirmed = m.c.Handle(keys.Confirm) == forms.Submitted
		return m, tea.Quit
	}
	return m, nil
}

func (m *confirmModel) View() string {
	return m.c.Render(layout.NewFrame(theme.BodyWidth)) + "\n"
}
