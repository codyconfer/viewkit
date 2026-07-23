package timefmt

import (
	"testing"
	"time"
)

func TestRelAt(t *testing.T) {
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name string
		at   time.Time
		want string
	}{
		{"just now", now.Add(-30 * time.Second), "just now"},
		{"minutes", now.Add(-5 * time.Minute), "5m ago"},
		{"hours", now.Add(-3 * time.Hour), "3h ago"},
		{"days", now.Add(-49 * time.Hour), "2d ago"},
		{"future minutes", now.Add(45 * time.Minute), "in 45m"},
		{"future days", now.Add(72 * time.Hour), "in 3d"},
		{"exact minute", now.Add(-time.Minute), "1m ago"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := RelAt(c.at, now); got != c.want {
				t.Errorf("RelAt = %q, want %q", got, c.want)
			}
		})
	}
}
