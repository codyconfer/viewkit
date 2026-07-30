package deck

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/theme"
)

// Scroll is a lazy-loaded scrollable text view (bubbles viewport).
// Hosted only in the deck module so tea stays out of viewkit core.
type Scroll struct {
	title string
	load  func() string
	hints [][2]string
	ctx   [][2]string

	// ChromeReserve is subtracted from window height for title/status chrome.
	ChromeReserve int
	// IsCancel reports whether a key string should pop the view.
	IsCancel func(string) bool
	// ReloadHint is the footer label for the reload key (default "reload").
	ReloadHint string

	vp     viewport.Model
	ready  bool
	body   string
	loaded bool
}

// NewScroll builds a Scroll view. load is invoked once on Init.
func NewScroll(title string, ctx, hints [][2]string, load func() string) *Scroll {
	return &Scroll{
		title:         title,
		load:          load,
		ctx:           ctx,
		hints:         hints,
		ChromeReserve: 7,
	}
}

type scrollViewLoadedMsg struct {
	own  *Scroll
	body string
}

func (m scrollViewLoadedMsg) recipient() View { return m.own }

func (c *Scroll) Title() string        { return c.title }
func (c *Scroll) Context() [][2]string { return c.ctx }
func (c *Scroll) Hints() [][2]string {
	km := navMap()
	hints := [][2]string{
		km.HintLabeled(keys.Up, "scroll"),
		km.HintLabeled(keys.PageUp, "page"),
	}
	if c.load != nil {
		hints = append(hints, km.HintLabeled(keys.Reload, c.reloadHint()))
	}
	return append(hints, c.hints...)
}

func (c *Scroll) reloadHint() string {
	if c.ReloadHint != "" {
		return c.ReloadHint
	}
	return "reload"
}

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

func (c *Scroll) Update(h *Model, msg tea.Msg) tea.Cmd {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		reserve := c.ChromeReserve
		if reserve <= 0 {
			reserve = 7
		}
		height := max(m.Height-reserve, 1)
		if !c.ready {
			c.vp = viewport.New(m.Width, height)
			c.ready = true
		} else {
			c.vp.Width, c.vp.Height = m.Width, height
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
		if c.IsCancel != nil && c.IsCancel(m.String()) {
			return h.Pop()
		}
		act, ok := navMap().Action(m.String())
		if !ok {
			return nil
		}
		if act == keys.Cancel && c.IsCancel == nil {
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

func (c *Scroll) Body(width, height int) string {
	if !c.ready {
		return theme.Cur().Dim.Render("loading…")
	}
	return c.vp.View()
}
