package panels

import "github.com/codyconfer/viewkit/layout"

// Target identifies where a panel is mounted.
//
// Inline shells (CLI dashboards, non-tea panes) use layout.Frame.
// Deck targets (viewkit/deck body slots) supply width×height.
// Panels that implement DualHost work in both without tea in viewkit core.
type Target int

const (
	// TargetInline is a non-tea shell using layout.Frame.
	TargetInline Target = iota
	// TargetDeck is a viewkit/deck body region (width×height).
	TargetDeck
)

// DualHost is the inline-shell vs deck panel contract.
// Implementations must not import bubbletea — tea stays in viewkit/deck.
type DualHost interface {
	RenderInline(f layout.Frame) string
	RenderDeck(f layout.Frame) string
}

// Render dispatches to the target-appropriate DualHost method.
func Render(p DualHost, target Target, f layout.Frame, width, height int) string {
	if p == nil {
		return ""
	}
	switch target {
	case TargetDeck:
		return p.RenderDeck(layout.Frame{Width: width, Height: height, UI: f.UI})
	default:
		return p.RenderInline(f)
	}
}

// StaticPanel is a DualHost that paints the same titled lines for both targets.
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
func (p StaticPanel) RenderDeck(f layout.Frame) string {
	body := f.WithWidth(f.Width).Panel(p.Title, p.Lines...)
	return layout.FillHeight(body, max(f.Height, 1))
}
