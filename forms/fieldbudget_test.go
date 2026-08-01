package forms

import (
	"testing"

	"github.com/codyconfer/viewkit/layout"
)

const fieldBudgetContract = "layout.NewFrame clamps any width below theme.MinBodyWidth up to it, so every " +
	"narrow body fieldRowBudget can receive through render() is identical. The narrow arms are only " +
	"reachable by calling the helper directly, which is what this test does: without it, changing " +
	"max(body-fieldRowChrome, 2) or the room/2 floor is invisible to the whole suite."

func TestFieldRowBudgetNarrowArms(t *testing.T) {
	t.Parallel()
	t.Log(fieldBudgetContract)
	const longLabel = 72

	for _, tc := range []struct {
		body, label      int
		wantLabel, wantV int
		arm              string
	}{
		{body: 0, label: longLabel, wantLabel: 1, wantV: 1, arm: "room floor of 2, then the room/2 floor"},
		{body: 4, label: longLabel, wantLabel: 1, wantV: 1, arm: "room floor of 2"},
		{body: 6, label: longLabel, wantLabel: 1, wantV: 1, arm: "room floor of 2"},
		{body: 8, label: longLabel, wantLabel: 2, wantV: 2, arm: "room/2 floor"},
		{body: 14, label: longLabel, wantLabel: 5, wantV: 5, arm: "room/2 floor"},
		{body: 20, label: longLabel, wantLabel: 8, wantV: 8, arm: "room/2 floor exactly meets fieldValueMin"},
		{body: 24, label: longLabel, wantLabel: 12, wantV: 8, arm: "room-fieldValueMin"},
		{body: 40, label: longLabel, wantLabel: 28, wantV: 8, arm: "room-fieldValueMin"},
		{body: 40, label: 5, wantLabel: 5, wantV: 31, arm: "short label passes through"},
	} {
		gotLabel, gotV := fieldRowBudget(tc.body, tc.label)
		if gotLabel != tc.wantLabel || gotV != tc.wantV {
			t.Errorf("fieldRowBudget(body=%d, label=%d) = (%d, %d), want (%d, %d) via the %s",
				tc.body, tc.label, gotLabel, gotV, tc.wantLabel, tc.wantV, tc.arm)
		}
		if gotLabel < 1 || gotV < 1 {
			t.Errorf("fieldRowBudget(body=%d, label=%d) = (%d, %d): neither side may reach zero or the "+
				"label or the value disappears entirely", tc.body, tc.label, gotLabel, gotV)
		}
	}
}

func TestFieldRowBudgetNeverNeedsAnUpperClamp(t *testing.T) {
	t.Parallel()
	t.Log("labelW = max(room-fieldValueMin, room/2) is always <= room-1 for room >= 2, which is why the " +
		"trailing min(labelW, room-1) was deleted as a proven no-op. If a future edit breaks that, the " +
		"value column goes to zero width and this test says so.")
	for body := 0; body <= 200; body++ {
		for _, label := range []int{0, 1, 2, 7, 8, 9, 40, 72, 500} {
			labelW, valW := fieldRowBudget(body, label)
			room := labelW + valW
			if labelW > room-1 {
				t.Fatalf("fieldRowBudget(body=%d, label=%d) gave labelW=%d of room=%d, leaving %d columns "+
					"for the value", body, label, labelW, room, valW)
			}
		}
	}
}

func TestFrameClampMakesNarrowWidthsIdentical(t *testing.T) {
	t.Parallel()
	t.Log(fieldBudgetContract)
	base := layout.NewFrame(1).BodyWidth()
	for _, w := range []int{1, 5, 10, 20, 24} {
		if got := layout.NewFrame(w).BodyWidth(); got != base {
			t.Fatalf("NewFrame(%d).BodyWidth() = %d, want %d; if the clamp changed, the width list in "+
				"TestLongLabelValueVisibleAcrossWidths is no longer redundant and should be widened again",
				w, got, base)
		}
	}
	if got := layout.NewFrame(base + 4).BodyWidth(); got == base {
		t.Fatalf("NewFrame(%d).BodyWidth() = %d, want a width above the clamp so the test list covers both "+
			"sides of it", base+4, got)
	}
}
