package layout

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/codyconfer/viewkit/theme"
	"github.com/codyconfer/viewkit/ui"
)

func TestThemeInjectionFlowsThroughPanels(t *testing.T) {
	t.Parallel()
	custom := theme.Default()
	custom.TooNarrowTitle = "SCREEN TOO SMALL"

	out := stripANSI(TooNarrowIn(custom, theme.MinScreenWidth-1))
	if !strings.Contains(out, "SCREEN TOO SMALL") {
		t.Fatalf("injected copy not rendered by TooNarrowIn: %q", out)
	}
	if strings.Contains(out, theme.DefaultTooNarrowTitle) {
		t.Fatalf("default copy leaked past the injected theme: %q", out)
	}
}

func TestScopedFrameRendersWithItsThemeNotTheDefault(t *testing.T) {
	t.Parallel()
	garish := theme.Default()
	garish.Title = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff00ff"))

	f := NewFrame(40).WithUI(&ui.Scope{Theme: garish})
	scoped := f.Cursor(true)
	fallback := NewFrame(40).Cursor(true)

	if !strings.Contains(scoped, "38;2;255;0;255") {
		t.Fatalf("scoped render missing the scope theme's SGR: %q", scoped)
	}
	if scoped == fallback {
		t.Fatalf("scoped render matched the built-in default theme's: %q", scoped)
	}
}
