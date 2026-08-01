package deck

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/list"
	"github.com/codyconfer/viewkit/theme"
)

func relayoutEditorKeys() EditorKeys {
	sc := keys.Cur()
	return EditorKeys{
		Map: keys.NewMap(
			sc.Binding(keys.Cancel),
			sc.Binding(keys.Up),
			sc.Binding(keys.Down),
			sc.Binding(keys.PageUp),
			sc.Binding(keys.PageDown),
			keys.Binding{Keys: []string{"ctrl+r"}, Action: "test.run", Glyph: "ctrl+r", Label: "run"},
			keys.Binding{Keys: []string{"ctrl+s"}, Action: "test.save", Glyph: "ctrl+s", Label: "save"},
		),
		Confirm: keys.NewMap(sc.Binding(keys.Confirm), sc.Binding(keys.Cancel)),
		Run:     "test.run",
		Save:    "test.save",
	}
}

type keyedResults struct {
	keys  []string
	binds *int
}

func (r keyedResults) Items(layout.Frame) []list.Item {
	if r.binds != nil {
		*r.binds++
	}
	out := make([]list.Item, 0, len(r.keys))
	for _, k := range r.keys {
		out = append(out, list.Item{Block: "row " + k, Key: k, Selectable: true})
	}
	return out
}

func (r keyedResults) Count() int { return len(r.keys) }

type keyedDoc struct {
	*stubDoc
	rows  []string
	binds int
}

func (d *keyedDoc) Run() (string, func() Results, error) {
	rows := append([]string(nil), d.rows...)
	return "draft", func() Results { return keyedResults{keys: rows, binds: &d.binds} }, nil
}

func ranEditor(t *testing.T, doc *keyedDoc) (*Editor, *Model) {
	t.Helper()
	e := NewEditor(doc, relayoutEditorKeys(), nil)
	h := New(e)
	h = driveHost(h, tea.WindowSizeMsg{Width: 80, Height: 40})
	h = driveSettled(h, tea.KeyMsg{Type: tea.KeyCtrlR})
	if !e.hasResults || e.Running() {
		t.Fatalf("precondition: run did not land (hasResults=%v running=%v)", e.hasResults, e.Running())
	}
	_ = h.View()
	if !e.OnResults() {
		t.Fatal("precondition: a completed run should focus the results pane")
	}
	return e, h
}

func editorMoveTo(t *testing.T, h *Model, e *Editor, key string) *Model {
	t.Helper()
	for range 8 {
		if it, ok := e.Selected(); ok && it.Key == key {
			return h
		}
		h = driveHost(h, tea.KeyMsg{Type: tea.KeyDown})
	}
	it, _ := e.Selected()
	t.Fatalf("could not reach %q, stuck on %q", key, it.Key)
	return h
}

func TestEditorKeepsResultCursorAcrossMessageRoundTrip(t *testing.T) {
	doc := &keyedDoc{stubDoc: &stubDoc{saved: "report"}, rows: []string{"u1", "u2", "u3"}}
	e, h := ranEditor(t, doc)
	h = editorMoveTo(t, h, e, "u3")

	h = driveHost(h, tea.KeyMsg{Type: tea.KeyCtrlS})
	if h.Top() == View(e) {
		t.Fatal("precondition: save should have pushed a message view")
	}
	h = driveHost(h, tea.WindowSizeMsg{Width: 80, Height: 40})
	cmd := h.Pop()
	if cmd == nil {
		t.Fatal("Pop returned no relayout command")
	}
	driveHost(h, cmd())

	if it, ok := e.Selected(); !ok || it.Key != "u3" {
		t.Fatalf("cursor = %q (%v), want u3: the pop relayout reset the result list", it.Key, ok)
	}
}

func TestEditorSameSizeRelayoutDoesNotRebindResults(t *testing.T) {
	doc := &keyedDoc{stubDoc: &stubDoc{}, rows: []string{"u1", "u2", "u3"}}
	e, h := ranEditor(t, doc)
	h = editorMoveTo(t, h, e, "u2")
	before := doc.binds

	h = driveHost(h, tea.WindowSizeMsg{Width: 80, Height: 40})
	if doc.binds != before {
		t.Fatalf("Results.Items ran %d extra times on an unchanged size, want 0", doc.binds-before)
	}
	if it, ok := e.Selected(); !ok || it.Key != "u2" {
		t.Fatalf("cursor = %q (%v), want u2 after a no-op relayout", it.Key, ok)
	}

	driveHost(h, tea.WindowSizeMsg{Width: 100, Height: 40})
	if doc.binds == before {
		t.Fatal("a real resize must re-bind the result rows")
	}
	if it, ok := e.Selected(); !ok || it.Key != "u2" {
		t.Fatalf("cursor = %q (%v), want u2 preserved across a real resize", it.Key, ok)
	}
}

