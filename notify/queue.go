package notify

import "time"

type Queue struct {
	items []queued
	cap   int
}

type queued struct {
	n      Notification
	ttl    int
	expiry time.Time
}

func NewQueue(cap int) *Queue { return &Queue{cap: cap} }

func (q *Queue) Push(n Notification, ttl int) {
	if ttl <= 0 {
		return
	}
	q.append(queued{n: n, ttl: ttl})
}

func (q *Queue) PushUntil(n Notification, expiry time.Time) {
	q.append(queued{n: n, expiry: expiry})
}

func (q *Queue) PushFor(n Notification, now time.Time, d time.Duration) {
	if d <= 0 {
		return
	}
	q.PushUntil(n, now.Add(d))
}

func (q *Queue) append(item queued) {
	q.items = append(q.items, item)
	if q.cap > 0 && len(q.items) > q.cap {
		q.items = append(q.items[:1], q.items[2:]...)
	}
}

func (q *Queue) Beat() {
	if len(q.items) == 0 || !q.items[0].expiry.IsZero() {
		return
	}
	if q.items[0].ttl--; q.items[0].ttl <= 0 {
		q.items = q.items[1:]
	}
}

func (q *Queue) Prune(now time.Time) {
	kept := q.items[:0]
	for _, it := range q.items {
		if !it.expiry.IsZero() && !it.expiry.After(now) {
			continue
		}
		kept = append(kept, it)
	}
	q.items = kept
}

func (q *Queue) Current() (Notification, bool) {
	if len(q.items) == 0 {
		return Notification{}, false
	}
	return q.items[0].n, true
}

func (q *Queue) Snapshot() []Notification {
	out := make([]Notification, len(q.items))
	for i, it := range q.items {
		out[i] = it.n
	}
	return out
}

func (q *Queue) Active() bool { return len(q.items) > 0 }

func (q *Queue) Len() int { return len(q.items) }
