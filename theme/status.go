package theme

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/codyconfer/viewkit/glyph"
)

// SeverityStyle maps glyph.Severity to the active theme's style for that
// severity. Use it when rendering text; use SeverityColor when only the
// foreground color is needed.
func SeverityStyle(s glyph.Severity) lipgloss.Style {
	th := Cur()
	switch s {
	case glyph.SeverityPositive:
		return th.Can
	case glyph.SeverityNegative:
		return th.Cant
	case glyph.SeverityWarning:
		if len(th.Series) > 2 {
			return th.Series[2]
		}
		return th.Cant
	default:
		return th.Dim
	}
}

// SeverityColor maps glyph.Severity to the active theme's terminal color.
// glyph.Severity is the sole severity vocabulary; theme only supplies color.
func SeverityColor(s glyph.Severity) lipgloss.TerminalColor {
	return SeverityStyle(s).GetForeground()
}

func stripBgOf(t Theme) lipgloss.TerminalColor {
	return t.Panel.GetBorderTopForeground()
}

// StripBg returns the background color for status strips: the active theme's
// panel border color, cached when the theme is installed.
func StripBg() lipgloss.TerminalColor {
	return Cur().stripBg
}

// StripText renders s in foreground fg on the strip background.
func StripText(fg lipgloss.TerminalColor, s string) string {
	return Cur().stripText.Foreground(fg).Render(s)
}

// StripBold renders s bold in foreground fg on the strip background.
func StripBold(fg lipgloss.TerminalColor, s string) string {
	return Cur().stripBold.Foreground(fg).Render(s)
}

// StripBlock wraps lines in one strip-background filler line above and below,
// each width columns wide.
func StripBlock(width int, lines ...string) string {
	return PadBlock(StripBg(), width, 1, lines...)
}

// IconBlank returns glyph.LeadWidth spaces, aligning rows that have no icon
// with rows rendered by Icon.
func IconBlank() string {
	return strings.Repeat(" ", glyph.LeadWidth)
}

// Icon renders icon as a fixed-width lead colored by the theme's Series style
// at index hue, falling back to Accent when hue is out of range. An empty
// icon yields an empty string.
func Icon(icon string, hue int) string {
	if icon == "" {
		return ""
	}
	th := Cur()
	sty := th.Accent
	if hue >= 0 && hue < len(th.Series) {
		sty = th.Series[hue]
	}
	return sty.Render(glyph.Lead(icon))
}

// Success renders a success-colored check mark followed by msg in the body
// text style.
func Success(msg string) string {
	th := Cur()
	return th.Can.Render(glyph.Check()) + " " + th.Val.Render(msg)
}

// Bullet renders an accent-colored bullet followed by msg in the body text
// style.
func Bullet(msg string) string {
	th := Cur()
	return th.Accent.Render(glyph.Bullet()) + " " + th.Val.Render(msg)
}
