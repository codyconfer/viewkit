package layout

// CursorRows returns rows with a blank line above and below the cursor row, so
// the selection reads clearly without spacing out the whole list. Rows stay
// tight when there are fewer than two of them, when cursor is out of range, or
// when maxLines (0 for unbounded) cannot fit the two extra lines.
func CursorRows(rows []string, cursor, maxLines int) []string {
	if len(rows) < 2 || cursor < 0 || cursor >= len(rows) {
		return rows
	}
	if maxLines > 0 && len(rows)+2 > maxLines {
		return rows
	}
	out := make([]string, 0, len(rows)+2)
	for i, row := range rows {
		if i == cursor && i > 0 {
			out = append(out, "")
		}
		out = append(out, row)
		if i == cursor && i < len(rows)-1 {
			out = append(out, "")
		}
	}
	return out
}
