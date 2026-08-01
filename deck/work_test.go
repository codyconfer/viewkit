package deck

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
)

func newTestWork(labels ...string) *workModel {
	m := &workModel{
		panels: make([]jobPanel, len(labels)),
		spin:   spinner.New(),
		left:   len(labels),
	}
	for i, l := range labels {
		m.panels[i].label = l
	}
	return m
}

func TestWorkPanelsLeaveFrameOnDone(t *testing.T) {
	m := newTestWork("alpha", "beta")

	v := m.View()
	if !strings.Contains(v, "alpha") || !strings.Contains(v, "beta") || !strings.Contains(v, "loading") {
		t.Fatalf("initial view should show both labels loading:\n%s", v)
	}

	next, cmd := m.Update(jobDoneMsg{idx: 0, content: "ALPHA-CONTENT"})
	m = next.(*workModel)
	if m.left != 1 {
		t.Fatalf("left = %d, want 1", m.left)
	}
	if cmd == nil {
		t.Fatal("expected a print command for the finished panel")
	}
	v = m.View()
	if strings.Contains(v, "ALPHA-CONTENT") || strings.Contains(v, "alpha") {
		t.Errorf("printed panel should be gone from the live frame:\n%s", v)
	}
	if !strings.Contains(v, "beta") || !strings.Contains(v, "loading") {
		t.Errorf("second panel should still be loading:\n%s", v)
	}

	next, cmd = m.Update(jobDoneMsg{idx: 1, content: "BETA-CONTENT"})
	m = next.(*workModel)
	if m.left != 0 {
		t.Fatalf("left = %d, want 0", m.left)
	}
	if cmd == nil {
		t.Fatal("expected a print+quit command once all panels are done")
	}
	v = m.View()
	if strings.TrimSpace(v) != "" {
		t.Errorf("final frame should be empty, all content printed:\n%s", v)
	}
}

func TestWorkDrainKeepsJobOrder(t *testing.T) {
	m := newTestWork("alpha", "beta", "gamma")

	m.panels[1].done, m.panels[1].content = true, "BETA"
	if body, ok := m.drain(); ok {
		t.Fatalf("drain before the first panel finished returned %q", body)
	}

	m.panels[0].done, m.panels[0].content = true, "ALPHA"
	body, ok := m.drain()
	if !ok || body != "ALPHA\nBETA" {
		t.Fatalf("drain = %q, %v; want \"ALPHA\\nBETA\", true", body, ok)
	}
	if v := m.View(); strings.Contains(v, "alpha") || strings.Contains(v, "beta") {
		t.Errorf("drained panels should leave the frame:\n%s", v)
	}
	if v := m.View(); !strings.Contains(v, "gamma") {
		t.Errorf("pending panel should remain:\n%s", v)
	}

	if _, ok := m.drain(); ok {
		t.Error("second drain should have nothing to print")
	}
}

func TestWorkQueuedPanelShowsNoContent(t *testing.T) {
	m := newTestWork("alpha", "beta")
	next, cmd := m.Update(jobDoneMsg{idx: 1, content: "BETA-CONTENT"})
	m = next.(*workModel)
	if cmd != nil {
		t.Fatal("out-of-order completion should not print yet")
	}
	v := m.View()
	if strings.Contains(v, "BETA-CONTENT") {
		t.Errorf("queued panel must not inflate the live frame:\n%s", v)
	}
	if !strings.Contains(v, "queued") {
		t.Errorf("queued panel should be marked as such:\n%s", v)
	}
}

func TestWorkDuplicateDoneIgnored(t *testing.T) {
	m := newTestWork("only")
	next, _ := m.Update(jobDoneMsg{idx: 0, content: "X"})
	m = next.(*workModel)

	next, _ = m.Update(jobDoneMsg{idx: 0, content: "X"})
	m = next.(*workModel)
	if m.left != 0 {
		t.Fatalf("left = %d, want 0 (no underflow)", m.left)
	}
}

func TestCollectErrgroupOrder(t *testing.T) {
	var n atomic.Int32
	w := Work{
		{Label: "a", Run: func(context.Context) (Content, error) {
			n.Add(1)
			time.Sleep(20 * time.Millisecond)
			return Text("A"), nil
		}},
		{Label: "b", Run: func(context.Context) (Content, error) {
			n.Add(1)
			return Text("B"), nil
		}},
	}
	out, err := w.Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if n.Load() != 2 {
		t.Fatalf("ran %d jobs", n.Load())
	}
	if out[0].Render(0) != "A" || out[1].Render(0) != "B" {
		t.Fatalf("order wrong: %#v %#v", out[0], out[1])
	}
}

func TestCollectPropagatesError(t *testing.T) {
	boom := errors.New("boom")
	_, err := Work{
		{Run: func(context.Context) (Content, error) { return nil, boom }},
	}.Collect(context.Background())
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v", err)
	}
}
