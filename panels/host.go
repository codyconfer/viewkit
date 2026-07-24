package panels

import "github.com/codyconfer/viewkit/layout"

// Host identifies where a panel is mounted.
//
// Inline shells (CLI dashboards, non-tea panes) use layout.Frame.
// Deck hosts (viewkit/deck body slots) supply width×height.
// Panels that implement DualHost work in both without tea in viewkit core.
type Host int

const (
	// Inline is a non-tea shell using layout.Frame.
	Inline Host = iota
	// Deck is a viewkit/deck body region (width×height).
	Deck
)

// DualHost is the inline-shell vs deck panel contract.
// Implementations must not import bubbletea — tea stays in viewkit/deck.
type DualHost interface {
	RenderInline(f layout.Frame) string
	RenderDeck(width, height int) string
}

// Render dispatches to the host-appropriate DualHost method.
func Render(p DualHost, host Host, f layout.Frame, width, height int) string {
	if p == nil {
		return ""
	}
	switch host {
	case Deck:
		return p.RenderDeck(width, height)
	default:
		return p.RenderInline(f)
	}
}

// StaticPanel is a DualHost that paints the same titled lines for both hosts.
// Deck height is filled by truncating/padding body lines.
type StaticPanel struct {
	Title string
	Lines []string
}

// RenderInline implements DualHost.
func (p StaticPanel) RenderInline(f layout.Frame) string {
	return f.Panel(p.Title, p.Lines...)
}

// RenderDeck implements DualHost.
func (p StaticPanel) RenderDeck(width, height int) string {
	f := layout.NewFrame(width)
	body := f.Panel(p.Title, p.Lines...)
	return layout.FillHeight(body, max(height, 1))
}
