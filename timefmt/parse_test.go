package timefmt

import (
	"errors"
	"strings"
	"testing"
	"time"
)

var parseNow = time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)

func TestParseWhenAtBlankInput(t *testing.T) {
	for _, in := range []string{"", " ", "\t", "\n  \t "} {
		got, err := ParseWhenAt(in, parseNow)
		if !errors.Is(err, ErrEmptyTime) {
			t.Errorf("ParseWhenAt(%q) err = %v, want ErrEmptyTime", in, err)
		}
		if !got.IsZero() {
			t.Errorf("ParseWhenAt(%q) = %v, want zero time", in, got)
		}
	}
}

func TestParseWhenAtOffsets(t *testing.T) {
	cases := []struct {
		in   string
		want time.Time
	}{
		{"+1h", parseNow.Add(time.Hour)},
		{"+30m", parseNow.Add(30 * time.Minute)},
		{"+90s", parseNow.Add(90 * time.Second)},
		{"+1h30m", parseNow.Add(90 * time.Minute)},
		{"+0s", parseNow},
		{"+-2h", parseNow.Add(-2 * time.Hour)},
		{"  +1h  ", parseNow.Add(time.Hour)},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got, err := ParseWhenAt(c.in, parseNow)
			if err != nil {
				t.Fatalf("ParseWhenAt(%q) err = %v", c.in, err)
			}
			if !got.Equal(c.want) {
				t.Errorf("ParseWhenAt(%q) = %v, want %v", c.in, got, c.want)
			}
			if got.Location() != time.UTC {
				t.Errorf("ParseWhenAt(%q) location = %v, want UTC", c.in, got.Location())
			}
		})
	}
}

func TestParseWhenAtOffsetIsRelativeToTheGivenNow(t *testing.T) {
	other := time.Date(1999, 12, 31, 23, 0, 0, 0, time.FixedZone("odd", 5*3600))
	got, err := ParseWhenAt("+1h", other)
	if err != nil {
		t.Fatalf("ParseWhenAt err = %v", err)
	}
	if want := other.Add(time.Hour); !got.Equal(want) {
		t.Errorf("ParseWhenAt(+1h) = %v, want %v", got, want)
	}
	if got.Location() != time.UTC {
		t.Errorf("location = %v, want UTC", got.Location())
	}
}

func TestParseWhenAtBadOffset(t *testing.T) {
	for _, in := range []string{"+bogus", "+", "+1x", "+ 1h", "+1hour"} {
		got, err := ParseWhenAt(in, parseNow)
		if err == nil {
			t.Errorf("ParseWhenAt(%q) = %v, want an error", in, got)
			continue
		}
		if errors.Is(err, ErrEmptyTime) {
			t.Errorf("ParseWhenAt(%q) err = %v, want a duration error not ErrEmptyTime", in, err)
		}
		if !strings.Contains(err.Error(), "parse duration") {
			t.Errorf("ParseWhenAt(%q) err = %v, want it to mention the duration", in, err)
		}
		if !got.IsZero() {
			t.Errorf("ParseWhenAt(%q) = %v, want zero time", in, got)
		}
	}
}

func TestParseWhenAtRFC3339(t *testing.T) {
	cases := []struct {
		in   string
		want time.Time
	}{
		{"2026-07-22T12:00:00Z", time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)},
		{"2026-07-22T12:00:00+02:00", time.Date(2026, 7, 22, 10, 0, 0, 0, time.UTC)},
		{"2026-07-22T12:00:00-05:00", time.Date(2026, 7, 22, 17, 0, 0, 0, time.UTC)},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got, err := ParseWhenAt(c.in, parseNow)
			if err != nil {
				t.Fatalf("ParseWhenAt(%q) err = %v", c.in, err)
			}
			if !got.Equal(c.want) {
				t.Errorf("ParseWhenAt(%q) = %v, want %v", c.in, got, c.want)
			}
			if got.Location() != time.UTC {
				t.Errorf("location = %v, want UTC", got.Location())
			}
		})
	}
}

