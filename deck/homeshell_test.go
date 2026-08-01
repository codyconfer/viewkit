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
	shell := NewHomeShell(HomeShellSpec{
		Title:     "home",
		Items:     []MenuItem{{Label: "Go"}},
		SideLabel: "home flight · morning",
		SideFetch: func() any { return n },
		SideBind: func(int, any) []list.Item {
			items := []list.Item{
				{Block: "morning  (1)"},
				{Block: "Open PRs  (" + fmt.Sprint(n) + ")"},
			}
			for i := range n {
				label := fmt.Sprintf("row-%03d", i)
				items = append(items, list.Item{Block: label, Key: "https://example.com/" + label, Selectable: true})
			}
			return items
		},
	})
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

func driveSettled(h *Model, msg tea.Msg) *Model {
	m, cmd := h.Update(msg)
	h = m.(*Model)
	for cmd != nil {
		next := cmd()
		if next == nil {
			break
		}
		m, cmd = h.Update(next)
		h = m.(*Model)
	}
	return h
}

func TestHomeShellReloadRefetchesAndFollowsLiveLabel(t *testing.T) {
	subject := "morning"
	fetches := 0
	var shell *HomeShell
	shell = NewHomeShell(HomeShellSpec{
		Title:       "home",
		Items:       []MenuItem{{Label: "Go"}},
		SideLabelFn: func() string { return "home flight · " + subject },
		SideFetch:   func() any { fetches++; return subject },
		SideBind: func(int, any) []list.Item {
			return []list.Item{{Block: "loaded " + fmt.Sprint(shell.fetched)}}
		},
	})

	h := New(shell)
	h = driveHost(h, tea.WindowSizeMsg{Width: 80, Height: 24})
	if cmd := shell.Init(); cmd != nil {
		h = driveHost(h, cmd())
	}
	if fetches != 1 {
		t.Fatalf("Init fetched %d times, want 1", fetches)
	}

	subject = "evening"
	h = driveSettled(h, ReloadMsg{})
	if fetches != 2 {
		t.Fatalf("ReloadMsg fetched %d times, want 2", fetches)
	}
	got := ansi.Strip(h.View())
	for _, want := range []string{"home flight · evening", "loaded evening"} {
		if !strings.Contains(got, want) {
			t.Errorf("after reload missing %q\n%s", want, got)
		}
	}

	subject = "night"
	h = driveSettled(h, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	if fetches != 3 {
		t.Fatalf("reload key fetched %d times, want 3", fetches)
	}
	if got := ansi.Strip(h.View()); !strings.Contains(got, "loaded night") {
		t.Errorf("reload key did not refetch\n%s", got)
	}
}

func TestHomeShellEmptyLiveLabelHidesSidePane(t *testing.T) {
	subject := ""
	shell := NewHomeShell(HomeShellSpec{
		Title:       "home",
		Items:       []MenuItem{{Label: "Go"}},
		SideLabelFn: func() string { return subject },
		SideFetch:   func() any { return subject },
		SideBind:    func(int, any) []list.Item { return []list.Item{{Block: "rows"}} },
	})

	h := New(shell)
	h = driveHost(h, tea.WindowSizeMsg{Width: 80, Height: 24})
	if got := ansi.Strip(h.View()); strings.Contains(got, "rows") {
		t.Fatalf("empty live label still rendered a side pane:\n%s", got)
	}

	subject = "morning"
	h = driveSettled(h, ReloadMsg{})
	if got := ansi.Strip(h.View()); !strings.Contains(got, "rows") {
		t.Fatalf("side pane did not appear once the live label filled in:\n%s", got)
	}
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
