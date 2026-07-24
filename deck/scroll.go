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

type scrollViewLoadedMsg struct{ body string }

func (c *Scroll) Title() string        { return c.title }
func (c *Scroll) Context() [][2]string { return c.ctx }
func (c *Scroll) Hints() [][2]string {
	return append([][2]string{{"↑/↓", "scroll"}, {"pgup/pgdn", "page"}}, c.hints...)
}

func (c *Scroll) Init() tea.Cmd {
	if c.load == nil {
		return nil
	}
	return func() tea.Msg { return scrollViewLoadedMsg{body: c.load()} }
}

func (c *Scroll) Update(h *Host, msg tea.Msg) tea.Cmd {
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
	case tea.KeyMsg:
		if c.IsCancel != nil && c.IsCancel(m.String()) {
			return h.Pop()
		}
		// Default cancel bindings when host did not inject a checker.
		if c.IsCancel == nil {
			sc := keys.Cur()
			if act, ok := keys.NewMap(sc.Binding(keys.Cancel)).Action(m.String()); ok && act == keys.Cancel {
				return h.Pop()
			}
		}
		if c.ready {
			var cmd tea.Cmd
			c.vp, cmd = c.vp.Update(msg)
			return cmd
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
