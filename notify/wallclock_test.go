package notify

import (
	"testing"
	"time"
)

func TestQueuePruneExpiresByWallClock(t *testing.T) {
	q := NewQueue(0)
	base := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	q.PushUntil(n("a"), base.Add(10*time.Second))
	q.PushUntil(n("b"), base.Add(30*time.Second))

	q.Prune(base.Add(20 * time.Second))
	snap := q.Snapshot()
	if len(snap) != 1 || snap[0].Title != "b" {
		t.Fatalf("expected only b to survive, got %v", snap)
	}

	q.Prune(base.Add(40 * time.Second))
	if q.Active() {
		t.Fatal("all wall-clock entries should be pruned")
	}
}

func TestPushForZeroDurationDropped(t *testing.T) {
	q := NewQueue(0)
	q.PushFor(n("a"), time.Now(), 0)
	if q.Active() {
		t.Fatal("zero-duration push should be dropped")
	}
}

func TestBeatIgnoresWallClockEntries(t *testing.T) {
	q := NewQueue(0)
	q.PushUntil(n("wall"), time.Now().Add(time.Hour))
	q.Beat()
	q.Beat()
	if !q.Active() {
		t.Fatal("Beat must not expire wall-clock entries")
	}
}

func TestSnapshotOrderOldestFirst(t *testing.T) {
	q := NewQueue(0)
	now := time.Now()
	q.PushUntil(n("first"), now.Add(time.Hour))
	q.PushUntil(n("second"), now.Add(time.Hour))
	snap := q.Snapshot()
	if len(snap) != 2 || snap[0].Title != "first" || snap[1].Title != "second" {
		t.Fatalf("snapshot should preserve insertion order, got %v", snap)
	}
}
