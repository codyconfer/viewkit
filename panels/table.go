package panels

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/codyconfer/viewkit/layout"
)

// TableMaxCell is the default per-column display-width cap used by Table when
// TableOpts.MaxCell is not set.
const TableMaxCell = 40

const (
	tableCellSep = " │ "
	tableRuleSep = "─┼─"
)

// TableOpts tunes Table rendering. The zero value selects the defaults.
type TableOpts struct {
	// MaxCell caps each column's display width. Values <= 0 select
	// TableMaxCell. The effective cap is further limited by the frame's body
	// width so a single column can never overflow the frame.
	MaxCell int
}

// Table renders headers and rows as an aligned terminal table.
//
// Column width is the widest display width among the header and its cells,
// capped by TableOpts.MaxCell; longer cells are truncated with an ellipsis.
// All measuring and padding is display-width based (lipgloss.Width), so
// wide runes (CJK, emoji) stay aligned with ASCII rows.
//
// Cells are flattened to a single line first: "\r\n", "\n" and "\t" each
// become a space. Rows may be ragged — missing cells render empty and cells
// past the last header are dropped.
//
// The header uses theme Key, body rows use Val, and the separator plus the
// trailing "(N rows)" footer use Dim. Table returns "(no columns)" when
// headers is empty, and a header, separator and "(0 rows)" when there are no
// rows.
func Table(f layout.Frame, headers []string, rows [][]string, opts ...TableOpts) string {
	th := f.Theme()
	if len(headers) == 0 {
		return th.Dim.Render("(no columns)")
	}

	maxCell := TableMaxCell
	if len(opts) > 0 && opts[0].MaxCell > 0 {
		maxCell = opts[0].MaxCell
	}
	if bw := f.BodyWidth(); bw > 0 && maxCell > bw {
		maxCell = bw
	}

	cols := len(headers)
	head := make([]string, cols)
	widths := make([]int, cols)
	for i, h := range headers {
		head[i] = tableOneLine(h)
		widths[i] = tableClamp(lipgloss.Width(head[i]), maxCell)
	}

	body := make([][]string, len(rows))
	for r, row := range rows {
		cells := make([]string, cols)
		for i := 0; i < cols && i < len(row); i++ {
			cells[i] = tableOneLine(row[i])
			if w := tableClamp(lipgloss.Width(cells[i]), maxCell); w > widths[i] {
				widths[i] = w
			}
		}
		body[r] = cells
	}

	var b strings.Builder
	b.WriteString(th.Key.Render(tableRow(head, widths)))
	b.WriteString("\n")

	rule := make([]string, cols)
	for i, w := range widths {
		rule[i] = strings.Repeat("─", w)
	}
	b.WriteString(th.Dim.Render(strings.Join(rule, tableRuleSep)))
	b.WriteString("\n")

	if len(body) == 0 {
		b.WriteString(th.Dim.Render("(0 rows)"))
		return b.String()
	}
	for _, cells := range body {
		b.WriteString(th.Val.Render(tableRow(cells, widths)))
		b.WriteString("\n")
	}
	b.WriteString(th.Dim.Render(fmt.Sprintf("(%d rows)", len(body))))
	return b.String()
}

func tableRow(cells []string, widths []int) string {
	out := make([]string, len(widths))
	for i, w := range widths {
		val := ""
		if i < len(cells) {
			val = layout.Fit(cells[i], w)
		}
		pad := w - lipgloss.Width(val)
		if pad < 0 {
			pad = 0
		}
		out[i] = val + strings.Repeat(" ", pad)
	}
	return strings.Join(out, tableCellSep)
}

func tableOneLine(s string) string {
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	return s
}

func tableClamp(n, maxCell int) int {
	if n > maxCell {
		return maxCell
	}
	return n
}
