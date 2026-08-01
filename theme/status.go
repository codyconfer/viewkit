package theme

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/codyconfer/viewkit/glyph"
)

// SeverityStyle maps glyph.Severity to this theme's style for it.
func (t Theme) SeverityStyle(s glyph.Severity) lipgloss.Style {
	switch s {
	case glyph.SeverityPositive:
		return t.Can
	case glyph.SeverityNegative:
		return t.Cant
	case glyph.SeverityWarning:
		if len(t.Series) > 2 {
			return t.Series[2]
		}
		return t.Cant
	default:
		return t.Dim
	}
}

// SeverityColor maps glyph.Severity to this theme's terminal color.
func (t Theme) SeverityColor(s glyph.Severity) lipgloss.TerminalColor {
	return t.SeverityStyle(s).GetForeground()
}

func stripBgOf(t Theme) lipgloss.TerminalColor {
	return t.Panel.GetBorderTopForeground()
}

// strip returns the cached strip styles, computing them when the theme was
// built as a struct literal that never passed through New.
func (t Theme) strip() (lipgloss.TerminalColor, lipgloss.Style, lipgloss.Style) {
	if t.stripBg == nil {
		bg := stripBgOf(t)
		text := lipgloss.NewStyle().Background(bg)
		return bg, text, text.Bold(true)
	}
	return t.stripBg, t.stripText, t.stripBold
}

// StripBg returns this theme's status-strip background color.
func (t Theme) StripBg() lipgloss.TerminalColor {
	bg, _, _ := t.strip()
	return bg
}

// StripText renders s in foreground fg on the strip background.
func (t Theme) StripText(fg lipgloss.TerminalColor, s string) string {
	_, text, _ := t.strip()
	return text.Foreground(fg).Render(s)
}

// StripBold renders s bold in foreground fg on the strip background.
func (t Theme) StripBold(fg lipgloss.TerminalColor, s string) string {
	_, _, bold := t.strip()
	return bold.Foreground(fg).Render(s)
}

// StripBlock wraps lines in one strip-background filler line above and below,
// each width columns wide.
func (t Theme) StripBlock(width int, lines ...string) string {
	return PadBlock(t.StripBg(), width, 1, lines...)
}

// Icon renders icon as a fixed-width lead colored by Series[hue], falling
// back to Accent when hue is out of range. Empty icon yields "".
func (t Theme) Icon(icon string, hue int) string {
	if icon == "" {
		return ""
	}
	sty := t.Accent
	if hue >= 0 && hue < len(t.Series) {
		sty = t.Series[hue]
	}
	return sty.Render(glyph.Lead(icon))
}

// IconBlank returns glyph.LeadWidth spaces, aligning rows that have no icon
// with rows rendered by Icon.
func IconBlank() string {
	return strings.Repeat(" ", glyph.LeadWidth)
}
