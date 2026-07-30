package forms

import (
	"strings"
	"testing"

	"github.com/codyconfer/viewkit/layout"
)

func TestLongLabelKeepsValueVisible(t *testing.T) {
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

func TestLongLabelValueVisibleAcrossWidths(t *testing.T) {
	for _, width := range []int{1, 10, 24, 30, 40, 81} {
		fm := NewForm(Field{Key: "a", Label: strings.Repeat("Label ", 12), Text: "VALUE"})
		out := stripANSI(fm.Render(layout.NewFrame(width), "form"))
		if !strings.Contains(out, "▎") {
			t.Errorf("width %d hid the caret:\n%s", width, out)
		}
		if !strings.Contains(out, "V") {
			t.Errorf("width %d hid the value entirely:\n%s", width, out)
		}
	}
}

func TestShortLabelValueBudgetUnchanged(t *testing.T) {
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
	fm := NewForm(
		Field{Key: "first", Label: "First"},
		Field{Key: "second", Label: strings.Repeat("L", 30), Text: "SECRETVALUE"},
	)
	out := stripANSI(fm.Render(layout.NewFrame(24), "form"))
	if !strings.Contains(out, "SECRET") {
		t.Fatalf("unfocused field with a long label hid its value:\n%s", out)
	}
}
