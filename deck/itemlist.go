package deck

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/viewkit/browser"
	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/list"
	"github.com/codyconfer/viewkit/ui"
)

// ItemListSpec configures an ItemList, following the FormSpec pattern: every
// field beyond Title is optional, and a struct keeps the call sites readable
// while letting the type grow a knob without breaking anyone.
type ItemListSpec struct {
	// Title heads the breadcrumb trail.
	Title string
	// Ctx supplies chrome context cues.
	Ctx []keys.Hint

	// Fetch loads opaque content once on Init (optional if Bind alone is enough).
	Fetch func() any
	// Bind turns fetched data into list rows. Called on each refresh after load.
	Bind func(width int, fetched any) []list.Item

	// IsCancel reports whether a key string should pop the view.
	IsCancel func(string) bool
	// IsAction maps a key string to a keys.Action. It is consulted first; keys
	// it does not claim fall back to the active scheme map (keys.Cur), so
	// paging and navigation keep working without re-mapping them.
	IsAction func(string) (keys.Action, bool)
	// OnOpen overrides browser.Open when an item Key is opened.
	OnOpen func(url string) error
	// OnSelect handles Confirm on the highlighted row. The full item is passed
	// so its Payload (and Key) are available without a side table.
	OnSelect func(h *Model, it list.Item) tea.Cmd
	// LoadingText shown before load completes (default "░▒▓ loading…").
	LoadingText string
	// ReloadHint is the footer label for the reload key (default "reload").
	ReloadHint string
}

// ItemList is a lazy-loaded selectable list View.
// Fetch runs once on Init; Bind maps fetched data → rows using the current width
// so callers can width-wrap without importing domain types into deck.
type ItemList struct {
	title string
	ctx   []keys.Hint

	// Fetch loads opaque content once. Exported so hosts can wrap or replace
	// the loader after construction (e.g. tests counting fetches).
	Fetch func() any
	// OnOpen overrides browser.Open when an item Key is opened.
	OnOpen func(url string) error
	// OnSelect handles Confirm on the highlighted row.
	OnSelect func(h *Model, it list.Item) tea.Cmd

	bind        func(width int, fetched any) []list.Item
	isCancel    func(string) bool
	isAction    func(string) (keys.Action, bool)
	loadingText string
	reloadHint  string

	lst     list.Model
	width   int
	height  int
	ready   bool
	loaded  bool
	fetched any
	scope   *ui.Scope

	boundUI *ui.Scope
	bound   bool
}

// NewItemList builds an ItemList from spec.
func NewItemList(spec ItemListSpec) *ItemList {
	r := &ItemList{
		title:       spec.Title,
		ctx:         spec.Ctx,
		Fetch:       spec.Fetch,
		OnOpen:      spec.OnOpen,
		OnSelect:    spec.OnSelect,
		bind:        spec.Bind,
		isCancel:    spec.IsCancel,
		isAction:    spec.IsAction,
		loadingText: spec.LoadingText,
		reloadHint:  spec.ReloadHint,
		lst:         list.New(),
	}
	if r.loadingText == "" {
		r.loadingText = "░▒▓ loading…"
	}
	r.lst.SetFocused(true)
	return r
}

type itemListLoadedMsg struct {
	own  *ItemList
	data any
}

func (m itemListLoadedMsg) recipient() View { return m.own }

// Title implements View.
func (r *ItemList) Title() string { return r.title }

// Context implements View.
func (r *ItemList) Context(scope *ui.Scope) []keys.Hint { return r.ctx }

// Selected returns the highlighted row, if any. Hosts that decorate an
// ItemList with their own row commands need it to know what the command
// applies to.
func (r *ItemList) Selected() (list.Item, bool) { return r.lst.Selected() }

// Hints implements View; the legend adapts to OnSelect and Fetch being set.
func (r *ItemList) Hints(scope *ui.Scope) []keys.Hint {
	km := navMapFor(schemeOf(scope))
	hints := []keys.Hint{km.HintLabeled(keys.Up, "move")}
	if r.OnSelect != nil {
		hints = append(hints,
			km.HintLabeled(keys.Confirm, "details"),
			km.HintLabeled(keys.Open, "open"))
	} else {
		hints = append(hints, km.HintLabeled(keys.Confirm, "open"))
	}
	hints = append(hints, km.HintLabeled(keys.PageUp, "page"))
	if r.Fetch != nil {
		hints = append(hints, km.HintLabeled(keys.Reload, r.reloadLabel()))
	}
	return hints
}

