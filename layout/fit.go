package layout

import "github.com/codyconfer/viewkit/theme"

// Tier buckets terminal height into three coarse size classes so views can
// drop optional content on short screens instead of measuring exact rows.
// Higher tiers include everything lower ones show.
type Tier int

// The three height tiers, from most to least constrained. TierForHeight maps
// a raw terminal height onto one of these.
const (
	TierShort Tier = iota
	TierMedium
	TierTall
)

// Section is a block of rendered content plus the minimum Tier at which it
// should appear; StackFit and StackTightFit drop sections below the tier.
type Section struct {
	Content string
	MinTier Tier
}

// StackFit stacks the sections visible at the given tier (tier >= MinTier)
// with blank lines between them, like Stack.
func StackFit(tier Tier, sections ...Section) string {
	contents := make([]string, 0, len(sections))
	for _, s := range sections {
		if tier >= s.MinTier {
			contents = append(contents, s.Content)
		}
	}
	return Stack(contents...)
}

// StackTightFit is StackFit joined with single newlines instead of blank
// lines.
func StackTightFit(tier Tier, sections ...Section) string {
	contents := make([]string, 0, len(sections))
	for _, s := range sections {
		if tier >= s.MinTier {
			contents = append(contents, s.Content)
		}
	}
	return StackTight(contents...)
}

// BodyBudget converts a terminal height into the rows available to the body
// after the vertical app margins, floored at 1. A non-positive height (size
// unknown) assumes the minimum body height rather than returning 0.
func BodyBudget(height int) int {
	if height <= 0 {
		return theme.MinBodyHeight - theme.AppMarginY*2
	}
	rows := height - theme.AppMarginY*2
	if rows < 1 {
		rows = 1
	}
	return rows
}

// ContentRows is BodyBudget for callers that would rather render nothing than
// guess: a non-positive height returns 0 instead of the minimum-height
// fallback.
func ContentRows(height int) int {
	if height <= 0 {
		return 0
	}
	rows := height - theme.AppMarginY*2
	if rows < 1 {
		rows = 1
	}
	return rows
}

// TierForHeight maps a terminal height to its Tier using the theme's body
// height thresholds. Unknown (non-positive) heights land on TierMedium via
// BodyBudget's minimum-height fallback.
func TierForHeight(height int) Tier {
	rows := BodyBudget(height)
	switch {
	case rows >= theme.TallBodyHeight-theme.AppMarginY*2:
		return TierTall
	case rows >= theme.MinBodyHeight-theme.AppMarginY*2:
		return TierMedium
	default:
		return TierShort
	}
}

// TierRows holds a per-tier row count for content that scales with screen
// height (for example, how many list rows to show).
type TierRows struct{ Short, Medium, Tall int }

// At returns the row count for the given tier; unknown tiers fall back to
// Short.
func (r TierRows) At(t Tier) int {
	switch t {
	case TierTall:
		return r.Tall
	case TierMedium:
		return r.Medium
	default:
		return r.Short
	}
}