func TestEditorFreshRunStartsAtTheTopOfTheResults(t *testing.T) {
	doc := &keyedDoc{stubDoc: &stubDoc{}, rows: []string{"u1", "u2", "u3"}}
	e, h := ranEditor(t, doc)
	h = editorMoveTo(t, h, e, "u3")

	h = driveSettled(h, tea.KeyMsg{Type: tea.KeyCtrlR})
	if it, ok := e.Selected(); !ok || it.Key != "u1" {
		t.Fatalf("cursor = %q (%v), want u1: a fresh run must start at the top", it.Key, ok)
	}

	h = editorMoveTo(t, h, e, "u2")
	doc.rows = []string{"u3", "u2", "u1"}
	driveSettled(h, tea.KeyMsg{Type: tea.KeyCtrlR})
	if it, ok := e.Selected(); !ok || it.Key != "u3" {
		t.Fatalf("cursor = %q (%v), want u3: a rerun must land on the first row of the new set", it.Key, ok)
	}
}

func TestEditorRebindsResultsWhenTheThemeChanges(t *testing.T) {
	t.Cleanup(func() { theme.Use(theme.Default()) })
	useTheme(t, "default")

	doc := &keyedDoc{stubDoc: &stubDoc{}, rows: []string{"u1", "u2", "u3"}}
	e, h := ranEditor(t, doc)
	h = editorMoveTo(t, h, e, "u2")
	before := doc.binds

	useTheme(t, "solarized-dark")
	driveHost(h, tea.WindowSizeMsg{Width: 80, Height: 40})
	if doc.binds == before {
		t.Fatal("a theme change must re-bind the result rows: Results.Items may style them from theme.Cur()")
	}
	if it, ok := e.Selected(); !ok || it.Key != "u2" {
		t.Fatalf("cursor = %q (%v), want u2 preserved across a theme re-bind", it.Key, ok)
	}
}

func homeSideKey(t *testing.T, shell *HomeShell) string {
	t.Helper()
	it, ok := shell.side.Selected()
	if !ok {
		t.Fatal("side pane has no selection")
	}
	return it.Key
}

func TestHomeShellKeepsSideCursorAcrossDetailRoundTrip(t *testing.T) {
	shell := homeShellWithRows(3)
	h := focusedHomeHost(t, shell)
	for range 2 {
		h = driveHost(h, tea.KeyMsg{Type: tea.KeyDown})
	}
	if got, want := homeSideKey(t, shell), "https://example.com/row-002"; got != want {
		t.Fatalf("precondition: side cursor = %q, want %q", got, want)
	}

	if cmd := h.Push(stubView{title: "detail"}); cmd != nil {
		h = driveHost(h, tea.WindowSizeMsg{Width: 80, Height: 24})
	}
	cmd := h.Pop()
	if cmd == nil {
		t.Fatal("Pop returned no relayout command")
	}
	driveHost(h, cmd())

	if got, want := homeSideKey(t, shell), "https://example.com/row-002"; got != want {
		t.Fatalf("side cursor = %q, want %q: the pop relayout reset the side list", got, want)
	}
}

func TestHomeShellSameWidthRelayoutDoesNotRebindSide(t *testing.T) {
	binds := 0
	shell := NewHomeShell(HomeShellSpec{
		Title:     "home",
		Items:     []MenuItem{{Label: "Go"}},
		SideLabel: "home flight",
		SideFetch: func() any { return "payload" },
		SideBind: func(int, any) []list.Item {
			binds++
			return []list.Item{
				{Block: "Open PRs  (3)"},
				{Block: "alpha", Key: "u1", Selectable: true},
				{Block: "beta", Key: "u2", Selectable: true},
				{Block: "gamma", Key: "u3", Selectable: true},
			}
		},
	})
	h := focusedHomeHost(t, shell)
	h = driveHost(h, tea.KeyMsg{Type: tea.KeyDown})
	if got := homeSideKey(t, shell); got != "u2" {
		t.Fatalf("precondition: side cursor = %q, want u2", got)
	}
	before := binds

	h = driveHost(h, tea.WindowSizeMsg{Width: 80, Height: 24})
	if binds != before {
		t.Fatalf("SideBind ran %d extra times on an unchanged width, want 0", binds-before)
	}
	if got := homeSideKey(t, shell); got != "u2" {
		t.Fatalf("side cursor = %q, want u2 after a no-op relayout", got)
	}

	driveHost(h, tea.WindowSizeMsg{Width: 100, Height: 24})
	if binds == before {
		t.Fatal("a real resize must re-bind the side rows")
	}
	if got := homeSideKey(t, shell); got != "u2" {
		t.Fatalf("side cursor = %q, want u2 preserved across a real resize", got)
	}
}

