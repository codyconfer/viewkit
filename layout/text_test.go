package layout

import (
	"strings"
	"testing"
)

func TestLines(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{"two lines", "a\nb", []string{"a", "b"}},
		{"single trailing newline dropped", "a\nb\n", []string{"a", "b"}},
		{"multiple trailing newlines dropped", "a\nb\n\n\n", []string{"a", "b"}},
		{"empty string", "", []string{""}},
		{"only a newline", "\n", []string{""}},
		{"interior blank preserved", "a\n\nb", []string{"a", "", "b"}},
		{"leading blank preserved", "\na", []string{"", "a"}},
		{"single line", "solo", []string{"solo"}},
		{"whitespace not trimmed", "  a  \n b ", []string{"  a  ", " b "}},
		{"carriage returns kept", "a\r\nb", []string{"a\r", "b"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Lines(c.in)
			if len(got) != len(c.want) || strings.Join(got, "|") != strings.Join(c.want, "|") {
				t.Errorf("Lines(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestLinesNeverReturnsEmptySlice(t *testing.T) {
	for _, in := range []string{"", "\n", "\n\n\n"} {
		if got := Lines(in); len(got) != 1 || got[0] != "" {
			t.Errorf("Lines(%q) = %q, want one empty element", in, got)
		}
	}
}

func TestFirstLine(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"first of many", "alpha\nbeta", "alpha"},
		{"trims the returned line", "   alpha   \nbeta", "alpha"},
		{"skips leading blank lines", "\n\nalpha\nbeta", "alpha"},
		{"skips whitespace-only lines", "  \n\t\n  alpha  ", "alpha"},
		{"empty string", "", ""},
		{"only newlines", "\n\n\n", ""},
		{"all whitespace", "   \n\t\t\n  ", ""},
		{"trailing newline ignored", "alpha\n", "alpha"},
		{"single line", "  solo  ", "solo"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := FirstLine(c.in); got != c.want {
				t.Errorf("FirstLine(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestDialogWidthInsetsUntilItHitsTheCap(t *testing.T) {
	cases := []struct {
		name   string
		screen int
		want   int
	}{
		{"narrow", 20, 12},
		{"medium", 40, 32},
		{"just under the cap", 63, 55},
		{"exactly the cap", 64, DialogMaxWidth},
		{"wide", 120, DialogMaxWidth},
		{"very wide", 4000, DialogMaxWidth},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := DialogWidth(c.screen); got != c.want {
				t.Errorf("DialogWidth(%d) = %d, want %d", c.screen, got, c.want)
			}
		})
	}
}

func TestDialogWidthNeverGoesBelowOne(t *testing.T) {
	for _, screen := range []int{9, 8, 5, 1, 0, -1, -100} {
		if got := DialogWidth(screen); got < 1 {
			t.Errorf("DialogWidth(%d) = %d, want >= 1", screen, got)
		}
	}
	if got := DialogWidth(9); got != 1 {
		t.Errorf("DialogWidth(9) = %d, want 1", got)
	}
	if got := DialogWidth(10); got != 2 {
		t.Errorf("DialogWidth(10) = %d, want 2", got)
	}
}

func TestDialogWidthIsMonotonic(t *testing.T) {
	prev := DialogWidth(0)
	for screen := 1; screen <= 200; screen++ {
		got := DialogWidth(screen)
		if got < prev {
			t.Fatalf("DialogWidth(%d) = %d, dropped below DialogWidth(%d) = %d", screen, got, screen-1, prev)
		}
		if got > DialogMaxWidth {
			t.Fatalf("DialogWidth(%d) = %d, exceeds DialogMaxWidth %d", screen, got, DialogMaxWidth)
		}
		prev = got
	}
}

func TestDialogMaxWidthValue(t *testing.T) {
	if DialogMaxWidth != 56 {
		t.Errorf("DialogMaxWidth = %d, want 56", DialogMaxWidth)
	}
}
