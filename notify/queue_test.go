package notify

import (
	"testing"
	"time"
)

func n(title string) Notification { return Notification{Title: title} }

var base = time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)

// at is base shifted by a whole number of seconds, so tests read like ticks.
func at(sec int) time.Time { return base.Add(time.Duration(sec) * time.Second) }

func push(q *Queue, title string, ttlSec int) {
	q.PushFor(n(title), base, time.Duration(ttlSec)*time.Second)
}

func TestQueueEmpty(t *testing.T) {
	q := NewQueue(0)
	if q.Active() {
		t.Fatal("fresh queue should be inactive")
	}
	if _, ok := q.Current(); ok {
		t.Fatal("empty queue should have no current notification")
	}
	q.Prune(at(1))
}

func TestQueueShowsHeadUntilExpiry(t *testing.T) {
	q := NewQueue(0)
	push(q, "a", 2)

	for sec := range 2 {
		q.Prune(at(sec))
		cur, ok := q.Current()
		if !ok || cur.Title != "a" {
			t.Fatalf("second %d: want head a, got %q ok=%v", sec, cur.Title, ok)
		}
	}
	q.Prune(at(2))
	if q.Active() {
		t.Fatal("queue should drain after the head's TTL elapses")
	}
}

func TestQueueAdvancesThroughBacklog(t *testing.T) {
	q := NewQueue(0)
	push(q, "a", 1)
	q.PushFor(n("b"), base, 2*time.Second)
	if q.Len() != 2 {
		t.Fatalf("want 2 queued, got %d", q.Len())
	}

	if cur, _ := q.Current(); cur.Title != "a" {
		t.Fatalf("want head a, got %q", cur.Title)
	}
	q.Prune(at(1))
	if cur, ok := q.Current(); !ok || cur.Title != "b" {
		t.Fatalf("want head b after a expires, got %q ok=%v", cur.Title, ok)
	}
	q.Prune(at(2))
	if q.Active() {
		t.Fatal("queue should be empty after both expire")
	}
}

func TestQueueDropsZeroTTL(t *testing.T) {
	q := NewQueue(0)
	push(q, "a", 0)
	push(q, "b", -5)
	if q.Active() {
		t.Fatal("non-positive TTL notifications should be dropped")
	}
}

func TestQueueCapDropsOldestPending(t *testing.T) {
	q := NewQueue(2)
	push(q, "head", 10)
	push(q, "old", 10)
	q.PushFor(n("new"), base, 20*time.Second)

	if q.Len() != 2 {
		t.Fatalf("want 2 after cap, got %d", q.Len())
	}
	if cur, _ := q.Current(); cur.Title != "head" {
		t.Fatalf("cap must not evict the on-screen head, got %q", cur.Title)
	}
	q.Prune(at(10))
	if cur, ok := q.Current(); !ok || cur.Title != "new" {
		t.Fatalf("want surviving pending to be new, got %q ok=%v", cur.Title, ok)
	}
}

func TestQueueCapOneShowsTheNewest(t *testing.T) {
	q := NewQueue(1)
	push(q, "first", 5)
	push(q, "second", 5)

	if q.Len() != 1 {
		t.Fatalf("cap 1 should hold one notification, got %d", q.Len())
	}
	cur, ok := q.Current()
	if !ok || cur.Title != "second" {
		t.Fatalf("cap 1 must show the newest push, got %q ok=%v", cur.Title, ok)
	}
}

func TestQueueCapOneRefreshesTheTTL(t *testing.T) {
	q := NewQueue(1)
	push(q, "first", 1)
	push(q, "second", 3)

	q.Prune(at(1))
	cur, ok := q.Current()
	if !ok || cur.Title != "second" {
		t.Fatalf("the replacing notification should carry its own TTL, got %q ok=%v", cur.Title, ok)
	}
	q.Prune(at(3))
	if q.Active() {
		t.Fatalf("queue should drain after the newest TTL elapses, got %v", q.Snapshot())
	}
}

func TestQueueNeverDropsTheNewestPush(t *testing.T) {
	for _, capacity := range []int{1, 2, 3, 5} {
		q := NewQueue(capacity)
		for i := range 12 {
			q.PushFor(Notification{Title: string(rune('a' + i))}, base, 10*time.Second)
			if q.Len() > capacity {
				t.Fatalf("cap %d: queue grew to %d", capacity, q.Len())
			}
			snap := q.Snapshot()
			newest := snap[len(snap)-1]
			if want := string(rune('a' + i)); newest.Title != want {
				t.Fatalf("cap %d: push %d was dropped, tail is %q want %q (%v)", capacity, i, newest.Title, want, snap)
			}
		}
	}
}

func TestQueueCapKeepsTheHeadAndTheNewest(t *testing.T) {
	q := NewQueue(3)
	push(q, "head", 10)
	push(q, "p1", 10)
	push(q, "p2", 10)
	push(q, "p3", 10)

	snap := q.Snapshot()
	if len(snap) != 3 || snap[0].Title != "head" || snap[1].Title != "p2" || snap[2].Title != "p3" {
		t.Fatalf("cap 3 should drop the oldest pending only, got %v", snap)
	}
}
