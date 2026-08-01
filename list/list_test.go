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

func flightSample() []Item {
	return []Item{
		{Block: "flight  (2)"},
		{Block: "Open PRs  (3)"},
		{Block: "pr one", Key: "u1", Selectable: true},
		{Block: "pr two", Key: "u2", Selectable: true},
		{Block: "pr three", Key: "u3", Selectable: true},
		{Block: "Reviews  (1)"},
		{Block: "rev one", Key: "u4", Selectable: true},
	}
}

func TestMoveBackToFirstSelectableRevealsLeadingHeaders(t *testing.T) {
	t.Parallel()
	m := New()
	m.SetItems(flightSample())
	m.SetSize(80, 5)

	for range 3 {
		m.Move(1)
	}
	for range 3 {
		m.Move(-1)
	}

	if it, _ := m.Selected(); it.Key != "u1" {
		t.Fatalf("cursor = %q, want u1 (back at first selectable)", it.Key)
	}
	if m.offset != 0 {
		t.Fatalf("offset = %d, want 0 so the trunk and first header are reachable", m.offset)
	}
	for _, want := range []string{"flight  (2)", "Open PRs  (3)"} {
		if !strings.Contains(m.View(), want) {
			t.Fatalf("%q truncated off the top of the panel:\n%s", want, m.View())
		}
	}
}

func TestFirstSelectableKeepsCursorVisibleUnderTallLeadingBlock(t *testing.T) {
	t.Parallel()
	m := New()
	m.SetItems([]Item{
		{Block: "h1\nh2\nh3\nh4\nh5\nh6"},
		{Block: "a", Key: "u1", Selectable: true},
		{Block: "b", Key: "u2", Selectable: true},
	})
	m.SetSize(80, 3)

	m.Move(1)
	m.Move(-1)

	if it, _ := m.Selected(); it.Key != "u1" {
		t.Fatalf("cursor = %q, want u1", it.Key)
	}
	if m.offset == 0 {
		t.Fatal("offset forced to 0: a leading block taller than the viewport pushed the cursor off screen")
	}
	if !strings.Contains(m.View(), "a") {
		t.Fatalf("cursor item not visible:\n%s", m.View())
	}
}

func TestLastSelectableRevealsTrailingNonSelectableRow(t *testing.T) {
	t.Parallel()
	m := New()
	m.SetItems([]Item{
		{Block: "hdr"},
		{Block: "a", Key: "u1", Selectable: true},
		{Block: "b", Key: "u2", Selectable: true},
		{Block: "nothing to show"},
	})
	m.SetSize(80, 3)

	m.Move(1)

	if !strings.Contains(m.View(), "nothing to show") {
		t.Fatalf("trailing row unreachable from the last selectable item:\n%s", m.View())
	}
}

func TestScrollDragsCursorSoMoveDoesNotSnapBack(t *testing.T) {
	t.Parallel()
	m := New()
	m.SetItems(flightSample())
	m.SetSize(80, 5)

	m.Scroll(8)
	if m.offset != 8 {
		t.Fatalf("offset after scroll = %d, want 8", m.offset)
	}
	if it, _ := m.Selected(); it.Key != "u3" {
		t.Fatalf("cursor after scroll = %q, want u3 (dragged into the window)", it.Key)
	}

	m.Move(1)
	if m.offset != 8 {
		t.Fatalf("offset = %d, want 8: move snapped the view back to the pre-scroll cursor", m.offset)
	}
}

func TestScrollLeavesCursorUnsetWhenNothingSelectable(t *testing.T) {
	t.Parallel()
	m := New()
	m.SetItems([]Item{{Block: "nothing to show"}})
	m.SetSize(80, 1)

	m.Scroll(1)
	if _, ok := m.Selected(); ok {
		t.Fatal("scroll assigned a cursor to a list with no selectable items")
	}
	m.Move(1)
	if _, ok := m.Selected(); ok {
		t.Fatal("move assigned a cursor to a list with no selectable items")
	}
}

func TestSetItemsSelectsFirstSelectable(t *testing.T) {
	t.Parallel()
	m := New()
	m.SetItems(sample())
	it, ok := m.Selected()
	if !ok || it.Key != "url1" {
		t.Fatalf("first selectable = %v (%v), want url1", it.Key, ok)
	}
}

