package theme

import (
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/codyconfer/viewkit/glyph"
)

func TestSeverityColorMapsByNamedConstant(t *testing.T) {
	th := Cur()
	cases := []struct {
		sev  glyph.Severity
		want any
		name string
	}{
		{glyph.SeverityPositive, th.Can.GetForeground(), "positive→Can"},
		{glyph.SeverityNegative, th.Cant.GetForeground(), "negative→Cant"},
		{glyph.SeverityNeutral, th.Dim.GetForeground(), "neutral→Dim"},
	}
	for _, tc := range cases {
		if got := SeverityColor(tc.sev); got != tc.want {
			t.Errorf("%s: SeverityColor(%v) = %v, want %v", tc.name, tc.sev, got, tc.want)
		}
	}
	warn := SeverityColor(glyph.SeverityWarning)
	if len(th.Series) > 2 {
		if warn != th.Series[2].GetForeground() {
			t.Errorf("warning→Series[2]: got %v, want %v", warn, th.Series[2].GetForeground())
		}
	} else if warn != th.Cant.GetForeground() {
		t.Errorf("warning fallback→Cant: got %v, want %v", warn, th.Cant.GetForeground())
	}
}

func TestSeverityGlyphMapsByNamedConstant(t *testing.T) {
	cases := []struct {
		sev  glyph.Severity
		want string
	}{
		{glyph.SeverityPositive, glyph.StatusOK()},
		{glyph.SeverityWarning, glyph.StatusWarn()},
		{glyph.SeverityNegative, glyph.StatusBad()},
		{glyph.SeverityNeutral, glyph.StatusMuted()},
		{glyph.Severity(99), glyph.StatusMuted()},
	}
	for _, tc := range cases {
		if got := SeverityGlyph(tc.sev); got != tc.want {
			t.Errorf("SeverityGlyph(%v) = %q, want %q", tc.sev, got, tc.want)
		}
	}
}

func TestSeverityStyleIsDistinctPerSeverity(t *testing.T) {
	th := Cur()
	cases := []struct {
		name string
		sev  glyph.Severity
		want lipgloss.TerminalColor
	}{
		{"positive", glyph.SeverityPositive, th.Can.GetForeground()},
		{"negative", glyph.SeverityNegative, th.Cant.GetForeground()},
		{"warning", glyph.SeverityWarning, th.Series[2].GetForeground()},
		{"neutral", glyph.SeverityNeutral, th.Dim.GetForeground()},
		{"unknown", glyph.Severity(99), th.Dim.GetForeground()},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := SeverityStyle(c.sev).GetForeground(); got != c.want {
				t.Errorf("SeverityStyle(%v) foreground = %v, want %v", c.sev, got, c.want)
			}
		})
	}

	seen := map[lipgloss.TerminalColor]glyph.Severity{}
	for _, sev := range []glyph.Severity{
		glyph.SeverityPositive,
		glyph.SeverityNegative,
		glyph.SeverityWarning,
		glyph.SeverityNeutral,
	} {
		fg := SeverityStyle(sev).GetForeground()
		if prev, dup := seen[fg]; dup {
			t.Errorf("severities %v and %v share foreground %v", prev, sev, fg)
		}
		seen[fg] = sev
	}
}

func TestSeverityStyleRendersWithSeverityColor(t *testing.T) {
	for _, sev := range []glyph.Severity{
		glyph.SeverityPositive,
		glyph.SeverityNegative,
		glyph.SeverityWarning,
		glyph.SeverityNeutral,
		glyph.Severity(42),
	} {
		want := SeverityStyle(sev).GetForeground()
		if got := SeverityColor(sev); got != want {
			t.Errorf("SeverityColor(%v) = %v, want SeverityStyle foreground %v", sev, got, want)
		}
	}
}

func TestSeverityStyleFollowsActiveTheme(t *testing.T) {
	orig := *Cur()
	defer Use(orig)

	th, ok := Named("solarized-dark")
	if !ok {
		t.Fatal("Named(solarized-dark) not found")
	}
	Use(th)

	if got, want := SeverityStyle(glyph.SeverityPositive).GetForeground(), th.Can.GetForeground(); got != want {
		t.Errorf("positive = %v, want %v after Use", got, want)
	}
	if got, want := SeverityColor(glyph.SeverityWarning), th.Series[2].GetForeground(); got != want {
		t.Errorf("warning = %v, want %v after Use", got, want)
	}
}

func TestSeverityStyleWarningFallsBackToCantWithShortSeries(t *testing.T) {
	orig := *Cur()
	defer Use(orig)

	short := orig
	short.Series = append([]lipgloss.Style(nil), orig.Series[:2]...)
	Use(short)

	if got, want := SeverityStyle(glyph.SeverityWarning).GetForeground(), short.Cant.GetForeground(); got != want {
		t.Errorf("warning fallback = %v, want Cant %v", got, want)
	}
	if got, want := SeverityColor(glyph.SeverityWarning), short.Cant.GetForeground(); got != want {
		t.Errorf("SeverityColor warning fallback = %v, want Cant %v", got, want)
	}

	empty := orig
	empty.Series = nil
	Use(empty)
	if got, want := SeverityStyle(glyph.SeverityWarning).GetForeground(), empty.Cant.GetForeground(); got != want {
		t.Errorf("warning with no series = %v, want Cant %v", got, want)
	}
}
