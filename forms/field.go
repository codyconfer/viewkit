package forms

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/charmbracelet/x/ansi"

	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/panels"
	"github.com/codyconfer/viewkit/theme"
)

// FieldKind selects the input widget a Field presents and which of its state
// fields carry the value.
type FieldKind int

const (
	// FieldText is a single-line text input; the value lives in Text.
	FieldText FieldKind = iota

	// FieldMultiline is a text input that also accepts newlines.
	FieldMultiline

	// FieldSelect picks one of Options, cycled in place with left/right.
	FieldSelect

	// FieldMultiselect toggles any subset of Options via Checked.
	FieldMultiselect

	// FieldRadio picks one of Options, listed vertically with radio marks.
	FieldRadio

	// FieldToggle is an on/off switch backed by On.
	FieldToggle
)

// Field is a single input in a Form. Key, Label, and Kind describe it; the
// remaining fields hold per-kind state and can be pre-set to seed a value.
type Field struct {
	// Key names the field's entry in Form.Values.
	Key string
	// Label is the caption rendered next to (or above) the input.
	Label string
	// Kind selects the widget; the zero value is FieldText.
	Kind FieldKind

	// Options are the choices for select, multiselect, and radio fields.
	Options []string

	// Text is the current content of text and multiline fields.
	Text string
	// On is the state of a toggle field.
	On bool
	// Selected indexes Options: the value of select and radio fields, and
	// the highlight cursor of multiselect fields.
	Selected int
	// Checked marks which Options indices a multiselect field has on.
	Checked map[int]bool

	// Secret masks text as bullets when rendering and disables suggestions.
	Secret bool

	// Suggest proposes completions for the token being typed. Text and
	// multiline fields only; a nil Suggester disables completion.
	Suggest Suggester
	// Delim splits the field into independently completed tokens, e.g. ","
	// for a comma-separated list or " " for a search expression. Empty means
	// the whole text is one token.
	Delim string
}

func (fd *Field) candidates() []string {
	if fd.Suggest == nil || fd.Secret {
		return nil
	}
	if fd.Kind != FieldText && fd.Kind != FieldMultiline {
		return nil
	}
	_, tail := splitTail(fd.Text, fd.Delim)
	return fd.Suggest(tail)
}

func (fd *Field) accept(pick string) {
	head, _ := splitTail(fd.Text, fd.Delim)
	fd.Text = joinTail(head, pick, fd.Delim)
}

// Value returns the field's current value by kind: bool for toggles, the
// selected option for select and radio ("" when Selected is out of range),
// []string of checked options for multiselect, and the raw Text otherwise.
func (fd *Field) Value() any {
	switch fd.Kind {
	case FieldToggle:
		return fd.On
	case FieldSelect, FieldRadio:
		if fd.Selected >= 0 && fd.Selected < len(fd.Options) {
			return fd.Options[fd.Selected]
		}
		return ""
	case FieldMultiselect:
		out := make([]string, 0, len(fd.Checked))
		for i, opt := range fd.Options {
			if fd.Checked[i] {
				out = append(out, opt)
			}
		}
		return out
	default:
		return fd.Text
	}
}

func (fd *Field) insert(s string) {
	if fd.Kind != FieldText && fd.Kind != FieldMultiline {
		return
	}
	var b strings.Builder
	for _, r := range s {
		if r == '\n' && fd.Kind == FieldMultiline {
			b.WriteRune(r)
			continue
		}
		if unicode.IsPrint(r) {
			b.WriteRune(r)
		}
	}
	fd.Text += b.String()
}

func (fd *Field) backspace() {
	if fd.Text == "" {
		return
	}
	r := []rune(fd.Text)
	fd.Text = string(r[:len(r)-1])
}

func (fd *Field) left() {
	switch fd.Kind {
	case FieldSelect, FieldRadio, FieldMultiselect:
		fd.Selected = panels.MoveIndex(fd.Selected, -1, len(fd.Options))
	case FieldToggle:
		fd.On = false
	}
}

func (fd *Field) right() {
	switch fd.Kind {
	case FieldSelect, FieldRadio, FieldMultiselect:
		fd.Selected = panels.MoveIndex(fd.Selected, +1, len(fd.Options))
	case FieldToggle:
		fd.On = true
	}
}

func (fd *Field) activate() bool {
	switch fd.Kind {
	case FieldMultiselect:
		if fd.Checked == nil {
			fd.Checked = map[int]bool{}
		}
		fd.Checked[fd.Selected] = !fd.Checked[fd.Selected]
		return true
	case FieldToggle:
		fd.On = !fd.On
		return true
	}
	return false
}

type suggState struct {
	cands []string
	idx   int
}

func (s suggState) pick() string {
	if s.idx < 0 || s.idx >= len(s.cands) {
		return ""
	}
	return s.cands[s.idx]
}

const suggListMax = 5

