package theme

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestNamedReturnsRegisteredTheme(t *testing.T) {
	th, ok := Named("monokai")
	if !ok {
		t.Fatal("Named(monokai) not found")
	}
	if got := th.Accent.GetForeground(); got != monokaiPalette.Accent {
		t.Fatalf("monokai accent = %v, want %v", got, monokaiPalette.Accent)
	}

	if _, ok := Named("does-not-exist"); ok {
		t.Fatal("Named(unknown) should report not found")
	}
}

func TestUseSyncsExportedVars(t *testing.T) {
	orig := *Cur()
	defer Use(orig)

	th, _ := Named("solarized-dark")
	Use(th)

	if Cur().Accent.GetForeground() != solarizedDarkPalette.Accent {
		t.Fatal("Cur() not updated by Use()")
	}
	if AccentSty.GetForeground() != solarizedDarkPalette.Accent {
		t.Fatalf("AccentSty not synced: %v", AccentSty.GetForeground())
	}
	if DimSty.GetForeground() != solarizedDarkPalette.Muted {
		t.Fatalf("DimSty not synced: %v", DimSty.GetForeground())
	}
}

func TestKeysDefaultFirst(t *testing.T) {
	keys := Keys()
	if len(keys) == 0 || keys[0] != "default" {
		t.Fatalf("Keys() = %v, want default first", keys)
	}
}

func TestRegisterAddsNamedTheme(t *testing.T) {
	orig := registry
	defer func() { registry = orig }()

	p := Palette{Accent: lipgloss.Color("#123456")}
	Register("custom", "My Custom", p)

	th, ok := Named("custom")
	if !ok {
		t.Fatal("Named(custom) not found after Register")
	}
	if got := th.Accent.GetForeground(); got != p.Accent {
		t.Fatalf("custom accent = %v, want %v", got, p.Accent)
	}
	if got := DisplayName("custom"); got != "My Custom" {
		t.Fatalf("DisplayName(custom) = %q, want %q", got, "My Custom")
	}

	var found bool
	for _, k := range Keys() {
		if k == "custom" {
			found = true
		}
	}
	if !found {
		t.Fatal("Keys() does not include registered key")
	}
}

func TestRegisterReplacesExistingKey(t *testing.T) {
	orig := registry
	defer func() { registry = orig }()

	before := len(registry)
	Register("dup", "First", Palette{Accent: lipgloss.Color("#111111")})
	Register("dup", "Second", Palette{Accent: lipgloss.Color("#222222")})

	if got := len(registry); got != before+1 {
		t.Fatalf("registry len = %d, want %d (no duplicate entries)", got, before+1)
	}
	if got := DisplayName("dup"); got != "Second" {
		t.Fatalf("DisplayName(dup) = %q, want %q", got, "Second")
	}
	th, _ := Named("dup")
	if got := th.Accent.GetForeground(); got != lipgloss.Color("#222222") {
		t.Fatalf("dup accent = %v, want the replacement palette", got)
	}
}
