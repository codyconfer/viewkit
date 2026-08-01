package deck

import (
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/viewkit/forms"
	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/ui"
)

const testSave = keys.Action("test.save")

func testFormKeys() FormKeys {
	sc := keys.Default()
	bs := sc.EditorBindings(keys.Up, keys.Down, keys.Left, keys.Right, keys.Cancel, keys.Erase)
	bs = append(bs, keys.Binding{Keys: []string{"ctrl+s"}, Action: testSave, Glyph: "ctrl+s", Label: "save"})
	return FormKeys{Map: keys.NewMap(bs...), Save: testSave}
}

func nameFields() []forms.Field {
	return []forms.Field{{Key: "name", Label: "name", Kind: forms.FieldText}}
}

func rune1(r rune) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}} }

func hostWith(v View) *Model {
	h := New(stubView{title: "root"})
	h.Push(v)
	return h
}

func TestFormViewCancelPops(t *testing.T) {
	v := NewFormView(FormSpec{Title: "edit", Fields: nameFields(), Keys: testFormKeys()})
	h := hostWith(v)

	if cmd := v.Update(h, tea.KeyMsg{Type: tea.KeyEsc}); cmd == nil {
		t.Error("cancel should return the host's resize command")
	}
	if got := h.Top().Title(); got != "root" {
		t.Fatalf("top view after cancel = %q, want root", got)
	}
}

func TestFormViewSaveInvokesOnSubmitWithValues(t *testing.T) {
	var got map[string]any
	var gotHost *Model
	v := NewFormView(FormSpec{
		Title:  "edit",
		Fields: nameFields(),
		Keys:   testFormKeys(),
		OnSubmit: func(a *Model, vals map[string]any) tea.Cmd {
			gotHost, got = a, vals
			return a.Pop()
		},
	})
	h := hostWith(v)

	v.Update(h, rune1('s'))
	v.Update(h, rune1('r'))
	v.Update(h, rune1('e'))
	if cmd := v.Update(h, tea.KeyMsg{Type: tea.KeyCtrlS}); cmd == nil {
		t.Error("save should return the OnSubmit command")
	}

	if gotHost != h {
		t.Error("OnSubmit should receive the host model")
	}
	if got["name"] != "sre" {
		t.Fatalf("submitted values = %v, want name=sre", got)
	}
	if top := h.Top().Title(); top != "root" {
		t.Errorf("OnSubmit's Pop did not run: top = %q", top)
	}
}

func TestFormViewNilOnSubmitDoesNotPanic(t *testing.T) {
	v := NewFormView(FormSpec{Title: "edit", Fields: nameFields(), Keys: testFormKeys()})
	h := hostWith(v)
	if cmd := v.Update(h, tea.KeyMsg{Type: tea.KeyCtrlS}); cmd != nil {
		t.Errorf("save without OnSubmit should be a no-op, got %v", cmd)
	}
}

func TestFormViewTypingReachesTheForm(t *testing.T) {
	v := NewFormView(FormSpec{Title: "edit", Fields: nameFields(), Keys: testFormKeys()})
	h := hostWith(v)

	for _, r := range "hi" {
		v.Update(h, rune1(r))
	}
	v.Update(h, tea.KeyMsg{Type: tea.KeySpace})
	for _, r := range "there" {
		v.Update(h, rune1(r))
	}

	if got := v.Values()["name"]; got != "hi there" {
		t.Fatalf("typed value = %q, want %q", got, "hi there")
	}
}

func TestFormViewSpaceIsNotSwallowedByConfirm(t *testing.T) {
	v := NewFormView(FormSpec{Title: "edit", Fields: nameFields()})
	h := hostWith(v)

	v.Update(h, rune1('a'))
	v.Update(h, tea.KeyMsg{Type: tea.KeySpace})
	v.Update(h, rune1('b'))
	if got := v.Values()["name"]; got != "a b" {
		t.Fatalf("typed value = %q, want %q", got, "a b")
	}
}

