package deck

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/list"
	"github.com/codyconfer/viewkit/theme"
	"github.com/codyconfer/viewkit/ui"
)

// TestScopedDeckViewUnderStrictScope drives a garish-scoped Model through
// View() with Menu, Message, and ItemList on the stack: any global-theme
// fallback in the render path panics.
func TestScopedDeckViewUnderStrictScope(t *testing.T) {
	prevStrict := layout.StrictScope
	layout.StrictScope = true
	t.Cleanup(func() { layout.StrictScope = prevStrict })

	garish := theme.New(theme.Palette{
		Accent:   lipgloss.Color("#ff00ff"),
		Border:   lipgloss.Color("#00ffff"),
		Muted:    lipgloss.Color("#00ff00"),
		Text:     lipgloss.Color("#ffff00"),
		Selected: lipgloss.Color("#ff0000"),
		Success:  lipgloss.Color("#00ff88"),
		Warning:  lipgloss.Color("#ff8800"),
		Failure:  lipgloss.Color("#8800ff"),
		Info:     lipgloss.Color("#0088ff"),
		Series2:  lipgloss.Color("#88ff00"),
		Series3:  lipgloss.Color("#ff0088"),
		Bg:       lipgloss.Color("#000000"),
	})
	scope := ui.Default()
	scope.Theme = garish

	menu := NewMenu("Main", nil,
		MenuItem{Label: "one", Desc: "first"},
		MenuItem{Label: "two", Icon: "◆"},
	)
	h := New(menu, WithScope(scope))
	h = driveHost(h, tea.WindowSizeMsg{Width: 100, Height: 40})

	view := func(stage string) string {
		t.Helper()
		var out string
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("%s View() hit a global fallback under StrictScope (scope leak): %v", stage, r)
				}
			}()
			out = h.View()
		}()
		if !strings.Contains(out, "38;2;255;0;255") {
			t.Errorf("%s View() missing garish accent SGR:\n%q", stage, out)
		}
		return out
	}

	view("Menu")

	_ = h.Push(NewMessage("Note", "hello there", nil))
	h = driveHost(h, tea.WindowSizeMsg{Width: 100, Height: 40})
	view("Message")
	_ = h.Pop()

	il := NewItemList(ItemListSpec{
		Title: "Items",
		Bind: func(width int, _ any) []list.Item {
			return []list.Item{
				{Block: "row one", Selectable: true},
				{Block: "row two", Selectable: true},
			}
		},
	})
	_ = h.Push(il)
	h = driveHost(h, tea.WindowSizeMsg{Width: 100, Height: 40})
	h = driveHost(h, itemListLoadedMsg{own: il})
	out := view("ItemList")
	if !strings.Contains(out, "row one") {
		t.Fatalf("ItemList rows missing from View():\n%q", out)
	}
}
