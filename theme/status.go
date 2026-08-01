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

// SeverityGlyph returns the status-strip glyph for s (StatusOK/Warn/Bad/Muted).
func SeverityGlyph(s glyph.Severity) string {
	switch s {
	case glyph.SeverityPositive:
		return glyph.StatusOK()
	case glyph.SeverityNegative:
		return glyph.StatusBad()
	case glyph.SeverityWarning:
		return glyph.StatusWarn()
	default:
		return glyph.StatusMuted()
	}
}

func stripBgOf(t Theme) lipgloss.TerminalColor {
	return t.Panel.GetBorderTopForeground()
}

func StripBg() lipgloss.TerminalColor {
	return Cur().stripBg
}

func StripText(fg lipgloss.TerminalColor, s string) string {
	return Cur().stripText.Foreground(fg).Render(s)
}

func StripBold(fg lipgloss.TerminalColor, s string) string {
	return Cur().stripBold.Foreground(fg).Render(s)
}

func StripBlock(width int, lines ...string) string {
	return PadBlock(StripBg(), width, 1, lines...)
}

func IconBlank() string {
	return strings.Repeat(" ", glyph.LeadWidth)
}

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

func Success(msg string) string {
	th := Cur()
	return th.Can.Render(glyph.Check()) + " " + th.Val.Render(msg)
}

func Bullet(msg string) string {
	th := Cur()
	return th.Accent.Render(glyph.Bullet()) + " " + th.Val.Render(msg)
}
