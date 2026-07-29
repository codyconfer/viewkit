package deck

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/codyconfer/viewkit/list"
)

func homeShellWithRows(n int) *HomeShell {
	shell := NewHomeShell("home", nil, []MenuItem{{Label: "Go"}}, "home flight · morning")
	shell.SideFetch = func() any { return n }
	shell.SideBind = func(int, any) []list.Item {
		items := []list.Item{
			{Block: "morning  (1)"},
			{Block: "Open PRs  (" + fmt.Sprint(n) + ")"},
		}
		for i := range n {
			label := fmt.Sprintf("row-%03d", i)
			items = append(items, list.Item{Block: label, Key: "https://example.com/" + label, Selectable: true})
		}
		return items
	}
	return shell
}

func focusedHomeHost(t *testing.T, shell *HomeShell) *Model {
	t.Helper()
	h := New(shell)
	h = driveHost(h, tea.WindowSizeMsg{Width: 80, Height: 24})
	if cmd := shell.Init(); cmd != nil {
		h = driveHost(h, cmd())
	}
	h = driveHost(h, tea.KeyMsg{Type: tea.KeyTab})
	if !shell.FocusSide() {
		t.Fatal("tab did not focus the side pane")
	}
	return h
}

func TestHomeShellPageKeysMoveAFullPage(t *testing.T) {
	shell := homeShellWithRows(40)
	h := focusedHomeHost(t, shell)

	if got := ansi.Strip(h.View()); !strings.Contains(got, "row-000") {
		t.Fatalf("want top of the side list before paging:\n%s", got)
	}

	h = driveHost(h, tea.KeyMsg{Type: tea.KeyPgDown})
	got := ansi.Strip(h.View())
	if strings.Contains(got, "row-000") {
		t.Fatalf("pgdown scrolled a single line, not a page:\n%s", got)
	}

	h = driveHost(h, tea.KeyMsg{Type: tea.KeyPgUp})
	if got := ansi.Strip(h.View()); !strings.Contains(got, "row-000") {
		t.Fatalf("pgup did not return to the top of the side list:\n%s", got)
	}
}

func TestHomeShellPageDownThenUpArrowKeepsTopReachable(t *testing.T) {
	shell := homeShellWithRows(40)
	h := focusedHomeHost(t, shell)

	for range 4 {
		h = driveHost(h, tea.KeyMsg{Type: tea.KeyPgDown})
	}
	for range 6 {
		h = driveHost(h, tea.KeyMsg{Type: tea.KeyPgUp})
	}
	for _, want := range []string{"morning  (1)", "Open PRs  (40)", "row-000"} {
		if got := ansi.Strip(h.View()); !strings.Contains(got, want) {
			t.Fatalf("paging down then back up left %q truncated off the top:\n%s", want, got)
		}
	}

	for range 6 {
		h = driveHost(h, tea.KeyMsg{Type: tea.KeyDown})
	}
	for range 6 {
		h = driveHost(h, tea.KeyMsg{Type: tea.KeyUp})
	}
	for _, want := range []string{"morning  (1)", "Open PRs  (40)", "row-000"} {
		if got := ansi.Strip(h.View()); !strings.Contains(got, want) {
			t.Fatalf("arrowing down then back up left %q truncated off the top:\n%s", want, got)
		}
	}
}
