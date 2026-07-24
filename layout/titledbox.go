package layout

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/codyconfer/viewkit/theme"
)

func (f Frame) TitledBox(title string, lines ...string) string {
	return f.TitledBoxIcon("", title, lines...)
}

func (f Frame) TitledBoxIcon(icon, title string, lines ...string) string {
	th := theme.Cur()
	inner := f.BodyWidth()
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