func TestSetItemsKeepingCursorSurvivesIdenticalRebind(t *testing.T) {
	t.Parallel()
	m := New()
	m.SetItems(flightSample())
	m.SetSize(80, 5)

	for range 3 {
		m.Move(1)
	}
	wantKey, wantCursor, wantOffset := "u4", m.cursor, m.offset
	if it, _ := m.Selected(); it.Key != wantKey {
		t.Fatalf("precondition: cursor = %q, want %q", it.Key, wantKey)
	}
	if wantCursor == 0 || wantOffset == 0 {
		t.Fatalf("precondition: want a non-zero cursor and offset, got %d/%d", wantCursor, wantOffset)
	}

	m.SetItemsKeepingCursor(flightSample())

	if it, _ := m.Selected(); it.Key != wantKey {
		t.Fatalf("cursor = %q, want %q: rebind discarded the selection", it.Key, wantKey)
	}
	if m.cursor != wantCursor {
		t.Fatalf("cursor index = %d, want %d", m.cursor, wantCursor)
	}
	if m.offset != wantOffset {
		t.Fatalf("offset = %d, want %d: rebind discarded the scroll position", m.offset, wantOffset)
	}
}

func TestSetItemsKeepingCursorTracksMovedItemAcrossChangedList(t *testing.T) {
	t.Parallel()
	m := New()
	m.SetItems(flightSample())
	m.SetSize(80, 5)
	m.Move(1)
	if it, _ := m.Selected(); it.Key != "u2" {
		t.Fatalf("precondition: cursor = %q, want u2", it.Key)
	}

	grown := append([]Item{
		{Block: "flight  (2)"},
		{Block: "Open PRs  (4)"},
		{Block: "pr zero", Key: "u0", Selectable: true},
	}, flightSample()[2:]...)
	m.SetItemsKeepingCursor(grown)

	if it, _ := m.Selected(); it.Key != "u2" {
		t.Fatalf("cursor = %q, want u2: a new row above the selection stole the cursor", it.Key)
	}
}

func TestSetItemsKeepingCursorFallsBackWhenKeyIsGone(t *testing.T) {
	t.Parallel()
	m := New()
	m.SetItems(flightSample())
	m.SetSize(80, 5)
	for range 3 {
		m.Move(1)
	}

	m.SetItemsKeepingCursor([]Item{
		{Block: "Open PRs  (1)"},
		{Block: "pr new", Key: "z1", Selectable: true},
	})

	if it, ok := m.Selected(); !ok || it.Key != "z1" {
		t.Fatalf("cursor = %q (%v), want z1 (first selectable)", it.Key, ok)
	}
	if m.cursor != m.firstSelectable() {
		t.Fatalf("cursor = %d, want firstSelectable %d", m.cursor, m.firstSelectable())
	}
	if m.offset != 0 {
		t.Fatalf("offset = %d, want 0 on fallback", m.offset)
	}
}

func TestSetItemsKeepingCursorHoldsIndexForKeylessRows(t *testing.T) {
	t.Parallel()
	rows := func(suffix string) []Item {
		return []Item{
			{Block: "header"},
			{Block: "first" + suffix, Selectable: true},
			{Block: "second" + suffix, Selectable: true},
			{Block: "third" + suffix, Selectable: true},
		}
	}
	m := New()
	m.SetItems(rows(""))
	m.SetSize(80, 5)
	m.Move(1)
	m.Move(1)
	if it, _ := m.Selected(); it.Block != "third" {
		t.Fatalf("precondition: cursor = %q, want third", it.Block)
	}

	m.SetItemsKeepingCursor(rows(" (wrapped)"))

	if it, ok := m.Selected(); !ok || it.Block != "third (wrapped)" {
		t.Fatalf("cursor = %q (%v), want third (wrapped): a keyless rebind reset the selection", it.Block, ok)
	}
	if m.cursor != 3 {
		t.Fatalf("cursor index = %d, want 3", m.cursor)
	}
}

