package deck

import (
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/notify"
	"github.com/codyconfer/viewkit/panels"
)

const toastInterval = time.Second

var toasterSeq atomic.Uint64

type toastPruneMsg struct {
	id uint64
	at time.Time
}

// Toaster owns a view's notification lifecycle: a TTL queue, the once-a-second
// prune tick, and the overlay that floats the current notification over a
// rendered body. Embed one in a View and forward Update / Body to it.
//
// Expiry does not depend on the tick arriving. Body prunes against the clock
// before it reads the queue, so a tick lost while another view was on top (see
// toastPruneMsg) delays nothing the user can see; the tick exists to repaint
// on time, not to be the only thing that expires a toast.
type Toaster struct {
	id    uint64
	queue *notify.Queue
	ttl   time.Duration
	pos   layout.OverlayPos
	armed bool
	now   func() time.Time
}

// NewToaster builds a Toaster holding at most capacity notifications (0 for
// unbounded), each shown for ttl.
func NewToaster(capacity int, ttl time.Duration) *Toaster {
	return &Toaster{
		id:    toasterSeq.Add(1),
		queue: notify.NewQueue(capacity),
		ttl:   ttl,
		pos:   layout.OverlayPos{XFrac: 0.5, YFrac: 0},
		now:   time.Now,
	}
}

// SetOverlayPos moves the toast card. It defaults to top centre.
func (t *Toaster) SetOverlayPos(pos layout.OverlayPos) { t.pos = pos }

// Push shows n for the Toaster's default TTL and returns the command that
// keeps the prune loop running. Return it from the view's Update.
func (t *Toaster) Push(n notify.Notification) tea.Cmd {
	return t.PushFor(n, t.ttl)
}

// PushFor shows n for ttl, overriding the Toaster's default.
func (t *Toaster) PushFor(n notify.Notification, ttl time.Duration) tea.Cmd {
	t.queue.PushFor(n, t.now(), ttl)
	return t.Tick()
}

// Tick arms the prune loop and returns the command to run, or nil when a tick
// is already in flight or there is nothing left to expire. Views that push
// notifications through Push / PushFor never need to call it; a view that
// restores a Toaster it filled earlier (after a Pop back onto it) can.
func (t *Toaster) Tick() tea.Cmd {
	if t.armed || !t.queue.Active() {
		return nil
	}
	t.armed = true
	id := t.id
	return tea.Tick(toastInterval, func(at time.Time) tea.Msg {
		return toastPruneMsg{id: id, at: at}
	})
}

// Update consumes this Toaster's prune tick. It reports handled=false for
// every other message — including another Toaster's tick — so the view can
// carry on dispatching. The returned command re-arms the loop while
// notifications remain.
func (t *Toaster) Update(msg tea.Msg) (cmd tea.Cmd, handled bool) {
	m, ok := msg.(toastPruneMsg)
	if !ok || m.id != t.id {
		return nil, false
	}
	t.armed = false
	t.queue.Prune(m.at)
	return t.Tick(), true
}

// Prune expires everything past its TTL. Body and the read accessors below
// already do this.
func (t *Toaster) Prune() { t.queue.Prune(t.now()) }

// Snapshot prunes expired toasts and returns the rest, oldest first.
func (t *Toaster) Snapshot() []notify.Notification {
	t.Prune()
	return t.queue.Snapshot()
}

// Len prunes expired toasts and returns how many remain.
func (t *Toaster) Len() int {
	t.Prune()
	return t.queue.Len()
}

// Current prunes expired toasts and returns the one showing now, if any.
func (t *Toaster) Current() (notify.Notification, bool) {
	t.Prune()
	return t.queue.Current()
}

// Active prunes expired toasts and reports whether any remain.
func (t *Toaster) Active() bool {
	t.Prune()
	return t.queue.Active()
}

// Body overlays the current notification on an already-rendered body, sized
// for the same screen width the body was rendered at. It returns body
// unchanged when nothing is showing.
func (t *Toaster) Body(body string, width int) string {
	t.Prune()
	n, ok := t.queue.Current()
	if !ok {
		return body
	}
	return panels.NotificationOverlay(body, layout.ScreenFrame(width), n, t.pos)
}
