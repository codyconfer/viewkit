package deck

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func sgrPrefix(t *testing.T, style lipgloss.Style) string {
	t.Helper()
	rendered := style.Render("Z")
	i := strings.Index(rendered, "Z")
	if i <= 0 {
		t.Fatalf("style emitted no SGR prefix: %q", rendered)
	}
	return rendered[:i]
}

func registryScope(t *testing.T) {
	t.Helper()

	regMu.Lock()
	prevViews := views
	views = map[string]func() View{}
	regMu.Unlock()

	compMu.Lock()
	prevComps := components
	components = map[string]func() Component{}
	compMu.Unlock()

	t.Cleanup(func() {
		regMu.Lock()
		views = prevViews
		regMu.Unlock()

		compMu.Lock()
		components = prevComps
		compMu.Unlock()
	})
}

func TestMain(m *testing.M) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	if got := lipgloss.ColorProfile(); got != termenv.TrueColor {
		fmt.Fprintf(os.Stderr, "TestMain: lipgloss.ColorProfile() = %v, want termenv.TrueColor\n", got)
		os.Exit(1)
	}
	probe := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff8800")).Render("probe")
	if !strings.Contains(probe, "\x1b[38;2;") {
		fmt.Fprintf(os.Stderr, "TestMain: TrueColor not in effect, probe render = %q\n", probe)
		os.Exit(1)
	}
	os.Exit(m.Run())
}
