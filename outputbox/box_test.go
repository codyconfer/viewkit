package outputbox

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestBoxStatusColorAndWidth(t *testing.T) {
	backgrounds := map[Status]string{Pending: "236", Success: "22", Warning: "58", Failure: "52"}
	for _, status := range []Status{Pending, Success, Warning, Failure} {
		for _, color := range []bool{false, true} {
			text := Render("gcx login ops", []string{strings.Repeat("界", 40), "complete"}, 40, color, status)
			if !strings.Contains(ansi.Strip(text), "["+string(status)+"]") || strings.Count(ansi.Strip(text), "界") != 40 {
				t.Fatal("missing status or truncated output", text)
			}
			if strings.ContainsAny(ansi.Strip(text), "╭╰│") || strings.Contains(text, "\x1b[J") {
				t.Fatal("unexpected chrome/cursor movement")
			}
			if strings.Contains(text, "\x1b[") != color {
				t.Fatal("wrong color mode")
			}
			for _, line := range strings.Split(text, "\n") {
				if ansi.StringWidth(line) != 40 {
					t.Fatalf("expected full-width shading: %q", line)
				}
			}
			if color && !strings.Contains(text, "48;5;"+backgrounds[status]+"m") {
				t.Fatalf("missing %s status background: %q", status, text)
			}
		}
	}
}
