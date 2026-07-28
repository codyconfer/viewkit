package deck

import (
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
	// Host margin padding may leave spaces on the blank separator line.
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
