package panels

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/codyconfer/viewkit/glyph"
	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/notify"
	"github.com/codyconfer/viewkit/theme"
	"github.com/codyconfer/viewkit/ui"
)

// garishPalette is deliberately unlike any registered theme so a scoped SGR
// can never be mistaken for a global one.
var garishPalette = theme.Palette{
	Accent:   lipgloss.Color("#ff00ff"),
	Border:   lipgloss.Color("#00ffff"),
	Muted:    lipgloss.Color("#00ff00"),
	Text:     lipgloss.Color("#ffff00"),
	Selected: lipgloss.Color("#ff0000"),
	Success:  lipgloss.Color("#00ff88"),
	Warning:  lipgloss.Color("#ff8800"),
	Failure:  lipgloss.Color("#8800ff"),
	Info:     lipgloss.Color("#0088ff"),
	Series2:  lipgloss.Color("#88ff00"),
	Series3:  lipgloss.Color("#ff0088"),
	Bg:       lipgloss.Color("#000000"),
}

// fgParams returns the SGR parameter string ("38;2;r;g;b") a foreground color
// renders as under the active profile.
func fgParams(t *testing.T, c lipgloss.Color) string {
	t.Helper()
	s := lipgloss.NewStyle().Foreground(c).Render("Z")
	i := strings.Index(s, "Z")
	if i <= 0 {
		t.Fatalf("style emitted no SGR prefix: %q", s)
	}
	return strings.TrimSuffix(strings.TrimPrefix(s[:i], "\x1b["), "m")
}

// TestGarishSweepUnderStrictScope renders representative panels through a
// garish-scoped Frame with layout.StrictScope on: any global-theme fallback
// panics, and any global SGR in the output is a scope leak.
func TestGarishSweepUnderStrictScope(t *testing.T) {
	prevProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(prevProfile) })

	prevStrict := layout.StrictScope
	layout.StrictScope = true
	t.Cleanup(func() { layout.StrictScope = prevStrict })

	scope := ui.Default()
	scope.Theme = theme.New(garishPalette)
	f := layout.NewFrame(60).WithUI(scope)

	garishAccent := fgParams(t, garishPalette.Accent)
	garishMuted := fgParams(t, garishPalette.Muted)
	forbidden := map[string]string{
		"default Accent": fgParams(t, lipgloss.Color("#6e9fff")),
		"default Muted":  fgParams(t, lipgloss.Color("#9c9fa3")),
		"default Text":   fgParams(t, lipgloss.Color("#ececed")),
	}

	renders := map[string]func() string{
		"Bar": func() string {
			return Bar(f, "T", []Datum{{"alpha", 3}, {"beta", -1}}, 10, fnum, "none")
		},
		"Ledger": func() string {
			rows := []LedgerRow{{"gain", 4}, {"loss", -2}, {"flat", 0}}
			return Ledger(f, "T", rows, "pts", fnum, 5, 0, "none")
		},
		"Table": func() string {
			return Table(f, []string{"name", "n"}, [][]string{{"a", "1"}, {"b", "2"}})
		},
		"NotificationPanel": func() string {
			ns := []notify.Notification{
				notify.Note(glyph.SeverityPositive, "ok", "done"),
				notify.Note(glyph.SeverityNegative, "bad", "broke"),
			}
			return NotificationPanel(f, "N", ns)
		},
		"Markdown": func() string {
			return Markdown(f, "# Head\n\nsome *text* with `code`\n\n```\nraw\n```\n")
		},
	}

	for name, render := range renders {
		t.Run(name, func(t *testing.T) {
			var out string
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("%s hit a global fallback under StrictScope (scope leak): %v", name, r)
					}
				}()
				out = render()
			}()
			if !strings.Contains(out, garishAccent) {
				t.Errorf("%s missing garish accent SGR %q:\n%q", name, garishAccent, out)
			}
			if !strings.Contains(out, garishMuted) {
				t.Errorf("%s missing garish muted SGR %q:\n%q", name, garishMuted, out)
			}
			for role, params := range forbidden {
				if strings.Contains(out, params) {
					t.Errorf("%s leaked %s SGR %q:\n%q", name, role, params, out)
				}
			}
		})
	}
}
