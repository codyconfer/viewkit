package theme

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const resetSeq = "\x1b[0m"

func AppMargin(body string) string {
	return lipgloss.NewStyle().Margin(0, AppMarginX).Render(body)
}

func FillLine(c lipgloss.TerminalColor, width int) string {
	return lipgloss.NewStyle().Background(c).Render(strings.Repeat(" ", width))
}

func PadBlock(c lipgloss.TerminalColor, width, rows int, lines ...string) string {
	pad := FillLine(c, width)
	out := make([]string, 0, len(lines)+2*rows)
	for range rows {
		out = append(out, pad)
	}
	out = append(out, lines...)
	for range rows {
		out = append(out, pad)
	}
	return strings.Join(out, "\n")
}

func Screen(body string, width, height int) string {
	bg := Cur().Bg
	if bg == "" {
		return body
	}
	if seq := bgSeq(bg); seq != "" {
		body = strings.ReplaceAll(body, resetSeq, resetSeq+seq)
	}
	return lipgloss.NewStyle().Background(bg).Width(width).Height(height).Render(body)
}

func bgSeq(c lipgloss.Color) string {
	s := lipgloss.NewStyle().Background(c).Render("\x00")
	if i := strings.IndexByte(s, 0); i > 0 {
		return s[:i]
	}
	return ""
}
