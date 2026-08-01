package deck

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/viewkit/forms"
	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/ui"
)

// ConfirmSpec configures a one-shot yes/no prompt.
type ConfirmSpec struct {
	Title    string
	Message  string
	YesLabel string
	NoLabel  string
	// UI is the rendering scope; nil snapshots the process defaults.
	UI *ui.Scope
}

// Confirm runs a yes/no tea prompt and returns whether the user confirmed.
// Besides the scope's scheme, y/Y and n/N always answer directly.
func Confirm(spec ConfirmSpec) (bool, error) {
	scope := spec.UI
	if scope == nil {
		scope = ui.Default()
	}
	m := &confirmModel{
		c:     forms.Confirm{Title: spec.Title, Message: spec.Message, YesLabel: spec.YesLabel, NoLabel: spec.NoLabel},
		km:    confirmMapFor(scope.Keys),
		scope: scope,
	}
	out, err := tea.NewProgram(m).Run()
	if err != nil {
		return false, err
	}
	fm := out.(*confirmModel)
	return fm.confirmed && fm.c.Yes, nil
}

func confirmMapFor(sc keys.Scheme) *keys.Map {
	return keys.NewMap(
		sc.Binding(keys.Left),
		sc.Binding(keys.Right),
		sc.Binding(keys.Confirm),
		sc.Binding(keys.Cancel),
		sc.Binding(keys.Quit),
		keys.Binding{Keys: []string{"y", "Y"}, Action: "confirm.yes", Glyph: "y", Label: "yes"},
		keys.Binding{Keys: []string{"n", "N"}, Action: "confirm.no", Glyph: "n", Label: "no"},
	)
}

type confirmModel struct {
	c         forms.Confirm
	km        *keys.Map
	scope     *ui.Scope
	confirmed bool
}

func (m *confirmModel) Init() tea.Cmd { return nil }

func (m *confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	act, ok := m.km.Action(key.String())
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
	return m.c.Render(layout.DocumentFrame().WithUI(m.scope)) + "\n"
}