func TestFormViewNavigationActionsDriveTheForm(t *testing.T) {
	v := NewFormView(FormSpec{
		Title: "edit",
		Fields: []forms.Field{
			{Key: "first", Label: "first", Kind: forms.FieldText},
			{Key: "second", Label: "second", Kind: forms.FieldText},
		},
		Keys: testFormKeys(),
	})
	h := hostWith(v)

	v.Update(h, rune1('a'))
	v.Update(h, tea.KeyMsg{Type: tea.KeyDown})
	v.Update(h, rune1('b'))

	vals := v.Values()
	if vals["first"] != "a" || vals["second"] != "b" {
		t.Fatalf("values = %v, want first=a second=b", vals)
	}
}

func TestFormViewUnknownKeysGoToOnKeyThenAreIgnored(t *testing.T) {
	var seen []string
	v := NewFormView(FormSpec{
		Title:  "edit",
		Fields: nameFields(),
		Keys:   testFormKeys(),
		OnKey: func(a *Model, key tea.KeyMsg) (tea.Cmd, bool) {
			seen = append(seen, key.String())
			if key.Type == tea.KeyCtrlX {
				return a.Pop(), true
			}
			return nil, false
		},
	})
	h := hostWith(v)

	if cmd := v.Update(h, tea.KeyMsg{Type: tea.KeyCtrlX}); cmd == nil {
		t.Error("OnKey should be able to return a command")
	}
	if h.Top().Title() != "root" {
		t.Error("OnKey's Pop did not run")
	}
	h.Push(v)

	if cmd := v.Update(h, tea.KeyMsg{Type: tea.KeyF5}); cmd != nil {
		t.Errorf("unhandled key returned %v, want nil", cmd)
	}
	if got := v.Values()["name"]; got != "" {
		t.Errorf("unhandled key typed %q into the form", got)
	}
	if !reflect.DeepEqual(seen, []string{"ctrl+x", "f5"}) {
		t.Fatalf("OnKey saw %v, want ctrl+x then f5", seen)
	}
}

func TestFormViewOnKeyCanPreemptSave(t *testing.T) {
	submitted := false
	v := NewFormView(FormSpec{
		Title:    "edit",
		Fields:   nameFields(),
		Keys:     testFormKeys(),
		OnSubmit: func(*Model, map[string]any) tea.Cmd { submitted = true; return nil },
		OnKey: func(*Model, tea.KeyMsg) (tea.Cmd, bool) {
			return nil, true
		},
	})
	v.Update(hostWith(v), tea.KeyMsg{Type: tea.KeyCtrlS})
	if submitted {
		t.Error("OnKey returning handled=true must suppress the save")
	}
}

func TestFormViewOnMsgSeesNonKeyMessages(t *testing.T) {
	type savedMsg struct{ err string }
	v := NewFormView(FormSpec{Title: "edit", Fields: nameFields(), Keys: testFormKeys()})
	v.spec.OnMsg = func(a *Model, msg tea.Msg) (tea.Cmd, bool) {
		m, ok := msg.(savedMsg)
		if !ok {
			return nil, false
		}
		v.Status(m.err)
		return nil, true
	}
	h := hostWith(v)

	v.Update(h, savedMsg{err: "title required"})
	if !strings.Contains(v.Body(layout.Frame{Width: 60, Height: 20}), "title required") {
		t.Fatalf("status line missing from body:\n%s", v.Body(layout.Frame{Width: 60, Height: 20}))
	}
	if cmd := v.Update(h, struct{}{}); cmd != nil {
		t.Errorf("unclaimed message returned %v, want nil", cmd)
	}
}

func TestFormViewBodyRendersFormAndPanelTitle(t *testing.T) {
	v := NewFormView(FormSpec{
		Title:      "status bar",
		PanelTitle: "status bar (show = visible chip)",
		Fields:     []forms.Field{{Key: "github", Label: "github", Kind: forms.FieldToggle, On: true}},
		Keys:       testFormKeys(),
	})
	body := v.Body(layout.Frame{Width: 70, Height: 20})
	if !strings.Contains(body, "github") {
		t.Errorf("body is missing the field:\n%s", body)
	}
	if !strings.Contains(body, "visible chip") {
		t.Errorf("body is missing the panel title:\n%s", body)
	}
	if strings.Contains(body, "\n\n\n") {
		t.Errorf("unexpected blank status line:\n%s", body)
	}
}

