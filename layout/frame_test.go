package layout

import (
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/theme"
)

var ansiRe = regexp.MustCompile("\x1b\\[[0-9;]*m")

func stripANSI(s string) string { return ansiRe.ReplaceAllString(s, "") }

func TestHeaderRendersTitleDetailAndRule(t *testing.T) {
	t.Parallel()
	out := stripANSI(DefaultFrame().Header("DASHBOARD", "live status"))
	if !strings.Contains(out, "DASHBOARD") || !strings.Contains(out, "live status") {
		t.Fatalf("header missing title or detail:\n%s", out)
	}
	if !strings.Contains(out, "─") {
		t.Fatalf("header missing rule:\n%s", out)
	}
}

func TestStackSkipsEmptySections(t *testing.T) {
	t.Parallel()
	got := Stack("one", "", "two")
	if got != "one\n\ntwo" {
		t.Fatalf("Stack = %q, want standard two-section join", got)
	}
}

func TestFrameSpreadFitsWidth(t *testing.T) {
	t.Parallel()
	f := NewFrame(24)
	out := f.Spread("a very long label that should not spill", "right")
	if got := ansi.StringWidth(out); got > f.Width {
		t.Fatalf("spread width=%d, want <= %d: %q", got, f.Width, stripANSI(out))
	}
	if !strings.Contains(stripANSI(out), "…") {
		t.Fatalf("spread did not truncate overflowing text: %q", stripANSI(out))
	}

	out = f.Spread(strings.Repeat("l", 12), strings.Repeat("r", 12))
	if got := ansi.StringWidth(out); got > f.Width {
		t.Fatalf("exact spread width=%d, want <= %d: %q", got, f.Width, stripANSI(out))
	}
}

func TestFrameHintLineWrapsToWidth(t *testing.T) {
	t.Parallel()
	out := NewFrame(24).HintLine(
		keys.Hint{Key: "enter", Label: "choose"},
		keys.Hint{Key: "pgup/pgdn", Label: "history"},
		keys.Hint{Key: "esc", Label: "back"},
	)
	if !strings.Contains(out, "\n") {
		t.Fatalf("hint line did not wrap:\n%s", stripANSI(out))
	}
	for line := range strings.SplitSeq(out, "\n") {
		if got := ansi.StringWidth(line); got > 24 {
			t.Fatalf("hint line width=%d, want <= 24: %q", got, stripANSI(line))
		}
	}
}

func TestFrameSelectableTruncatesLabel(t *testing.T) {
	t.Parallel()
	out := NewFrame(24).Selectable("a very long selectable label that should fit", true)
	if got := ansi.StringWidth(out); got > 24 {
		t.Fatalf("selectable width=%d, want <= 24: %q", got, stripANSI(out))
	}
}

func TestScreenFrameUsesMinimumSupportedWidth(t *testing.T) {
	t.Parallel()
	f := ScreenFrame(theme.MinScreenWidth)
	if f.Width != theme.MinScreenBodyWidth {
		t.Fatalf("screen frame width=%d, want %d", f.Width, theme.MinScreenBodyWidth)
	}
	if FitsScreenWidth(theme.MinScreenWidth - 1) {
		t.Fatalf("screen width %d should be rejected", theme.MinScreenWidth-1)
	}
}

func TestTooNarrowFitsCurrentScreenWidth(t *testing.T) {
	t.Parallel()
	width := theme.MinScreenWidth - 1
	out := TooNarrowIn(theme.Default(), width)
	if !strings.Contains(stripANSI(out), "80") || !strings.Contains(stripANSI(out), "79") {
		t.Fatalf("too-narrow message missing expected dimensions: %q", stripANSI(out))
	}
	for line := range strings.SplitSeq(out, "\n") {
		if got := ansi.StringWidth(line); got > width-theme.AppMarginX*2 {
			t.Fatalf("too-narrow line width=%d, want <= %d: %q", got, width-theme.AppMarginX*2, stripANSI(line))
		}
	}
}
