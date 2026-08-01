package theme

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// uncachedStripText / uncachedStripBold are the pre-cache implementations. The
// cached helpers must stay byte-identical to them: a cache that changes
// rendering is a regression.
func uncachedStripText(t Theme, fg lipgloss.TerminalColor, s string) string {
	return lipgloss.NewStyle().Background(stripBgOf(t)).Foreground(fg).Render(s)
}

func uncachedStripBold(t Theme, fg lipgloss.TerminalColor, s string) string {
	return lipgloss.NewStyle().Background(stripBgOf(t)).Foreground(fg).Bold(true).Render(s)
}

func TestStripHelpersMatchUncachedRendering(t *testing.T) {
	t.Parallel()
	samples := []string{"", " ", "munin · audit", "12:34:56 ", "◈ deck"}

	for _, key := range Keys() {
		th, ok := Named(key)
		if !ok {
			t.Fatalf("Named(%q) not found", key)
		}

		for _, fg := range []lipgloss.TerminalColor{
			th.Val.GetForeground(),
			th.Accent.GetForeground(),
			th.Dim.GetForeground(),
			lipgloss.NoColor{},
		} {
			for _, s := range samples {
				if got, want := th.StripText(fg, s), uncachedStripText(th, fg, s); got != want {
					t.Errorf("%s: StripText(%v, %q) = %q, want %q", key, fg, s, got, want)
				}
				if got, want := th.StripBold(fg, s), uncachedStripBold(th, fg, s); got != want {
					t.Errorf("%s: StripBold(%v, %q) = %q, want %q", key, fg, s, got, want)
				}
			}
		}

		if got, want := th.StripBg(), stripBgOf(th); got != want {
			t.Errorf("%s: StripBg() = %v, want %v", key, got, want)
		}
	}
}

// The cached styles must not freeze the color profile: switching profiles
// after New has to keep affecting output.
func TestCachedStripStylesFollowColorProfile(t *testing.T) {
	prev := lipgloss.ColorProfile()
	t.Cleanup(func() { lipgloss.SetColorProfile(prev) })

	th := Default()
	fg := th.Val.GetForeground()

	lipgloss.SetColorProfile(termenv.Ascii)
	ascii := th.StripText(fg, "probe")
	lipgloss.SetColorProfile(termenv.TrueColor)
	truecolor := th.StripText(fg, "probe")

	if ascii == truecolor {
		t.Fatalf("cached style ignored the profile switch: %q == %q", ascii, truecolor)
	}
	if want := "probe"; ascii != want {
		t.Errorf("Ascii profile output = %q, want %q", ascii, want)
	}
}

// The strip cache lives on the Theme value: two themes must not share a strip
// background.
func TestStripCacheIsPerTheme(t *testing.T) {
	t.Parallel()
	dark, _ := Named("solarized-dark")
	light, _ := Named("solarized-light")

	if dark.StripBg() == light.StripBg() {
		t.Fatalf("StripBg() identical across themes: %v", dark.StripBg())
	}
	if got, want := light.StripBg(), stripBgOf(light); got != want {
		t.Fatalf("light.StripBg() = %v, want %v", got, want)
	}
}

// A Theme built as a literal (not via New) still has to compute its strip
// styles on the fly, or StripText would render with a zero background.
func TestLiteralThemeComputesStripStyles(t *testing.T) {
	t.Parallel()
	lit := Theme{
		Panel: lipgloss.NewStyle().BorderForeground(lipgloss.Color("#123456")),
		Dim:   lipgloss.NewStyle().Foreground(lipgloss.Color("#654321")),
	}

	if got, want := lit.StripBg(), stripBgOf(lit); got != want {
		t.Fatalf("StripBg() = %v, want %v", got, want)
	}
	fg := lipgloss.Color("#ffffff")
	if got, want := lit.StripText(fg, "x"), uncachedStripText(lit, fg, "x"); got != want {
		t.Fatalf("StripText = %q, want %q", got, want)
	}
}
