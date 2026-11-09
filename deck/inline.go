package deck

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"

	"github.com/codyconfer/viewkit/console"
	"github.com/codyconfer/viewkit/forms"
	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/ui"
)

// InlineClear is the action that clears an Inline text response. The default
// binding is ctrl+u when not already used by the scoped editor map; override
// it through InlineSpec.Keys.
const InlineClear keys.Action = "inline.clear"

// InlineState identifies the lifecycle of a rendered workflow panel.
type InlineState uint8

const (
	// InlineContent is a caller-supplied panel, with no implied status.
	InlineContent InlineState = iota
	// InlineEditing is a text prompt awaiting a response.
	InlineEditing
	// InlineSubmitted is the final view of a submitted prompt.
	InlineSubmitted
	// InlineCancelled is the final view of a cancelled prompt.
	InlineCancelled
	// InlineError is an error displayed by Report.
	InlineError
)

// InlinePanel is the neutral input to a custom workflow renderer. Title and
// Lines contain sanitized text; Lines are wrapped to the preferred body width.
// Width includes the four columns reserved for default panel chrome.
type InlinePanel struct {
	Title string
	Lines []string
	Width int
	State InlineState
}

// InlineSpec configures a sequential, inline CLI workflow. Domain checks,
// validation, persistence, and external commands belong to the caller.
type InlineSpec struct {
	// Input and Output default to os.Stdin and os.Stdout.
	Input  io.Reader
	Output io.Writer
	// UI defaults to a snapshot of the built-in rendering scope.
	UI *ui.Scope
	// Width is the preferred body width, excluding the four columns of panel
	// chrome. Zero uses layout.DocumentFrame's width. Terminal width caps it.
	Width int
	// Plain disables panels, cursor movement, and interactive editing.
	Plain bool
	// AppendOnly keeps completed panels in scrollback even on terminals.
	// The active text prompt still edits and redraws interactively.
	AppendOnly bool
	// Render replaces the default panel presentation. It must be a pure
	// renderer, fit within InlinePanel.Width, and return no trailing newline.
	// It is not called in Plain mode or on terminals narrower than 5 cells.
	// ANSI styles are allowed; terminal control commands are not. The host
	// strips ANSI unless output supports it or AllowANSI is true.
	Render func(InlinePanel) string
	// AllowANSI preserves renderer styles even for redirected/dumb output
	// (for example, an application's forced-color option). It never enables
	// interaction or cursor movement, and Plain still strips all ANSI.
	AllowANSI bool
	// Keys overrides the text editor map, including submit and cancel. Nil
	// uses text-safe scoped confirm/cancel/quit/erase and ctrl+u to clear.
	Keys *keys.Map
}

// Inline renders one workflow panel at a time, in normal terminal scrollback
// (never the alternate screen). Panel, Ask, and Report replace the preceding
// view only when both streams are terminals and output supports ANSI. Plain,
// dumb-terminal, and redirected streams are append-only transcripts.
//
// An Inline owns its terminal region. Use it sequentially from one goroutine,
// do not nest it inside a deck session, and call Release before any other code
// writes to its output. It does not install global theme or key state.
type Inline struct {
	in          io.Reader
	out         io.Writer
	lines       *bufio.Reader
	scope       *ui.Scope
	keyMap      *keys.Map
	width       int
	plain       bool
	appendOnly  bool
	renderPanel func(InlinePanel) string
	interactive bool
	ansi        bool
	title       string
	body        []string
	rows        int
	lastWidth   int
	size        func() (int, int)
}

// NewInline constructs a workflow without reading input or writing output.
func NewInline(spec InlineSpec) *Inline {
	in, out := spec.Input, spec.Output
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}
	scope := spec.UI
	if scope == nil {
		scope = ui.Default()
	}
	caps := console.Detect(out)
	u := &Inline{
		in: in, out: out, lines: bufio.NewReader(in), scope: scope,
		keyMap: spec.Keys, width: layout.NewFrame(spec.Width).Width,
		plain: spec.Plain, ansi: (caps.ANSI || spec.AllowANSI) && !spec.Plain,
		appendOnly: spec.AppendOnly, renderPanel: spec.Render,
		interactive: inlineTerminal(in) && caps.ANSI && !spec.Plain,
	}
	u.size = func() (int, int) {
		if f, ok := out.(interface{ Fd() uintptr }); ok {
			w, h, err := term.GetSize(f.Fd())
			if err == nil {
				return w, h
			}
		}
		return 0, 0
	}
	return u
}

func inlineTerminal(v any) bool {
	f, ok := v.(interface{ Fd() uintptr })
	return ok && term.IsTerminal(f.Fd())
}

// inlineText treats caller strings as text, not terminal commands. Newlines
// are retained so multiline instructions and errors still wrap naturally.
func inlineText(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' {
			return ' '
		}
		return r
	}, ansi.Strip(s))
}

