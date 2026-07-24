package panels

import (
	"strings"
	"testing"

	"github.com/codyconfer/viewkit/layout"
)

func TestDualHostInlineAndDeck(t *testing.T) {
	p := StaticPanel{Title: "STATUS", Lines: []string{"ok", "ready"}}
	inline := Render(p, Inline, layout.NewFrame(40), 0, 0)
	if !strings.Contains(inline, "STATUS") || !strings.Contains(inline, "ok") {
		t.Fatalf("inline missing content:\n%s", inline)
	}
	deck := Render(p, Deck, layout.Frame{}, 40, 8)
	if !strings.Contains(deck, "STATUS") || !strings.Contains(deck, "ready") {
		t.Fatalf("deck missing content:\n%s", deck)
	}
	if layout.CountLines(deck) < 8 {
		t.Fatalf("deck should fill height, got %d lines:\n%s", layout.CountLines(deck), deck)
	}
}

func TestPanelRegistry(t *testing.T) {
	id := "test.static"
	// clean slate for this id if re-run in same process
	regMu.Lock()
	delete(panels, id)
	regMu.Unlock()

	Register(id, func() DualHost {
		return StaticPanel{Title: "T", Lines: []string{"x"}}
	})
	p, ok := Lookup(id)
	if !ok {
		t.Fatal("lookup failed")
	}
	out := p.RenderInline(layout.NewFrame(32))
	if !strings.Contains(out, "T") || !strings.Contains(out, "x") {
		t.Fatalf("render = %q", out)
	}
	found := false
	for _, k := range IDs() {
		if k == id {
			found = true
		}
	}
	if !found {
		t.Fatalf("IDs missing %q: %v", id, IDs())
	}
}
