package layout

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/codyconfer/viewkit/theme"
)

// TitledBox renders a rounded box with the title embedded in the top border.
// The body is f.BodyWidth() wide, so like Box it clamps narrow frames *up* to
// theme.MinBodyWidth and can overflow a tiling layout's rect — use
// Frame.CellTitledBox inside grids.
func (f Frame) TitledBox(title string, lines ...string) string {
	return f.TitledBoxIcon("", title, lines...)
}

// TitledBoxIcon is TitledBox with an icon rendered before the title in the
// top border. It shares TitledBox's clamp-up behavior.
func (f Frame) TitledBoxIcon(icon, title string, lines ...string) string {
	return f.titledBoxAt(f.BodyWidth(), icon, title, lines...)
}

// CellTitledBox is TitledBox sized to fit inside f.Width rather than around it.
// TitledBox routes through Frame.BodyWidth, which clamps a narrow body *up* to
// theme.MinBodyWidth, so a titled pane needs theme.MinBodyWidth+4 columns before
// it stops spilling past the rect its layout gave it — MinTrackWidth alone is not
// enough. CellTitledBox clamps the body down instead, exactly as CellBox does,
// and so keeps its right border in any track.
func (f Frame) CellTitledBox(title string, lines ...string) string {
	return f.CellTitledBoxIcon("", title, lines...)
}

// CellTitledBoxIcon is CellTitledBox with an icon before the title, clamped
// down to f.Width like the rest of the Cell* family.
func (f Frame) CellTitledBoxIcon(icon, title string, lines ...string) string {
	return f.fitCell(f.titledBoxAt(f.cellBody(), icon, title, lines...))
}

func (f Frame) titledBoxAt(inner int, icon, title string, lines ...string) string {
	th := theme.Cur()
	span := inner + 2

	border := th.Dim
	if f.Focused {
		border = th.Accent
	}

	out := make([]string, 0, len(lines)+2)
	out = append(out, titledTopBorder(border, th.PanelTitle, icon, title, span))

	edge := border.Render("│")
	for _, ln := range lines {
		for sub := range strings.SplitSeq(ansi.Hardwrap(ln, inner, false), "\n") {
			pad := max(inner-ansi.StringWidth(sub), 0)
			out = append(out, edge+" "+sub+strings.Repeat(" ", pad)+" "+edge)
		}
	}

	out = append(out, border.Render("╰"+strings.Repeat("─", span)+"╯"))
	return strings.Join(out, "\n")
}

func titledTopBorder(border, titleSty lipgloss.Style, icon, title string, span int) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return border.Render("╭" + strings.Repeat("─", span) + "╮")
	}
	var seg string
	if icon != "" {
		seg = " " + icon + titleSty.Render(ansi.Truncate(title, span-4, "…")) + " "
	} else {
		seg = " " + titleSty.Render(ansi.Truncate(title, span-2, "…")) + " "
	}
	fill := max(span-ansi.StringWidth(seg), 0)
	return border.Render("╭") + seg + border.Render(strings.Repeat("─", fill)+"╮")
}