func (u *Inline) frame(width int) layout.Frame {
	bodyWidth := u.width
	if width > 0 {
		bodyWidth = min(bodyWidth, max(1, width-4))
	}
	// Do not use WithWidth here: its minimum-width clamp can exceed a narrow
	// terminal. CellBox is the layout helper with clamp-down semantics.
	return layout.Frame{Width: bodyWidth + 4, UI: u.scope}
}

func (u *Inline) render(title string, body []string, width int, state InlineState) string {
	f := u.frame(width)
	bodyWidth := max(1, f.Width-4)
	if width > 0 && width < 5 {
		bodyWidth = width
	}
	lines := make([]string, len(body))
	for i, line := range body {
		lines[i] = inlineWrap(inlineText(line), bodyWidth)
	}
	var text string
	if u.plain || (width > 0 && width < 5) {
		titleWidth := u.width
		if width > 0 {
			titleWidth = min(titleWidth, width)
		}
		text = inlineWrap(inlineText(title), titleWidth) + "\n" + strings.Join(lines, "\n")
	} else if u.renderPanel != nil {
		text = u.renderPanel(InlinePanel{Title: inlineText(title), Lines: lines, Width: f.Width, State: state})
	} else {
		text = f.CellBox(inlineText(title), lines...)
	}
	if !u.ansi {
		text = ansi.Strip(text)
	}
	return text
}

func inlineWrap(text string, width int) string {
	wrapped := ansi.Hardwrap(ansi.Wrap(text, width, ""), width, false)
	lines := strings.Split(wrapped, "\n")
	for i, line := range lines {
		// A single wide grapheme cannot fit a one-cell body. Never allow it
		// to force a terminal wrap and invalidate our owned-region height.
		lines[i] = ansi.Truncate(line, width, "")
	}
	return strings.Join(lines, "\n")
}

func (u *Inline) clear(width, height int) error {
	if u.interactive && !u.appendOnly && u.rows > 0 && width == u.lastWidth {
		rows := u.rows
		if height > 0 {
			rows = min(rows, height-1)
		}
		if rows > 0 {
			if _, err := fmt.Fprintf(u.out, "\r\x1b[%dA\x1b[J", rows); err != nil {
				return err
			}
		}
	}
	// A terminal width change can reflow previous output. Keep that view as
	// transcript rather than guessing and erasing unrelated scrollback.
	u.rows = 0
	return nil
}

// Panel replaces the current view, or appends it to a transcript. All strings
// are untrusted text: terminal control sequences are stripped. A matching Ask
// retains these instructions inside the same panel.
func (u *Inline) Panel(title string, lines ...string) error {
	return u.panel(InlineContent, title, lines...)
}

func (u *Inline) panel(state InlineState, title string, lines ...string) error {
	width, height := u.size()
	text := u.render(title, lines, width, state)
	if err := u.clear(width, height); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(u.out, text); err != nil {
		return err
	}
	u.title, u.body = title, append([]string(nil), lines...)
	u.rows, u.lastWidth = strings.Count(text, "\n")+1, width
	return nil
}

// Ask collects one text response in the current panel, retaining instructions
// when title matches the preceding Panel. The value is an editable default,
// not a secret: it is displayed in clear text. Cancellation returns
// ErrCancelled. Scripted input accepts an empty line as the default and a final
// nonempty line without a newline. EOF without an answer returns io.EOF.
//
// Context cancellation interrupts the interactive program. Arbitrary fallback
// readers cannot be interrupted: ctx is checked before and after a line read.
func (u *Inline) Ask(ctx context.Context, title, label, value string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	width, height := u.size()
	var body []string
	if u.title == title {
		body = append([]string(nil), u.body...)
	}
	value = strings.ReplaceAll(inlineText(value), "\n", " ")
	m := &inlineModel{
		owner: u, form: forms.NewForm(forms.Text("value", label).WithText(value)),
		title: title, label: label, body: body, width: width,
	}
	if u.interactive {
		if err := u.clear(width, height); err != nil {
			return "", err
		}
		_, err := tea.NewProgram(m, tea.WithContext(ctx), tea.WithInput(u.in), tea.WithOutput(u.out)).Run()
		u.title, u.body = title, body
		u.rows, u.lastWidth = strings.Count(m.View(), "\n"), m.width
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		if err != nil {
			return "", err
		}
		if !m.submitted {
			return "", ErrCancelled
		}
		return m.form.Focused().Text, nil
	}
	// No Tea program or cursor escapes for transcripts. Reuse one buffered
	// reader across prompts so read-ahead remains available to the next step.
	if err := u.panel(InlineEditing, title, append(body, inlineText(label)+" ["+value+"]:")...); err != nil {
		return "", err
	}
	// A prompt is not an instruction for the following Ask/Report.
	u.body = body
	answer, err := u.lines.ReadString('\n')
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	if err != nil && (!errors.Is(err, io.EOF) || len(answer) == 0) {
		return "", err
	}
	answer = strings.TrimSpace(answer)
	if answer == "" {
		answer = value
	}
	return answer, nil
}

