package panels

import (
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/notify"
	"github.com/codyconfer/viewkit/theme"
)

// NotificationCard renders n as a multi-line card — styled title over the
// message wrapped to the full frame width — with the severity style applied
// to the whole block. It is the card NotificationOverlay floats over a
// background.
func NotificationCard(f layout.Frame, n notify.Notification) string {
	t := theme.Cur()
	body := lipgloss.JoinVertical(lipgloss.Left,
		t.NotifTitle.Render(n.Title),
		lipgloss.NewStyle().Width(f.Width).Render(n.Message),
	)
	return severityStyle(n.Severity).Render(body)
}

func finite(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	return v
}

// seriesStyle picks the i-th series colour. theme.New always fills Series, but
// theme.Use accepts any Theme literal, and i%len(series) on an empty slice is an
// integer divide by zero, so a Series-less theme falls back to the value style.
func seriesStyle(i int) lipgloss.Style {
	return seriesAt(theme.Cur().Series, i)
}

func seriesAt(series []lipgloss.Style, i int) lipgloss.Style {
	if len(series) == 0 {
		return theme.Cur().Val
	}
	if i < 0 {
		i = 0
	}
	return series[i%len(series)]
}

// ProgressBar renders a width-cell bar filled to frac: accent █ cells then
// dim ░ cells. frac is clamped to [0, 1] (NaN/Inf become 0), the filled
// count truncates rather than rounds, and a negative width yields an empty
// string.
func ProgressBar(frac float64, width int) string {
	if width < 0 {
		width = 0
	}
	frac = finite(frac)
	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}
	filled := min(max(int(frac*float64(width)), 0), width)
	return theme.Cur().Accent.Render(strings.Repeat("█", filled)) + theme.Cur().Dim.Render(strings.Repeat("░", width-filled))
}

// Meter is ProgressBar wrapped in unstyled square brackets, so its rendered
// width is width+2 cells.
func Meter(frac float64, width int) string {
	return "[" + ProgressBar(frac, width) + "]"
}

// MeterWidth caps a desired meter width (in cells) to a third of frameWidth,
// but never below 8 — so on narrow frames it can return more than
// frameWidth/3. Desired widths under 1 return 1.
func MeterWidth(frameWidth, desired int) int {
	if desired < 1 {
		return 1
	}
	max := frameWidth / 3
	if max < 8 {
		max = 8
	}
	if desired > max {
		return max
	}
	return desired
}

// Flash renders a transient status message in dim italics; an empty message
// returns "" so callers can drop the line entirely.
func Flash(message string) string {
	if message == "" {
		return ""
	}
	return theme.Cur().Dim.Italic(true).Render(message)
}

// Toggle renders a two-option switch as "left  /  right" with the active side
// in the accent style and the inactive side in the value style.
func Toggle(left, right string, leftActive bool) string {
	leftSty, rightSty := theme.Cur().Val, theme.Cur().Val
	if leftActive {
		leftSty = theme.Cur().Accent
	} else {
		rightSty = theme.Cur().Accent
	}
	return leftSty.Render(left) + theme.Cur().Dim.Render("  /  ") + rightSty.Render(right)
}

// ClampIndex clamps a selection index into [0, total-1], returning 0 when
// total is not positive.
func ClampIndex(index, total int) int {
	if total <= 0 {
		return 0
	}
	if index < 0 {
		return 0
	}
	if index >= total {
		return total - 1
	}
	return index
}

// MoveIndex moves a selection index by delta, clamped into [0, total-1]; it
// stops at the ends rather than wrapping.
func MoveIndex(index, delta, total int) int {
	return ClampIndex(index+delta, total)
}
