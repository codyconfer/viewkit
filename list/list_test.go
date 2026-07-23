package list

import (
	"strings"
	"testing"
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
