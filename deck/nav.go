package deck

import "github.com/codyconfer/viewkit/keys"

// navMap builds the generic-view action map from the active scheme (keys.Use).
//
// Generic views resolve input and render hints through this single source so an
// installed scheme drives both — no raw key literals in view code, and no
// footer hint that contradicts the bindings actually in effect.
func navMap() *keys.Map {
	sc := keys.Cur()
	return keys.NewMap(
		sc.Binding(keys.Up),
		sc.Binding(keys.Down),
		sc.Binding(keys.Confirm),
		sc.Binding(keys.Cancel),
		sc.Binding(keys.Quit),
		sc.Binding(keys.PageUp),
		sc.Binding(keys.PageDown),
		sc.Binding(keys.FocusNext),
		sc.Binding(keys.FocusPrev),
	)
}
