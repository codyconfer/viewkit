package deck

import (
	"strings"
	"testing"
)

func rowIndex(lines []string, want string) int {
	for i, ln := range lines {
		if strings.Contains(ln, want) {
			return i
		}
	}
	return -1
}

func TestMenuGivesOnlyTheCursorRowBreathingRoom(t *testing.T) {
	m := NewMenu("queries", nil,
		MenuItem{Label: "first"},
		MenuItem{Label: "second"},
		MenuItem{Label: "third"},
	)
	lines := strings.Split(m.Body(60, 30), "\n")

	firstAt := rowIndex(lines, "first")
	secondAt := rowIndex(lines, "second")
	thirdAt := rowIndex(lines, "third")
	if firstAt < 0 || secondAt < 0 || thirdAt < 0 {
		t.Fatalf("items missing:\n%s", strings.Join(lines, "\n"))
	}
	if gap := secondAt - firstAt; gap != 2 {
		t.Errorf("cursor row should have a blank line after it, gap = %d:\n%s", gap, strings.Join(lines, "\n"))
	}
	if gap := thirdAt - secondAt; gap != 1 {
		t.Errorf("rows away from the cursor should stay tight, gap = %d:\n%s", gap, strings.Join(lines, "\n"))
	}
}

func TestMenuBreathingRoomFollowsTheCursor(t *testing.T) {
	m := NewMenu("queries", nil,
		MenuItem{Label: "first"},
		MenuItem{Label: "second"},
		MenuItem{Label: "third"},
	)
	m.cursor = 1
	lines := strings.Split(m.Body(60, 30), "\n")

	firstAt := rowIndex(lines, "first")
	secondAt := rowIndex(lines, "second")
	thirdAt := rowIndex(lines, "third")
	if secondAt-firstAt != 2 || thirdAt-secondAt != 2 {
		t.Errorf("a middle cursor should be blank-separated on both sides:\n%s", strings.Join(lines, "\n"))
	}
}

func TestMenuDropsSpacingRatherThanOverflow(t *testing.T) {
	items := make([]MenuItem, 12)
	for i := range items {
		items[i] = MenuItem{Label: string(rune('a' + i))}
	}
	m := NewMenu("queries", nil, items...)

	tight := m.Body(60, len(items)+menuBoxChrome)
	if got := strings.Count(tight, "\n") + 1; got != len(items)+menuBoxChrome {
		t.Errorf("exact-fit menu rendered %d lines, want %d:\n%s", got, len(items)+menuBoxChrome, tight)
	}
	firstAt := rowIndex(strings.Split(tight, "\n"), " a ")
	secondAt := rowIndex(strings.Split(tight, "\n"), " b ")
	if firstAt < 0 || secondAt-firstAt != 1 {
		t.Errorf("exact-fit menu should drop the breathing room:\n%s", tight)
	}

	roomy := m.Body(60, 40)
	if strings.Count(roomy, "\n") <= strings.Count(tight, "\n") {
		t.Error("a tall terminal should get the cursor breathing room")
	}
}
