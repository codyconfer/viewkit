// Package deck is the bubbletea runtime for viewkit (nested module).
//
// Invariant: tea lives ONLY here — viewkit core (glyph/theme/layout/panels/…)
// must not import bubbletea. Apps and plugins implement View and register via
// RegisterView; Host owns stack navigation + chrome.
//
// See INTERFACE.md for the goose-facing contract review notes (ADR-2).
package deck