func (r *ItemList) reloadLabel() string {
	if r.reloadHint != "" {
		return r.reloadHint
	}
	return "reload"
}

// Init implements View, kicking off Fetch — or reporting an immediate empty
// load when there is nothing to fetch.
func (r *ItemList) Init() tea.Cmd {
	if r.Fetch == nil {
		return func() tea.Msg { return itemListLoadedMsg{own: r} }
	}
	return r.fetchCmd()
}

// Reload drops fetched content and re-runs Fetch, showing the loading text
// until the new data lands. No-op when there is nothing to fetch.
func (r *ItemList) Reload() tea.Cmd {
	if r.Fetch == nil {
		return nil
	}
	r.fetched, r.loaded = nil, false
	r.refresh()
	return r.fetchCmd()
}

func (r *ItemList) fetchCmd() tea.Cmd {
	fetch := r.Fetch
	return func() tea.Msg { return itemListLoadedMsg{own: r, data: fetch()} }
}

// Update implements View, handling resize, load completion, ReloadMsg and
// navigation keys.
func (r *ItemList) Update(h *Model, msg tea.Msg) tea.Cmd {
	r.scope = h.UI()
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		if r.ready && m.Width == r.width && r.bound && r.boundUI == r.scope {
			return nil
		}
		r.width, r.ready = m.Width, true
		r.refresh()
		return nil
	case itemListLoadedMsg:
		r.fetched, r.loaded = m.data, true
		r.refresh()
		return nil
	case ReloadMsg:
		return r.Reload()
	case tea.KeyMsg:
		return r.handleKey(h, m)
	}
	return nil
}

func (r *ItemList) handleKey(h *Model, m tea.KeyMsg) tea.Cmd {
	act, ok := r.action(schemeOf(h.UI()), m.String())
	if !ok {
		return nil
	}
	switch act {
	case keys.Up:
		r.lst.Move(-1)
	case keys.Down:
		r.lst.Move(1)
	case keys.PageUp:
		r.lst.Scroll(-max(r.height, 1))
	case keys.PageDown:
		r.lst.Scroll(max(r.height, 1))
	case keys.Confirm:
		return r.confirmSelected(h)
	case keys.Open:
		return r.openSelected()
	case keys.Reload:
		return r.Reload()
	case keys.Cancel:
		return h.Pop()
	}
	return nil
}

func (r *ItemList) confirmSelected(h *Model) tea.Cmd {
	if r.OnSelect == nil {
		return r.openSelected()
	}
	it, ok := r.lst.Selected()
	if !ok || (it.Key == "" && it.Payload == nil) {
		return nil
	}
	return r.OnSelect(h, it)
}

func (r *ItemList) action(sc keys.Scheme, key string) (keys.Action, bool) {
	if r.isCancel != nil && r.isCancel(key) {
		return keys.Cancel, true
	}
	if r.isAction != nil {
		if act, ok := r.isAction(key); ok {
			return act, true
		}
	}
	return navMapFor(sc).Action(key)
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
	r.boundUI, r.bound = r.scope, true
	if r.bind != nil {
		r.lst.SetItemsKeepingCursor(r.bind(r.width, r.fetched))
		return
	}
	if items, ok := r.fetched.([]list.Item); ok {
		r.lst.SetItemsKeepingCursor(items)
	}
}

// Body implements View, rendering the loading text until Fetch lands.
func (r *ItemList) Body(f layout.Frame) string {
	width, height := f.Width, f.Height
	r.lst.SetTheme(f.Theme())
	if !r.loaded {
		txt := r.loadingText
		if txt == "" {
			txt = "░▒▓ loading…"
		}
		return f.Theme().Dim.Render(txt)
	}
	if height > 0 {
		changed := height != r.height
		r.height = height
		r.lst.SetSize(width, height)
		if changed {
			r.lst.EnsureVisible()
		}
	}
	return r.lst.View()
}
