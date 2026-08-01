package theme

import (
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/codyconfer/viewkit/glyph"
)

func TestSeverityColorMapsByNamedConstant(t *testing.T) {
	t.Parallel()
	th := Default()
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
		if got := th.SeverityColor(tc.sev); got != tc.want {
			t.Errorf("%s: SeverityColor(%v) = %v, want %v", tc.name, tc.sev, got, tc.want)
		}
	}
	warn := th.SeverityColor(glyph.SeverityWarning)
	if len(th.Series) > 2 {
		if warn != th.Series[2].GetForeground() {
			t.Errorf("warning→Series[2]: got %v, want %v", warn, th.Series[2].GetForeground())
		}
	} else if warn != th.Cant.GetForeground() {
		t.Errorf("warning fallback→Cant: got %v, want %v", warn, th.Cant.GetForeground())
	}
}

func TestSeverityStyleIsDistinctPerSeverity(t *testing.T) {
	t.Parallel()
	th := Default()
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
			if got := th.SeverityStyle(c.sev).GetForeground(); got != c.want {
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
		fg := th.SeverityStyle(sev).GetForeground()
		if prev, dup := seen[fg]; dup {
			t.Errorf("severities %v and %v share foreground %v", prev, sev, fg)
		}
		seen[fg] = sev
	}
}

func TestSeverityStyleRendersWithSeverityColor(t *testing.T) {
	t.Parallel()
	th := Default()
	for _, sev := range []glyph.Severity{
		glyph.SeverityPositive,
		glyph.SeverityNegative,
		glyph.SeverityWarning,
		glyph.SeverityNeutral,
		glyph.Severity(42),
	} {
		want := th.SeverityStyle(sev).GetForeground()
		if got := th.SeverityColor(sev); got != want {
			t.Errorf("SeverityColor(%v) = %v, want SeverityStyle foreground %v", sev, got, want)
		}
	}
}

func TestSeverityStyleFollowsItsTheme(t *testing.T) {
	t.Parallel()
	th, ok := Named("solarized-dark")
	if !ok {
		t.Fatal("Named(solarized-dark) not found")
	}

	if got, want := th.SeverityStyle(glyph.SeverityPositive).GetForeground(), th.Can.GetForeground(); got != want {
		t.Errorf("positive = %v, want that theme's Can %v", got, want)
	}
	if got, want := th.SeverityColor(glyph.SeverityWarning), th.Series[2].GetForeground(); got != want {
		t.Errorf("warning = %v, want that theme's Series[2] %v", got, want)
	}
}

func TestSeverityStyleWarningFallsBackToCantWithShortSeries(t *testing.T) {
	t.Parallel()
	orig := Default()

	short := orig
	short.Series = append([]lipgloss.Style(nil), orig.Series[:2]...)

	if got, want := short.SeverityStyle(glyph.SeverityWarning).GetForeground(), short.Cant.GetForeground(); got != want {
		t.Errorf("warning fallback = %v, want Cant %v", got, want)
	}
	if got, want := short.SeverityColor(glyph.SeverityWarning), short.Cant.GetForeground(); got != want {
		t.Errorf("SeverityColor warning fallback = %v, want Cant %v", got, want)
	}

	empty := orig
	empty.Series = nil
	if got, want := empty.SeverityStyle(glyph.SeverityWarning).GetForeground(), empty.Cant.GetForeground(); got != want {
		t.Errorf("warning with no series = %v, want Cant %v", got, want)
	}
}
