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
	t := f.Theme()
	body := lipgloss.JoinVertical(lipgloss.Left,
		t.NotifTitle.Render(n.Title),
		lipgloss.NewStyle().Width(f.Width).Render(n.Message),
	)
	return severityStyle(t, n.Severity).Render(body)
}

func finite(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	return v
}

// seriesAt picks the i-th series colour from th. theme.New always fills
// Series, but theme.Use accepts any Theme literal, and i%len(series) on an
// empty slice is an integer divide by zero, so a Series-less theme falls back
// to the value style.
func seriesAt(th theme.Theme, i int) lipgloss.Style {
	if len(th.Series) == 0 {
		return th.Val
	}
	if i < 0 {
		i = 0
	}
	return th.Series[i%len(th.Series)]
}

// ProgressBar renders a width-cell bar filled to frac in th's styles: accent
// █ cells then dim ░ cells. frac is clamped to [0, 1] (NaN/Inf become 0),
// the filled count truncates rather than rounds, and a negative width yields
// an empty string.
func ProgressBar(th theme.Theme, frac float64, width int) string {
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
	return th.Accent.Render(strings.Repeat("█", filled)) + th.Dim.Render(strings.Repeat("░", width-filled))
}

// Meter is ProgressBar wrapped in unstyled square brackets, so its rendered
// width is width+2 cells.
func Meter(th theme.Theme, frac float64, width int) string {
	return "[" + ProgressBar(th, frac, width) + "]"
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

// Flash renders a transient status message in th's dim italics; an empty
// message returns "" so callers can drop the line entirely.
func Flash(th theme.Theme, message string) string {
	if message == "" {
		return ""
	}
	return th.Dim.Italic(true).Render(message)
}

// Toggle renders a two-option switch as "left  /  right" with the active side
// in th's accent style and the inactive side in the value style.
func Toggle(th theme.Theme, left, right string, leftActive bool) string {
	leftSty, rightSty := th.Val, th.Val
	if leftActive {
		leftSty = th.Accent
	} else {
		rightSty = th.Accent
	}
	return leftSty.Render(left) + th.Dim.Render("  /  ") + rightSty.Render(right)
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
