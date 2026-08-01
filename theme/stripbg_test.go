package theme

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

const stripBgContract = "the strip background is the panel border colour and nothing else. " +
	"lipgloss.Style.GetBorderTopForeground never returns a nil interface — an unset border yields " +
	"lipgloss.NoColor{} — so the `if bg != nil` fallback to Dim that stripBgOf used to carry could " +
	"never run. It was deleted as a verified no-op. A Theme literal with no panel border therefore " +
	"gets NoColor{}, NOT Dim's foreground; if Dim is the intent, that is a product change, not a test fix."

func TestStripBgIsThePanelBorderColour(t *testing.T) {
	t.Log(stripBgContract)
	orig := Cur()
	t.Cleanup(func() { Use(orig) })

	for _, key := range Keys() {
		th, ok := Named(key)
		if !ok {
			t.Fatalf("Named(%q) not found", key)
		}
		want := th.Panel.GetBorderTopForeground()
		if want == nil {
			t.Fatalf("%s: panel border foreground is a nil interface, which lipgloss is not supposed to "+
				"produce; %s", key, stripBgContract)
		}
		if _, isNone := want.(lipgloss.NoColor); isNone {
			t.Errorf("%s: panel border foreground is NoColor, so the status strip renders with no "+
				"background at all", key)
		}
		Use(th)
		if got := StripBg(); got != want {
			t.Errorf("%s: StripBg() = %v, want the panel border %v", key, got, want)
		}
	}
}

func TestStripBgOfIgnoresDim(t *testing.T) {
	t.Log(stripBgContract)
	border := lipgloss.Color("#111111")
	dim := lipgloss.Color("#999999")

	withBorder := Theme{
		Panel: lipgloss.NewStyle().BorderForeground(border),
		Dim:   lipgloss.NewStyle().Foreground(dim),
	}
	if got := stripBgOf(withBorder); got != border {
		t.Errorf("stripBgOf = %v, want the panel border %v", got, border)
	}

	noBorder := Theme{Dim: lipgloss.NewStyle().Foreground(dim)}
	got := stripBgOf(noBorder)
	if got == dim {
		t.Fatalf("stripBgOf fell back to Dim (%v); %s", dim, stripBgContract)
	}
	if _, isNone := got.(lipgloss.NoColor); !isNone {
		t.Errorf("stripBgOf with no panel border = %#v, want lipgloss.NoColor{}", got)
	}
}