func themeSGR(t *testing.T) string {
	t.Helper()
	return sgrPrefix(t, theme.Cur().Val)
}

func useTheme(t *testing.T, key string) {
	t.Helper()
	th, ok := theme.Named(key)
	if !ok {
		t.Fatalf("theme %q is not registered", key)
	}
	theme.Use(th)
}

func themedRows() []list.Item {
	th := theme.Cur()
	return []list.Item{
		{Block: th.Val.Render("alpha"), Key: "u1", Selectable: true},
		{Block: th.Val.Render("beta"), Key: "u2", Selectable: true},
	}
}

func TestItemListRebindsWhenTheThemeChanges(t *testing.T) {
	t.Cleanup(func() { theme.Use(theme.Default()) })
	useTheme(t, "default")
	stale := themeSGR(t)

	il := NewItemList(ItemListSpec{
		Title: "results",
		Fetch: func() any { return "payload" },
		Bind:  func(int, any) []list.Item { return themedRows() },
	})
	h := loadedItemList(t, il)
	if !strings.Contains(h.View(), stale) {
		t.Fatalf("precondition: rows do not carry the theme SGR %q", stale)
	}

	useTheme(t, "solarized-dark")
	fresh := themeSGR(t)
	if fresh == stale {
		t.Fatalf("precondition: both themes render the same SGR %q", fresh)
	}

	driveHost(h, tea.WindowSizeMsg{Width: 80, Height: 24})
	body := il.Body(80, 24)
	if strings.Contains(body, stale) {
		t.Fatalf("a same-size relayout stranded the previous theme's SGR %q:\n%q", stale, body)
	}
	if !strings.Contains(body, fresh) {
		t.Fatalf("rows missing the new theme's SGR %q:\n%q", fresh, body)
	}
	if _, ok := il.Selected(); !ok {
		t.Fatal("theme re-bind lost the selection")
	}
}

func TestHomeShellRebindsWhenTheThemeChanges(t *testing.T) {
	t.Cleanup(func() { theme.Use(theme.Default()) })
	useTheme(t, "default")
	stale := themeSGR(t)

	shell := NewHomeShell(HomeShellSpec{
		Title:     "home",
		Items:     []MenuItem{{Label: "Go"}},
		SideLabel: "home flight",
		SideFetch: func() any { return "payload" },
		SideBind:  func(int, any) []list.Item { return themedRows() },
	})
	h := focusedHomeHost(t, shell)
	if !strings.Contains(h.View(), stale) {
		t.Fatalf("precondition: side rows do not carry the theme SGR %q", stale)
	}

	useTheme(t, "solarized-dark")
	fresh := themeSGR(t)
	if fresh == stale {
		t.Fatalf("precondition: both themes render the same SGR %q", fresh)
	}

	driveHost(h, tea.WindowSizeMsg{Width: 80, Height: 24})
	body := shell.Body(80, 24)
	if strings.Contains(body, stale) {
		t.Fatalf("a same-width relayout stranded the previous theme's SGR %q:\n%q", stale, body)
	}
	if !strings.Contains(body, fresh) {
		t.Fatalf("side rows missing the new theme's SGR %q:\n%q", fresh, body)
	}
	if _, ok := shell.side.Selected(); !ok {
		t.Fatal("theme re-bind lost the side selection")
	}
}

func TestHomeShellRelayoutBindsASidePaneRevealedByItsLiveLabel(t *testing.T) {
	subject := ""
	shell := NewHomeShell(HomeShellSpec{
		Title:       "home",
		Items:       []MenuItem{{Label: "Go"}},
		SideLabelFn: func() string { return subject },
		SideFetch:   func() any { return "payload" },
		SideBind: func(int, any) []list.Item {
			return []list.Item{{Block: "row-one", Selectable: true}}
		},
		SideLoading: "side-loading-marker",
	})

	h := New(shell)
	h = driveHost(h, tea.WindowSizeMsg{Width: 80, Height: 24})
	if got := ansi.Strip(h.View()); strings.Contains(got, "side-loading-marker") {
		t.Fatalf("an empty live label should hide the side pane entirely:\n%s", got)
	}

	subject = "home flight"
	h = driveHost(h, tea.WindowSizeMsg{Width: 80, Height: 24})
	if got := ansi.Strip(h.View()); !strings.Contains(got, "side-loading-marker") {
		t.Fatalf("the relayout that revealed the side pane bound nothing into it:\n%s", got)
	}

	h = driveSettled(h, ReloadMsg{})
	if got := ansi.Strip(h.View()); !strings.Contains(got, "row-one") {
		t.Fatalf("side rows never reached the revealed pane:\n%s", got)
	}
}
