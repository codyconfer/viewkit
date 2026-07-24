package deck

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/viewkit/list"
)

func driveHost(h *Host, msg tea.Msg) *Host {
	m, _ := h.Update(msg)
	return m.(*Host)
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
	for _, hint := range menuOnly.Hints() {
		if hint[0] == "tab" {
			t.Fatal("menu-only should not offer tab")
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
	if !strings.Contains(h.View(), "row-one") {
		t.Fatalf("missing side row\n%s", h.View())
	}
	if shell.FocusSide() {
		t.Fatal("want menu focus initially")
	}
	h = driveHost(h, tea.KeyMsg{Type: tea.KeyTab})
	if !shell.FocusSide() {
		t.Fatal("want side focus after tab")
	}
}
