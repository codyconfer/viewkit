package deck

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/theme"
)

// ScrollSpec configures a Scroll, following the FormSpec pattern: every field
// beyond Title and Load is optional.
type ScrollSpec struct {
	// Title heads the breadcrumb trail.
	Title string
	// Ctx supplies chrome context cues.
	Ctx []keys.Hint
	// Hints are appended to the footer legend after the built-in ones.
	Hints []keys.Hint
	// Load produces the scrollable text; invoked once on Init.
	Load func() string
	// IsCancel reports whether a key string should pop the view.
	IsCancel func(string) bool
	// ReloadHint is the footer label for the reload key (default "reload").
	ReloadHint string
}

// Scroll is a lazy-loaded scrollable text view (bubbles viewport).
// Hosted only in the deck module so tea stays out of viewkit core.
type Scroll struct {
	title string
	load  func() string
	hints []keys.Hint
	ctx   []keys.Hint

	isCancel   func(string) bool
	reloadHint string

	vp     viewport.Model
	ready  bool
	body   string
	loaded bool
}

// NewScroll builds a Scroll view from spec.
func NewScroll(spec ScrollSpec) *Scroll {
	return &Scroll{
		title:      spec.Title,
		load:       spec.Load,
		ctx:        spec.Ctx,
		hints:      spec.Hints,
		isCancel:   spec.IsCancel,
		reloadHint: spec.ReloadHint,
	}
}

type scrollViewLoadedMsg struct {
	own  *Scroll
	body string
}

func (m scrollViewLoadedMsg) recipient() View { return m.own }

// Title implements View.
func (c *Scroll) Title() string { return c.title }

// Context implements View.
func (c *Scroll) Context() []keys.Hint { return c.ctx }

// Hints implements View: scrolling legend plus any spec-supplied hints.
func (c *Scroll) Hints() []keys.Hint {
	km := navMap()
	hints := []keys.Hint{
		km.HintLabeled(keys.Up, "scroll"),
		km.HintLabeled(keys.PageUp, "page"),
	}
	if c.load != nil {
		hints = append(hints, km.HintLabeled(keys.Reload, c.reloadLabel()))
	}
	return append(hints, c.hints...)
}

func (c *Scroll) reloadLabel() string {
	if c.reloadHint != "" {
		return c.reloadHint
	}
	return "reload"
}

// Init implements View, kicking off Load when one was supplied.
func (c *Scroll) Init() tea.Cmd {
	if c.load == nil {
		return nil
	}
	return c.loadCmd()
}

// Reload discards the rendered body and re-runs load, showing the loading
// placeholder until the new text lands. No-op when there is nothing to load.
func (c *Scroll) Reload() tea.Cmd {
	if c.load == nil {
		return nil
	}
	c.body, c.loaded = "", false
	c.refresh()
	return c.loadCmd()
}

func (c *Scroll) loadCmd() tea.Cmd {
	load := c.load
	return func() tea.Msg { return scrollViewLoadedMsg{own: c, body: load()} }
}

// Update implements View: navigation keys scroll the viewport, Reload
// re-runs Load, Cancel (or an IsCancel match) pops the view.
func (c *Scroll) Update(h *Model, msg tea.Msg) tea.Cmd {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		if !c.ready {
			c.vp = viewport.New(m.Width, m.Height)
			c.ready = true
		} else {
			c.vp.Width = m.Width
		}
		c.refresh()
		return nil
	case scrollViewLoadedMsg:
		c.body, c.loaded = m.body, true
		c.refresh()
		return nil
	case ReloadMsg:
		return c.Reload()
	case tea.KeyMsg:
		if c.isCancel != nil && c.isCancel(m.String()) {
			return h.Pop()
		}
		act, ok := navMap().Action(m.String())
		if !ok {
			return nil
		}
		if act == keys.Cancel && c.isCancel == nil {
			return h.Pop()
		}
		if act == keys.Reload {
			return c.Reload()
		}
		if !c.ready {
			return nil
		}
		switch act {
		case keys.Up:
			c.vp.ScrollUp(1)
		case keys.Down:
			c.vp.ScrollDown(1)
		case keys.PageUp:
			c.vp.PageUp()
		case keys.PageDown:
			c.vp.PageDown()
		}
	}
	return nil
}

func (c *Scroll) refresh() {
	if !c.ready {
		return
	}
	if !c.loaded {
		c.vp.SetContent(theme.Cur().Dim.Render("░▒▓ loading…"))
		return
	}
	c.vp.SetContent(c.body)
}

// Body implements View, resizing the viewport to the frame before rendering
// so the scroll offset stays within the content.
func (c *Scroll) Body(width, height int) string {
	if !c.ready {
		return theme.Cur().Dim.Render("loading…")
	}
	c.vp.Width = width
	if height > 0 {
		c.vp.Height = height
	}
	return c.vp.View()
}
