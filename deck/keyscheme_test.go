package deck

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/list"
)

func useScheme(t *testing.T, overrides ...keys.Binding) {
	t.Helper()
	keys.Use(keys.Default().With(overrides...))
	t.Cleanup(func() { keys.Use(keys.Default()) })
}

func TestFooterLegendFollowsScheme(t *testing.T) {
	useScheme(t,
		keys.Binding{Keys: []string{"q"}, Action: keys.Cancel, Glyph: "q"},
		keys.Binding{Keys: []string{"ctrl+x"}, Action: keys.Quit, Glyph: "ctrl+x"},
	)
	h := New(stubView{title: "root"})
	h = driveHost(h, tea.WindowSizeMsg{Width: 100, Height: 40})

	view := ansi.Strip(h.View())
	for _, want := range []string{"q back", "ctrl+x quit"} {
		if !strings.Contains(view, want) {
			t.Fatalf("footer legend missing %q:\n%s", want, view)
		}
	}
	for _, stale := range []string{"esc back", "ctrl+c quit"} {
		if strings.Contains(view, stale) {
			t.Fatalf("footer legend still hard-codes %q:\n%s", stale, view)
		}
	}
}

func TestHostQuitFollowsScheme(t *testing.T) {
	useScheme(t, keys.Binding{Keys: []string{"ctrl+x"}, Action: keys.Quit, Glyph: "ctrl+x"})
	h := New(stubView{title: "root"})

	_, cmd := h.Update(tea.KeyMsg{Type: tea.KeyCtrlX})
	if cmd == nil {
		t.Fatal("scheme Quit key did not quit")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("want tea.QuitMsg from scheme Quit key, got %T", cmd())
	}

	if _, cmd = h.Update(tea.KeyMsg{Type: tea.KeyCtrlC}); cmd != nil {
		if _, ok := cmd().(tea.QuitMsg); ok {
			t.Fatal("ctrl+c quit even though the scheme rebound Quit to ctrl+x")
		}
	}
}

func TestWithQuitHintRendersCustomMatcher(t *testing.T) {
	h := New(stubView{title: "root"},
		WithQuitCheck(func(k string) bool { return k == "ctrl+d" }),
		WithQuitHint("ctrl+d"),
	)
	h = driveHost(h, tea.WindowSizeMsg{Width: 100, Height: 40})

	view := ansi.Strip(h.View())
	if !strings.Contains(view, "ctrl+d quit") {
		t.Fatalf("footer legend missing WithQuitHint glyph:\n%s", view)
	}
	if strings.Contains(view, "ctrl+c quit") {
		t.Fatalf("footer legend contradicts the injected quit matcher:\n%s", view)
	}
}

func TestHomeShellFocusFollowsScheme(t *testing.T) {
	useScheme(t, keys.Binding{Keys: []string{"ctrl+n"}, Action: keys.FocusNext, Glyph: "ctrl+n"})
	shell := NewHomeShell("home", nil, []MenuItem{{Label: "Go"}}, "side panel")
	shell.SideFetch = func() any { return "x" }
	shell.SideBind = func(int, any) []list.Item {
		return []list.Item{{Block: "row-one", Selectable: true}}
	}
	h := New(shell)
	h = driveHost(h, tea.WindowSizeMsg{Width: 80, Height: 24})

	h = driveHost(h, tea.KeyMsg{Type: tea.KeyTab})
	if shell.FocusSide() {
		t.Fatal("raw tab switched panes even though the scheme rebound FocusNext")
	}
	driveHost(h, tea.KeyMsg{Type: tea.KeyCtrlN})
	if !shell.FocusSide() {
		t.Fatal("rebound FocusNext key did not switch panes")
	}

	glyphs := make([]string, 0, len(shell.Hints()))
	for _, hint := range shell.Hints() {
		glyphs = append(glyphs, hint[0])
	}
	if !slices.Contains(glyphs, "ctrl+n") {
		t.Fatalf("hints %v do not advertise the rebound focus key", glyphs)
	}
}

func TestScrollPagesViaScheme(t *testing.T) {
	useScheme(t, keys.Binding{Keys: []string{"ctrl+f"}, Action: keys.PageDown, Glyph: "ctrl+f"})
	lines := make([]string, 200)
	for i := range lines {
		lines[i] = fmt.Sprintf("line-%03d", i)
	}
	sc := NewScroll("log", nil, nil, func() string { return strings.Join(lines, "\n") })
	h := New(sc)
	h = driveHost(h, tea.WindowSizeMsg{Width: 80, Height: 24})
	if cmd := sc.Init(); cmd != nil {
		h = driveHost(h, cmd())
	}
	if !strings.Contains(ansi.Strip(h.View()), "line-000") {
		t.Fatalf("want top of content before paging:\n%s", h.View())
	}

	h = driveHost(h, tea.KeyMsg{Type: tea.KeyCtrlF})
	if got := ansi.Strip(h.View()); strings.Contains(got, "line-000") {
		t.Fatalf("rebound PageDown did not page the viewport:\n%s", got)
	}

	h = driveHost(h, tea.KeyMsg{Type: tea.KeyPgDown})
	first := firstContentLine(t, ansi.Strip(h.View()))
	h = driveHost(h, tea.KeyMsg{Type: tea.KeyPgDown})
	if got := firstContentLine(t, ansi.Strip(h.View())); got != first {
		t.Fatalf("pgdown still pages after being unbound: %q then %q", first, got)
	}
}

func firstContentLine(t *testing.T, view string) string {
	t.Helper()
	for ln := range strings.SplitSeq(view, "\n") {
		if i := strings.Index(ln, "line-"); i >= 0 {
			return strings.TrimSpace(ln[i:])
		}
	}
	t.Fatalf("no body line found in view:\n%s", view)
	return ""
}
