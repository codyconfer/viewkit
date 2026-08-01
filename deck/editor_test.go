package deck

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/codyconfer/viewkit/forms"
	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/list"
)

const (
	testCopy  keys.Action = "test.copy"
	testWrite keys.Action = "test.write"
)

type stubResults struct{ n int }

func (r stubResults) Items(layout.Frame) []list.Item {
	out := make([]list.Item, 0, r.n)
	for i := 0; i < r.n; i++ {
		out = append(out, list.Item{Block: "row", Selectable: true})
	}
	return out
}

func (r stubResults) Count() int { return r.n }

type erroredResults struct {
	stubResults
	errored int
}

func (r erroredResults) Errored() int { return r.errored }

type stubDoc struct {
	saved    string
	copies   int
	writes   int
	copyErr  error
	writeErr error
}

func (d *stubDoc) Kind() string           { return "report" }
func (d *stubDoc) Title() string          { return "build report" }
func (d *stubDoc) Context() [][2]string   { return nil }
func (d *stubDoc) SavedName() string      { return d.saved }
func (d *stubDoc) Sync() bool             { return false }
func (d *stubDoc) Summary() string        { return "draft" }
func (d *stubDoc) PreviewLines() []string { return []string{"name: draft"} }
func (d *stubDoc) Remove() string         { return "removed" }

func (d *stubDoc) Fields(prev map[string]any) []forms.Field {
	return []forms.Field{{Key: "name", Label: "name", Kind: forms.FieldText, Text: forms.Raw(prev, "name")}}
}

func (d *stubDoc) ValidateLines() ([]string, error) { return []string{"ok"}, nil }

func (d *stubDoc) Run() (string, func() Results, error) {
	return "draft", func() Results { return stubResults{n: 1} }, nil
}

func (d *stubDoc) Persist() (string, error) { return "saved", nil }

type outputDoc struct{ *stubDoc }

func (d outputDoc) CopyOutput() (string, error) {
	if d.copyErr != nil {
		return "", d.copyErr
	}
	d.copies++
	return "copied 12 bytes", nil
}

func (d outputDoc) WriteOutput() (string, error) {
	if d.writeErr != nil {
		return "", d.writeErr
	}
	d.writes++
	return "wrote /tmp/report.md", nil
}

func editorTestKeys() EditorKeys {
	sc := keys.Cur()
	return EditorKeys{
		Map: keys.NewMap(
			sc.Binding(keys.Cancel),
			keys.Binding{Keys: []string{"ctrl+r"}, Action: "test.run", Glyph: "ctrl+r", Label: "run"},
			keys.Binding{Keys: []string{"ctrl+g"}, Action: testCopy, Glyph: "ctrl+g", Label: "copy"},
			keys.Binding{Keys: []string{"ctrl+w"}, Action: testWrite, Glyph: "ctrl+w", Label: "write"},
		),
		Confirm: keys.NewMap(sc.Binding(keys.Confirm), sc.Binding(keys.Cancel)),
		Run:     "test.run",
		Copy:    testCopy,
		Write:   testWrite,
	}
}

func hasHint(hints [][2]string, glyph string) bool {
	for _, h := range hints {
		if h[0] == glyph {
			return true
		}
	}
	return false
}

func TestEditorOutputKeysPushMessages(t *testing.T) {
	doc := outputDoc{stubDoc: &stubDoc{}}
	e := NewEditor(doc, editorTestKeys(), nil)
	h := New(e)
	h = driveHost(h, tea.WindowSizeMsg{Width: 80, Height: 24})

	h = driveHost(h, tea.KeyMsg{Type: tea.KeyCtrlG})
	if doc.copies != 1 {
		t.Fatalf("CopyOutput calls = %d, want 1", doc.copies)
	}
	if view := h.View(); !strings.Contains(view, "copied 12 bytes") {
		t.Errorf("copy summary not shown:\n%s", view)
	}

	h = driveHost(h, tea.KeyMsg{Type: tea.KeyEsc})
	h = driveHost(h, tea.KeyMsg{Type: tea.KeyCtrlW})
	if doc.writes != 1 {
		t.Fatalf("WriteOutput calls = %d, want 1", doc.writes)
	}
	if view := h.View(); !strings.Contains(view, "wrote /tmp/report.md") {
		t.Errorf("write summary not shown:\n%s", view)
	}
}

func TestEditorOutputErrorBecomesStatus(t *testing.T) {
	doc := outputDoc{stubDoc: &stubDoc{copyErr: errors.New("run ctrl+r first")}}
	e := NewEditor(doc, editorTestKeys(), nil)
	h := New(e)
	h = driveHost(h, tea.WindowSizeMsg{Width: 80, Height: 24})
	driveHost(h, tea.KeyMsg{Type: tea.KeyCtrlG})

	if e.Status() != "run ctrl+r first" {
		t.Fatalf("status = %q, want the copy error", e.Status())
	}
	if !hasHint(e.Hints(), "ctrl+g") {
		t.Errorf("hints should offer copy: %v", e.Hints())
	}
}

func TestEditorCollapsedResultsSummary(t *testing.T) {
	cases := []struct {
		name    string
		set     Results
		want    string
		wantNot string
	}{
		{"empty success", stubResults{n: 0}, "no items", "item(s)"},
		{"plain items", stubResults{n: 3}, "3 item(s)  ·  tab to view", ""},
		{"errors only", erroredResults{stubResults{n: 2}, 2}, "errors  ·  tab to view", "item(s)"},
		{"mixed", erroredResults{stubResults{n: 5}, 2}, "3 item(s)  ·  tab to view", "5 item(s)"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := NewEditor(&stubDoc{}, editorTestKeys(), nil)
			e.hasResults = true
			e.set = c.set
			got := ansi.Strip(e.collapsedResults(layout.NewFrame(60)))
			if !strings.Contains(got, c.want) {
				t.Errorf("collapsed summary missing %q:\n%s", c.want, got)
			}
			if c.wantNot != "" && strings.Contains(got, c.wantNot) {
				t.Errorf("collapsed summary should not contain %q:\n%s", c.wantNot, got)
			}
		})
	}
}

func TestEditorWithoutOutputIgnoresCopyAndWrite(t *testing.T) {
	e := NewEditor(&stubDoc{}, editorTestKeys(), nil)
	h := New(e)
	h = driveHost(h, tea.WindowSizeMsg{Width: 80, Height: 24})
	driveHost(h, tea.KeyMsg{Type: tea.KeyCtrlG})

	if e.Status() != "" {
		t.Errorf("plain doc should ignore copy, got status %q", e.Status())
	}
	for _, glyph := range []string{"ctrl+g", "ctrl+w"} {
		if hasHint(e.Hints(), glyph) {
			t.Errorf("plain doc should not offer %s: %v", glyph, e.Hints())
		}
	}
}
