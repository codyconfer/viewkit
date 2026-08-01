// Package spin renders a single-line CLI spinner directly to a writer. It is a
// plain io.Writer animation with no bubbletea program behind it, so it suits
// startup work that happens before (or entirely outside of) a TUI.
package spin

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/x/term"

	"github.com/codyconfer/viewkit/theme"
)

// DefaultInterval is the frame interval used when Options.Interval is unset.
const DefaultInterval = time.Second / 12

// DefaultMessage is the in-progress text used when Options.Message is empty.
const DefaultMessage = "starting…"

// DefaultDoneMessage is the settle text used when Options.DoneMessage is empty.
const DefaultDoneMessage = "ready"

// DefaultDoneGlyph is the check mark printed by Done when Options.DoneGlyph is
// empty.
const DefaultDoneGlyph = "✓"

var defaultFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// DefaultFrames returns a copy of the frame set used when Options.Frames is
// empty.
func DefaultFrames() []string { return append([]string(nil), defaultFrames...) }

// Spinner is a running single-line animation. Create one with Start and end it
// with Stop or Done; every method is safe on a nil receiver.
type Spinner struct {
	w         io.Writer
	prefix    string
	msg       string
	doneMsg   string
	doneGlyph string
	frames    []string
	every     time.Duration
	th        theme.Theme
	stopCh    chan struct{}
	doneCh    chan struct{}
	mu        sync.Mutex
	stopped   bool
	animated  bool
}

// Options configures a Spinner. The zero value is usable: it animates
// DefaultMessage on os.Stderr whenever that is a terminal.
type Options struct {
	// Writer receives the animation. Defaults to os.Stderr.
	Writer io.Writer
	// Prefix is rendered dimmed ahead of the frame, e.g. "myapp ▸". When
	// empty no prefix (and no leading space) is written.
	Prefix string
	// Message is the dimmed in-progress text. Defaults to DefaultMessage.
	Message string
	// DoneMessage is the dimmed text Done settles on. Defaults to
	// DefaultDoneMessage.
	DoneMessage string
	// DoneGlyph is the check mark Done prints. Defaults to DefaultDoneGlyph.
	DoneGlyph string
	// Interval is the time between frames. Defaults to DefaultInterval.
	Interval time.Duration
	// Frames is the animation cycle. Defaults to DefaultFrames.
	Frames []string
	// Theme styles the spinner. Nil snapshots the process default theme at
	// Start.
	Theme *theme.Theme
	// Force animates even when Writer is not a terminal. Without it a
	// non-terminal Writer makes the Spinner a silent no-op.
	Force bool
}

// Start paints the first frame and animates until Stop or Done. When the
// writer is not a terminal and Force is unset it returns an inert Spinner that
// writes nothing.
func Start(opts Options) *Spinner {
	w := opts.Writer
	if w == nil {
		w = os.Stderr
	}
	msg := opts.Message
	if msg == "" {
		msg = DefaultMessage
	}
	doneMsg := opts.DoneMessage
	if doneMsg == "" {
		doneMsg = DefaultDoneMessage
	}
	doneGlyph := opts.DoneGlyph
	if doneGlyph == "" {
		doneGlyph = DefaultDoneGlyph
	}
	frames := opts.Frames
	if len(frames) == 0 {
		frames = DefaultFrames()
	}
	every := opts.Interval
	if every <= 0 {
		every = DefaultInterval
	}
	th := theme.Default()
	if opts.Theme != nil {
		th = *opts.Theme
	}

	s := &Spinner{
		w:         w,
		prefix:    opts.Prefix,
		msg:       msg,
		doneMsg:   doneMsg,
		doneGlyph: doneGlyph,
		frames:    frames,
		every:     every,
		th:        th,
		stopCh:    make(chan struct{}),
		doneCh:    make(chan struct{}),
	}
	if !opts.Force && !isTerminal(w) {
		s.stopped = true
		close(s.doneCh)
		return s
	}
	s.animated = true
	s.paint(0)
	go s.loop()
	return s
}

// Stop ends the animation and erases the line. Calls after the first are
// no-ops.
func (s *Spinner) Stop() { s.finish(false) }

// Done ends the animation and replaces the line with a settled check mark and
// the done message, followed by a newline. Calls after the first are no-ops.
func (s *Spinner) Done() { s.finish(true) }

func (s *Spinner) finish(settle bool) {
	if s == nil {
		return
	}
	s.mu.Lock()
	if s.stopped {
		s.mu.Unlock()
		return
	}
	s.stopped = true
	s.mu.Unlock()

	if !s.animated {
		return
	}
	close(s.stopCh)
	<-s.doneCh
	if !settle {
		_, _ = fmt.Fprint(s.w, "\r\033[K")
		return
	}
	_, _ = fmt.Fprintf(s.w, "\r\033[K%s%s %s\n",
		s.renderPrefix(), s.th.Can.Render(s.doneGlyph), s.th.Dim.Render(s.doneMsg))
}

func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(f.Fd())
}

func (s *Spinner) loop() {
	defer close(s.doneCh)
	t := time.NewTicker(s.every)
	defer t.Stop()
	i := 0
	for {
		select {
		case <-s.stopCh:
			return
		case <-t.C:
			i++
			s.paint(i)
		}
	}
}

func (s *Spinner) paint(i int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stopped || len(s.frames) == 0 {
		return
	}
	frame := s.th.Accent.Render(s.frames[i%len(s.frames)])
	msg := s.th.Dim.Render(s.msg)
	_, _ = fmt.Fprintf(s.w, "\r%s%s %s", s.renderPrefix(), frame, msg)
}

func (s *Spinner) renderPrefix() string {
	if s.prefix == "" {
		return ""
	}
	return s.th.Dim.Render(s.prefix) + " "
}
