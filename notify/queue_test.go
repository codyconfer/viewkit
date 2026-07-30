package notify

import "testing"

func n(title string) Notification { return Notification{Title: title, Tone: ToneNeutral} }

func TestQueueEmpty(t *testing.T) {
	q := NewQueue(0)
	if q.Active() {
		t.Fatal("fresh queue should be inactive")
	}
	if _, ok := q.Current(); ok {
		t.Fatal("empty queue should have no current notification")
	}
	q.Beat()
}

func TestQueueShowsHeadUntilExpiry(t *testing.T) {
	q := NewQueue(0)
	q.Push(n("a"), 2)

	for i := range 2 {
		cur, ok := q.Current()
		if !ok || cur.Title != "a" {
			t.Fatalf("beat %d: want head a, got %q ok=%v", i, cur.Title, ok)
		}
		q.Beat()
	}
	if q.Active() {
		t.Fatal("queue should drain after the head's TTL elapses")
	}
}

func TestQueueAdvancesThroughBacklog(t *testing.T) {
	q := NewQueue(0)
	q.Push(n("a"), 1)
	q.Push(n("b"), 1)
	if q.Len() != 2 {
		t.Fatalf("want 2 queued, got %d", q.Len())
	}

	if cur, _ := q.Current(); cur.Title != "a" {
		t.Fatalf("want head a, got %q", cur.Title)
	}
	q.Beat()
	if cur, ok := q.Current(); !ok || cur.Title != "b" {
		t.Fatalf("want head b after a expires, got %q ok=%v", cur.Title, ok)
	}
	q.Beat()
	if q.Active() {
		t.Fatal("queue should be empty after both expire")
	}
}

func TestQueueDropsZeroTTL(t *testing.T) {
	q := NewQueue(0)
	q.Push(n("a"), 0)
	q.Push(n("b"), -5)
	if q.Active() {
		t.Fatal("non-positive TTL notifications should be dropped")
	}
}

func TestQueueCapDropsOldestPending(t *testing.T) {
	q := NewQueue(2)
	q.Push(n("head"), 10)
	q.Push(n("old"), 10)
	q.Push(n("new"), 10)

	if q.Len() != 2 {
		t.Fatalf("want 2 after cap, got %d", q.Len())
	}
	if cur, _ := q.Current(); cur.Title != "head" {
		t.Fatalf("cap must not evict the on-screen head, got %q", cur.Title)
	}
	q.Beat()
	q.Beat()
	q.Beat()
	q.Beat()
	q.Beat()
	q.Beat()
	q.Beat()
	q.Beat()
	q.Beat()
	q.Beat()
	if cur, ok := q.Current(); !ok || cur.Title != "new" {
		t.Fatalf("want surviving pending to be new, got %q ok=%v", cur.Title, ok)
	}
}

func TestQueueCapOneShowsTheNewest(t *testing.T) {
	q := NewQueue(1)
	q.Push(n("first"), 5)
	q.Push(n("second"), 5)

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
	q.Push(n("first"), 1)
	q.Push(n("second"), 3)

	q.Beat()
	cur, ok := q.Current()
	if !ok || cur.Title != "second" {
		t.Fatalf("the replacing notification should carry its own TTL, got %q ok=%v", cur.Title, ok)
	}
	q.Beat()
	q.Beat()
	if q.Active() {
		t.Fatalf("queue should drain after the newest TTL elapses, got %v", q.Snapshot())
	}
}

func TestQueueNeverDropsTheNewestPush(t *testing.T) {
	for _, cap := range []int{1, 2, 3, 5} {
		q := NewQueue(cap)
		for i := range 12 {
			q.Push(Notification{Title: string(rune('a' + i))}, 10)
			if q.Len() > cap {
				t.Fatalf("cap %d: queue grew to %d", cap, q.Len())
			}
			snap := q.Snapshot()
			newest := snap[len(snap)-1]
			if want := string(rune('a' + i)); newest.Title != want {
				t.Fatalf("cap %d: push %d was dropped, tail is %q want %q (%v)", cap, i, newest.Title, want, snap)
			}
		}
	}
}

func TestQueueCapKeepsTheHeadAndTheNewest(t *testing.T) {
	q := NewQueue(3)
	q.Push(n("head"), 10)
	q.Push(n("p1"), 10)
	q.Push(n("p2"), 10)
	q.Push(n("p3"), 10)

	snap := q.Snapshot()
	if len(snap) != 3 || snap[0].Title != "head" || snap[1].Title != "p2" || snap[2].Title != "p3" {
		t.Fatalf("cap 3 should drop the oldest pending only, got %v", snap)
	}
}
