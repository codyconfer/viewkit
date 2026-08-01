package panels

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/theme"
	"github.com/codyconfer/viewkit/ui"
)

// TestScopedThemeWinsOverGlobal renders through a Frame carrying a garish
// scoped theme and asserts the scope's SGR shows up instead of the global's.
func TestScopedThemeWinsOverGlobal(t *testing.T) {
	t.Parallel()
	garish := theme.Default()
	garish.Accent = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff00ff"))

	scope := ui.Default()
	scope.Theme = garish
	f := layout.NewFrame(60).WithUI(scope)

	out := Clock(f, "T", time.Date(2026, 8, 1, 12, 30, 0, 0, time.UTC))
	scoped := garish.Accent.Render("12:30:00")
	global := theme.Default().Accent.Render("12:30:00")
	if !strings.Contains(out, scoped) {
		t.Fatalf("scoped Accent SGR missing from output:\n%q", out)
	}
	if scoped != global && strings.Contains(out, global) {
		t.Fatalf("global Accent SGR leaked into scoped render:\n%q", out)
	}
}
