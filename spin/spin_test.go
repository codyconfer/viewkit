package spin

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/codyconfer/viewkit/theme"
)

type safeBuf struct {
	mu sync.Mutex
	b  strings.Builder
}

func (s *safeBuf) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *safeBuf) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

func (s *safeBuf) Len() int { return len(s.String()) }

func waitFor(t *testing.T, cond func() bool) bool {
	t.Helper()
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(2 * time.Millisecond)
	}
	return cond()
}

func stripANSI(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' {
			i++
			for i < len(s) && ((s[i] >= '0' && s[i] <= '9') || s[i] == '[' || s[i] == ';' || s[i] == '?') {
				i++
			}
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func TestStartStopClearsLine(t *testing.T) {
	buf := &safeBuf{}
	s := Start(Options{
		Writer:   buf,
		Prefix:   "viewkit ▸",
		Message:  "starting…",
		Force:    true,
		Interval: 5 * time.Millisecond,
		Frames:   []string{"⠁", "⠂"},
	})

	waitFor(t, func() bool { return strings.Contains(stripANSI(buf.String()), "starting…") })

	got := stripANSI(buf.String())
	if !strings.Contains(got, "viewkit ▸") {
		t.Fatalf("expected prefix, got %q", got)
	}
	if !strings.Contains(got, "starting…") {
		t.Fatalf("expected message, got %q", got)
	}
	if !strings.Contains(got, "⠁") && !strings.Contains(got, "⠂") {
		t.Fatalf("expected a spinner frame, got %q", got)
	}

	s.Stop()
	s.Stop()

	out := buf.String()
	if !strings.Contains(out, "\r\033[K") {
		t.Fatalf("expected clear sequence after Stop, got %q", out)
	}
}

func TestStartAnimatesThroughFrames(t *testing.T) {
	buf := &safeBuf{}
	s := Start(Options{
		Writer:   buf,
		Force:    true,
		Interval: 5 * time.Millisecond,
		Frames:   []string{"⠁", "⠂"},
	})
	defer s.Stop()

	if !waitFor(t, func() bool {
		got := stripANSI(buf.String())
		return strings.Contains(got, "⠁") && strings.Contains(got, "⠂")
	}) {
		t.Fatalf("expected both frames to be painted, got %q", stripANSI(buf.String()))
	}
	if !strings.Contains(stripANSI(buf.String()), DefaultMessage) {
		t.Errorf("expected default message %q, got %q", DefaultMessage, stripANSI(buf.String()))
	}
}

func TestEmptyPrefixHasNoLeadingSpace(t *testing.T) {
	buf := &safeBuf{}
	s := Start(Options{
		Writer:   buf,
		Message:  "starting…",
		Force:    true,
		Interval: time.Hour,
		Frames:   []string{"⠁"},
	})
	defer s.Stop()

	got := stripANSI(buf.String())
	if got != "\r⠁ starting…" {
		t.Fatalf("empty prefix should paint frame right after the carriage return, got %q", got)
	}
}

func TestPrefixIsFollowedBySingleSpace(t *testing.T) {
	buf := &safeBuf{}
	s := Start(Options{
		Writer:   buf,
		Prefix:   "app ▸",
		Message:  "working",
		Force:    true,
		Interval: time.Hour,
		Frames:   []string{"⠁"},
	})
	defer s.Stop()

	got := stripANSI(buf.String())
	if got != "\rapp ▸ ⠁ working" {
		t.Fatalf("unexpected painted line: %q", got)
	}
}

func TestEmptyPrefixOnDoneLine(t *testing.T) {
	buf := &safeBuf{}
	s := Start(Options{
		Writer:      buf,
		DoneMessage: "ready",
		Force:       true,
		Interval:    time.Hour,
		Frames:      []string{"⠁"},
	})
	s.Done()

	tail := stripANSI(buf.String())
	tail = tail[strings.LastIndex(tail, "\r"):]
	if tail != "\r"+DefaultDoneGlyph+" ready\n" {
		t.Fatalf("unexpected settled line without a prefix: %q", tail)
	}
}

func TestNonTTYNoOp(t *testing.T) {
	buf := &safeBuf{}
	s := Start(Options{Writer: buf, Message: "starting…"})
	time.Sleep(20 * time.Millisecond)
	s.Stop()
	if buf.Len() != 0 {
		t.Fatalf("non-TTY should not write, got %q", buf.String())
	}
}

func TestDoneOnNonTTYStaysSilent(t *testing.T) {
	buf := &safeBuf{}
	s := Start(Options{Writer: buf, Message: "starting…", DoneMessage: "ready"})
	s.Done()
	if buf.Len() != 0 {
		t.Errorf("non-TTY Done should print nothing, got %q", buf.String())
	}
}

func TestForceAnimatesOnNonTTYWriter(t *testing.T) {
	buf := &safeBuf{}
	s := Start(Options{
		Writer:   buf,
		Message:  "forced",
		Force:    true,
		Interval: 5 * time.Millisecond,
		Frames:   []string{"⠁", "⠂"},
	})
	if !waitFor(t, func() bool { return strings.Contains(stripANSI(buf.String()), "forced") }) {
		t.Fatalf("Force should animate on a non-TTY writer, got %q", buf.String())
	}
	s.Stop()
}

func TestDoneSettlesIntoACheckmark(t *testing.T) {
	buf := &safeBuf{}
	s := Start(Options{
		Writer:      buf,
		Prefix:      "app ▸",
		Message:     "starting…",
		DoneMessage: "ready",
		Force:       true,
		Interval:    5 * time.Millisecond,
		Frames:      []string{"⠁", "⠂"},
	})

	waitFor(t, func() bool { return strings.Contains(stripANSI(buf.String()), "⠂") })
	s.Done()

	out := buf.String()
	if !strings.HasSuffix(out, "\n") {
		t.Errorf("settled line should end with a newline so later output does not overwrite it: %q", out)
	}
	tailAt := strings.LastIndex(out, "\033[K")
	if tailAt < 0 {
		t.Fatalf("no line clear before the settled line: %q", out)
	}
	tail := stripANSI(out[tailAt:])
	if !strings.Contains(tail, "ready") {
		t.Errorf("settled line missing the done message: %q", tail)
	}
	if !strings.Contains(tail, DefaultDoneGlyph) {
		t.Errorf("settled line missing the check glyph: %q", tail)
	}
	if !strings.Contains(tail, "app ▸") {
		t.Errorf("settled line missing the prefix: %q", tail)
	}
	for _, frame := range []string{"⠁", "⠂"} {
		if strings.Contains(tail, frame) {
			t.Errorf("settled line still shows spinner frame %q: %q", frame, tail)
		}
	}
	if strings.Contains(tail, "starting…") {
		t.Errorf("settled line should replace the in-progress message: %q", tail)
	}
}

func TestCustomDoneGlyph(t *testing.T) {
	buf := &safeBuf{}
	s := Start(Options{
		Writer:    buf,
		DoneGlyph: "**",
		Force:     true,
		Interval:  time.Hour,
		Frames:    []string{"⠁"},
	})
	s.Done()

	got := stripANSI(buf.String())
	if !strings.Contains(got, "** "+DefaultDoneMessage) {
		t.Fatalf("expected custom glyph and default done message, got %q", got)
	}
}

func TestStopVersusDoneOutput(t *testing.T) {
	newSpinner := func(buf *safeBuf) *Spinner {
		return Start(Options{
			Writer:      buf,
			Message:     "working",
			DoneMessage: "ready",
			Force:       true,
			Interval:    time.Hour,
			Frames:      []string{"⠁"},
		})
	}

	stopBuf := &safeBuf{}
	newSpinner(stopBuf).Stop()
	stopped := stripANSI(stopBuf.String())
	if !strings.HasSuffix(stopped, "\r") {
		t.Errorf("Stop should end with an erased line and no trailing text: %q", stopBuf.String())
	}
	if strings.Contains(stopped, "ready") || strings.Contains(stopped, DefaultDoneGlyph) {
		t.Errorf("Stop should not print a settle line: %q", stopped)
	}
	if strings.Contains(stopBuf.String(), "\n") {
		t.Errorf("Stop should not emit a newline: %q", stopBuf.String())
	}

	doneBuf := &safeBuf{}
	newSpinner(doneBuf).Done()
	done := stripANSI(doneBuf.String())
	if !strings.HasSuffix(done, DefaultDoneGlyph+" ready\n") {
		t.Errorf("Done should settle into glyph + done message: %q", done)
	}
}

func TestDoneAfterStopIsANoOp(t *testing.T) {
	buf := &safeBuf{}
	s := Start(Options{
		Writer:      buf,
		DoneMessage: "ready",
		Force:       true,
		Interval:    time.Hour,
		Frames:      []string{"⠁"},
	})
	s.Stop()
	before := buf.String()
	s.Done()
	if buf.String() != before {
		t.Fatalf("Done after Stop should write nothing extra, got %q", buf.String())
	}
	if strings.Contains(buf.String(), "ready") {
		t.Fatalf("Done after Stop should not settle: %q", buf.String())
	}
}

func TestDoubleStopWritesEraseOnce(t *testing.T) {
	buf := &safeBuf{}
	s := Start(Options{
		Writer:   buf,
		Force:    true,
		Interval: time.Hour,
		Frames:   []string{"⠁"},
	})
	s.Stop()
	s.Stop()
	s.Stop()
	if n := strings.Count(buf.String(), "\033[K"); n != 1 {
		t.Fatalf("expected exactly one erase sequence, got %d in %q", n, buf.String())
	}
}

func TestConcurrentStopIsSafe(t *testing.T) {
	buf := &safeBuf{}
	s := Start(Options{
		Writer:   buf,
		Force:    true,
		Interval: 2 * time.Millisecond,
		Frames:   []string{"⠁", "⠂"},
	})

	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			if i%2 == 0 {
				s.Stop()
				return
			}
			s.Done()
		}(i)
	}
	close(start)
	wg.Wait()

	if n := strings.Count(buf.String(), "\033[K"); n != 1 {
		t.Fatalf("expected exactly one terminating write, got %d in %q", n, buf.String())
	}
}

