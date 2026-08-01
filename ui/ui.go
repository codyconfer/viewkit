// Package ui bundles the per-program rendering context — theme, key scheme,
// and glyph set. A Scope is immutable after construction; to change
// appearance, build a new Scope and swap the pointer. There is no process
// state to install: components read the Scope they are handed.
package ui

import (
	"github.com/codyconfer/viewkit/glyph"
	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/theme"
)

// Scope is one program's rendering context.
type Scope struct {
	Theme  theme.Theme
	Keys   keys.Scheme
	Glyphs glyph.Set
}

// Default returns a Scope over the built-in defaults: the default theme and
// scheme, and the glyph set for the default mode (the one piece of process
// state left, by design — see glyph.SetMode).
func Default() *Scope {
	return &Scope{Theme: theme.Default(), Keys: keys.Default(), Glyphs: glyph.DefaultSet()}
}

// Success renders a success-colored check followed by msg.
func (s *Scope) Success(msg string) string {
	return s.Theme.Can.Render(s.Glyphs.Check()) + " " + s.Theme.Val.Render(msg)
}

// Bullet renders an accent-colored bullet followed by msg.
func (s *Scope) Bullet(msg string) string {
	return s.Theme.Accent.Render(s.Glyphs.Bullet()) + " " + s.Theme.Val.Render(msg)
}
