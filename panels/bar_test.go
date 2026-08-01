package panels

import (
	"math"
	"strings"
	"testing"

	"github.com/codyconfer/viewkit/layout"
)

func barRowCells(out, label string) (int, bool) {
	for _, l := range strings.Split(stripANSI(out), "\n") {
		if strings.Contains(l, label) {
			return strings.Count(l, "█"), true
		}
	}
	return 0, false
}

func TestBarNonFiniteValues(t *testing.T) {
	t.Parallel()
	data := []Datum{{"posinf", math.Inf(1)}, {"nan", math.NaN()}, {"neginf", math.Inf(-1)}, {"ok", 5}}
	out := Bar(layout.DefaultFrame(), "Flow", data, 20, fnum, "")
	if len(out) > 1<<16 {
		t.Fatalf("Bar with non-finite data rendered %d bytes", len(out))
	}
	for _, label := range []string{"posinf", "nan", "neginf"} {
		got, ok := barRowCells(out, label)
		if !ok {
			t.Fatalf("Bar output has no row for %q:\n%s", label, out)
		}
		if got != 0 {
			t.Errorf("Bar row %q drew %d cells, want 0", label, got)
		}
	}
	got, ok := barRowCells(out, "ok")
	if !ok {
		t.Fatalf("Bar output has no row for finite datum:\n%s", out)
	}
	if got < 1 || got > 20 {
		t.Errorf("Bar row for finite datum drew %d cells, want 1..20", got)
	}
}

func TestBarAllNonFiniteValues(t *testing.T) {
	t.Parallel()
	data := []Datum{{"a", math.NaN()}, {"b", math.Inf(1)}}
	out := Bar(layout.DefaultFrame(), "Flow", data, 20, fnum, "")
	if len(out) > 1<<16 {
		t.Fatalf("Bar with all non-finite data rendered %d bytes", len(out))
	}
	if got := strings.Count(stripANSI(out), "█"); got != 0 {
		t.Errorf("Bar with all non-finite data drew %d cells, want 0", got)
	}
}

func TestBarScrollNonFiniteValues(t *testing.T) {
	t.Parallel()
	data := []Datum{{"posinf", math.Inf(1)}, {"ok", 5}}
	out := BarScroll(layout.DefaultFrame(), "Flow", data, 20, fnum, "", 2, 0)
	if len(out) > 1<<16 {
		t.Fatalf("BarScroll with non-finite data rendered %d bytes", len(out))
	}
	if got, _ := barRowCells(out, "posinf"); got != 0 {
		t.Errorf("BarScroll row posinf drew %d cells, want 0", got)
	}
}

func TestBarEmpty(t *testing.T) {
	t.Parallel()
	out := Bar(layout.DefaultFrame(), "Flow", nil, 20, fnum, "no data")
	if !strings.Contains(out, "no data") {
		t.Fatalf("empty bar missing placeholder:\n%s", out)
	}
}

func TestBarShowsLabelsAndValues(t *testing.T) {
	t.Parallel()
	data := []Datum{{"laying", 10}, {"selling", 4}, {"deficit", -2}}
	out := Bar(layout.DefaultFrame(), "Flow", data, 20, fnum, "")
	for _, want := range []string{"laying", "selling", "deficit", "10", "4", "-2", "█"} {
		if !strings.Contains(out, want) {
			t.Errorf("bar output missing %q:\n%s", want, out)
		}
	}
}

func TestBarScalesToLargestMagnitude(t *testing.T) {
	t.Parallel()
	data := []Datum{{"big", 100}, {"small", 10}}
	lines := strings.Split(stripANSI(Bar(layout.DefaultFrame(), "F", data, 20, fnum, "")), "\n")
	var big, small int
	for _, l := range lines {
		switch {
		case strings.Contains(l, "big"):
			big = strings.Count(l, "█")
		case strings.Contains(l, "small"):
			small = strings.Count(l, "█")
		}
	}
	if big != 20 {
		t.Errorf("largest bar = %d cells, want full width 20", big)
	}
	if small >= big {
		t.Errorf("smaller value drew %d cells, not fewer than %d", small, big)
	}
}
