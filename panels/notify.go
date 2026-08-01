package panels

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/codyconfer/viewkit/glyph"
	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/notify"
	"github.com/codyconfer/viewkit/theme"
)

func severityStyle(sev glyph.Severity) lipgloss.Style {
	t := theme.Cur()
	switch sev {
	case glyph.SeverityPositive:
		return t.NotifPositive
	case glyph.SeverityWarning:
		return t.NotifWarning
	case glyph.SeverityNegative:
		return t.NotifNegative
	default:
		return t.NotifNeutral
	}
}

// NotificationToast renders n as a single severity-styled line — its severity
// glyph, title, and " — message" when the message is non-empty — truncated to
// the frame body width.
func NotificationToast(f layout.Frame, n notify.Notification) string {
	line := glyph.GlyphFor(n.Severity) + " " + n.Title
	if n.Message != "" {
		line += " — " + n.Message
	}
	return severityStyle(n.Severity).Render(f.Fit(line))
}

// NotificationPanel renders a panel listing ns, one severity-glyph-plus-title
// line per notification and an indented dim message line beneath it when the
// message is non-empty. Only the glyph takes the severity color, using the
// severity style's foreground. With no notifications the panel shows a dim
// "no notifications".
func NotificationPanel(f layout.Frame, title string, ns []notify.Notification) string {
	if len(ns) == 0 {
		return f.Panel(title, theme.Cur().Dim.Render("no notifications"))
	}
	t := theme.Cur()
	lines := make([]string, 0, len(ns))
	for _, n := range ns {
		sty := severityStyle(n.Severity)
		head := sty.GetForeground()
		marker := lipgloss.NewStyle().Foreground(head).Render(glyph.GlyphFor(n.Severity) + " ")
		lines = append(lines, f.Fit(marker+t.NotifTitle.Render(n.Title)))
		if n.Message != "" {
			lines = append(lines, f.Fit(t.Dim.Render("  "+n.Message)))
		}
	}
	return f.Panel(title, lines...)
}

// NotificationOverlay composites a NotificationCard for n on top of the
// already-rendered background bg at pos (layout.Overlay's default position
// when omitted).
func NotificationOverlay(bg string, f layout.Frame, n notify.Notification, pos ...layout.OverlayPos) string {
	return layout.Overlay(bg, NotificationCard(f, n), pos...)
}
