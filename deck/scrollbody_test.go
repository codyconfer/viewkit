package deck

import (
	"fmt"
	"strings"
	"testing"

	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/layout"
)

func scrollBodyLines(n int) string {
	lines := make([]string, n)
	for i := range lines {
		lines[i] = fmt.Sprintf("line-%03d", i)
	}
	return strings.Join(lines, "\n")
}

func TestScrollBodyHandlesNavAndWindows(t *testing.T) {
	var s ScrollBody
	body := scrollBodyLines(50)

	if got := s.View(layout.Frame{Width: 80}, body, 10); !strings.Contains(got, "line-000") {
		t.Fatalf("initial window missing top line:\n%s", got)
	}
	if s.Handle(keys.Confirm) {
		t.Fatal("Handle claimed a non-scroll action")
	}
	if !s.Handle(keys.Down) {
		t.Fatal("Handle did not claim Down")
	}
	if s.Offset != 1 {
		t.Fatalf("offset after Down = %d, want 1", s.Offset)
	}
	if !s.Handle(keys.PageDown) {
		t.Fatal("Handle did not claim PageDown")
	}
	if got := s.View(layout.Frame{Width: 80}, body, 10); strings.Contains(got, "line-000") {
		t.Fatalf("window did not advance after PageDown:\n%s", got)
	}
	for range 20 {
		s.Handle(keys.PageDown)
	}
	if got := s.View(layout.Frame{Width: 80}, body, 10); !strings.Contains(got, "line-049") {
		t.Fatalf("clamped window missing last line:\n%s", got)
	}
	for range 30 {
		s.Handle(keys.PageUp)
	}
	if s.Offset != 0 {
		t.Fatalf("offset after paging back = %d, want 0", s.Offset)
	}
}

func TestScrollBodyShortContentIsUntouched(t *testing.T) {
	var s ScrollBody
	body := scrollBodyLines(3)
	if got := s.View(layout.Frame{Width: 80}, body, 10); got != body {
		t.Fatalf("short content was windowed:\n%s", got)
	}
	if s.Handle(keys.Down); s.Offset != 0 {
		t.Fatalf("offset moved on short content: %d", s.Offset)
	}
}
