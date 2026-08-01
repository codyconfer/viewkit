package timefmt

import (
	"fmt"
	"time"
)

// Rel formats t relative to now ("3h ago", "in 2d").
func Rel(t time.Time) string {
	return RelAt(t, time.Now())
}

// RelAt formats t relative to now; now is explicit for tests.
func RelAt(t, now time.Time) string {
	d := now.Sub(t)
	future := d < 0
	if future {
		d = -d
	}
	if d < time.Minute {
		return "just now"
	}
	var s string
	switch {
	case d < time.Hour:
		s = fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		s = fmt.Sprintf("%dh", int(d.Hours()))
	default:
		s = fmt.Sprintf("%dd", int(d.Hours()/24))
	}
	if future {
		return "in " + s
	}
	return s + " ago"
}
