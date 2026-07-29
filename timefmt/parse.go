package timefmt

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrEmptyTime is returned by ParseWhen for blank input. Callers that want a
// default instant should check for it and substitute their own.
var ErrEmptyTime = errors.New("timefmt: empty time")

var whenLayouts = []string{time.RFC3339, "2006-01-02 15:04", "2006-01-02"}

// ParseWhen parses a user-entered instant relative to the current time.
// See ParseWhenAt for the accepted forms.
func ParseWhen(s string) (time.Time, error) {
	return ParseWhenAt(s, time.Now())
}

// ParseWhenAt parses a user-entered instant relative to now, returning UTC.
//
// It accepts a "+" followed by a Go duration ("+1h", "+30m") for an offset
// from now, or an absolute RFC3339, "2006-01-02 15:04", or "2006-01-02"
// timestamp. Absolute forms without a zone are read in the local zone.
// Blank input returns ErrEmptyTime.
//
// It is the inverse of Rel, which formats an instant back into prose.
func ParseWhenAt(s string, now time.Time) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, ErrEmptyTime
	}
	if after, ok := strings.CutPrefix(s, "+"); ok {
		d, err := time.ParseDuration(after)
		if err != nil {
			return time.Time{}, fmt.Errorf("timefmt: parse duration %q: %w", after, err)
		}
		return now.UTC().Add(d), nil
	}
	for _, layout := range whenLayouts {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("timefmt: parse %q: want RFC3339, YYYY-MM-DD, YYYY-MM-DD HH:MM, or +1h", s)
}
