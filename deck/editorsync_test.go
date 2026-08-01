package deck

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/viewkit/forms"
	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/list"
)

type syncDoc struct{}

func (d *syncDoc) Kind() string           { return "query" }
func (d *syncDoc) Title() string          { return "new query" }
func (d *syncDoc) Context() []keys.Hint   { return nil }
func (d *syncDoc) SavedName() string      { return "" }
func (d *syncDoc) Sync() bool             { return true }
func (d *syncDoc) Summary() string        { return "draft" }
func (d *syncDoc) PreviewLines() []string { return []string{"kind: query"} }
func (d *syncDoc) Remove() string         { return "removed" }
func (d *syncDoc) ValidateLines() ([]string, error) {
	return []string{"ok"}, nil
}
func (d *syncDoc) Persist() (string, error) { return "saved", nil }

func (d *syncDoc) Run() (string, func() Results, error) {
	return "draft", func() Results { return syncResults{} }, nil
}

func (d *syncDoc) Fields(prev map[string]any) []forms.Field {
	fields := []forms.Field{
		{Key: "type", Label: "type", Kind: forms.FieldSelect, Options: []string{"(infer)", "query", "filter"}},
		{Key: "signal", Label: "signal", Kind: forms.FieldSelect, Options: []string{"none", "prs", "issues"}},
	}
	fields[0].Selected = forms.SelectIndex(fields[0].Options, forms.Str(prev, "type"))
	fields[1].Selected = forms.SelectIndex(fields[1].Options, forms.Str(prev, "signal"))
	return fields
}

type syncResults struct{}

func (syncResults) Items(layout.Frame) []list.Item { return nil }
func (syncResults) Count() int                     { return 0 }

func syncTestKeys() EditorKeys {
	sc := keys.Cur()
	return EditorKeys{
		Map: keys.NewMap(
			sc.Binding(keys.Cancel),
			keys.Binding{Keys: []string{"up"}, Action: keys.Up, Glyph: "↑", Label: "up"},
			keys.Binding{Keys: []string{"down"}, Action: keys.Down, Glyph: "↓", Label: "down"},
			keys.Binding{Keys: []string{"left"}, Action: keys.Left, Glyph: "←", Label: "left"},
			keys.Binding{Keys: []string{"right"}, Action: keys.Right, Glyph: "→", Label: "right"},
		),
		Confirm: keys.NewMap(sc.Binding(keys.Confirm), sc.Binding(keys.Cancel)),
	}
}

func pressSpecial(e *Editor, m *Model, kt tea.KeyType) {
	e.Update(m, tea.KeyMsg{Type: kt})
}

func focusedKey(t *testing.T, e *Editor) string {
	t.Helper()
	fd := e.Form().Focused()
	if fd == nil {
		t.Fatal("form has no focused field")
	}
	return fd.Key
}

func TestSyncFieldsKeepsFocusOnSelectPath(t *testing.T) {
	doc := &syncDoc{}
	e := NewEditor(doc, syncTestKeys(), nil)
	m := New(e)

	pressSpecial(e, m, tea.KeyDown)
	if got := focusedKey(t, e); got != "signal" {
		t.Fatalf("after down, focus = %q, want signal", got)
	}

	pressSpecial(e, m, tea.KeyRight)
	if got := focusedKey(t, e); got != "signal" {
		t.Fatalf("adjusting signal moved focus to %q; a rebuild must not reset the caret", got)
	}
	if got := e.Value("signal"); got != "prs" {
		t.Fatalf("signal = %q, want prs", got)
	}
	if got := e.Value("type"); got != "(infer)" {
		t.Fatalf("type = %q, want it untouched", got)
	}

	pressSpecial(e, m, tea.KeyRight)
	if got := e.Value("signal"); got != "issues" {
		t.Fatalf("second right should keep adjusting signal, got %q", got)
	}
	if got := e.Value("type"); got != "(infer)" {
		t.Fatalf("type = %q, want it still untouched", got)
	}
}

func TestSyncFieldsKeepsFocusOnTypePath(t *testing.T) {
	doc := &syncDoc{}
	e := NewEditor(doc, syncTestKeys(), nil)
	m := New(e)

	pressSpecial(e, m, tea.KeyRight)
	if got := focusedKey(t, e); got != "type" {
		t.Fatalf("after adjusting type, focus = %q, want type", got)
	}
	if got := e.Value("type"); got != "query" {
		t.Fatalf("type = %q, want query", got)
	}

	pressSpecial(e, m, tea.KeyDown)
	if got := focusedKey(t, e); got != "signal" {
		t.Fatalf("down after a rebuild should reach signal, got %q", got)
	}
}

func TestSyncFieldsDropsFocusWhenFieldDisappears(t *testing.T) {
	doc := &shrinkDoc{}
	e := NewEditor(doc, syncTestKeys(), nil)
	m := New(e)

	pressSpecial(e, m, tea.KeyDown)
	if got := focusedKey(t, e); got != "extra" {
		t.Fatalf("focus = %q, want extra", got)
	}

	doc.hide = true
	e.SyncFields()
	if got := focusedKey(t, e); got != "type" {
		t.Fatalf("a vanished field should fall back to the first, got %q", got)
	}
}

type shrinkDoc struct {
	syncDoc
	hide bool
}

func (d *shrinkDoc) Fields(map[string]any) []forms.Field {
	out := []forms.Field{{Key: "type", Label: "type", Kind: forms.FieldText}}
	if !d.hide {
		out = append(out, forms.Field{Key: "extra", Label: "extra", Kind: forms.FieldText})
	}
	return out
}

func TestFormFocusKeyReportsMatch(t *testing.T) {
	fm := forms.NewForm(
		forms.Field{Key: "a", Kind: forms.FieldText},
		forms.Field{Key: "b", Kind: forms.FieldText},
	)
	if fm.FocusedKey() != "a" {
		t.Fatalf("a new form focuses the first field, got %q", fm.FocusedKey())
	}
	if !fm.FocusKey("b") {
		t.Fatal("FocusKey(b) should report a match")
	}
	if fm.FocusedKey() != "b" {
		t.Fatalf("FocusedKey = %q, want b", fm.FocusedKey())
	}
	if fm.FocusKey("nope") {
		t.Fatal("FocusKey on an absent key should report false")
	}
	if fm.FocusedKey() != "b" {
		t.Fatalf("a missed FocusKey must not move focus, got %q", fm.FocusedKey())
	}
	if fm.FocusKey("") {
		t.Fatal("FocusKey(\"\") should never match")
	}
	if forms.NewForm().FocusedKey() != "" {
		t.Fatal("an empty form has no focused key")
	}
}

func TestRememberedKeepsHiddenFieldValues(t *testing.T) {
	doc := &shrinkDoc{}
	e := NewEditor(doc, syncTestKeys(), nil)
	m := New(e)

	pressSpecial(e, m, tea.KeyDown)
	e.Form().Insert("kept")
	if got := e.Value("extra"); got != "kept" {
		t.Fatalf("extra = %q, want kept", got)
	}

	doc.hide = true
	e.SyncFields()

	if _, ok := e.Form().Values()["extra"]; ok {
		t.Fatal("the hidden field should be gone from the live form")
	}
	if got := e.Remembered()["extra"]; got != "kept" {
		t.Fatalf("Remembered()[extra] = %v, want kept", got)
	}
	if _, ok := e.Remembered()["type"]; !ok {
		t.Fatal("Remembered should also carry live fields")
	}
}
