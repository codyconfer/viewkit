package forms

import (
	"strings"
	"testing"

	"github.com/codyconfer/viewkit/layout"
)

func TestLongLabelKeepsValueVisible(t *testing.T) {
	t.Parallel()
	fm := NewForm(Field{Key: "a", Label: strings.Repeat("L", 30), Text: "SECRETVALUE"})
	out := stripANSI(fm.Render(layout.NewFrame(24), "form"))

	if !strings.Contains(out, "SECRET") {
		t.Fatalf("a 30 column label hid the value at width 24:\n%s", out)
	}
	if !strings.Contains(out, "▎") {
		t.Fatalf("a 30 column label hid the caret at width 24:\n%s", out)
	}
	if !strings.Contains(out, "…") {
		t.Fatalf("the label should be clipped, not the whole value:\n%s", out)
	}
	if strings.Contains(out, strings.Repeat("L", 20)) {
		t.Fatalf("label should be truncated to leave room for the value:\n%s", out)
	}
}

const clampedWidthsContract = "widths 1, 10 and 24 all render identically: layout.NewFrame clamps every " +
	"width below theme.MinBodyWidth up to it, so they were three copies of one case. One below-clamp " +
	"representative is kept; TestFrameClampMakesNarrowWidthsIdentical fails if the clamp ever changes " +
	"and this list needs widening again."

func TestLongLabelValueVisibleAcrossWidths(t *testing.T) {
	t.Parallel()
	t.Log(clampedWidthsContract)
	for _, tc := range []struct {
		name    string
		suggest Suggester
	}{
		{name: "no suggestions"},
		{name: "with ghost", suggest: Static("VALUEVALUEVALUEEXTENDEDCOMPLETION")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, width := range []int{1, 30, 40, 81} {
				fm := NewForm(Field{
					Key:     "a",
					Label:   strings.Repeat("Label ", 12),
					Text:    "VALUEVALUEVALUE",
					Suggest: tc.suggest,
				})
				out := stripANSI(fm.Render(layout.NewFrame(width), "form"))
				if !strings.Contains(out, "▎") {
					t.Errorf("width %d hid the caret:\n%s", width, out)
				}
				if !strings.Contains(out, "V") {
					t.Errorf("width %d hid the value entirely:\n%s", width, out)
				}
			}
		})
	}
}

func TestGhostNeverEvictsTheCaret(t *testing.T) {
	t.Parallel()
	wide := NewForm(Field{
		Key:     "a",
		Label:   "LLL",
		Text:    "abc",
		Suggest: Static("abcdefghijklmnopqrstuvwxyz"),
	})
	if row := valueRow(t, wide.Render(layout.NewFrame(60), "form")); !strings.Contains(row, "abc▎defghij") {
		t.Fatalf("no ghost drawn after the caret at width 60:\n%s", row)
	}

	for _, typed := range []string{"", "abc", "abcdefgh", "abcdefghijklmnop"} {
		fm := NewForm(Field{
			Key:     "a",
			Label:   strings.Repeat("L", 30),
			Text:    typed,
			Suggest: Static("abcdefghijklmnopqrstuvwxyz"),
		})
		row := valueRow(t, fm.Render(layout.NewFrame(30), "form"))
		if !strings.Contains(row, "▎") {
			t.Errorf("a ghost evicted the caret with %d chars typed:\n%s", len(typed), row)
		}
	}
}

func valueRow(t *testing.T, rendered string) string {
	t.Helper()
	for _, ln := range strings.Split(stripANSI(rendered), "\n") {
		if strings.Contains(ln, "LLL") {
			return ln
		}
	}
	t.Fatalf("no field row in:\n%s", stripANSI(rendered))
	return ""
}

func TestShortLabelValueBudgetUnchanged(t *testing.T) {
	t.Parallel()
	fm := NewForm(Field{Key: "a", Label: "Name", Text: strings.Repeat("v", 40)})
	out := stripANSI(fm.Render(layout.NewFrame(24), "form"))
	if !strings.Contains(out, "Name") {
		t.Fatalf("a short label should not be clipped:\n%s", out)
	}
	if !strings.Contains(out, strings.Repeat("v", 14)) {
		t.Fatalf("value should fill the remaining 15 columns:\n%s", out)
	}
}

func TestUnfocusedLongLabelStillShowsValue(t *testing.T) {
	t.Parallel()
	fm := NewForm(
		Field{Key: "first", Label: "First"},
		Field{Key: "second", Label: strings.Repeat("L", 30), Text: "SECRETVALUE"},
	)
	out := stripANSI(fm.Render(layout.NewFrame(24), "form"))
	if !strings.Contains(out, "SECRET") {
		t.Fatalf("unfocused field with a long label hid its value:\n%s", out)
	}
}
