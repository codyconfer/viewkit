package deck

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/viewkit/browser"
	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/list"
	"github.com/codyconfer/viewkit/theme"
)

// ItemList is a lazy-loaded selectable list View.
// Fetch runs once on Init; Bind maps fetched data → rows using the current width
// so callers can width-wrap without importing domain types into deck.
type ItemList struct {
	title string
	ctx   [][2]string

	// Fetch loads opaque content once (optional if Bind alone is enough).
	Fetch func() any
	// Bind turns fetched data into list rows. Called on each refresh after load.
	Bind func(width int, fetched any) []list.Item

	// ChromeReserve is subtracted from window height for title/status chrome.
	ChromeReserve int
	// IsCancel reports whether a key string should pop the view.
	IsCancel func(string) bool
	// IsAction maps a key string to a keys.Action (defaults to keys.Cur menu map).
	IsAction func(string) (keys.Action, bool)
	// OnOpen overrides browser.Open when an item Key is confirmed.
	OnOpen func(url string) error
	// LoadingText shown before load completes.
	LoadingText string

	lst     list.Model
	width   int
	height  int
	ready   bool
	loaded  bool
	fetched any
}

// NewItemList builds an ItemList with Fetch+Bind.
func NewItemList(title string, ctx [][2]string, fetch func() any, bind func(width int, fetched any) []list.Item) *ItemList {
	r := &ItemList{
		title:         title,
		ctx:           ctx,
		Fetch:         fetch,
		Bind:          bind,
		ChromeReserve: 7,
		LoadingText:   "░▒▓ loading…",
		lst:           list.New(),
	}
	r.lst.SetFocused(true)
	return r
}

type itemListLoadedMsg struct{ data any }

func (r *ItemList) Title() string        { return r.title }
func (r *ItemList) Context() [][2]string { return r.ctx }
func (r *ItemList) Hints() [][2]string {
	return [][2]string{{"↑/↓", "move"}, {"enter", "open"}, {"pgup/pgdn", "page"}}
}

func (r *ItemList) Init() tea.Cmd {
	if r.Fetch == nil {
		return func() tea.Msg { return itemListLoadedMsg{} }
	}
	return func() tea.Msg { return itemListLoadedMsg{data: r.Fetch()} }
}

func (r *ItemList) Update(h *Model, msg tea.Msg) tea.Cmd {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		r.width = m.Width
		r.height = max(m.Height-r.ChromeReserve, 1)
		r.ready = true
		r.refresh()
		return nil
	case itemListLoadedMsg:
		r.fetched, r.loaded = m.data, true
		r.refresh()
		return nil
	case tea.KeyMsg:
		return r.handleKey(h, m)
	}
	return nil
}

func (r *ItemList) handleKey(h *Model, m tea.KeyMsg) tea.Cmd {
	switch m.String() {
	case "pgup":
		r.lst.Scroll(-r.height)
		return nil
	case "pgdown":
		r.lst.Scroll(r.height)
		return nil
	}
	act, ok := r.action(m.String())
	if !ok {
		return nil
	}
	switch act {
	case keys.Up:
		r.lst.Move(-1)
	case keys.Down:
		r.lst.Move(1)
	case keys.Confirm:
		return r.openSelected()
	case keys.Cancel:
		return h.Pop()
	}
	return nil
}

func (r *ItemList) action(key string) (keys.Action, bool) {
	if r.IsCancel != nil && r.IsCancel(key) {
		return keys.Cancel, true
	}
	if r.IsAction != nil {
		return r.IsAction(key)
	}
	sc := keys.Cur()
	km := keys.NewMap(
		sc.Binding(keys.Up),
		sc.Binding(keys.Down),
		sc.Binding(keys.Confirm),
		sc.Binding(keys.Cancel),
		sc.Binding(keys.Quit),
	)
	return km.Action(key)
}

func (r *ItemList) openSelected() tea.Cmd {
	it, ok := r.lst.Selected()
	if !ok || it.Key == "" {
		return nil
	}
	url := it.Key
	open := r.OnOpen
	if open == nil {
		open = browser.Open
	}
	return func() tea.Msg {
		_ = open(url)
		return nil
	}
}

func (r *ItemList) refresh() {
	if !r.ready || !r.loaded {
		return
	}
	r.lst.SetSize(r.width, r.height)
	if r.Bind != nil {
		r.lst.SetItems(r.Bind(r.width, r.fetched))
		return
	}
	if items, ok := r.fetched.([]list.Item); ok {
		r.lst.SetItems(items)
	}
}

func (r *ItemList) Body(width, height int) string {
	if !r.loaded {
		txt := r.LoadingText
		if txt == "" {
			txt = "░▒▓ loading…"
		}
		return theme.Cur().Dim.Render(txt)
	}
	return r.lst.View()
}
