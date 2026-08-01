package layout

// Focusable names a candidate for the focus ring; only entries with
// Interactive set actually join it.
type Focusable struct {
	Name        string
	Interactive bool
}

// Ring is the ordered list of interactive pane names focus cycles through.
// Callers keep an integer index and use At/Step; both tolerate out-of-range
// indexes by clamping.
type Ring []string

// NewRing builds a Ring from the interactive entries, preserving order and
// skipping non-interactive ones.
func NewRing(all ...Focusable) Ring {
	out := make(Ring, 0, len(all))
	for _, f := range all {
		if f.Interactive {
			out = append(out, f.Name)
		}
	}
	return out
}

// At returns the name at idx, clamping out-of-range indexes to the nearest
// end; an empty ring returns "".
func (r Ring) At(idx int) string {
	if len(r) == 0 {
		return ""
	}
	return r[r.clamp(idx)]
}

// Step returns the index delta positions from idx, wrapping around both ends
// of the ring (delta may be negative). An empty ring returns 0.
func (r Ring) Step(idx, delta int) int {
	n := len(r)
	if n == 0 {
		return 0
	}
	return ((r.clamp(idx)+delta)%n + n) % n
}

func (r Ring) clamp(idx int) int {
	if idx < 0 {
		return 0
	}
	if idx >= len(r) {
		return len(r) - 1
	}
	return idx
}
