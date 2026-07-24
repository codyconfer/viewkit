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

	// SideLabel is shown above the side list when SideFetch is set.
	SideLabel string
	// SideHint is the tab-target label when menu-focused (default "side").
	SideHint string
	// SideFetch loads opaque side content once (nil disables the side pane).
	SideFetch func() any
	// SideBind maps fetched data → rows using current width.
	SideBind func(width int, fetched any) []list.Item
	// SideLoading shown while SideFetch has not completed.
	SideLoading string

	// IsAction maps a key string to a keys.Action (defaults to keys.Cur menu map).
	IsAction func(string) (keys.Action, bool)
	// OnOpen overrides browser.Open when a side item Key is confirmed.
	OnOpen func(url string) error

	cursor  int
	focus   int
	side    list.Model
	width   int
	loaded  bool
	fetched any
}

// NewHomeShell builds a HomeShell. sideLabel empty + SideFetch nil → menu-only.
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

type homeShellLoadedMsg struct{ data any }

func (h *HomeShell) Title() string        { return h.title }
func (h *HomeShell) Context() [][2]string { return h.ctx }

// FocusSide reports whether the side pane has keyboard focus (for tests/adapters).
func (h *HomeShell) FocusSide() bool { return h.focus == homeFocusSide }

func (h *HomeShell) Hints() [][2]string {
	if h.focus == homeFocusSide {
		return [][2]string{{"↑/↓", "move"}, {"enter", "open"}, {"pgup/pgdn", "page"}, {"tab", "menu"}}
	}
	hints := [][2]string{{"↑/↓", "move"}, {"enter", "open"}}
	if h.hasSide() {
		hint := h.SideHint
		if hint == "" {
			hint = "side"
		}
		hints = append(hints, [2]string{"tab", hint})
	}
	return hints
}

func (h *HomeShell) hasSide() bool { return h.SideLabel != "" && h.SideFetch != nil }

func (h *HomeShell) Init() tea.Cmd {
	if !h.hasSide() {
		return nil
	}
	return func() tea.Msg { return homeShellLoadedMsg{data: h.SideFetch()} }
}

func (h *HomeShell) Update(host *Model, msg tea.Msg) tea.Cmd {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		h.width = m.Width
		h.refresh()
		return nil
	case homeShellLoadedMsg:
		h.fetched, h.loaded = m.data, true
		h.refresh()
		return nil
	case tea.KeyMsg:
		return h.handleKey(host, m)
	}
	return nil
}

func (h *HomeShell) handleKey(host *Model, m tea.KeyMsg) tea.Cmd {
	if h.hasSide() && m.String() == "tab" {
		if h.focus == homeFocusMenu {
			h.focus = homeFocusSide
		} else {
			h.focus = homeFocusMenu
		}
		h.side.SetFocused(h.focus == homeFocusSide)
		return nil
	}
	if h.focus == homeFocusSide {
		switch m.String() {
		case "pgup":
			h.side.Scroll(-1)
			return nil
		case "pgdown":
			h.side.Scroll(1)
			return nil
		}
		act, ok := h.action(m.String())
		if !ok {
			return nil
		}
		switch act {
		case keys.Up:
			h.side.Move(-1)
		case keys.Down:
			h.side.Move(1)
		case keys.Confirm:
			return h.openSelected()
		case keys.Cancel:
			h.focus = homeFocusMenu
			h.side.SetFocused(false)
		}
		return nil
	}
	act, ok := h.action(m.String())
	if !ok {
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

func (h *HomeShell) action(key string) (keys.Action, bool) {
	if h.IsAction != nil {
		return h.IsAction(key)
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

func (h *HomeShell) refresh() {
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
		h.side.SetItems(h.SideBind(h.width, h.fetched))
		return
	}
	if items, ok := h.fetched.([]list.Item); ok {
		h.side.SetItems(items)
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
	menuBox := f.TitledBox("MAIN MENU", h.menuRows(f)...)
	if !h.hasSide() {
		return menuBox
	}
	th := theme.Cur()
	label := "◈ " + h.SideLabel
	if h.focus == homeFocusSide {
		label = th.Accent.Render(label)
	} else {
		label = th.Dim.Render(label)
	}
	h.side.SetSize(width, max(height-layout.CountLines(menuBox)-2, 1))
	return menuBox + "\n\n" + label + "\n" + h.side.View()
}
