package deck

import (
	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/theme"
	"github.com/codyconfer/viewkit/ui"
)

// schemeOf returns the scope's scheme, or the built-in default when unscoped.
func schemeOf(scope *ui.Scope) keys.Scheme {
	if scope != nil {
		return scope.Keys
	}
	return keys.Default()
}

// themeOf returns the scope's theme, or the built-in default when unscoped.
func themeOf(scope *ui.Scope) theme.Theme {
	if scope != nil {
		return scope.Theme
	}
	return theme.Default()
}

func navMapFor(sc keys.Scheme) *keys.Map {
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
