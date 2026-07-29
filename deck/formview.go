package deck

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/viewkit/forms"
	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/theme"
)

// FormKeys binds a FormView's commands to a host application's key scheme.
// Only cancel comes from viewkit's vocabulary (keys.Cancel, resolved through
// Map); the save action is supplied by the caller because command vocabularies
// are app-specific — an app may name it "app.save" and bind it to ctrl+s, or
// name it something else entirely. This mirrors EditorKeys.
type FormKeys struct {
	// Map resolves keys while the form has focus. A nil Map falls back to the
	// active scheme's editor bindings (navigation, confirm, cancel, erase,
	// paging), resolved fresh on every key so a later keys.Use still applies.
	Map *keys.Map
	// Save is the action that triggers OnSubmit. Leave it empty for a form
	// with no save command (submit then only happens through OnKey).
	Save keys.Action
}

// FormSpec configures a FormView. Every field beyond Fields is optional, and
// most views set only two or three of them; a struct keeps those call sites
// readable and lets the type grow a knob without breaking anyone, which a
// ten-argument constructor could not.
type FormSpec struct {
	// Fields build the form, in display order.
	Fields []forms.Field
	// Keys binds cancel and save.
	Keys FormKeys

	// Title heads the breadcrumb trail. TitleFunc wins when both are set,
	// for views whose title depends on state ("new role" vs "edit role").
	Title string
	// TitleFunc computes the title on every read.
	TitleFunc func() string
	// PanelTitle heads the rendered form panel. Defaults to the view title,
	// which is what most forms want; set it when the panel wants a longer
	// caption than the breadcrumb.
	PanelTitle string

	// Context supplies chrome context cues. ContextFunc wins when both are
	// set; use it when the cues can change while the form is open.
	Context [][2]string
	// ContextFunc computes context cues on every read.
	ContextFunc func() [][2]string
	// Hints overrides the footer legend. When nil, the legend is derived from
	// Keys: field movement, value change, and the save binding.
	Hints [][2]string

	// OnSubmit runs on the save action with the form's current values. It
	// receives the Model so it can Pop back and Push a result view. A nil
	// OnSubmit makes save a no-op.
	OnSubmit func(a *Model, vals map[string]any) tea.Cmd
	// OnKey sees every key before the form does. Return handled=true to
	// suppress all default handling — including cancel, save and text entry —
	// which is how a view adds its own commands (delete, preview) on top of
	// the shared form behaviour.
	OnKey func(a *Model, key tea.KeyMsg) (cmd tea.Cmd, handled bool)
	// OnMsg sees every message before anything else, keys included. Views
	// whose save runs as a tea.Cmd use it to catch the completion message and
	// either Pop or Status the error.
	OnMsg func(a *Model, msg tea.Msg) (cmd tea.Cmd, handled bool)
}

// FormView is a single-panel form View: the boilerplate sibling of Menu,
// Message and Scroll. It owns the form, key routing (cancel pops, save calls
// back, everything else is navigation or text entry) and an error line under
// the panel. Domain concerns stay in the spec's callbacks.
type FormView struct {
	spec   FormSpec
	form   *forms.Form
	status string
}

// NewFormView builds a form View from spec.
func NewFormView(spec FormSpec) *FormView {
	return &FormView{spec: spec, form: forms.NewForm(spec.Fields...)}
}

// Form exposes the live form so hosts can read or seed entered values.
func (v *FormView) Form() *forms.Form { return v.form }

// Values returns the form's current values, the same map OnSubmit receives.
func (v *FormView) Values() map[string]any { return v.form.Values() }

// Status sets the error line shown beneath the form; "" clears it.
func (v *FormView) Status(msg string) { v.status = msg }

// Title reports the breadcrumb title.
func (v *FormView) Title() string {
	if v.spec.TitleFunc != nil {
		return v.spec.TitleFunc()
	}
	return v.spec.Title
}

// Init satisfies View; a form has nothing to load.
func (v *FormView) Init() tea.Cmd { return nil }

// Context reports the chrome context cues.
func (v *FormView) Context() [][2]string {
	if v.spec.ContextFunc != nil {
		return v.spec.ContextFunc()
	}
	return v.spec.Context
}

// Hints reports the footer legend: the spec's override, else one derived from
// the bound keys so the legend cannot advertise a binding the view ignores.
func (v *FormView) Hints() [][2]string {
	if v.spec.Hints != nil {
		return v.spec.Hints
	}
	km := v.keyMap()
	hints := [][2]string{
		km.HintLabeled(keys.Up, "field"),
		km.HintLabeled(keys.Left, "change"),
	}
	if save := v.spec.Keys.Save; save != "" && km.Has(save) {
		hints = append(hints, km.Hint(save))
	}
	return hints
}

// Body renders the form panel, with the status line appended when set.
func (v *FormView) Body(width, _ int) string {
	body := v.form.Render(layout.NewFrame(width), v.panelTitle())
	if v.status != "" {
		return body + "\n" + theme.Cur().Cant.Render(v.status)
	}
	return body
}

func (v *FormView) panelTitle() string {
	if v.spec.PanelTitle != "" {
		return v.spec.PanelTitle
	}
	return v.Title()
}

// keyMap resolves the map to route keys through. It is rebuilt per call when
// the spec left Map nil so a scheme swap mid-session takes effect.
func (v *FormView) keyMap() *keys.Map {
	if v.spec.Keys.Map != nil {
		return v.spec.Keys.Map
	}
	return formMap()
}

// formMap is the default form action map: the active scheme's editor bindings,
// which drop single-character keys so ordinary typing is not swallowed.
func formMap() *keys.Map {
	sc := keys.Cur()
	return keys.NewMap(sc.EditorBindings(
		keys.Up, keys.Down, keys.Left, keys.Right,
		keys.Confirm, keys.Cancel, keys.Erase, keys.PageUp, keys.PageDown,
	)...)
}

// Update routes a message: OnMsg first, then OnKey, then cancel / save /
// form navigation, and finally text entry for unbound runes.
func (v *FormView) Update(a *Model, msg tea.Msg) tea.Cmd {
	if v.spec.OnMsg != nil {
		if cmd, handled := v.spec.OnMsg(a, msg); handled {
			return cmd
		}
	}
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}
	if v.spec.OnKey != nil {
		if cmd, handled := v.spec.OnKey(a, key); handled {
			return cmd
		}
	}
	act, ok := v.keyMap().Action(key.String())
	if !ok {
		v.insert(key)
		return nil
	}
	switch act {
	case keys.Cancel:
		return a.Pop()
	case v.spec.Keys.Save:
		return v.submit(a)
	default:
		v.form.Handle(act)
	}
	return nil
}

// insert types an unbound key into the focused field. Space arrives as a key
// with no runes on some terminals, so it is matched by name.
func (v *FormView) insert(key tea.KeyMsg) {
	switch {
	case key.String() == " ":
		v.form.Insert(" ")
	case key.Type == tea.KeyRunes:
		v.form.Insert(string(key.Runes))
	}
}

func (v *FormView) submit(a *Model) tea.Cmd {
	if v.spec.OnSubmit == nil {
		return nil
	}
	return v.spec.OnSubmit(a, v.form.Values())
}
