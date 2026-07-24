package theme

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/codyconfer/viewkit/glyph"
)

type Severity int

const (
	SevMuted Severity = iota
	SevOK
	SevWarn
	SevBad
)

func SeverityColor(s Severity) lipgloss.TerminalColor {
	th := Cur()
	switch s {
	case SevOK:
		return th.Can.GetForeground()
	case SevBad:
		return th.Cant.GetForeground()
	case SevWarn:
		if len(th.Series) > 2 {
			return th.Series[2].GetForeground()
		}
		return th.Cant.GetForeground()
	default:
		return th.Dim.GetForeground()
	}
}

func SeverityGlyph(s Severity) string {
	switch s {
	case SevOK:
		return glyph.StatusOK()
	case SevBad:
		return glyph.StatusBad()
	case SevWarn:
		return glyph.StatusWarn()
	default:
		return glyph.StatusMuted()
	}
}

func StripBg() lipgloss.TerminalColor {
	if bg := Cur().Panel.GetBorderTopForeground(); bg != nil {
		return bg
	}
	return Cur().Dim.GetForeground()
}

func StripText(fg lipgloss.TerminalColor, s string) string {
	return lipgloss.NewStyle().Background(StripBg()).Foreground(fg).Render(s)
}

func StripBold(fg lipgloss.TerminalColor, s string) string {
	return lipgloss.NewStyle().Background(StripBg()).Foreground(fg).Bold(true).Render(s)
}

func StripBlock(width int, lines ...string) string {
	return PadBlock(StripBg(), width, 1, lines...)
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
