package deck

type hintView struct {
	View
	extra [][2]string
}

// WithExtraHints wraps inner so the footer legend carries extra after the
// view's own hints. Every other method passes through untouched. inner is
// returned unchanged when extra is empty, so the wrapper never costs a level
// of indirection it does not earn.
func WithExtraHints(inner View, extra [][2]string) View {
	if len(extra) == 0 {
		return inner
	}
	return &hintView{View: inner, extra: extra}
}

func (h *hintView) Hints() [][2]string {
	return append(append([][2]string{}, h.View.Hints()...), h.extra...)
}

type liveContextView struct {
	View
	ctx func() [][2]string
}

// WithLiveContext wraps inner so its chrome context cues are recomputed on
// every render rather than captured when the view was built — the shape a
// long-lived view needs when the cue (active role, selection, count) can
// change beneath it. Every other method passes through untouched. inner is
// returned unchanged when fn is nil.
func WithLiveContext(inner View, fn func() [][2]string) View {
	if fn == nil {
		return inner
	}
	return &liveContextView{View: inner, ctx: fn}
}

func (v *liveContextView) Context() [][2]string { return v.ctx() }
