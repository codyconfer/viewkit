package notify

import "time"

// Queue holds pending notifications with wall-clock expiries, capped at a
// maximum length that never evicts the on-screen head or the newest push.
type Queue struct {
	items []queued
	cap   int
}

type queued struct {
	n      Notification
	expiry time.Time
}

// NewQueue returns a queue holding at most capacity notifications (0 = unbounded).
func NewQueue(capacity int) *Queue { return &Queue{cap: capacity} }

// PushUntil enqueues n until an absolute expiry.
func (q *Queue) PushUntil(n Notification, expiry time.Time) {
	q.append(queued{n: n, expiry: expiry})
}

// PushFor enqueues n for d measured from now. Non-positive durations drop n.
func (q *Queue) PushFor(n Notification, now time.Time, d time.Duration) {
	if d <= 0 {
		return
	}
	q.PushUntil(n, now.Add(d))
}

func (q *Queue) append(item queued) {
	q.items = append(q.items, item)
	for q.cap > 0 && len(q.items) > q.cap {
		drop := 1
		if len(q.items) < 3 {
			drop = 0
		}
		q.items = append(q.items[:drop], q.items[drop+1:]...)
	}
}

// Prune drops every notification whose expiry is at or before now.
func (q *Queue) Prune(now time.Time) {
	kept := q.items[:0]
	for _, it := range q.items {
		if !it.expiry.After(now) {
			continue
		}
		kept = append(kept, it)
	}
	q.items = kept
}

// Current returns the notification being shown: the oldest unexpired entry.
func (q *Queue) Current() (Notification, bool) {
	if len(q.items) == 0 {
		return Notification{}, false
	}
	return q.items[0].n, true
}

// Snapshot copies the queued notifications, oldest first.
func (q *Queue) Snapshot() []Notification {
	out := make([]Notification, len(q.items))
	for i, it := range q.items {
		out[i] = it.n
	}
	return out
}

// Active reports whether anything is queued.
func (q *Queue) Active() bool { return len(q.items) > 0 }

// Len returns the number of queued notifications.
func (q *Queue) Len() int { return len(q.items) }
