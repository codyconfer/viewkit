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

type FieldKind int

const (
	FieldText FieldKind = iota

	FieldMultiline

	FieldSelect

	FieldMultiselect

	FieldRadio

	FieldToggle
)

type Field struct {
	Key   string
	Label string
	Kind  FieldKind

	Options []string

	Text     string
	On       bool
	Selected int
	Checked  map[int]bool

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

func (fd *Field) render(f layout.Frame, focused bool, sg suggState) []string {
	t := theme.Cur()
	label := layout.Cursor(false) + t.Dim.Render(fd.Label)
	if focused {
		label = layout.Cursor(true) + t.Accent.Render(fd.Label)
	}

	switch fd.Kind {
	case FieldToggle:
		return []string{label + "  " + panels.Toggle("on", "off", fd.On)}

	case FieldSelect:
		return []string{label + "  " + selectGlyph(fd, focused)}

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
				cursor = layout.Cursor(true)
			}
			lines = append(lines, cursor+box+t.Val.Render(f.Fit(opt)))
		}
		return lines

	case FieldMultiline:
		lines := []string{label}
		body := fd.display()
		if focused {
			body += "▎" + fd.ghost(sg)
		}
		if body == "" {
			body = t.Dim.Render("…")
		}
		for _, ln := range strings.Split(body, "\n") {
			lines = append(lines, "  "+t.Val.Render(f.Fit(ln)))
		}
		return append(lines, fd.suggestions(f, focused, sg)...)

	default:
		val := fd.display()
		if focused {
			val += "▎" + fd.ghost(sg)
		}
		shown := t.Val.Render(ansi.Truncate(val, f.BodyWidth()-ansi.StringWidth(fd.Label)-4, "…"))
		if fd.Text == "" && !focused {
			shown = t.Dim.Render("…")
		}
		return append([]string{label + "  " + shown}, fd.suggestions(f, focused, sg)...)
	}
}

func (fd *Field) ghost(sg suggState) string {
	pick := sg.pick()
	if pick == "" || fd.Secret {
		return ""
	}
	_, tail := splitTail(fd.Text, fd.Delim)
	rest := ghostOf(pick, tail)
	if rest == "" {
		return ""
	}
	return theme.Cur().Dim.Render(rest)
}

func (fd *Field) suggestions(f layout.Frame, focused bool, sg suggState) []string {
	if !focused || len(sg.cands) == 0 {
		return nil
	}
	t := theme.Cur()
	shown := min(len(sg.cands), suggListMax)
	out := make([]string, 0, shown+1)
	for i := range shown {
		style := t.Dim
		mark := "  "
		if i == sg.idx {
			style, mark = t.Val, layout.Cursor(true)
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

func selectGlyph(fd *Field, focused bool) string {
	t := theme.Cur()
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
