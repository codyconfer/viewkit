package layout

import (
	"strings"

	"github.com/charmbracelet/x/ansi"

	"github.com/codyconfer/viewkit/theme"
)

// FlexSections is FlexColumns applied per pane Group: panes sharing a Group
// value are flowed together under a titled section header (a rule line
// labeled with the group name), sections stacked in first-appearance order.
// Panes with an empty Group render without a header.
type FlexSections struct {
	FlexBounds
}

// Arrange implements Arranger by grouping tier-visible panes by Group and
// laying each group out with FlexColumns.
func (g FlexSections) Arrange(f Frame, tier Tier, panes []Pane, focusedName string) string {
	width, visible := flexVisible(f, tier, panes)
	if len(visible) == 0 {
		return ""
	}

	order := make([]string, 0)
	groups := map[string][]Pane{}
	for _, p := range visible {
		if _, ok := groups[p.Group]; !ok {
			order = append(order, p.Group)
		}
		groups[p.Group] = append(groups[p.Group], p)
	}

	inner := FlexColumns{FlexBounds: g.FlexBounds} //nolint:staticcheck // explicit over conversion
	blocks := make([]string, 0, len(order))
	for _, name := range order {
		body := inner.Arrange(f, tier, groups[name], focusedName)
		if body == "" {
			continue
		}
		if name == "" {
			blocks = append(blocks, body)
			continue
		}
		blocks = append(blocks, StackTight(sectionHeader(width, name), body))
	}
	return Stack(blocks...)
}

func sectionHeader(width int, title string) string {
	label := theme.Cur().PanelTitle.Render(ansi.Truncate(title, width, "…"))
	rule := theme.Cur().Dim.Render(strings.Repeat("─", width))
	return label + "\n" + rule
}