func fieldLabel(f layout.Frame, text string, focused bool) string {
	t := f.Theme()
	if focused {
		return f.Cursor(true) + t.Accent.Render(text)
	}
	return f.Cursor(false) + t.Dim.Render(text)
}

const (
	fieldRowChrome = 4
	fieldValueMin  = 8
)

func fieldRowBudget(body, label int) (labelW, valW int) {
	room := max(body-fieldRowChrome, 2)
	labelW = max(label, 1)
	if labelW > room-fieldValueMin {
		labelW = max(room-fieldValueMin, room/2)
	}
	return labelW, room - labelW
}

func (fd *Field) render(f layout.Frame, focused bool, sg suggState) []string {
	t := f.Theme()
	label := fieldLabel(f, fd.Label, focused)

	switch fd.Kind {
	case FieldToggle:
		return []string{label + "  " + panels.Toggle(t, "on", "off", fd.On)}

	case FieldSelect:
		return []string{label + "  " + selectGlyph(t, fd, focused)}

	case FieldRadio:
		lines := []string{label}
		for i, opt := range fd.Options {
			mark := t.Dim.Render("( ) ")
			if i == fd.Selected {
				mark = t.Accent.Render("(•) ")
			}
			lines = append(lines, "  "+mark+t.Val.Render(f.Fit(opt)))
		}
		return lines

	case FieldMultiselect:
		lines := []string{label}
		for i, opt := range fd.Options {
			box := t.Dim.Render("[ ] ")
			if fd.Checked[i] {
				box = t.Can.Render("[x] ")
			}
			cursor := "  "
			if focused && i == fd.Selected {
				cursor = f.Cursor(true)
			}
			lines = append(lines, cursor+box+t.Val.Render(f.Fit(opt)))
		}
		return lines

	case FieldMultiline:
		lines := []string{label}
		body := fd.display()
		if focused {
			body += "▎" + fd.ghost(t, sg)
		}
		if body == "" {
			body = t.Dim.Render("…")
		}
		for _, ln := range strings.Split(body, "\n") {
			lines = append(lines, "  "+t.Val.Render(f.Fit(ln)))
		}
		return append(lines, fd.suggestions(f, focused, sg)...)

	default:
		labelW, valW := fieldRowBudget(f.BodyWidth(), ansi.StringWidth(fd.Label))
		if labelW < ansi.StringWidth(fd.Label) {
			label = fieldLabel(f, ansi.Truncate(fd.Label, labelW, "…"), focused)
		}
		val := ansi.Truncate(fd.display(), valW, "…")
		if focused {
			val = fd.caretRow(t, valW, sg)
		}
		shown := t.Val.Render(val)
		if fd.Text == "" && !focused {
			shown = t.Dim.Render("…")
		}
		return append([]string{label + "  " + shown}, fd.suggestions(f, focused, sg)...)
	}
}

func (fd *Field) caretRow(th theme.Theme, valW int, sg suggState) string {
	body := ansi.Truncate(fd.display(), max(valW-1, 1), "…")
	room := valW - ansi.StringWidth(body) - 1
	if room <= 0 {
		return body + "▎"
	}
	return body + "▎" + ansi.Truncate(fd.ghost(th, sg), room, "")
}

func (fd *Field) ghost(th theme.Theme, sg suggState) string {
	pick := sg.pick()
	if pick == "" || fd.Secret {
		return ""
	}
	_, tail := splitTail(fd.Text, fd.Delim)
	rest := ghostOf(pick, tail)
	if rest == "" {
		return ""
	}
	return th.Dim.Render(rest)
}

func (fd *Field) suggestions(f layout.Frame, focused bool, sg suggState) []string {
	if !focused || len(sg.cands) == 0 {
		return nil
	}
	t := f.Theme()
	shown := min(len(sg.cands), suggListMax)
	out := make([]string, 0, shown+1)
	for i := range shown {
		style := t.Dim
		mark := "  "
		if i == sg.idx {
			style, mark = t.Val, f.Cursor(true)
		}
		out = append(out, "  "+mark+style.Render(f.Fit(sg.cands[i])))
	}
	if rest := len(sg.cands) - shown; rest > 0 {
		out = append(out, "    "+t.Dim.Render(moreMarker+" +"+strconv.Itoa(rest)))
	}
	return out
}

func (fd *Field) display() string {
	if !fd.Secret {
		return fd.Text
	}
	var b strings.Builder
	for _, r := range fd.Text {
		if r == '\n' {
			b.WriteRune('\n')
		} else {
			b.WriteRune('•')
		}
	}
	return b.String()
}

func selectGlyph(t theme.Theme, fd *Field, focused bool) string {
	cur := ""
	if fd.Selected >= 0 && fd.Selected < len(fd.Options) {
		cur = fd.Options[fd.Selected]
	}
	arrow := t.Dim
	if focused {
		arrow = t.Accent
	}
	return arrow.Render("◂ ") + t.Val.Render(cur) + arrow.Render(" ▸")
}
