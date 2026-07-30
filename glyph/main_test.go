package glyph

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

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
