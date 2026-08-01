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
	orig := Cur()
	t.Cleanup(func() { Use(orig) })

	samples := []string{"", " ", "munin · audit", "12:34:56 ", "◈ deck"}

	for _, key := range Keys() {
		th, ok := Named(key)
		if !ok {
			t.Fatalf("Named(%q) not found", key)
		}
		Use(th)
		cur := Cur()

		for _, fg := range []lipgloss.TerminalColor{
			cur.Val.GetForeground(),
			cur.Accent.GetForeground(),
			cur.Dim.GetForeground(),
			lipgloss.NoColor{},
		} {
			for _, s := range samples {
				if got, want := StripText(fg, s), uncachedStripText(th, fg, s); got != want {
					t.Errorf("%s: StripText(%v, %q) = %q, want %q", key, fg, s, got, want)
				}
				if got, want := StripBold(fg, s), uncachedStripBold(th, fg, s); got != want {
					t.Errorf("%s: StripBold(%v, %q) = %q, want %q", key, fg, s, got, want)
				}
			}
		}

		if got, want := StripBg(), stripBgOf(th); got != want {
			t.Errorf("%s: StripBg() = %v, want %v", key, got, want)
		}
	}
}

// The cached styles must not freeze the color profile: switching profiles after
// Use has to keep affecting output.
func TestCachedStripStylesFollowColorProfile(t *testing.T) {
	prev := lipgloss.ColorProfile()
	t.Cleanup(func() { lipgloss.SetColorProfile(prev) })

	fg := Cur().Val.GetForeground()

	lipgloss.SetColorProfile(termenv.Ascii)
	ascii := StripText(fg, "probe")
	lipgloss.SetColorProfile(termenv.TrueColor)
	truecolor := StripText(fg, "probe")

	if ascii == truecolor {
		t.Fatalf("cached style ignored the profile switch: %q == %q", ascii, truecolor)
	}
	if want := "probe"; ascii != want {
		t.Errorf("Ascii profile output = %q, want %q", ascii, want)
	}
}

// Use must refresh the cache, not leave a stale strip background behind.
func TestUseRefreshesStripCache(t *testing.T) {
	orig := Cur()
	t.Cleanup(func() { Use(orig) })

	dark, _ := Named("solarized-dark")
	light, _ := Named("solarized-light")

	Use(dark)
	first := StripBg()
	Use(light)
	second := StripBg()

	if first == second {
		t.Fatalf("StripBg() unchanged across Use: %v", first)
	}
	if want := stripBgOf(light); second != want {
		t.Fatalf("StripBg() = %v after Use(light), want %v", second, want)
	}
}

// A Theme built as a literal (not via New) still has to get its cache filled by
// Use, or StripText would render with a zero background.
func TestUseCachesForLiteralTheme(t *testing.T) {
	orig := Cur()
	t.Cleanup(func() { Use(orig) })

	lit := Theme{
		Panel: lipgloss.NewStyle().BorderForeground(lipgloss.Color("#123456")),
		Dim:   lipgloss.NewStyle().Foreground(lipgloss.Color("#654321")),
	}
	Use(lit)

	if got, want := StripBg(), stripBgOf(lit); got != want {
		t.Fatalf("StripBg() = %v, want %v", got, want)
	}
	fg := lipgloss.Color("#ffffff")
	if got, want := StripText(fg, "x"), uncachedStripText(lit, fg, "x"); got != want {
		t.Fatalf("StripText = %q, want %q", got, want)
	}
}
