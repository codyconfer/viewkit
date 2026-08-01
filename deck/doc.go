// Package deck is the bubbletea runtime for viewkit.
//
// Invariant: tea lives ONLY in this package — viewkit core (glyph/theme/layout/panels/…)
// must not import bubbletea. Apps and plugins implement View and register via
// RegisterView; Model owns stack navigation + chrome.
//
// Rendering context is a ui.Scope (theme, keys, glyphs) passed via WithScope
// and swapped with Model.SetScope; there is no process-global theme or scheme.
// Contribution registries (theme.Register / keys.Register / RegisterView) and
// the glyph default mode (glyph.SetMode, write-once) remain process level.
//
// See INTERFACE.md for the Model + scope contract.
package deck
