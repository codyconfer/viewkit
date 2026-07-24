package list

import (
	"strings"
	"testing"

	"github.com/codyconfer/viewkit/theme"
)

func sample() []Item {
	return []Item{
		{Block: "Section A"},
		{Block: "item1\nurl1", Key: "url1", Selectable: true},
		{Block: "item2", Key: "url2", Selectable: true},
	}
}

func TestSetItemsSelectsFirstSelectable(t *testing.T) {
	m := New()
	m.SetItems(sample())
	it, ok := m.Selected()
	if !ok || it.Key != "url1" {
		t.Fatalf("first selectable = %v (%v), want url1", it.Key, ok)
	}
}

func TestMoveSkipsHeadersAndClamps(t *testing.T) {
	m := New()
	m.SetItems(sample())

	m.Move(1)
	if it, _ := m.Selected(); it.Key != "url2" {
		t.Fatalf("after down = %q, want url2", it.Key)
	}
	m.Move(1)
	if it, _ := m.Selected(); it.Key != "url2" {
		t.Fatalf("down past end = %q, want url2 (clamped)", it.Key)
	}
	m.Move(-1)
	if it, _ := m.Selected(); it.Key != "url1" {
		t.Fatalf("after up = %q, want url1", it.Key)
	}
	m.Move(-1)
	if it, _ := m.Selected(); it.Key != "url1" {
		t.Fatalf("up onto header = %q, want url1 (clamped, header skipped)", it.Key)
	}
}

func TestNoSelectableYieldsNoSelection(t *testing.T) {
	m := New()
	m.SetItems([]Item{{Block: "nothing to show"}})
	if _, ok := m.Selected(); ok {
		t.Fatal("expected no selection when no item is selectable")
	}
}

func TestViewWindowsToHeightAndKeepsCursorVisible(t *testing.T) {
	m := New()
	m.SetItems(sample())
	m.SetSize(80, 2)
	m.Move(1)

	lines := strings.Split(m.View(), "\n")
	if len(lines) != 2 {
		t.Fatalf("view height = %d lines, want 2", len(lines))
	}
	if !strings.Contains(m.View(), "item2") {
		t.Fatalf("selected item2 not visible in window:\n%s", m.View())
	}
}

func TestRenderInsertsItemGapBetweenNodes(t *testing.T) {
	m := New()
	m.SetItems([]Item{
		{Block: "a", Key: "a", Selectable: true},
		{Block: "b", Key: "b", Selectable: true},
	})
	m.SetFocused(true)

	lines := strings.Split(m.View(), "\n")
	wantGap := theme.ListItemGapY
	if wantGap < 1 {
		t.Fatal("ListItemGapY must be >= 1")
	}
	if len(lines) != 2+wantGap {
		t.Fatalf("line count = %d, want %d (2 items + %d gap):\n%q", len(lines), 2+wantGap, wantGap, m.View())
	}
	for i := 1; i <= wantGap; i++ {
		if lines[i] != "" {
			t.Fatalf("gap line %d = %q, want blank", i, lines[i])
		}
	}
	if !strings.Contains(lines[0], "a") || !strings.Contains(lines[len(lines)-1], "b") {
		t.Fatalf("expected items around gap:\n%s", m.View())
	}
}

func TestRenderGapStemContinuesTreeThroughItemGap(t *testing.T) {
	m := New()
	m.SetItems([]Item{
		{Block: "│  ├─ a", Key: "a", Selectable: true, GapStem: "│  │  "},
		{Block: "│  └─ b", Key: "b", Selectable: true, GapStem: "│     "},
	})

	lines := strings.Split(m.View(), "\n")
	wantGap := theme.ListItemGapY
	if wantGap < 1 {
		t.Fatal("ListItemGapY must be >= 1 for gap-stem coverage")
	}
	if len(lines) != 2+wantGap {
		t.Fatalf("line count = %d, want %d:\n%q", len(lines), 2+wantGap, m.View())
	}
	for i := 1; i <= wantGap; i++ {
		got := lines[i]
		want := "  │  │  "
		if got != want {
			t.Fatalf("gap line %d = %q, want %q (stem through ListItemGapY)", i, got, want)
		}
	}
	if !strings.Contains(lines[0], "├─ a") || !strings.Contains(lines[len(lines)-1], "└─ b") {
		t.Fatalf("expected tree items around gap stem:\n%s", m.View())
	}
}
