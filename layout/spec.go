package layout

import (
	"encoding/json"
	"fmt"
)

// Params carries layout options from a ScreenSpec (typically decoded JSON or
// YAML) into a LayoutFactory.
type Params map[string]any

// Int returns the value under key as an int, accepting int, int64, float64
// (truncated), and json.Number; missing keys and any other type return def.
func (p Params) Int(key string, def int) int {
	v, ok := p[key]
	if !ok {
		return def
	}
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	case json.Number:
		if i, err := n.Int64(); err == nil {
			return int(i)
		}
	}
	return def
}

// PaneRef names a registered pane in a ScreenSpec and optionally overrides
// the placement fields (Pos, MinTier, Slim) of the Pane its factory builds.
// Slim only sets the flag — a spec cannot clear a factory's Slim.
type PaneRef struct {
	Key     string   `json:"key"`
	Pos     *GridPos `json:"pos,omitempty"`
	MinTier *Tier    `json:"minTier,omitempty"`
	Slim    bool     `json:"slim,omitempty"`
}

// ScreenSpec is the declarative, serializable description of a Screen: a
// layout key with its params plus the panes to place. BuildScreen resolves it
// against a Registry.
type ScreenSpec struct {
	Layout       string    `json:"layout"`
	LayoutParams Params    `json:"layoutParams,omitempty"`
	Panes        []PaneRef `json:"panes"`
}

// BuildScreen resolves a ScreenSpec against the registry. Unknown layout or
// pane keys (and a nil registry) are errors, but a pane factory returning
// ok=false just omits that pane. Spec-level Pos/MinTier/Slim overrides are
// applied to each built pane.
func BuildScreen[Ctx any](s ScreenSpec, ctx Ctx, r *Registry[Ctx]) (Screen, error) {
	if r == nil {
		return Screen{}, fmt.Errorf("layout: nil registry")
	}

	lf, ok := r.layouts[s.Layout]
	if !ok {
		return Screen{}, fmt.Errorf("layout: unknown layout %q", s.Layout)
	}
	l, err := lf(s.LayoutParams)
	if err != nil {
		return Screen{}, fmt.Errorf("layout: build layout %q: %w", s.Layout, err)
	}
	if l == nil {
		return Screen{}, fmt.Errorf("layout: layout %q produced nil Arranger", s.Layout)
	}

	panes := make([]Pane, 0, len(s.Panes))
	for _, ref := range s.Panes {
		pf, ok := r.panes[ref.Key]
		if !ok {
			return Screen{}, fmt.Errorf("layout: unknown pane %q", ref.Key)
		}
		p, ok := pf(ctx)
		if !ok {
			continue
		}
		if ref.Pos != nil {
			pos := *ref.Pos
			p.Pos = &pos
		}
		if ref.MinTier != nil {
			p.MinTier = *ref.MinTier
		}
		if ref.Slim {
			p.Slim = true
		}
		panes = append(panes, p)
	}

	return Screen{Layout: l, Panes: panes}, nil
}
