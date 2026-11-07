package deck

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/codyconfer/viewkit/forms"
	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/ui"
)

func TestInlinePanelReplacement(t *testing.T) {
	for _, mode := range []string{"interactive", "plain", "redirected"} {
		t.Run(mode, func(t *testing.T) {
			var out bytes.Buffer
			u := NewInline(InlineSpec{Input: strings.NewReader(""), Output: &out, Plain: mode == "plain"})
			u.interactive = mode == "interactive"
			u.ansi = u.interactive
			u.size = func() (int, int) { return 60, 30 }
			if err := u.Panel("First", "first body"); err != nil {
				t.Fatal(err)
			}
			rows := u.rows
			out.Reset()
			if err := u.Panel("Next", "next body"); err != nil {
				t.Fatal(err)
			}
			wantPrefix := fmt.Sprintf("\r\x1b[%dA\x1b[J", rows)
			if strings.HasPrefix(out.String(), wantPrefix) != u.interactive {
				t.Fatalf("unexpected replacement: %q", out.String())
			}
			if !u.interactive && strings.Contains(out.String(), "\x1b") {
				t.Fatalf("escape in transcript: %q", out.String())
			}
			if !strings.Contains(out.String(), "Next") || !strings.Contains(out.String(), "next body") {
				t.Fatal(out.String())
			}
			if strings.Contains(out.String(), "First") {
				t.Fatal("previous panel repeated")
			}
			if strings.Contains(out.String(), "╭") == u.plain {
				t.Fatal("incorrect plain mode")
			}
		})
	}
}

func TestInlineRegionBoundsAndHandoff(t *testing.T) {
	var out bytes.Buffer
	u := NewInline(InlineSpec{Input: strings.NewReader(""), Output: &out})
	u.interactive, u.ansi = true, true
	width, height := 60, 4
	u.size = func() (int, int) { return width, height }
	u.Panel("Tall", strings.Repeat("line\n", 15))
	out.Reset()
	u.Panel("Next", "body")
	if !strings.HasPrefix(out.String(), "\r\x1b[3A\x1b[J") {
		t.Fatalf("clearing must stay within the screen: %q", out.String())
	}
	width = 40
	out.Reset()
	u.Panel("Resized", "body")
	if strings.Contains(out.String(), "\x1b[J") {
		t.Fatal("must preserve reflowed output instead of guessing its height")
	}
	u.Release()
	fmt.Fprintln(&out, "child output")
	out.Reset()
	u.Panel("After child", "body")
	if strings.Contains(out.String(), "\x1b[J") {
		t.Fatal("must not erase child output after Release")
	}
}

func TestInlineTranscriptInputAndChildHandoff(t *testing.T) {
	for _, plain := range []bool{false, true} {
		t.Run(fmt.Sprint(plain), func(t *testing.T) {
			var out bytes.Buffer
			u := NewInline(InlineSpec{Input: strings.NewReader("alpha\r\n\nchild\n"), Output: &out, Plain: plain})
			u.Panel("Setup", "retained instructions")
			for _, want := range []string{"alpha", "default"} {
				got, err := u.Ask(context.Background(), "Setup", "Name", "default")
				if err != nil || got != want {
					t.Fatalf("got=%q want=%q err=%v", got, want, err)
				}
			}
			u.Release()
			remaining, err := io.ReadAll(u.Reader())
			if err != nil || string(remaining) != "child\n" {
				t.Fatalf("buffered child input lost: %q %v", remaining, err)
			}
			if strings.Contains(out.String(), "\x1b") {
				t.Fatal("transcript has escapes")
			}
		})
	}
}