func TestParseWhenAtAbsoluteFormsUseTheLocalZone(t *testing.T) {
	cases := []struct {
		name   string
		in     string
		layout string
	}{
		{"date and time", "2026-07-22 15:04", "2006-01-02 15:04"},
		{"date only", "2026-07-22", "2006-01-02"},
		{"padded date", "  2026-07-22  ", "2006-01-02"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			want, err := time.ParseInLocation(c.layout, strings.TrimSpace(c.in), time.Local)
			if err != nil {
				t.Fatalf("fixture parse: %v", err)
			}
			got, err := ParseWhenAt(c.in, parseNow)
			if err != nil {
				t.Fatalf("ParseWhenAt(%q) err = %v", c.in, err)
			}
			if !got.Equal(want.UTC()) {
				t.Errorf("ParseWhenAt(%q) = %v, want %v", c.in, got, want.UTC())
			}
			if got.Location() != time.UTC {
				t.Errorf("location = %v, want UTC", got.Location())
			}
			if y, m, d := want.Date(); got.In(time.Local).Year() != y || got.In(time.Local).Month() != m || got.In(time.Local).Day() != d {
				t.Errorf("local calendar date = %v, want %04d-%02d-%02d", got.In(time.Local), y, m, d)
			}
		})
	}
}

func TestParseWhenAtGarbage(t *testing.T) {
	for _, in := range []string{"next tuesday", "22/07/2026", "2026-13-45", "-1h", "1h", "12:00", "2026-07-22T12:00:00"} {
		got, err := ParseWhenAt(in, parseNow)
		if err == nil {
			t.Errorf("ParseWhenAt(%q) = %v, want an error", in, got)
			continue
		}
		if !got.IsZero() {
			t.Errorf("ParseWhenAt(%q) = %v, want zero time", in, got)
		}
		msg := err.Error()
		for _, want := range []string{in, "RFC3339", "YYYY-MM-DD", "YYYY-MM-DD HH:MM", "+1h"} {
			if !strings.Contains(msg, want) {
				t.Errorf("ParseWhenAt(%q) err = %q, want it to mention %q", in, msg, want)
			}
		}
	}
}

func TestParseWhenAtIsDeterministic(t *testing.T) {
	for _, in := range []string{"+2h", "2026-07-22", "2026-07-22 15:04", "2026-07-22T12:00:00Z"} {
		first, err1 := ParseWhenAt(in, parseNow)
		second, err2 := ParseWhenAt(in, parseNow)
		if err1 != nil || err2 != nil {
			t.Fatalf("ParseWhenAt(%q) errs = %v, %v", in, err1, err2)
		}
		if !first.Equal(second) {
			t.Errorf("ParseWhenAt(%q) = %v then %v, want the same instant", in, first, second)
		}
	}
}

func TestParseWhenDelegatesToNow(t *testing.T) {
	before := time.Now()
	got, err := ParseWhen("+1h")
	after := time.Now()
	if err != nil {
		t.Fatalf("ParseWhen(+1h) err = %v", err)
	}
	if got.Before(before.Add(time.Hour)) || got.After(after.Add(time.Hour)) {
		t.Errorf("ParseWhen(+1h) = %v, want between %v and %v", got, before.Add(time.Hour), after.Add(time.Hour))
	}
	if got.Location() != time.UTC {
		t.Errorf("location = %v, want UTC", got.Location())
	}
}

func TestParseWhenMatchesParseWhenAtForAbsoluteInput(t *testing.T) {
	const in = "2026-07-22T12:00:00Z"
	got, err := ParseWhen(in)
	if err != nil {
		t.Fatalf("ParseWhen err = %v", err)
	}
	want, err := ParseWhenAt(in, parseNow)
	if err != nil {
		t.Fatalf("ParseWhenAt err = %v", err)
	}
	if !got.Equal(want) {
		t.Errorf("ParseWhen = %v, ParseWhenAt = %v, want the same instant", got, want)
	}
}

func TestParseWhenBlankInput(t *testing.T) {
	if _, err := ParseWhen("   "); !errors.Is(err, ErrEmptyTime) {
		t.Errorf("ParseWhen(blank) err = %v, want ErrEmptyTime", err)
	}
}

func TestErrEmptyTimeMessage(t *testing.T) {
	if got := ErrEmptyTime.Error(); got != "timefmt: empty time" {
		t.Errorf("ErrEmptyTime = %q", got)
	}
}
