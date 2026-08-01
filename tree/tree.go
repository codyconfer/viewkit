// Package tree draws hierarchical rows as an ASCII or Unicode tree.
//
// Callers render their own row bodies and hand them to this package together
// with structural flags (the parent's stem, whether the row is the last child).
// The package supplies the connector glyphs, the dim styling from the current
// theme, and the gap stems that keep vertical rules unbroken when a list widget
// inserts blank lines between rows.
//
// A tree is built from the top down: a trunk row the caller renders itself,
// then one Branch per child of the trunk, then Leaf rows under each branch
// using that branch's stem.
package tree

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/codyconfer/viewkit/glyph"
	"github.com/codyconfer/viewkit/list"
	"github.com/codyconfer/viewkit/theme"
)

// Row is one rendered node of a tree: its lines already carry the connector
// prefixes and are ready to print.
type Row struct {
	// Lines are the rendered lines of the row, connectors included.
	Lines []string
	// Key identifies the row to the caller, for example the URL a leaf opens.
	// An empty Key marks a structural row that carries no payload.
	Key string
	// Selectable reports whether a cursor may land on the row.
	Selectable bool
	// GapStem is the stem to repeat in blank lines drawn after the row, so the
	// tree's vertical rules survive the gap. It matches the field of the same
	// name on list.Item.
	GapStem string
	// Payload carries the caller's domain value for the row; Item copies it
	// through so list selection hands the value back without a side table.
	Payload any
}

// Item converts the row to a list.Item, joining its rendered lines into one
// block. Key, Selectable, GapStem and Payload copy through unchanged.
func (r Row) Item() list.Item {
	return list.Item{
		Block:      strings.Join(r.Lines, "\n"),
		Key:        r.Key,
		Selectable: r.Selectable,
		GapStem:    r.GapStem,
		Payload:    r.Payload,
	}
}

// Connectors is the glyph set used to draw tree edges. Mid, End, Vert and Space
// must all have the same display width so that columns line up.
type Connectors struct {
	// Trunk marks the root of the tree.
	Trunk string
	// Mid joins a child that still has siblings below it.
	Mid string
	// End joins the last child of a parent.
	End string
	// Vert continues the parent's rule past a Mid child.
	Vert string
	// Space blanks the parent's rule past an End child.
	Space string
	// Empty marks a parent that has no children at all.
	Empty string
	// Dim styles the connectors and gap stems. A zero value falls back to the
	// process default theme's Dim; build with ConnectorsIn to scope it.
	Dim lipgloss.Style
}

var (
	unicodeConnectors = Connectors{Trunk: "●", Mid: "├─ ", End: "└─ ", Vert: "│  ", Space: "   ", Empty: "∅ "}
	asciiConnectors   = Connectors{Trunk: "*", Mid: "+- ", End: "`- ", Vert: "|  ", Space: "   ", Empty: "o "}
)

// DefaultConnectors returns the connector set for the default glyph mode and
// the built-in default theme: a plain ASCII set when glyphs are off,
// box-drawing characters otherwise.
func DefaultConnectors() Connectors {
	return ConnectorsIn(glyph.DefaultSet(), theme.Default())
}

// ConnectorsIn returns the connector set for g's mode, styled with th's Dim.
func ConnectorsIn(g glyph.Set, th theme.Theme) Connectors {
	c := unicodeConnectors
	if g.Mode() == glyph.ModeNone {
		c = asciiConnectors
	}
	c.Dim = th.Dim
	return c
}

// Edge returns the connector that joins a child to its parent and the stem that
// continues underneath that child. The last child closes the rule with End and
// leaves blank space below it; every other child keeps the rule running.
func (c Connectors) Edge(last bool) (connector, stem string) {
	if last {
		return c.End, c.Space
	}
	return c.Mid, c.Vert
}

// Indent reports the display width a child's prefix occupies when it hangs off
// stem. Subtract it from the available width before rendering child bodies so
// they wrap inside the space the tree leaves them.
func (c Connectors) Indent(stem string) int {
	return lipgloss.Width(stem) + lipgloss.Width(c.Mid)
}

// Leaf renders body as a child hanging off stem, the accumulated prefix of the
// parent's own connectors. The first body line is joined with the child's
// connector; later lines are indented with the continuation stem so a
// multi-line body stays under its connector. A non-empty key makes the row
// selectable.
func Leaf(c Connectors, stem string, last bool, body []string, key string) Row {
	dim := c.dim()
	connector, cont := c.Edge(last)
	lines := make([]string, len(body))
	for i, line := range body {
		if i == 0 {
			lines[i] = dim.Render(stem+connector) + line
			continue
		}
		lines[i] = dim.Render(stem+cont) + line
	}
	return Row{
		Lines:      lines,
		Key:        key,
		Selectable: key != "",
		GapStem:    dim.Render(stem + cont),
	}
}

// dim returns c.Dim, falling back to the built-in default theme's Dim when
// the field was left zero (callers constructing Connectors literals).
func (c Connectors) dim() lipgloss.Style {
	if c.Dim.GetForeground() == (lipgloss.NoColor{}) {
		return theme.Default().Dim
	}
	return c.Dim
}

// Branch renders body as a direct child of the trunk: a Leaf with no parent
// stem. Its GapStem is the stem its own children must hang off, so pass the
// unstyled result of Edge to those Leaf calls.
func Branch(c Connectors, last bool, body []string, key string) Row {
	return Leaf(c, "", last, body, key)
}
