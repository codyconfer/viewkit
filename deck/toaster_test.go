package deck

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/notify"
)

var toastEpoch = time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)

func fixedToaster(ttl time.Duration, clock *time.Time) *Toaster {
	tst := NewToaster(4, ttl)
	tst.now = func() time.Time { return *clock }
	return tst
}

func toastBody() string {
	return strings.Repeat("body line\n", 8)
}

func tick(tst *Toaster, at time.Time) tea.Msg {
	return toastPruneMsg{id: tst.id, at: at}
}

func TestToasterOverlaysAPushedNotification(t *testing.T) {
	now := toastEpoch
	tst := fixedToaster(3*time.Second, &now)

	plain := tst.Body(toastBody(), 80)
	if strings.Contains(plain, "Saved") {
		t.Fatal("an empty toaster should leave the body alone")
	}

	tst.Push(notify.Positive("Saved", "wrote config"))
	got := tst.Body(toastBody(), 80)
	if !strings.Contains(got, "Saved") || !strings.Contains(got, "wrote config") {
		t.Fatalf("body is missing the toast:\n%s", got)
	}
}

func TestToasterPruneTickExpiresTheToast(t *testing.T) {
	now := toastEpoch
	tst := fixedToaster(3*time.Second, &now)
	tst.Push(notify.Neutral("Installed", "ntr"))

	now = toastEpoch.Add(time.Second)
	if _, handled := tst.Update(tick(tst, now)); !handled {
		t.Fatal("the toaster should claim its own tick")
	}
	if !strings.Contains(tst.Body(toastBody(), 80), "Installed") {
		t.Error("a toast should survive a tick inside its TTL")
	}

	now = toastEpoch.Add(4 * time.Second)
	if _, handled := tst.Update(tick(tst, now)); !handled {
		t.Fatal("the toaster should claim its own tick")
	}
	if got := tst.Body(toastBody(), 80); strings.Contains(got, "Installed") {
		t.Fatalf("the toast outlived its TTL:\n%s", got)
	}
	if tst.Queue().Active() {
		t.Error("queue should be empty after the TTL elapsed")
	}
}

func TestToasterReArmsWhileNotificationsRemain(t *testing.T) {
	now := toastEpoch
	tst := fixedToaster(3*time.Second, &now)

	if cmd := tst.Push(notify.Neutral("a", "")); cmd == nil {
		t.Fatal("the first push should arm the prune loop")
	}
	if cmd := tst.Push(notify.Neutral("b", "")); cmd != nil {
		t.Error("a second push should not start a competing tick chain")
	}

	now = toastEpoch.Add(time.Second)
	cmd, handled := tst.Update(tick(tst, now))
	if !handled || cmd == nil {
		t.Fatalf("a tick with toasts left should re-arm: cmd=%v handled=%v", cmd, handled)
	}

	now = toastEpoch.Add(5 * time.Second)
	cmd, handled = tst.Update(tick(tst, now))
	if !handled {
		t.Fatal("the toaster should claim its own tick")
	}
	if cmd != nil {
		t.Error("the loop should stop once the queue drains")
	}
	if cmd := tst.Tick(); cmd != nil {
		t.Error("Tick on a drained queue should be nil")
	}
}

func TestToastersDoNotConsumeEachOthersTicks(t *testing.T) {
	now := toastEpoch
	a := fixedToaster(3*time.Second, &now)
	b := fixedToaster(3*time.Second, &now)
	a.Push(notify.Neutral("alpha", ""))
	b.Push(notify.Neutral("beta", ""))

	late := toastEpoch.Add(10 * time.Second)
	if cmd, handled := b.Update(tick(a, late)); handled || cmd != nil {
		t.Fatalf("b claimed a's tick: cmd=%v handled=%v", cmd, handled)
	}
	if !b.Queue().Active() {
		t.Fatal("b's queue was pruned by a's tick")
	}
	if !strings.Contains(b.Body(toastBody(), 80), "beta") {
		t.Error("b's toast should still be showing")
	}

	if _, handled := a.Update(tick(a, late)); !handled {
		t.Fatal("a should still claim its own tick")
	}
	if a.Queue().Active() {
		t.Error("a's own tick should have pruned a's queue")
	}
}

func TestToasterExpiresOnRenderWhenATickWasLost(t *testing.T) {
	now := toastEpoch
	tst := fixedToaster(3*time.Second, &now)
	tst.Push(notify.Neutral("stale", ""))

	now = toastEpoch.Add(time.Minute)
	if got := tst.Body(toastBody(), 80); strings.Contains(got, "stale") {
		t.Fatalf("Body should prune against the clock:\n%s", got)
	}
}

func TestToasterPruneClearsTheSnapshotForPanelViews(t *testing.T) {
	now := toastEpoch
	tst := fixedToaster(3*time.Second, &now)
	tst.Push(notify.Neutral("one", ""))
	tst.PushFor(notify.Neutral("two", ""), time.Hour)

	now = toastEpoch.Add(10 * time.Second)
	tst.Prune()
	snap := tst.Queue().Snapshot()
	if len(snap) != 1 || snap[0].Title != "two" {
		t.Fatalf("snapshot after Prune = %v, want only the long-lived toast", snap)
	}
}

func TestToasterPushForOverridesTheDefaultTTL(t *testing.T) {
	now := toastEpoch
	tst := fixedToaster(time.Second, &now)
	tst.PushFor(notify.Neutral("long", ""), time.Hour)

	now = toastEpoch.Add(10 * time.Minute)
	if !strings.Contains(tst.Body(toastBody(), 80), "long") {
		t.Error("PushFor's TTL should win over the Toaster default")
	}
}

func TestToasterIgnoresOtherMessages(t *testing.T) {
	now := toastEpoch
	tst := fixedToaster(3*time.Second, &now)
	if cmd, handled := tst.Update(tea.KeyMsg{Type: tea.KeyEsc}); handled || cmd != nil {
		t.Fatalf("a key should pass through: cmd=%v handled=%v", cmd, handled)
	}
	if cmd, handled := tst.Update(struct{}{}); handled || cmd != nil {
		t.Fatalf("an unrelated message should pass through: cmd=%v handled=%v", cmd, handled)
	}
}

func TestToasterOverlayPosIsConfigurable(t *testing.T) {
	now := toastEpoch
	tst := fixedToaster(3*time.Second, &now)
	if tst.pos.XFrac != 0.5 || tst.pos.YFrac != 0 {
		t.Fatalf("default overlay position = %+v, want top centre", tst.pos)
	}
	tst.Push(notify.Neutral("Saved", ""))
	top := tst.Body(toastBody(), 80)

	tst.SetOverlayPos(layout.OverlayPos{XFrac: 0, YFrac: 1})
	bottom := tst.Body(toastBody(), 80)
	if top == bottom {
		t.Error("SetOverlayPos should move the card")
	}
	if !strings.Contains(bottom, "Saved") {
		t.Error("the toast should still render after a move")
	}
}