func TestSetItemsKeepingCursorDropsKeylessCursorWhenTheIndexIsGone(t *testing.T) {
	t.Parallel()
	m := New()
	m.SetItems([]Item{
		{Block: "header"},
		{Block: "first", Selectable: true},
		{Block: "second", Selectable: true},
	})
	m.SetSize(80, 5)
	m.Move(1)
	if it, _ := m.Selected(); it.Block != "second" {
		t.Fatalf("precondition: cursor = %q, want second", it.Block)
	}

	m.SetItemsKeepingCursor([]Item{
		{Block: "header"},
		{Block: "only", Selectable: true},
		{Block: "footer"},
	})

	if it, ok := m.Selected(); !ok || it.Block != "only" {
		t.Fatalf("cursor = %q (%v), want only (first selectable)", it.Block, ok)
	}
	if m.offset != 0 {
		t.Fatalf("offset = %d, want 0 on fallback", m.offset)
	}
}

func TestSetItemsKeepingCursorResolvesDuplicateKeysToTheFirstMatch(t *testing.T) {
	t.Parallel()
	dupes := []Item{
		{Block: "Open PRs  (3)"},
		{Block: "pr one", Key: "dup", Selectable: true},
		{Block: "pr two", Key: "dup", Selectable: true},
		{Block: "pr three", Key: "dup", Selectable: true},
	}
	m := New()
	m.SetItems(dupes)
	m.SetSize(80, 6)
	m.Move(1)
	m.Move(1)
	if m.cursor != 3 {
		t.Fatalf("precondition: cursor index = %d, want 3", m.cursor)
	}

	m.SetItemsKeepingCursor(dupes)

	if m.cursor != 1 {
		t.Fatalf("cursor index = %d, want 1: duplicate keys resolve to the first match", m.cursor)
	}
	if it, ok := m.Selected(); !ok || it.Key != "dup" {
		t.Fatalf("selection = %q (%v), want the dup key", it.Key, ok)
	}
}

func TestSetItemsKeepingCursorHandlesEmptyAndUnselectableLists(t *testing.T) {
	t.Parallel()
	m := New()
	m.SetItems(flightSample())
	m.SetSize(80, 5)
	m.Move(1)

	m.SetItemsKeepingCursor(nil)
	if _, ok := m.Selected(); ok {
		t.Fatal("empty rebind kept a selection")
	}

	m.SetItemsKeepingCursor([]Item{{Block: "nothing to show"}})
	if _, ok := m.Selected(); ok {
		t.Fatal("rebind onto an unselectable list assigned a cursor")
	}

	m.SetItemsKeepingCursor(flightSample())
	if it, ok := m.Selected(); !ok || it.Key != "u1" {
		t.Fatalf("cursor = %q (%v), want u1 once rows are selectable again", it.Key, ok)
	}
}

func TestSetItemsKeepingCursorKeepsScrollWithoutSelectables(t *testing.T) {
	t.Parallel()
	rows := []Item{{Block: "a\nb\nc\nd\ne\nf"}}
	m := New()
	m.SetItems(rows)
	m.SetSize(80, 2)
	m.Scroll(3)
	if m.offset != 3 {
		t.Fatalf("precondition: offset = %d, want 3", m.offset)
	}

	m.SetItemsKeepingCursor(rows)
	if m.offset != 3 {
		t.Fatalf("offset = %d, want 3: rebind reset a scroll-only list", m.offset)
	}
	if _, ok := m.Selected(); ok {
		t.Fatal("rebind assigned a cursor to a list with no selectable items")
	}
}

func TestSetItemsStillResetsSelection(t *testing.T) {
	t.Parallel()
	m := New()
	m.SetItems(flightSample())
	m.SetSize(80, 5)
	for range 3 {
		m.Move(1)
	}

	m.SetItems(flightSample())

	if it, _ := m.Selected(); it.Key != "u1" {
		t.Fatalf("cursor = %q, want u1: SetItems must keep its reset semantics", it.Key)
	}
	if m.offset != 0 {
		t.Fatalf("offset = %d, want 0", m.offset)
	}
}

func TestMoveSkipsHeadersAndClamps(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	m := New()
	m.SetItems([]Item{{Block: "nothing to show"}})
	if _, ok := m.Selected(); ok {
		t.Fatal("expected no selection when no item is selectable")
	}
}

func TestViewWindowsToHeightAndKeepsCursorVisible(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
