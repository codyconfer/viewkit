package layout

import (
	"strings"
	"testing"
)

func TestCursorRowsSurroundsOnlyTheCursor(t *testing.T) {
	rows := []string{"a", "b", "c", "d"}
	got := CursorRows(rows, 2, 0)
	want := []string{"a", "b", "", "c", "", "d"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("CursorRows = %q, want %q", got, want)
	}
}

func TestCursorRowsAtTheEndsAddsOneGap(t *testing.T) {
	rows := []string{"a", "b", "c"}

	first := CursorRows(rows, 0, 0)
	if strings.Join(first, "|") != "a||b|c" {
		t.Errorf("cursor at first = %q", first)
	}
	last := CursorRows(rows, 2, 0)
	if strings.Join(last, "|") != "a|b||c" {
		t.Errorf("cursor at last = %q", last)
	}
}

func TestCursorRowsLeavesShortAndUnfocusedListsAlone(t *testing.T) {
	single := []string{"only"}
	if got := CursorRows(single, 0, 0); len(got) != 1 {
		t.Errorf("single row = %q, want untouched", got)
	}
	rows := []string{"a", "b", "c"}
	for _, cursor := range []int{-1, 3, 99} {
		if got := CursorRows(rows, cursor, 0); len(got) != len(rows) {
			t.Errorf("cursor %d = %q, want untouched", cursor, got)
		}
	}
}

func TestCursorRowsDropsGapsRatherThanOverflow(t *testing.T) {
	rows := []string{"a", "b", "c", "d"}

	if got := CursorRows(rows, 1, len(rows)); len(got) != len(rows) {
		t.Errorf("exact-fit budget should stay tight, got %q", got)
	}
	if got := CursorRows(rows, 1, len(rows)+1); len(got) != len(rows) {
		t.Errorf("a budget one short of both gaps should stay tight, got %q", got)
	}
	if got := CursorRows(rows, 1, len(rows)+2); len(got) != len(rows)+2 {
		t.Errorf("a budget that fits both gaps should space, got %q", got)
	}
}
