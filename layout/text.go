package layout

import "strings"

// Lines splits s into display lines. Every trailing newline is dropped, so a
// block ending in "\n" yields no spurious empty last line; interior blank
// lines are preserved. Input is assumed to use "\n" endings, not CRLF.
func Lines(s string) []string {
	return strings.Split(strings.TrimRight(s, "\n"), "\n")
}

// FirstLine returns the first non-empty line of s with surrounding whitespace
// trimmed, or "" when s holds no text. It measures nothing — pair it with Fit
// when the result also needs to respect a width budget.
func FirstLine(s string) string {
	for _, line := range Lines(s) {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

// DialogWidth returns the width for a modal drawn over a screen of the given
// width: inset from both edges, but never wider than DialogMaxWidth.
func DialogWidth(screenWidth int) int {
	return max(min(screenWidth-dialogInset, DialogMaxWidth), 1)
}

// DialogMaxWidth caps how wide DialogWidth will grow on a wide screen.
const DialogMaxWidth = 56

const dialogInset = 8
