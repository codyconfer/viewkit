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

func TestUseUpdatesCur(t *testing.T) {
	orig := *Cur()
	defer Use(orig)

	th, _ := Named("solarized-dark")
	Use(th)

	if Cur().Accent.GetForeground() != solarizedDarkPalette.Accent {
		t.Fatal("Cur() not updated by Use()")
	}
	if Cur().Dim.GetForeground() != solarizedDarkPalette.Muted {
		t.Fatalf("Cur().Dim = %v, want %v", Cur().Dim.GetForeground(), solarizedDarkPalette.Muted)
	}
}

// The exported style vars are a write-once snapshot of the default theme: that
// is what makes them safe to read from a render goroutine while Use runs.
const exportedVarsContract = "the package-level *Sty vars are a one-time snapshot taken at init and are " +
	"deliberately NOT refreshed by Use. Use is safe to call while other goroutines render, which is why " +
	"the active theme lives behind an atomic.Pointer; re-adding syncExported to Use would write these " +
	"plain package vars with no synchronisation at all while renderers read them. If this test fails " +
	"because Use started syncing again, the fix is to stop syncing, not to update the expectation."

func TestExportedVarsSnapshotDefaultTheme(t *testing.T) {
	t.Log(exportedVarsContract)
	orig := *Cur()
	defer Use(orig)

	th, ok := Named("solarized-dark")
	if !ok {
		t.Fatal("solarized-dark is not registered; this test needs a theme that differs from the default")
	}
	if th.Accent.GetForeground() == defaultPalette.Accent {
		t.Fatal("solarized-dark's accent equals the default's, so switching to it proves nothing")
	}
	Use(th)

	if got := AccentSty.GetForeground(); got != defaultPalette.Accent {
		t.Fatalf("AccentSty = %v, want the default-theme snapshot %v: %s",
			got, defaultPalette.Accent, exportedVarsContract)
	}
	if got := DimSty.GetForeground(); got != defaultPalette.Muted {
		t.Fatalf("DimSty = %v, want the default-theme snapshot %v: %s",
			got, defaultPalette.Muted, exportedVarsContract)
	}
	if got := Cur().Accent.GetForeground(); got != th.Accent.GetForeground() {
		t.Fatalf("Cur().Accent = %v after Use, want %v; the active theme must still follow Use even though "+
			"the exported snapshot does not", got, th.Accent.GetForeground())
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
