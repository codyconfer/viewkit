package layout

import "sort"

// PaneFactory builds a Pane from the caller's context. Returning ok=false
// omits the pane from the screen (BuildScreen skips it silently).
type PaneFactory[Ctx any] func(ctx Ctx) (pane Pane, ok bool)

// LayoutFactory builds an Arranger from spec parameters.
type LayoutFactory func(params Params) (Arranger, error) //nolint:revive // name predates the lint rule; renaming would break callers

// PaneInfo is a registered pane's key and human-readable title, as returned
// by Registry.PaneKeys.
type PaneInfo struct {
	Key   string
	Title string
}

// Registry maps spec keys to pane and layout factories for BuildScreen. Ctx is
// whatever state pane factories need at build time. Use NewRegistry — it
// pre-registers the built-in layouts.
type Registry[Ctx any] struct {
	panes    map[string]PaneFactory[Ctx]
	paneInfo map[string]PaneInfo
	order    []string
	layouts  map[string]LayoutFactory
}

// NewRegistry returns a Registry with the built-in layouts pre-registered:
// "single" (SingleColumn), "flex-columns", "flex-rows", "sections" (all
// honoring minWidth/maxCols params), and "grid" (cols/rows params).
func NewRegistry[Ctx any]() *Registry[Ctx] {
	r := &Registry[Ctx]{
		panes:    map[string]PaneFactory[Ctx]{},
		paneInfo: map[string]PaneInfo{},
		layouts:  map[string]LayoutFactory{},
	}
	r.LayoutFn("single", func(Params) (Arranger, error) { return SingleColumn{}, nil })
	r.LayoutFn("flex-columns", func(p Params) (Arranger, error) {
		return FlexColumns{
			MinWidth: p.Int("minWidth", DefaultFlexMinWidth),
			MaxCols:  p.Int("maxCols", DefaultFlexMaxCols),
		}, nil
	})
	r.LayoutFn("flex-rows", func(p Params) (Arranger, error) {
		return FlexRows{
			MinWidth: p.Int("minWidth", DefaultFlexMinWidth),
			MaxCols:  p.Int("maxCols", DefaultFlexMaxCols),
		}, nil
	})
	r.LayoutFn("sections", func(p Params) (Arranger, error) {
		return FlexSections{
			MinWidth: p.Int("minWidth", DefaultFlexMinWidth),
			MaxCols:  p.Int("maxCols", DefaultFlexMaxCols),
		}, nil
	})
	r.LayoutFn("grid", func(p Params) (Arranger, error) {
		return Grid{Cols: p.Int("cols", 1), Rows: p.Int("rows", 0)}, nil
	})
	return r
}

// Pane registers a pane factory under key and returns the registry for
// chaining. Registering an existing key overwrites the previous factory and
// title in place; the key keeps its original position in PaneKeys order.
func (r *Registry[Ctx]) Pane(key, title string, f PaneFactory[Ctx]) *Registry[Ctx] {
	if _, exists := r.panes[key]; !exists {
		r.order = append(r.order, key)
	}
	r.panes[key] = f
	r.paneInfo[key] = PaneInfo{Key: key, Title: title}
	return r
}

// LayoutFn registers a layout factory under key and returns the registry for
// chaining. Registering an existing key — including a built-in like "grid" —
// replaces it.
func (r *Registry[Ctx]) LayoutFn(key string, f LayoutFactory) *Registry[Ctx] {
	r.layouts[key] = f
	return r
}

// PaneKeys lists registered panes in first-registration order.
func (r *Registry[Ctx]) PaneKeys() []PaneInfo {
	out := make([]PaneInfo, 0, len(r.order))
	for _, k := range r.order {
		out = append(out, r.paneInfo[k])
	}
	return out
}

// LayoutKeys lists registered layout keys, sorted alphabetically.
func (r *Registry[Ctx]) LayoutKeys() []string {
	out := make([]string, 0, len(r.layouts))
	for k := range r.layouts {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
