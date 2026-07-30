package deck

import "github.com/codyconfer/viewkit/keys"

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
		sc.Binding(keys.Open),
		sc.Binding(keys.Reload),
	)
}
