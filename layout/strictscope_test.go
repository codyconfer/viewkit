package layout

import (
	"strings"
	"testing"

	"github.com/codyconfer/viewkit/ui"
)

func TestStrictScopeAccessors(t *testing.T) {
	prev := StrictScope
	StrictScope = true
	t.Cleanup(func() { StrictScope = prev })

	reads := map[string]func(Frame){
		"Theme":  func(f Frame) { _ = f.Theme() },
		"Glyphs": func(f Frame) { _ = f.Glyphs() },
		"Scheme": func(f Frame) { _ = f.Scheme() },
	}

	scoped := NewFrame(40).WithUI(ui.Default())
	for name, read := range reads {
		t.Run("scoped/"+name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("scoped %s read panicked: %v", name, r)
				}
			}()
			read(scoped)
		})
	}

	unscoped := NewFrame(40)
	for name, read := range reads {
		t.Run("unscoped/"+name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r == nil {
					t.Fatalf("unscoped %s read did not panic under StrictScope", name)
				}
				msg, ok := r.(string)
				if !ok || !strings.Contains(msg, "unscoped Frame read") {
					t.Fatalf("unscoped %s read panicked with %v, want the StrictScope message", name, r)
				}
			}()
			read(unscoped)
		})
	}
}
