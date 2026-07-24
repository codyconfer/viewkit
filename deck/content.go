package deck

// Content is domain-agnostic flight/panel body text. Apps adapt their types
// (e.g. munin signals.Section trees) into Content before crossing into deck.
type Content interface {
	// Render returns the painted body for the given width (may ignore width).
	Render(width int) string
}

// Text is a Content adapter for pre-rendered strings.
type Text string

// Render implements Content.
func (t Text) Render(int) string { return string(t) }

// ContentFunc adapts a function to Content.
type ContentFunc func(width int) string

// Render implements Content.
func (f ContentFunc) Render(width int) string { return f(width) }
