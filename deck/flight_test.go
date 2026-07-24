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

func newTestFlight(labels ...string) *flightModel {
	m := &flightModel{
		panels: make([]flightPanel, len(labels)),
		spin:   spinner.New(),
		left:   len(labels),
	}
	for i, l := range labels {
		m.panels[i].label = l
	}
	return m
}

func TestFlightPanelsReplaceOnDone(t *testing.T) {
	m := newTestFlight("alpha", "beta")

	v := m.View()
	if !strings.Contains(v, "alpha") || !strings.Contains(v, "beta") || !strings.Contains(v, "loading") {
		t.Fatalf("initial view should show both labels loading:\n%s", v)
	}

	next, _ := m.Update(flightDoneMsg{idx: 0, content: "ALPHA-CONTENT"})
	m = next.(*flightModel)
	if m.left != 1 {
		t.Fatalf("left = %d, want 1", m.left)
	}
	v = m.View()
	if !strings.Contains(v, "ALPHA-CONTENT") {
		t.Errorf("view should show completed content:\n%s", v)
	}
	if !strings.Contains(v, "beta") || !strings.Contains(v, "loading") {
		t.Errorf("second panel should still be loading:\n%s", v)
	}

	next, cmd := m.Update(flightDoneMsg{idx: 1, content: "BETA-CONTENT"})
	m = next.(*flightModel)
	if m.left != 0 {
		t.Fatalf("left = %d, want 0", m.left)
	}
	if cmd == nil {
		t.Fatal("expected a quit command once all panels are done")
	}
	v = m.View()
	if !strings.Contains(v, "ALPHA-CONTENT") || !strings.Contains(v, "BETA-CONTENT") {
		t.Errorf("final view should show all content:\n%s", v)
	}
	if strings.Contains(v, "loading") {
		t.Errorf("final view should have no loading panels:\n%s", v)
	}
}

func TestFlightDuplicateDoneIgnored(t *testing.T) {
	m := newTestFlight("only")
	next, _ := m.Update(flightDoneMsg{idx: 0, content: "X"})
	m = next.(*flightModel)

	next, _ = m.Update(flightDoneMsg{idx: 0, content: "X"})
	m = next.(*flightModel)
	if m.left != 0 {
		t.Fatalf("left = %d, want 0 (no underflow)", m.left)
	}
}

func TestExecuteErrgroupOrder(t *testing.T) {
	var n atomic.Int32
	tasks := []Task{
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
	out, err := Execute(context.Background(), tasks)
	if err != nil {
		t.Fatal(err)
	}
	if n.Load() != 2 {
		t.Fatalf("ran %d tasks", n.Load())
	}
	if out[0].Render(0) != "A" || out[1].Render(0) != "B" {
		t.Fatalf("order wrong: %#v %#v", out[0], out[1])
	}
}

func TestExecutePropagatesError(t *testing.T) {
	boom := errors.New("boom")
	_, err := Execute(context.Background(), []Task{
		{Run: func(context.Context) (Content, error) { return nil, boom }},
	})
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v", err)
	}
}
