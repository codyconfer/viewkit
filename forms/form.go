package forms

import (
	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/panels"
	"github.com/codyconfer/viewkit/theme"
)

type Form struct {
	Fields []Field
	cursor int
}

func NewForm(fields ...Field) *Form {
	return &Form{Fields: fields}
}

func (fm *Form) Focused() *Field {
	if len(fm.Fields) == 0 {
		return nil
	}
	fm.cursor = panels.ClampIndex(fm.cursor, len(fm.Fields))
	return &fm.Fields[fm.cursor]
}

func (fm *Form) Handle(a keys.Action) bool {
	if len(fm.Fields) == 0 {
		return false
	}
	fm.cursor = panels.ClampIndex(fm.cursor, len(fm.Fields))
	fd := &fm.Fields[fm.cursor]

	switch a {
	case keys.Up, keys.FocusPrev:
		fm.cursor = panels.MoveIndex(fm.cursor, -1, len(fm.Fields))
	case keys.Down, keys.FocusNext:
		fm.cursor = panels.MoveIndex(fm.cursor, +1, len(fm.Fields))
	case keys.Left, keys.Dec:
		fd.left()
	case keys.Right, keys.Inc:
		fd.right()
	case keys.Erase:
		fd.backspace()
	case keys.Confirm:
		return fd.activate()
	default:
		return false
	}
	return true
}

func (fm *Form) Insert(s string) {
	if fd := fm.Focused(); fd != nil {
		fd.insert(s)
	}
}

func (fm *Form) Values() map[string]any {
	out := make(map[string]any, len(fm.Fields))
	for i := range fm.Fields {
		out[fm.Fields[i].Key] = fm.Fields[i].Value()
	}
	return out
}

func (fm *Form) Render(f layout.Frame, title string) string {
	return fm.render(f, title, 0)
}

// RenderWindow renders at most maxLines of field rows, scrolled to keep the
// focused field visible. A maxLines of zero renders every field.
func (fm *Form) RenderWindow(f layout.Frame, title string, maxLines int) string {
	return fm.render(f, title, maxLines)
}

func (fm *Form) render(f layout.Frame, title string, maxLines int) string {
	fm.cursor = panels.ClampIndex(fm.cursor, len(fm.Fields))
	lines := []string{""}
	focusStart, focusEnd := 0, 0
	for i := range fm.Fields {
		if i > 0 {
			lines = append(lines, "")
		}
		block := fm.Fields[i].render(f, i == fm.cursor)
		if i == fm.cursor {
			focusStart = len(lines)
			focusEnd = focusStart + len(block)
		}
		lines = append(lines, block...)
	}
	if maxLines > 0 && len(lines) > maxLines {
		lines = windowAround(lines, focusStart, focusEnd, maxLines)
	}
	return f.Panel(title, lines...)
}

const moreMarker = "⋯"

// windowAround returns exactly maxLines of lines containing [focusStart,
// focusEnd), marking clipped edges so the field list does not look complete.
func windowAround(lines []string, focusStart, focusEnd, maxLines int) []string {
	if maxLines < 3 {
		maxLines = 3
	}
	span := focusEnd - focusStart
	start := focusStart - (maxLines-span)/2
	if start < 0 {
		start = 0
	}
	end := start + maxLines
	if end > len(lines) {
		end = len(lines)
		start = end - maxLines
		if start < 0 {
			start = 0
		}
	}
	out := make([]string, 0, end-start)
	out = append(out, lines[start:end]...)
	marker := theme.Cur().Dim.Render(moreMarker)
	if start > 0 {
		out[0] = marker
	}
	if end < len(lines) {
		out[len(out)-1] = marker
	}
	return out
}

func (fm *Form) Overlay(bg string, f layout.Frame, title string, pos ...layout.OverlayPos) string {
	return layout.Overlay(bg, fm.Render(f, title), pos...)
}
