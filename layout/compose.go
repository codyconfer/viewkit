package layout

// Pane is one composable region of a Screen. Render is called with the Frame
// (width/height/focus) the Arranger allotted; MinTier hides the pane on
// screens shorter than its tier; Pos pins it to grid coordinates; Slim lets
// row-mates take width it does not need; Interactive panes with a Name join
// the focus Ring.
type Pane struct {
	Name        string
	Title       string
	Group       string
	Interactive bool
	MinTier     Tier
	Pos         *GridPos
	Slim        bool
	Render      func(Frame) string
}

// Focused reports whether name selects this pane for focus.
func (p Pane) Focused(name string) bool {
	return p.Interactive && p.Name != "" && p.Name == name
}

// PaneRing builds the focus Ring for a pane set: only interactive panes are
// included, in declaration order.
func PaneRing(panes []Pane) Ring {
	out := make(Ring, 0, len(panes))
	for _, p := range panes {
		if p.Interactive {
			out = append(out, p.Name)
		}
	}
	return out
}

// Arranger positions and renders a set of panes into a single block. f is the
// full area available, tier filters panes by MinTier, and focusedName marks
// which interactive pane renders with a focused Frame. Implementations include
// SingleColumn, Grid, FlexColumns, FlexRows, and FlexSections.
type Arranger interface {
	Arrange(f Frame, tier Tier, panes []Pane, focusedName string) string
}

// Screen pairs a pane set with the Arranger that lays it out. A nil Layout
// falls back to SingleColumn.
type Screen struct {
	Layout Arranger
	Panes  []Pane
}

// Ring returns the focus ring over the screen's interactive panes.
func (s Screen) Ring() Ring { return PaneRing(s.Panes) }

// Render arranges the screen's panes for the given frame and tier, focusing
// the interactive pane named focused. A nil Layout uses SingleColumn.
func (s Screen) Render(f Frame, tier Tier, focused string) string {
	l := s.Layout
	if l == nil {
		l = SingleColumn{}
	}
	return l.Arrange(f, tier, s.Panes, focused)
}

// SingleColumn is the simplest Arranger: every visible pane renders at the
// full frame width and the results are stacked vertically with blank lines
// between them. It ignores Frame.Height, Pos, and Slim.
type SingleColumn struct{}

// Arrange implements Arranger by stacking tier-visible panes top to bottom.
func (SingleColumn) Arrange(f Frame, tier Tier, panes []Pane, focusedName string) string {
	sections := make([]Section, 0, len(panes))
	for _, p := range panes {
		pf := f
		if p.Focused(focusedName) {
			pf = f.Focus()
		}
		sections = append(sections, Section{Content: p.Render(pf), MinTier: p.MinTier})
	}
	return StackFit(tier, sections...)
}
