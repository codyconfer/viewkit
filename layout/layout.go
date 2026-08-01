package layout

import (
	"strings"

	"github.com/codyconfer/viewkit/theme"
)

// ViewportLayout scrolls body inside rows of height while keeping a sticky
// footer pinned to the bottom, hint lines styled with the built-in default
// theme. The footer is whatever follows the last blank line (see
// SplitStickyFooter); only the content above it scrolls with offset. If the
// footer alone fills the budget the content is dropped entirely, and rows <= 0
// returns the body unscrolled.
func ViewportLayout(body string, rows, offset int) string {
	if rows <= 0 {
		return body
	}

	content, footer := SplitStickyFooter(body)
	if footer == "" {
		return viewportIn(theme.Default(), body, rows, offset)
	}

	footerRows := CountLines(footer)
	if footerRows >= rows {
		return viewportIn(theme.Default(), footer, rows, 0)
	}

	contentRows := rows - footerRows
	separator := ""
	if content != "" && contentRows > 0 {
		contentRows--
		separator = "\n\n"
	}

	contentView := ""
	if contentRows > 0 && content != "" {
		contentView = viewportIn(theme.Default(), content, contentRows, offset)
		contentView = PadLines(contentView, contentRows)
	}

	switch {
	case contentView == "":
		return PadLines("", rows-footerRows) + footer
	case separator == "":
		return contentView + footer
	default:
		return contentView + separator + footer
	}
}

// ScrollableBody returns the part of body that ViewportLayout would scroll:
// the content above the sticky footer. A body with no footer scrolls whole
// and is returned as is; if the footer leaves no room to scroll it returns "".
func ScrollableBody(body string, rows int) string {
	content, footer := SplitStickyFooter(body)
	if footer == "" {
		return body
	}
	if ScrollableRows(body, rows) < 1 {
		return ""
	}
	return content
}

// ScrollableRows returns how many of the given rows ViewportLayout would give
// the scrolling content once the sticky footer (and the separator line before
// it) is paid for. Callers use it to clamp scroll offsets.
func ScrollableRows(body string, rows int) int {
	if rows <= 0 {
		return 0
	}

	content, footer := SplitStickyFooter(body)
	if footer == "" {
		return rows
	}

	footerRows := CountLines(footer)
	if footerRows >= rows {
		return 0
	}

	contentRows := rows - footerRows
	if content != "" && contentRows > 0 {
		contentRows--
	}
	return max(contentRows, 0)
}

// SplitStickyFooter splits body at its *last* blank line: everything after it
// is the sticky footer, everything before it the scrollable content. A body
// with no blank line has no footer.
func SplitStickyFooter(body string) (content, footer string) {
	idx := strings.LastIndex(body, "\n\n")
	if idx < 0 {
		return body, ""
	}
	return body[:idx], body[idx+2:]
}

// CountLines reports how many display lines s occupies. Unlike
// strings.Count+1 conventions elsewhere, an empty string still counts as one
// line, and a trailing newline adds a final (empty) line.
func CountLines(s string) int {
	lines := 0
	for range strings.SplitSeq(s, "\n") {
		lines++
	}
	if lines == 0 {
		return 1
	}
	return lines
}

// PadLines pads body with trailing newlines until it spans rows display lines.
// It never trims: a body already at or over the budget is returned unchanged.
// rows <= 0 returns "".
func PadLines(body string, rows int) string {
	if rows <= 0 {
		return ""
	}
	if body == "" {
		return strings.Repeat("\n", max(rows-1, 0))
	}

	lines := CountLines(body)
	if lines >= rows {
		return body
	}

	var b strings.Builder
	if body != "" {
		b.WriteString(body)
	}
	for i := lines; i < rows; i++ {
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}
