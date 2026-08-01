package deck

import (
	"errors"
	"maps"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/viewkit/forms"
	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/ui"
)

// PromptSpec configures a one-shot form prompt. It is the Editor's shape
// minus everything a command-line caller has no use for — no results pane,
// no preview, no persistence — so a CLI command can collect the same fields,
// with the same suggestions, without entering the deck.
type PromptSpec struct {
	// Title heads the rendered panel.
	Title string
	// Fields builds the form, seeded from previously entered values. It is
	// called again after every Sync, exactly like EditorDoc.Fields.
	Fields func(prev map[string]any) []forms.Field
	// Sync reports whether the last edit changed a value the field set
	// depends on, and so whether to rebuild. A nil Sync never rebuilds.
	Sync func(vals map[string]any) bool
	// Seed pre-fills the form.
	Seed map[string]any
	// Save is the action that submits. It defaults to keys.Confirm when left
	// empty; hosts with a save binding of their own pass it here.
	Save keys.Action
	// Keys resolves keys. A nil Map falls back to the scope's editor
	// bindings, like FormView.
	Keys *keys.Map
	// UI is the rendering scope; nil snapshots the process defaults.
	UI *ui.Scope
}

// ErrCancelled reports that the user dismissed a prompt without submitting.
var ErrCancelled = errors.New("prompt cancelled")

// Prompt runs spec as a standalone tea program and returns the entered
// values. A user cancel returns ErrCancelled; a nil Fields is an error.
func Prompt(spec PromptSpec) (map[string]any, error) {
	if spec.Fields == nil {
		return nil, errors.New("deck: PromptSpec.Fields is nil")
	}
	seed := spec.Seed
	if seed == nil {
		seed = map[string]any{}
	}
	scope := spec.UI
	if scope == nil {
		scope = ui.Default()
	}
	m := &promptModel{spec: spec, sticky: seed, scope: scope}
	m.form = forms.NewForm(spec.Fields(seed)...)

	out, runErr := tea.NewProgram(m).Run()
	if runErr != nil {
		return nil, runErr
	}
	pm := out.(*promptModel)
	if !pm.submitted {
		return nil, ErrCancelled
	}
	return pm.form.Values(), nil
}

type promptModel struct {
	spec      PromptSpec
	form      *forms.Form
	sticky    map[string]any
	scope     *ui.Scope
	submitted bool
}

func (m *promptModel) Init() tea.Cmd { return nil }

func (m *promptModel) keyMap() *keys.Map {
	if m.spec.Keys != nil {
		return m.spec.Keys
	}
	return formMapFor(m.scope.Keys)
}

func (m *promptModel) save() keys.Action {
	if m.spec.Save != "" {
		return m.spec.Save
	}
	return keys.Confirm
}

func (m *promptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	act, ok := m.keyMap().Action(key.String())
	if !ok {
		switch {
		case key.String() == " ":
			m.form.Insert(" ")
		case key.Type == tea.KeyRunes:
			m.form.Insert(string(key.Runes))
		}
		return m, nil
	}
	switch act {
	case keys.Cancel, keys.Quit:
		return m, tea.Quit
	case keys.Complete:
		m.form.AcceptSuggestion()
	case m.save():
		m.submitted = true
		return m, tea.Quit
	default:
		m.form.Handle(act)
		m.resync()
	}
	return m, nil
}

func (m *promptModel) resync() {
	if m.spec.Sync == nil {
		return
	}
	vals := m.form.Values()
	if !m.spec.Sync(vals) {
		return
	}
	maps.Copy(m.sticky, vals)
	m.form = forms.NewForm(m.spec.Fields(m.sticky)...)
}

func (m *promptModel) View() string {
	f := layout.DocumentFrame().WithUI(m.scope)
	hint := m.scope.Theme.Dim.Render("↑/↓ field · tab accept · ctrl+n/ctrl+p suggestion · esc cancel")
	return m.form.Render(f, m.spec.Title) + "\n" + hint + "\n"
}