func TestNilSpinnerIsSafe(t *testing.T) {
	var s *Spinner
	s.Stop()
	s.Done()
}

func TestDefaultFramesReturnsACopy(t *testing.T) {
	a := DefaultFrames()
	if len(a) != 10 {
		t.Fatalf("expected 10 default frames, got %d", len(a))
	}
	if a[0] != "⠋" || a[len(a)-1] != "⠏" {
		t.Fatalf("unexpected default frames: %q", a)
	}
	a[0] = "x"
	if DefaultFrames()[0] != "⠋" {
		t.Fatal("DefaultFrames must return a copy")
	}
}

func TestDefaultsAreApplied(t *testing.T) {
	buf := &safeBuf{}
	s := Start(Options{Writer: buf, Force: true, Interval: time.Hour})
	defer s.Stop()

	if s.every != time.Hour {
		t.Errorf("interval = %v, want %v", s.every, time.Hour)
	}

	buf2 := &safeBuf{}
	s2 := Start(Options{Writer: buf2, Force: true})
	defer s2.Stop()
	if s2.every != DefaultInterval {
		t.Errorf("default interval = %v, want %v", s2.every, DefaultInterval)
	}
	if s2.msg != DefaultMessage || s2.doneMsg != DefaultDoneMessage || s2.doneGlyph != DefaultDoneGlyph {
		t.Errorf("unexpected defaults: %q %q %q", s2.msg, s2.doneMsg, s2.doneGlyph)
	}
	if len(s2.frames) != len(defaultFrames) {
		t.Errorf("expected default frames, got %q", s2.frames)
	}
}

func TestZeroOptionsFallsBackToStderr(t *testing.T) {
	s := Start(Options{})
	if s == nil {
		t.Fatal("Start returned nil")
	}
	s.Stop()
}

func TestOptionsThemeIsSnapshotted(t *testing.T) {
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(prev) })

	th := theme.Default()
	th.Dim = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ffee"))
	buf := &safeBuf{}
	s := Start(Options{Writer: buf, Prefix: "pfx", Theme: &th, Force: true, Interval: time.Hour})
	s.Done()
	if want := th.Dim.Render("pfx"); !strings.Contains(buf.String(), want) {
		t.Fatalf("scoped Dim prefix %q missing from output %q", want, buf.String())
	}
}
