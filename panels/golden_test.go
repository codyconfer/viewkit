package panels

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/theme"
)

// goldenPalette pins every colour the digest below depends on, so the golden is
// a statement about panels/ alone: a theme palette edit cannot move it, while a
// change in how a panel renders can.
var goldenPalette = theme.Palette{
	Accent:   lipgloss.Color("#101010"),
	Border:   lipgloss.Color("#202020"),
	Muted:    lipgloss.Color("#303030"),
	Text:     lipgloss.Color("#404040"),
	Selected: lipgloss.Color("#505050"),
	Success:  lipgloss.Color("#606060"),
	Warning:  lipgloss.Color("#707070"),
	Failure:  lipgloss.Color("#808080"),
	Info:     lipgloss.Color("#909090"),
	Series2:  lipgloss.Color("#a0a0a0"),
	Series3:  lipgloss.Color("#b0b0b0"),
	Bg:       lipgloss.Color("#c0c0c0"),
}

// goldenPanelCases is a deterministic matrix of finite-input renders across the
// numeric panels the NaN/Inf hardening touched.
func goldenPanelCases() []struct{ name, out string } {
	datasets := map[string][]Datum{
		"empty":    {},
		"single":   {{"only", 1}},
		"pair":     {{"a", 3}, {"b", 1}},
		"zeros":    {{"a", 0}, {"b", 0}},
		"signed":   {{"pos", 10}, {"neg", -4}, {"zero", 0}},
		"scale":    {{"tiny", 0.0001}, {"huge", 1e12}},
		"wide":     {{"日本語", 5}, {"a much longer label", 2}},
		"maxes":    {{"max", math.MaxFloat64}, {"one", 1}},
		"overflow": {{"max1", math.MaxFloat64}, {"max2", math.MaxFloat64}},
	}
	dataOrder := []string{"empty", "single", "pair", "zeros", "signed", "scale", "wide", "maxes", "overflow"}
	levelSets := map[string][]float64{
		"none":    nil,
		"silent":  {0},
		"full":    {1},
		"ramp":    {0.25, 0.5, 0.75, 1},
		"edges":   {0.0001, 0.9999},
		"clamped": {-0.5, 1.5},
	}
	levelOrder := []string{"none", "silent", "full", "ramp", "edges", "clamped"}

	var cases []struct{ name, out string }
	add := func(name, out string) {
		cases = append(cases, struct{ name, out string }{name, out})
	}
	for _, width := range []int{24, 40, 81} {
		f := layout.NewFrame(width)
		for _, dk := range dataOrder {
			data := datasets[dk]
			for _, bw := range []int{1, 5, 20} {
				add(fmt.Sprintf("bar/%d/%s/%d", width, dk, bw), Bar(f, "T", data, bw, fnum, "none"))
				add(fmt.Sprintf("pie/%d/%s/%d", width, dk, bw), Pie(f, "T", data, bw, fnum, "none"))
			}
			add(fmt.Sprintf("barscroll/%d/%s", width, dk), BarScroll(f, "T", data, 10, fnum, "none", 2, 1))
		}
		for _, lk := range levelOrder {
			for _, height := range []int{1, 4, 6} {
				add(fmt.Sprintf("spectrum/%d/%s/%d", width, lk, height),
					Spectrum(f, "T", levelSets[lk], height, "none"))
			}
		}
	}
	return cases
}

// TestPanelRenderingIsByteStable is the byte-identity guard the NaN/Inf hardening
// wave claimed but never committed. It hashes every case verbatim — escapes
// included — so any unintended rendering churn in the numeric panels shows up
// here. A deliberate change means re-recording goldenPanelDigest and saying so.
func TestPanelRenderingIsByteStable(t *testing.T) {
	const (
		goldenPanelCount  = 243
		goldenPanelDigest = "b9bd40a7dd486c0c16814592e3b5837865ceab25a15102a280f5c64a79c6ede7"
	)

	prevProfile := lipgloss.ColorProfile()
	prevTheme := *theme.Cur()
	t.Cleanup(func() {
		theme.Use(prevTheme)
		lipgloss.SetColorProfile(prevProfile)
	})
	lipgloss.SetColorProfile(termenv.TrueColor)
	theme.Use(theme.New(goldenPalette))

	cases := goldenPanelCases()
	if len(cases) != goldenPanelCount {
		t.Fatalf("case matrix has %d cases, want %d — update goldenPanelCount and the digest together",
			len(cases), goldenPanelCount)
	}
	h := sha256.New()
	for _, c := range cases {
		fmt.Fprintf(h, "%s\x00%s\x00", c.name, c.out)
	}
	if got := hex.EncodeToString(h.Sum(nil)); got != goldenPanelDigest {
		t.Errorf("panel rendering digest = %s, want %s (%d cases)", got, goldenPanelDigest, len(cases))
	}
}