func TestFormViewBodyTitleDefaultsToViewTitle(t *testing.T) {
	v := NewFormView(FormSpec{Title: "appearance", Fields: nameFields(), Keys: testFormKeys()})
	if !strings.Contains(v.Body(layout.Frame{Width: 60, Height: 20}), "appearance") {
		t.Errorf("panel title should default to the view title:\n%s", v.Body(layout.Frame{Width: 60, Height: 20}))
	}
}

func TestFormViewTitleAndContextFuncsWinAndReEvaluate(t *testing.T) {
	name := ""
	v := NewFormView(FormSpec{
		Title:       "ignored",
		Context:     []keys.Hint{{Key: "role", Label: "ignored"}},
		TitleFunc:   func() string { return "edit " + name },
		ContextFunc: func() []keys.Hint { return []keys.Hint{{Key: "name", Label: name}} },
		Fields:      nameFields(),
		Keys:        testFormKeys(),
	})

	if v.Title() != "edit " {
		t.Errorf("title = %q", v.Title())
	}
	name = "sre"
	if v.Title() != "edit sre" {
		t.Errorf("title = %q, want edit sre", v.Title())
	}
	if got := v.Context(ui.Default()); !reflect.DeepEqual(got, []keys.Hint{{Key: "name", Label: "sre"}}) {
		t.Errorf("context = %v", got)
	}
}

func TestFormViewStaticTitleAndContextPlumbThrough(t *testing.T) {
	ctx := []keys.Hint{{Key: "role", Label: "sre"}}
	v := NewFormView(FormSpec{Title: "edit config", Context: ctx, Fields: nameFields(), Keys: testFormKeys()})
	if v.Title() != "edit config" {
		t.Errorf("title = %q", v.Title())
	}
	if !reflect.DeepEqual(v.Context(ui.Default()), ctx) {
		t.Errorf("context = %v, want %v", v.Context(ui.Default()), ctx)
	}
	if v.Init() != nil {
		t.Error("Init should be nil")
	}
}

func TestFormViewHintsOverrideWins(t *testing.T) {
	want := []keys.Hint{{Key: "↑/↓", Label: "field"}, {Key: "ctrl+s", Label: "save"}}
	v := NewFormView(FormSpec{Title: "edit", Fields: nameFields(), Keys: testFormKeys(), Hints: want})
	if got := v.Hints(ui.Default()); !reflect.DeepEqual(got, want) {
		t.Fatalf("hints = %v, want %v", got, want)
	}
}

func TestFormViewDefaultHintsAdvertiseTheSaveBinding(t *testing.T) {
	v := NewFormView(FormSpec{Title: "edit", Fields: nameFields(), Keys: testFormKeys()})
	hints := v.Hints(ui.Default())
	if len(hints) != 3 {
		t.Fatalf("hints = %v, want field/change/save", hints)
	}
	if hints[0].Label != "field" || hints[1].Label != "change" {
		t.Errorf("hints = %v, want field then change", hints)
	}
	if hints[2] != (keys.Hint{Key: "ctrl+s", Label: "save"}) {
		t.Errorf("save hint = %v, want the bound glyph and label", hints[2])
	}
}

func TestFormViewDefaultHintsOmitAnUnboundSave(t *testing.T) {
	v := NewFormView(FormSpec{Title: "edit", Fields: nameFields()})
	for _, h := range v.Hints(ui.Default()) {
		if h.Label == "save" {
			t.Fatalf("hints advertise a save the view cannot honour: %v", v.Hints(ui.Default()))
		}
	}
}

func TestFormViewWithoutAKeyMapStillNavigatesAndCancels(t *testing.T) {
	v := NewFormView(FormSpec{
		Title: "edit",
		Fields: []forms.Field{
			{Key: "first", Label: "first", Kind: forms.FieldText},
			{Key: "second", Label: "second", Kind: forms.FieldText},
		},
	})
	h := hostWith(v)

	v.Update(h, rune1('a'))
	v.Update(h, tea.KeyMsg{Type: tea.KeyDown})
	v.Update(h, rune1('b'))
	if vals := v.Values(); vals["first"] != "a" || vals["second"] != "b" {
		t.Fatalf("values = %v, want first=a second=b", vals)
	}
	v.Update(h, tea.KeyMsg{Type: tea.KeyEsc})
	if h.Top().Title() != "root" {
		t.Error("the default map should still bind cancel")
	}
}
