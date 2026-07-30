package theme

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestFillLineClampsNegativeWidth(t *testing.T) {
	c := Cur().Dim.GetForeground()
	for _, width := range []int{-8, -1, 0, 1} {
		got := FillLine(c, width)
		want := max(width, 0)
		if w := ansi.StringWidth(got); w != want {
			t.Errorf("FillLine(width=%d) width = %d, want %d", width, w, want)
		}
	}
}

func TestPadBlockClampsNegativeWidthAndRows(t *testing.T) {
	c := Cur().Dim.GetForeground()
	for _, width := range []int{-8, -1, 0, 1} {
		for _, rows := range []int{-3, -1, 0, 1} {
			got := PadBlock(c, width, rows, "body")
			wantLines := 1 + 2*max(rows, 0)
			if n := strings.Count(got, "\n") + 1; n != wantLines {
				t.Errorf("PadBlock(width=%d rows=%d) = %d lines, want %d", width, rows, n, wantLines)
			}
		}
	}
}

func TestStripBlockSurvivesNarrowScreen(t *testing.T) {
	for _, screenWidth := range []int{1, 2, 8, 9, 80} {
		got := StripBlock(screenWidth-8, "body")
		if got == "" && screenWidth > 8 {
			t.Errorf("StripBlock(%d) returned empty", screenWidth-8)
		}
	}
}
