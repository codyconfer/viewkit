package deck

import (
	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/layout"
)

// ScrollBody owns the scroll offset and row math that hand-rolled scrolling
// Views repeat: embed it, route navigation keys through Handle, and render
// through View with the height that View.Body receives (the deck passes the
// usable body height, so no chrome constants are needed).
//
//	func (v *myView) Body(width, height int) string {
//		return v.scroll.View(render(width), height)
//	}
type ScrollBody struct {
	layout.ScrollState

	rows  int
	total int
}

// Handle applies a navigation action to the scroll offset — Up/Down move one
// line, PageUp/PageDown move one window — and reports whether act was a
// scroll action. Row math comes from the last View call.
func (s *ScrollBody) Handle(act keys.Action) bool {
	rows := max(layout.ViewportContentRows(s.rows), 1)
	switch act {
	case keys.Up:
		s.Scroll(-1, s.total, rows)
	case keys.Down:
		s.Scroll(1, s.total, rows)
	case keys.PageUp:
		s.Scroll(-rows, s.total, rows)
	case keys.PageDown:
		s.Scroll(rows, s.total, rows)
	default:
		return false
	}
	return true
}

// Total returns the line count of the last body given to View.
func (s *ScrollBody) Total() int { return s.total }

// View windows body to height rows at the current offset and records the
// totals Handle needs. height is the usable body height View.Body receives.
func (s *ScrollBody) View(body string, height int) string {
	s.rows = max(height, 1)
	s.total = layout.CountLines(body)
	return layout.Viewport(body, s.rows, s.Offset)
}
