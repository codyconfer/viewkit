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

	sugg suggState
}

func NewForm(fields ...Field) *Form {
	fm := &Form{Fields: fields}
	fm.resuggest()
	return fm
}

func (fm *Form) Focused() *Field {
	if len(fm.Fields) == 0 {
		return nil
	}
	fm.cursor = panels.ClampIndex(fm.cursor, len(fm.Fields))
	return &fm.Fields[fm.cursor]
}

// Suggestions lists the completions offered for the focused field, most
// relevant first. It is empty when the field has no Suggester or nothing
// matches what has been typed.
func (fm *Form) Suggestions() []string { return fm.sugg.cands }

// AcceptSuggestion writes the active candidate into the focused field,
// replacing the token being typed. It reports false when there is nothing to
// accept, which lets a host give the accept key a second meaning.
func (fm *Form) AcceptSuggestion() bool {
	pick := fm.sugg.pick()
	if pick == "" {
		return false
	}
	fd := fm.Focused()
	if fd == nil {
		return false
	}
	fd.accept(pick)
	fm.resuggest()
	return true
}

// CycleSuggestion moves through the candidate list. It wraps at both ends
// rather than clamping like field navigation does: a short list is meant to
// be cycled repeatedly, and stopping at the last entry hides the first.
func (fm *Form) CycleSuggestion(delta int) {
	n := len(fm.sugg.cands)
	if n == 0 {
		return
	}
	fm.sugg.idx = ((fm.sugg.idx+delta)%n + n) % n
}

// resuggest recomputes the focused field's candidates and returns to the
// first of them, so a fresh keystroke never leaves a stale ghost on screen.
func (fm *Form) resuggest() {
	fm.sugg = suggState{}
	if fd := fm.Focused(); fd != nil {
		fm.sugg.cands = fd.candidates()
	}
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
		fm.resuggest()
	case keys.Down, keys.FocusNext:
		fm.cursor = panels.MoveIndex(fm.cursor, +1, len(fm.Fields))
		fm.resuggest()
	case keys.Left, keys.Dec:
		fd.left()
	case keys.Right, keys.Inc:
		fd.right()
	case keys.Erase:
		fd.backspace()
		fm.resuggest()
	case keys.CompleteNext:
		fm.CycleSuggestion(+1)
	case keys.CompletePrev:
		fm.CycleSuggestion(-1)
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
		fm.resuggest()
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
		var sg suggState
		if i == fm.cursor {
			sg = fm.sugg
		}
		block := fm.Fields[i].render(f, i == fm.cursor, sg)
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
