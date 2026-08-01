package panels

import (
	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/theme"
)

// LedgerRow is one labeled signed change for Ledger.
type LedgerRow struct {
	Label string
	Delta float64
}

// Ledger renders rows of label/delta pairs in a scroll panel showing visible
// rows starting at row offset. Positive deltas render as "+<n> <unit>" in the
// positive style, negative as "<n> <unit>" in the negative style, and exactly
// zero as a dim em dash with no unit. fmtNum formats the delta magnitude.
// With no rows the panel shows empty in the dim style.
func Ledger(f layout.Frame, title string, rows []LedgerRow, unit string, fmtNum func(float64) string, visible, offset int, empty string) string {
	th := f.Theme()
	if len(rows) == 0 {
		return f.Panel(title, th.Dim.Render(empty))
	}
	lines := make([]string, len(rows))
	for i, r := range rows {
		lines[i] = f.Spread(th.Val.Render(r.Label), delta(th, r.Delta, unit, fmtNum))
	}
	return f.ScrollPanel(title, lines, visible, offset)
}

func delta(th theme.Theme, v float64, unit string, fmtNum func(float64) string) string {
	switch {
	case v > 0:
		return th.Can.Render("+" + fmtNum(v) + " " + unit)
	case v < 0:
		return th.Cant.Render(fmtNum(v) + " " + unit)
	default:
		return th.Dim.Render("—")
	}
}
