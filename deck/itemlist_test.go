package deck

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/list"
)

func driveHost(h *Model, msg tea.Msg) *Model {
	m, _ := h.Update(msg)
	return m.(*Model)
}

func TestItemListLoadsAndShows(t *testing.T) {
	il := NewItemList("results", nil,
		func() any { return "payload" },
		func(width int, fetched any) []list.Item {
			if fetched != "payload" {
				t.Fatalf("fetched = %v", fetched)
			}
			return []list.Item{
				{Block: "alpha", Key: "https://example.com/a", Selectable: true},
				{Block: "beta", Selectable: true},
			}
		},
	)
	h := New(il)
	h = driveHost(h, tea.WindowSizeMsg{Width: 80, Height: 24})
	if cmd := il.Init(); cmd != nil {
		h = driveHost(h, cmd())
	}
	view := h.View()
	for _, want := range []string{"alpha", "beta"} {
		if !strings.Contains(view, want) {
			t.Errorf("missing %q\n%s", want, view)
		}
	}
}

func TestHomeShellMenuOnlyAndSideFocus(t *testing.T) {
	menuOnly := NewHomeShell("home", nil, []MenuItem{{Label: "Quit"}}, "")
	h := New(menuOnly)
	h = driveHost(h, tea.WindowSizeMsg{Width: 80, Height: 24})
	if strings.Contains(h.View(), "◈") {
		t.Fatal("menu-only should not show side label")
	}
	focusGlyph := keys.Cur().Binding(keys.FocusNext).DisplayGlyph()
	for _, hint := range menuOnly.Hints() {
		if hint[0] == focusGlyph {
			t.Fatalf("menu-only should not offer pane switching (%q)", focusGlyph)
		}
	}

	shell := NewHomeShell("home", nil, []MenuItem{{Label: "Go"}}, "side panel")
	shell.SideFetch = func() any { return "x" }
	shell.SideBind = func(width int, fetched any) []list.Item {
		return []list.Item{{Block: "row-one", Selectable: true}}
	}
	h = New(shell)
	h = driveHost(h, tea.WindowSizeMsg{Width: 80, Height: 24})
	if cmd := shell.Init(); cmd != nil {
		h = driveHost(h, cmd())
	}
	view := ansi.Strip(h.View())
	if !strings.Contains(view, "row-one") {
		t.Fatalf("missing side row\n%s", view)
	}
	labelAt := strings.Index(view, "side panel")
	rowAt := strings.Index(view, "row-one")
	if labelAt < 0 || rowAt <= labelAt {
		t.Fatalf("side title/results missing or out of order\n%s", view)
	}
	blank := false
	for _, ln := range strings.Split(view[labelAt:rowAt], "\n")[1:] {
		if strings.TrimSpace(ln) == "" {
			blank = true
			break
		}
	}
	if !blank {
		t.Fatalf("want blank line between side title and results\n%s", view)
	}
	if shell.FocusSide() {
		t.Fatal("want menu focus initially")
	}
	driveHost(h, tea.KeyMsg{Type: tea.KeyTab})
	if !shell.FocusSide() {
		t.Fatal("want side focus after tab")
	}
}

func TestHomeShellBoxTitle(t *testing.T) {
	shell := NewHomeShell("home", nil, []MenuItem{{Label: "Quit"}}, "")
	h := New(shell)
	h = driveHost(h, tea.WindowSizeMsg{Width: 80, Height: 24})
	if strings.Contains(h.View(), "MAIN MENU") {
		t.Fatalf("default BoxTitle should omit titled-box title\n%s", h.View())
	}

	shell.BoxTitle = "MAIN MENU"
	view := h.View()
	if !strings.Contains(view, "MAIN MENU") {
		t.Fatalf("BoxTitle should appear in titled box\n%s", view)
	}
}

func TestItemListReloadRefetches(t *testing.T) {
	payload := "first"
	fetches := 0
	il := NewItemList("results", nil,
		func() any { fetches++; return payload },
		func(_ int, fetched any) []list.Item {
			return []list.Item{{Block: "body " + fmt.Sprint(fetched)}}
		})
	il.ReloadHint = "rerun"

	h := New(il)
	h = driveHost(h, tea.WindowSizeMsg{Width: 80, Height: 24})
	if cmd := il.Init(); cmd != nil {
		h = driveHost(h, cmd())
	}
	if got := ansi.Strip(h.View()); !strings.Contains(got, "body first") {
		t.Fatalf("initial load missing:\n%s", got)
	}
	if !strings.Contains(ansi.Strip(h.View()), "rerun") {
		t.Errorf("footer missing the reload hint:\n%s", ansi.Strip(h.View()))
	}

	payload = "second"
	h = driveSettled(h, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	if fetches != 2 {
		t.Fatalf("reload key fetched %d times, want 2", fetches)
	}
	if got := ansi.Strip(h.View()); !strings.Contains(got, "body second") {
		t.Fatalf("reload key did not refetch:\n%s", got)
	}

	payload = "third"
	h = driveSettled(h, ReloadMsg{})
	if fetches != 3 {
		t.Fatalf("ReloadMsg fetched %d times, want 3", fetches)
	}
	if got := ansi.Strip(h.View()); !strings.Contains(got, "body third") {
		t.Fatalf("ReloadMsg did not refetch:\n%s", got)
	}
}

func TestScrollReloadReRendersBody(t *testing.T) {
	body := "first"
	loads := 0
	sc := NewScroll("report", nil, nil, func() string { loads++; return body })
	sc.ReloadHint = "rerun report"

	h := New(sc)
	h = driveHost(h, tea.WindowSizeMsg{Width: 80, Height: 24})
	if cmd := sc.Init(); cmd != nil {
		h = driveHost(h, cmd())
	}
	if got := ansi.Strip(h.View()); !strings.Contains(got, "first") {
		t.Fatalf("initial load missing:\n%s", got)
	}
	if got := ansi.Strip(h.View()); !strings.Contains(got, "rerun report") {
		t.Errorf("footer missing the reload hint:\n%s", got)
	}

	body = "second"
	h = driveSettled(h, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	if loads != 2 {
		t.Fatalf("reload key loaded %d times, want 2", loads)
	}
	if got := ansi.Strip(h.View()); !strings.Contains(got, "second") {
		t.Fatalf("reload key did not re-render:\n%s", got)
	}

	body = "third"
	h = driveSettled(h, ReloadMsg{})
	if loads != 3 {
		t.Fatalf("ReloadMsg loaded %d times, want 3", loads)
	}
	if got := ansi.Strip(h.View()); !strings.Contains(got, "third") {
		t.Fatalf("ReloadMsg did not re-render:\n%s", got)
	}
}
