package deck

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/viewkit/browser"
	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/list"
	"github.com/codyconfer/viewkit/theme"
)

const (
	homeFocusMenu = iota
	homeFocusSide
)

// HomeShell is a dual-pane View: navigable menu + optional side Item list.
// SideFetch/SideBind keep domain types out of deck (caller maps → list.Item).
type HomeShell struct {
	title string
	ctx   [][2]string
	items []MenuItem

	// BoxTitle is the menu TitledBox title. Empty (default) omits the title.
	BoxTitle string
	// SideLabel is shown above the side list when SideFetch is set.
	SideLabel string
	// SideLabelFn, when set, supplies the side label on every read instead of
	// SideLabel — the shape a long-lived shell needs when the side pane's
	// subject can change beneath it. Returning "" hides the side pane.
	SideLabelFn func() string
	// SideHint is the tab-target label when menu-focused (default "side").
	SideHint string
	// SideFetch loads opaque side content once (nil disables the side pane).
	SideFetch func() any
	// SideBind maps fetched data → rows using current width.
	SideBind func(width int, fetched any) []list.Item
	// SideLoading shown while SideFetch has not completed.
	SideLoading string
	// ReloadHint is the footer label for the reload key (default "reload").
	ReloadHint string

	// IsAction maps a key string to a keys.Action, fully replacing the active
	// scheme map (keys.Cur); map PageUp/PageDown and FocusNext to keep paging
	// and pane switching.
	IsAction func(string) (keys.Action, bool)
	// OnOpen overrides browser.Open when a side item Key is opened.
	OnOpen   func(url string) error
	OnSelect func(h *Model, key string) tea.Cmd

	cursor  int
	focus   int
	side    list.Model
	width   int
	ready   bool
	loaded  bool
	fetched any
	bound   sideBind
}

type sideBind struct {
	width int
	label string
	theme *theme.Theme
}

// NewHomeShell builds a HomeShell. sideLabel empty + SideFetch nil → menu-only.
// BoxTitle defaults empty (no menu titled-box title); set after construction if desired.
func NewHomeShell(title string, ctx [][2]string, items []MenuItem, sideLabel string) *HomeShell {
	return &HomeShell{
		title:       title,
		ctx:         ctx,
		items:       items,
		SideLabel:   sideLabel,
		SideHint:    "side",
		SideLoading: "░▒▓ loading…",
		side:        list.New(),
	}
}

type homeShellLoadedMsg struct {
	own  *HomeShell
	data any
}

func (m homeShellLoadedMsg) recipient() View { return m.own }

// ReloadMsg asks the active view to discard what it fetched and fetch again.
// Views that load lazily honour it; the rest ignore it, so it is safe to emit
// from a global key hook without knowing what is on top of the stack.
type ReloadMsg struct{}

func (h *HomeShell) Title() string        { return h.title }
func (h *HomeShell) Context() [][2]string { return h.ctx }

// FocusSide reports whether the side pane has keyboard focus (for tests/adapters).
func (h *HomeShell) FocusSide() bool { return h.focus == homeFocusSide }

func (h *HomeShell) Hints() [][2]string {
	km := navMap()
	if h.focus == homeFocusSide {
		hints := [][2]string{km.HintLabeled(keys.Up, "move")}
		if h.OnSelect != nil {
			hints = append(hints,
				km.HintLabeled(keys.Confirm, "details"),
				km.HintLabeled(keys.Open, "open"))
		} else {
			hints = append(hints, km.HintLabeled(keys.Confirm, "open"))
		}
		hints = append(hints,
			km.HintLabeled(keys.PageUp, "page"),
			km.HintLabeled(keys.FocusNext, "menu"))
		if h.SideFetch != nil {
			hints = append(hints, km.HintLabeled(keys.Reload, h.reloadHint()))
		}
		return hints
	}
	hints := [][2]string{
		km.HintLabeled(keys.Up, "move"),
		km.HintLabeled(keys.Confirm, "open"),
	}
	if h.hasSide() {
		hint := h.SideHint
		if hint == "" {
			hint = "side"
		}
		hints = append(hints, km.HintLabeled(keys.FocusNext, hint))
	}
	if h.SideFetch != nil {
		hints = append(hints, km.HintLabeled(keys.Reload, h.reloadHint()))
	}
	return hints
}

func (h *HomeShell) hasSide() bool { return h.sideLabel() != "" && h.SideFetch != nil }

func (h *HomeShell) reloadHint() string {
	if h.ReloadHint != "" {
		return h.ReloadHint
	}
	return "reload"
}

func (h *HomeShell) sideLabel() string {
	if h.SideLabelFn != nil {
		return h.SideLabelFn()
	}
	return h.SideLabel
}

func (h *HomeShell) Init() tea.Cmd {
	if !h.hasSide() {
		return nil
	}
	return h.fetchSide()
}

// Reload drops fetched side content and re-runs SideFetch, showing SideLoading
// until the new data lands. No-op when there is nothing to fetch.
func (h *HomeShell) Reload() tea.Cmd {
	if h.SideFetch == nil {
		return nil
	}
	h.fetched, h.loaded = nil, false
	h.refresh()
	return h.fetchSide()
}

func (h *HomeShell) fetchSide() tea.Cmd {
	fetch := h.SideFetch
	return func() tea.Msg { return homeShellLoadedMsg{own: h, data: fetch()} }
}

