package glyph

import (
	"sync"
	"testing"
)

func TestSetModeIsRaceFreeWhileResolving(t *testing.T) {
	prev := CurrentMode()
	t.Cleanup(func() { SetMode(prev) })

	Register("test.race", Variants{Nerd: "N", Uni: "U", ASCII: "A"})

	const iterations = 20000
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		<-start
		for range iterations {
			_ = StatusOK()
			_ = Lead(Check())
			_ = ResolveID("test.race")
			_ = CurrentMode()
		}
	}()

	go func() {
		defer wg.Done()
		<-start
		for i := range iterations {
			switch i % 3 {
			case 0:
				SetMode(ModeNerd)
			case 1:
				SetMode(ModeUnicode)
			default:
				SetMode(ModeNone)
			}
		}
	}()

	close(start)
	wg.Wait()
}

func TestPartialVariantsDegradeAcrossModes(t *testing.T) {
	prev := CurrentMode()
	t.Cleanup(func() { SetMode(prev) })

	cases := []struct {
		name string
		v    Variants
		want map[Mode]string
	}{
		{
			name: "nerd only",
			v:    Variants{Nerd: "N"},
			want: map[Mode]string{ModeNerd: "N", ModeUnicode: "N", ModeNone: "N"},
		},
		{
			name: "ascii only",
			v:    Variants{ASCII: "A"},
			want: map[Mode]string{ModeNerd: "A", ModeUnicode: "A", ModeNone: "A"},
		},
		{
			name: "unicode only",
			v:    Variants{Uni: "U"},
			want: map[Mode]string{ModeNerd: "U", ModeUnicode: "U", ModeNone: "U"},
		},
		{
			name: "nerd and ascii",
			v:    Variants{Nerd: "N", ASCII: "A"},
			want: map[Mode]string{ModeNerd: "N", ModeUnicode: "A", ModeNone: "A"},
		},
		{
			name: "empty stays empty",
			v:    Variants{},
			want: map[Mode]string{ModeNerd: "", ModeUnicode: "", ModeNone: ""},
		},
	}

	for _, tc := range cases {
		for m, want := range tc.want {
			SetMode(m)
			if got := tc.v.String(); got != want {
				t.Errorf("%s: mode %v: String() = %q, want %q", tc.name, m, got, want)
			}
		}
	}
}

func TestRegisteredPartialVariantDegradesInAsciiMode(t *testing.T) {
	prev := CurrentMode()
	t.Cleanup(func() { SetMode(prev) })

	const nerdOnly = "\uf111"
	Register("plugin.x", Variants{Nerd: nerdOnly})
	SetMode(ModeNone)

	if got := ResolveID("plugin.x"); got != nerdOnly {
		t.Fatalf("ResolveID(plugin.x) = %q in ModeNone, want %q; a partially filled Variants must degrade, not vanish", got, nerdOnly)
	}
}

func TestFullVariantsUnaffectedByFallback(t *testing.T) {
	prev := CurrentMode()
	t.Cleanup(func() { SetMode(prev) })

	v := Variants{Nerd: "N", Uni: "U", ASCII: "A"}
	for m, want := range map[Mode]string{ModeNerd: "N", ModeUnicode: "U", ModeNone: "A"} {
		SetMode(m)
		if got := v.String(); got != want {
			t.Errorf("mode %v: String() = %q, want %q", m, got, want)
		}
	}
}
