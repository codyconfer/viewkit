package deck

import (
	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/ui"
)

type hintView struct {
	View
	extra []keys.Hint
}

// WithExtraHints wraps inner so the footer legend carries extra after the
// view's own hints. Every other method passes through untouched. inner is
// returned unchanged when extra is empty, so the wrapper never costs a level
// of indirection it does not earn.
func WithExtraHints(inner View, extra []keys.Hint) View {
	if len(extra) == 0 {
		return inner
	}
	return &hintView{View: inner, extra: extra}
}

func (h *hintView) Hints(scope *ui.Scope) []keys.Hint {
	return append(append([]keys.Hint{}, h.View.Hints(scope)...), h.extra...)
}

type liveContextView struct {
	View
	ctx func() []keys.Hint
}

// WithLiveContext wraps inner so its chrome context cues are recomputed on
// every render rather than captured when the view was built — the shape a
// long-lived view needs when the cue (active role, selection, count) can
// change beneath it. Every other method passes through untouched. inner is
// returned unchanged when fn is nil.
func WithLiveContext(inner View, fn func() []keys.Hint) View {
	if fn == nil {
		return inner
	}
	return &liveContextView{View: inner, ctx: fn}
}

func (v *liveContextView) Context(scope *ui.Scope) []keys.Hint { return v.ctx() }