func TestInlineInputEndAndContext(t *testing.T) {
	for _, tc := range []struct {
		input, want string
		err         error
	}{
		{"last", "last", nil}, {"", "", io.EOF}, {"\n", "seed", nil},
	} {
		u := NewInline(InlineSpec{Input: strings.NewReader(tc.input), Output: io.Discard})
		got, err := u.Ask(context.Background(), "Setup", "Value", "seed")
		if got != tc.want || !errors.Is(err, tc.err) {
			t.Fatalf("input=%q got=%q err=%v", tc.input, got, err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var out bytes.Buffer
	u := NewInline(InlineSpec{Input: strings.NewReader("answer\n"), Output: &out})
	if _, err := u.Ask(ctx, "Setup", "Value", ""); !errors.Is(err, context.Canceled) || out.Len() != 0 {
		t.Fatalf("cancelled ask wrote output: %q err=%v", out.String(), err)
	}
	// A general reader may block; cancellation after reading must still win.
	ctx, cancel = context.WithCancel(context.Background())
	u = NewInline(InlineSpec{Input: inlineReaderFunc(func(p []byte) (int, error) {
		cancel()
		return copy(p, "answer\n"), nil
	}), Output: io.Discard})
	if _, err := u.Ask(ctx, "Setup", "Value", ""); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

type inlineReaderFunc func([]byte) (int, error)

func (f inlineReaderFunc) Read(p []byte) (int, error) { return f(p) }

type inlineBrokenWriter struct{ err error }

func (w inlineBrokenWriter) Write([]byte) (int, error) { return 0, w.err }

func TestInlineReport(t *testing.T) {
	var out bytes.Buffer
	u := NewInline(InlineSpec{Input: strings.NewReader(""), Output: &out})
	u.interactive, u.ansi = true, true
	u.size = func() (int, int) { return 80, 30 }
	u.Panel("Setup", "instructions")
	out.Reset()
	cause := errors.New("cannot continue")
	err := u.Report(cause)
	if !ErrorReported(fmt.Errorf("wrapped: %w", err)) || !errors.Is(err, cause) {
		t.Fatal("reported error lost identity", err)
	}
	text := ansi.Strip(out.String())
	if strings.Count(text, "Setup") != 1 || !strings.Contains(text, "instructions") || !strings.Contains(text, cause.Error()) {
		t.Fatal(text)
	}
	before := out.Len()
	if u.Report(nil) != nil || u.Report(err) != err || out.Len() != before {
		t.Fatal("nil/already-reported errors must not print twice")
	}
	writeErr := errors.New("broken output")
	u.out = inlineBrokenWriter{writeErr}
	err = u.Report(cause)
	if ErrorReported(err) || !errors.Is(err, writeErr) || !errors.Is(err, cause) {
		t.Fatal("failed rendering must retain both errors and not claim reported", err)
	}
	u = NewInline(InlineSpec{Input: strings.NewReader(""), Output: io.Discard})
	if err := u.Panel("Setup"); err != nil {
		t.Fatal(err)
	}
	u.out = inlineBrokenWriter{writeErr}
	if _, err := u.Ask(context.Background(), "Setup", "Value", ""); !errors.Is(err, writeErr) {
		t.Fatal(err)
	}
}

func TestInlineNarrowWrappingAndSanitization(t *testing.T) {
	for _, width := range []int{3, 12, 28, 80} {
		var out bytes.Buffer
		u := NewInline(InlineSpec{Input: strings.NewReader(""), Output: &out})
		u.size = func() (int, int) { return width, 30 }
		u.Panel("Long title\x1b[2J", strings.Repeat("界text ", 20), "multiline\nbody\r\a\x1b]0;bad title\a")
		if strings.ContainsAny(out.String(), "\x1b\r\a") {
			t.Fatalf("control characters survived: %q", out.String())
		}
		for _, line := range strings.Split(out.String(), "\n") {
			if ansi.StringWidth(line) > width {
				t.Fatalf("width %d exceeded by %q", width, line)
			}
		}
		if !strings.Contains(out.String(), "body") && width >= 12 {
			t.Fatal("multiline instructions lost")
		}
	}
}

func TestInlineModelEditingSubmissionAndCancel(t *testing.T) {
	makeModel := func() *inlineModel {
		u := NewInline(InlineSpec{Input: strings.NewReader(""), Output: io.Discard})
		return &inlineModel{owner: u, title: "Setup", label: "Name", body: []string{"instructions"}, width: 50,
			form: forms.NewForm(forms.Text("value", "Name").WithText("seed"))}
	}
	m := makeModel()
	m.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("界a")})
	m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	if got := m.form.Focused().Text; got != "界 b" {
		t.Fatalf("editing=%q", got)
	}
	view := ansi.Strip(m.View())
	for _, want := range []string{"Setup", "instructions", "Name: 界 b", "enter submit", "ctrl+u clear", "esc cancel"} {
		if !strings.Contains(view, want) {
			t.Fatalf("missing %q in %s", want, view)
		}
	}
	m.Update(tea.WindowSizeMsg{Width: 35, Height: 20})
	for _, line := range strings.Split(m.View(), "\n") {
		if ansi.StringWidth(line) > 35 {
			t.Fatal("model ignored resize")
		}
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !m.submitted || cmd == nil || !strings.Contains(m.View(), "Saved response.") {
		t.Fatal("not submitted")
	}
	for _, key := range []tea.KeyType{tea.KeyEsc, tea.KeyCtrlC} {
		m = makeModel()
		_, cmd = m.Update(tea.KeyMsg{Type: key})
		if !m.cancelled || m.submitted || cmd == nil || !strings.Contains(m.View(), "Cancelled.") {
			t.Fatal("not cancelled")
		}
	}
}

func TestInlineScopedKeysAndHints(t *testing.T) {
	scope := ui.Default()
	scope.Keys = scope.Keys.With(keys.Binding{Action: keys.Confirm, Keys: []string{"ctrl+s", "s"}, Glyph: "wrong glyph"})
	u := NewInline(InlineSpec{Input: strings.NewReader(""), Output: io.Discard, UI: scope})
	m := &inlineModel{owner: u, title: "Setup", label: "Name", width: 70,
		form: forms.NewForm(forms.Text("value", "Name"))}
	if view := m.View(); !strings.Contains(view, "ctrl+s submit") || strings.Contains(view, "wrong glyph") || strings.Contains(view, "enter submit") {
		t.Fatal(view)
	}
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	if m.form.Focused().Text != "s" || m.submitted {
		t.Fatal("printable scoped binding swallowed text")
	}
	m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if !m.submitted {
		t.Fatal("scoped submit ignored")
	}
	scope.Keys = scope.Keys.With(keys.Binding{Action: keys.Confirm, Keys: []string{"ctrl+u"}})
	if action, _ := m.owner.editorMap().Action("ctrl+u"); action != keys.Confirm || m.owner.editorMap().Has(InlineClear) {
		t.Fatal("default clear must not shadow a scoped editor binding")
	}
	custom := keys.NewMap(keys.Binding{Action: keys.Confirm, Keys: []string{"ctrl+d"}, Glyph: "custom"})
	u = NewInline(InlineSpec{Input: strings.NewReader(""), Output: io.Discard, Keys: custom})
	if u.editorMap() != custom {
		t.Fatal("custom map ignored")
	}
	if m.owner.frame(40).UI != scope {
		t.Fatal("frame dropped its scope")
	}
}

func ExampleInline() {
	var transcript bytes.Buffer
	flow := NewInline(InlineSpec{
		Input: strings.NewReader("workspace\n"), Output: &transcript, Plain: true,
	})
	if err := flow.Panel("Workspace", "Choose a local directory."); err != nil {
		return
	}
	path, err := flow.Ask(context.Background(), "Workspace", "Directory", ".")
	if err != nil {
		_ = flow.Report(err)
		return
	}
	flow.Release() // hand output back to the caller without erasing it later
	fmt.Println(path)
	// Output: workspace
}
