// Package deck is the bubbletea runtime for viewkit (nested module).
//
// Invariant: tea lives ONLY here — viewkit core (glyph/theme/layout/panels/…)
// must not import bubbletea. Apps and plugins implement View and register via
// RegisterView; Model (alias Host) owns stack navigation + chrome.
//
// Process-global singletons (install before Run):
//   - theme.Use / theme.Cur — active palette
//   - keys.Use / keys.Cur — active key scheme
//   - theme.Register / keys.Register / RegisterView — contribution registries
//
// See INTERFACE.md for the Model + singleton contract.
package deck
