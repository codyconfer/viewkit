package deck

import "github.com/codyconfer/viewkit/layout"

// Content is domain-agnostic job/panel body text. Apps adapt their types
// (e.g. domain section trees) into Content before crossing into deck.
type Content interface {
	// Render returns the painted body for the given frame (may ignore it).
	Render(f layout.Frame) string
}

// Text is a Content adapter for pre-rendered strings.
type Text string

// Render implements Content.
func (t Text) Render(layout.Frame) string { return string(t) }

// ContentFunc adapts a function to Content.
type ContentFunc func(f layout.Frame) string

// Render implements Content.
func (f ContentFunc) Render(fr layout.Frame) string { return f(fr) }
