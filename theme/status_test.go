package theme

import (
	"testing"

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