func (h *HomeShell) Update(host *Model, msg tea.Msg) tea.Cmd {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		h.width = m.Width
		if h.ready && h.bound == h.bindKey() {
			return nil
		}
		h.ready = true
		h.refresh()
		return nil
	case homeShellLoadedMsg:
		h.fetched, h.loaded = m.data, true
		h.refresh()
		return nil
	case ReloadMsg:
		return h.Reload()
	case tea.KeyMsg:
		return h.handleKey(host, m)
	}
	return nil
}

func (h *HomeShell) handleKey(host *Model, m tea.KeyMsg) tea.Cmd {
	act, ok := h.action(m.String())
	if !ok {
		return nil
	}
	if h.hasSide() && (act == keys.FocusNext || act == keys.FocusPrev) {
		h.toggleFocus()
		return nil
	}
	if act == keys.Reload && h.SideFetch != nil {
		return h.Reload()
	}
	if h.focus == homeFocusSide {
		switch act {
		case keys.Up:
			h.side.Move(-1)
		case keys.Down:
			h.side.Move(1)
		case keys.PageUp:
			h.side.Scroll(-max(h.side.Height(), 1))
		case keys.PageDown:
			h.side.Scroll(max(h.side.Height(), 1))
		case keys.Confirm:
			return h.confirmSelected(host)
		case keys.Open:
			return h.openSelected()
		case keys.Cancel:
			h.focus = homeFocusMenu
			h.side.SetFocused(false)
		}
		return nil
	}
	switch act {
	case keys.Up:
		if h.cursor > 0 {
			h.cursor--
		}
	case keys.Down:
		if h.cursor < len(h.items)-1 {
			h.cursor++
		}
	case keys.Confirm:
		if len(h.items) > 0 && h.items[h.cursor].Do != nil {
			return h.items[h.cursor].Do(host)
		}
	case keys.Cancel:
		return host.Pop()
	}
	return nil
}

func (h *HomeShell) toggleFocus() {
	if h.focus == homeFocusMenu {
		h.focus = homeFocusSide
	} else {
		h.focus = homeFocusMenu
	}
	h.side.SetFocused(h.focus == homeFocusSide)
}

func (h *HomeShell) action(key string) (keys.Action, bool) {
	if h.IsAction != nil {
		return h.IsAction(key)
	}
	return navMap().Action(key)
}

func (h *HomeShell) confirmSelected(host *Model) tea.Cmd {
	if h.OnSelect == nil {
		return h.openSelected()
	}
	it, ok := h.side.Selected()
	if !ok || it.Key == "" {
		return nil
	}
	return h.OnSelect(host, it.Key)
}

func (h *HomeShell) openSelected() tea.Cmd {
	it, ok := h.side.Selected()
	if !ok || it.Key == "" {
		return nil
	}
	url := it.Key
	open := h.OnOpen
	if open == nil {
		open = browser.Open
	}
	return func() tea.Msg {
		_ = open(url)
		return nil
	}
}

func (h *HomeShell) bindKey() sideBind {
	return sideBind{width: h.width, label: h.sideLabel(), theme: theme.Cur()}
}

func (h *HomeShell) refresh() {
	h.bound = h.bindKey()
	if !h.hasSide() || h.width == 0 {
		return
	}
	th := theme.Cur()
	if !h.loaded {
		txt := h.SideLoading
		if txt == "" {
			txt = "░▒▓ loading…"
		}
		h.side.SetItems([]list.Item{{Block: th.Dim.Render(txt)}})
		return
	}
	if h.SideBind != nil {
		h.side.SetItemsKeepingCursor(h.SideBind(h.width, h.fetched))
		return
	}
	if items, ok := h.fetched.([]list.Item); ok {
		h.side.SetItemsKeepingCursor(items)
	}
}

func (h *HomeShell) menuRows(f layout.Frame) []string {
	th := theme.Cur()
	rows := make([]string, len(h.items))
	for i, it := range h.items {
		cursor := "  "
		label := th.Val.Render(it.Label)
		switch {
		case i == h.cursor && h.focus == homeFocusMenu:
			cursor = th.Key.Render("▸ ")
			label = th.Key.Render(it.Label)
		case i == h.cursor:
			cursor = th.Dim.Render("▸ ")
		}
		row := cursor
		if it.Icon != "" {
			row += theme.Icon(it.Icon, it.Hue)
		}
		row += label
		if it.Desc != "" {
			row = f.Spread(row, th.Dim.Render(it.Desc))
		}
		rows[i] = row
	}
	return rows
}

func (h *HomeShell) Body(width, height int) string {
	f := layout.ScreenFrame(width)
	f.Focused = h.focus == homeFocusMenu
	menuBox := f.TitledBox(h.BoxTitle, h.menuRows(f)...)
	if !h.hasSide() {
		return menuBox
	}
	th := theme.Cur()
	label := "◈ " + h.sideLabel()
	if h.focus == homeFocusSide {
		label = th.Accent.Render(label)
	} else {
		label = th.Dim.Render(label)
	}
	h.side.SetSize(width, max(height-layout.CountLines(menuBox)-3, 1))
	return menuBox + "\n\n" + label + "\n\n" + h.side.View()
}