// Release relinquishes the current terminal region without erasing it. Call
// before external command output or other writers; the next Panel must not
// erase that output. Instructions are forgotten as well.
func (u *Inline) Release() {
	u.rows = 0
	u.title, u.body = "", nil
}

// Reader returns the input stream for an external command after Release. For
// transcripts it includes any bytes buffered by Ask. Do not bypass this reader
// when handing scripted input to a child process.
func (u *Inline) Reader() io.Reader {
	if u.interactive {
		return u.in
	}
	return u.lines
}

type inlineReportedError struct{ err error }

func (e inlineReportedError) Error() string { return e.err.Error() }
func (e inlineReportedError) Unwrap() error { return e.err }

// ErrorReported reports whether err (possibly wrapped) has already been
// displayed by Inline.Report. CLIs can suppress a duplicate stderr message
// while retaining their nonzero exit status and errors.Is/errors.As identity.
func ErrorReported(err error) bool {
	var reported inlineReportedError
	return errors.As(err, &reported)
}

// Report displays err inside the current panel and returns a marked error
// preserving its identity. Nil and already-reported errors are unchanged. If
// rendering fails, both errors are returned, without marking either reported.
// Error strings are shown verbatim as text; callers must redact secrets first.
func (u *Inline) Report(err error) error {
	if err == nil || ErrorReported(err) {
		return err
	}
	title := u.title
	if title == "" {
		title = "Error"
	}
	body := append(append([]string(nil), u.body...), err.Error())
	if renderErr := u.panel(InlineError, title, body...); renderErr != nil {
		return errors.Join(err, renderErr)
	}
	return inlineReportedError{err: err}
}

func (u *Inline) editorMap() *keys.Map {
	if u.keyMap != nil {
		return u.keyMap
	}
	bindings := u.scope.Keys.EditorBindings(keys.Confirm, keys.Cancel, keys.Quit, keys.Erase)
	textBindings := make([]keys.Binding, 0, len(bindings)+1)
	for _, binding := range bindings {
		// "spacebar" is a printable-space alias in some schemes. Spaces
		// must remain text, and hints must name only keys we dispatch.
		var control []string
		for _, key := range binding.Keys {
			if key != "spacebar" {
				control = append(control, key)
			}
		}
		if len(control) > 0 {
			binding.Keys = control
			binding.Glyph = strings.Join(control, "/")
			textBindings = append(textBindings, binding)
		}
	}
	km := keys.NewMap(textBindings...)
	if _, bound := km.Action("ctrl+u"); bound {
		// A scoped control binding wins over the optional clear shortcut.
		return km
	}
	textBindings = append(textBindings, keys.Binding{Action: InlineClear, Keys: []string{"ctrl+u"}, Glyph: "ctrl+u"})
	return keys.NewMap(textBindings...)
}

type inlineModel struct {
	owner                *Inline
	form                 *forms.Form
	title, label         string
	body                 []string
	width                int
	cancelled, submitted bool
}

func (m *inlineModel) Init() tea.Cmd { return nil }

func (m *inlineModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch key := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = key.Width
	case tea.KeyMsg:
		act, bound := m.owner.editorMap().Action(key.String())
		if bound {
			switch act {
			case keys.Cancel, keys.Quit:
				m.cancelled = true
				return m, tea.Quit
			case keys.Confirm:
				m.submitted = true
				return m, tea.Quit
			case InlineClear:
				m.form.Focused().Text = ""
			case keys.Erase:
				m.form.Handle(keys.Erase)
			}
		} else if key.Type == tea.KeyRunes {
			m.form.Insert(strings.ReplaceAll(inlineText(string(key.Runes)), "\n", " "))
		} else if key.Type == tea.KeySpace {
			m.form.Insert(" ")
		}
	}
	return m, nil
}

func (m *inlineModel) View() string {
	if m.cancelled {
		return m.owner.render(m.title, []string{"Cancelled."}, m.width, InlineCancelled) + "\n"
	}
	lines := append([]string(nil), m.body...)
	lines = append(lines, "", inlineText(m.label)+": "+m.form.Focused().Text, "")
	if m.submitted {
		lines = append(lines, "Saved response.")
	} else {
		km := m.owner.editorMap()
		var hints []string
		for _, item := range []struct {
			action keys.Action
			label  string
		}{{keys.Confirm, "submit"}, {InlineClear, "clear"}, {keys.Cancel, "cancel"}, {keys.Quit, "cancel"}} {
			if km.Has(item.action) {
				h := km.HintLabeled(item.action, item.label)
				hints = append(hints, h.Key+" "+h.Label)
			}
		}
		lines = append(lines, strings.Join(hints, " · "))
	}
	state := InlineEditing
	if m.submitted {
		state = InlineSubmitted
	}
	return m.owner.render(m.title, lines, m.width, state) + "\n"
}
