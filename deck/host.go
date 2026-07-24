package deck

import (
	"context"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/theme"
)

const statusRefreshInterval = 60 * time.Second

type tickMsg time.Time
type statusMsg struct{ info StatusInfo }
type statusRefreshMsg struct{}

// Option configures Model (Host alias).
type Option func(*Model)

// WithStatus installs an async status loader for the footer strip.
func WithStatus(fn StatusFunc) Option {
	return func(h *Host) { h.statusFn = fn }
}

// WithChrome sets brand chrome.
func WithChrome(c Chrome) Option {
	return func(h *Host) { h.chrome = c }
}

// WithQuitCheck overrides quit key matcher.
func WithQuitCheck(fn func(string) bool) Option {
	return func(h *Host) { h.quitCheck = fn }
}

// WithKeyMapQuit installs a quit matcher from keys.Cur() Quit binding.
func WithKeyMapQuit() Option {
	return func(h *Host) {
		h.quitCheck = func(k string) bool {
			for _, q := range keys.Cur().Binding(keys.Quit).Keys {
				if k == q {
					return true
				}
			}
			return false
		}
	}
}

// Host is the tea model: stack navigation + chrome.
// Prefer the Model alias in new code; Host remains for compatibility.
type Host struct {
	stack  []View
	width  int
	height int
	clock  string

	chrome    Chrome
	statusFn  StatusFunc
	status    StatusInfo
	hasStatus bool
	quitCheck func(string) bool
}

// New builds a Model with root view.
func New(root View, opts ...Option) *Model {
	h := &Model{
		stack: []View{root},
		clock: time.Now().Format("15:04:05"),
		chrome: Chrome{
			Brand:    "APP",
			Subtitle: "deck",
		},
		quitCheck: func(k string) bool { return k == "ctrl+c" },
	}
	for _, o := range opts {
		o(h)
	}
	return h
}

// Run starts the tea program with alt screen.
func Run(root View, opts ...Option) error {
	_, err := tea.NewProgram(New(root, opts...), tea.WithAltScreen()).Run()
	return err
}

func (h *Host) top() View { return h.stack[len(h.stack)-1] }

// Top returns the current view.
func (h *Host) Top() View { return h.top() }

// Width returns the current terminal width.
func (h *Host) Width() int { return h.width }

// SetStatus applies chrome status immediately (tests / non-async hosts).
func (h *Host) SetStatus(info StatusInfo) {
	h.status, h.hasStatus = info, true
}

// Height returns the current terminal height.
func (h *Host) Height() int { return h.height }

// Push navigates to v.
func (h *Host) Push(v View) tea.Cmd {
	h.stack = append(h.stack, v)
	return tea.Batch(v.Init(), h.resizeCmd())
}

// Pop leaves the current view (quits on root).
func (h *Host) Pop() tea.Cmd {
	if len(h.stack) <= 1 {
		return tea.Quit
	}
	h.stack = h.stack[:len(h.stack)-1]
	return h.resizeCmd()
}

func (h *Host) resizeCmd() tea.Cmd {
	return func() tea.Msg { return tea.WindowSizeMsg{Width: h.width, Height: h.height} }
}

func (h *Host) Init() tea.Cmd {
	cmds := []tea.Cmd{h.tick(), h.top().Init()}
	if h.statusFn != nil {
		cmds = append(cmds, h.fetchStatus())
	}
	return tea.Batch(cmds...)
}

func (h *Host) tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (h *Host) fetchStatus() tea.Cmd {
	fn := h.statusFn
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		defer cancel()
		return statusMsg{info: fn(ctx)}
	}
}

func (h *Host) scheduleStatusRefresh() tea.Cmd {
	return tea.Tick(statusRefreshInterval, func(time.Time) tea.Msg { return statusRefreshMsg{} })
}

func (h *Host) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		h.width, h.height = m.Width, m.Height
		return h, h.top().Update(h, msg)
	case tickMsg:
		h.clock = time.Time(m).Format("15:04:05")
		return h, h.tick()
	case statusMsg:
		h.status, h.hasStatus = m.info, true
		return h, h.scheduleStatusRefresh()
	case statusRefreshMsg:
		return h, h.fetchStatus()
	case tea.KeyMsg:
		if h.quitCheck != nil && h.quitCheck(m.String()) {
			return h, tea.Quit
		}
		return h, h.top().Update(h, msg)
	default:
		return h, h.top().Update(h, msg)
	}
}

func (h *Host) View() string {
	if h.width == 0 {
		return "initializing deck…"
	}
	if !layout.FitsScreenWidth(h.width) {
		return theme.AppMargin(layout.TooNarrow(h.width))
	}
	v := h.top()
	header := h.header(v)
	footer := h.footer(v)
	bodyHeight := max(h.height-layout.CountLines(header)-layout.CountLines(footer)-2, 1)
	body := layout.FillHeight(v.Body(h.width, bodyHeight), bodyHeight)
	return theme.AppMargin(layout.Stack(header, body, footer))
}

func (h *Host) header(v View) string {
	f := layout.ScreenFrame(h.width)
	full := f.BodyWidth() + 4
	th := theme.Cur()
	muted := th.Dim.GetForeground()
	label := h.chrome.Brand
	if h.chrome.BrandGlyph != "" {
		label = h.chrome.BrandGlyph + " " + h.chrome.Brand
	}
	brand := st(muted, " ") + sb(th.Accent.GetForeground(), label)
	if h.chrome.Subtitle != "" {
		brand += st(muted, " · "+h.chrome.Subtitle)
	}
	clockGlyph := h.chrome.ClockGlyph
	if clockGlyph != "" {
		clockGlyph += " "
	}
	clock := sb(th.Accent.GetForeground(), clockGlyph+h.clock)
	right := clock
	if h.hasStatus && h.status.Identity != "" {
		right = h.status.Identity + st(muted, "   ") + clock
	}
	return theme.StripBlock(full,
		layout.SpreadBG(theme.StripBg(), brand, right+st(muted, " "), full),
		layout.SpreadBG(theme.StripBg(), h.breadcrumbs(), h.contextCues(v), full),
	)
}

func (h *Host) breadcrumbs() string {
	th := theme.Cur()
	muted := th.Dim.GetForeground()
	sep := st(muted, " ⟩ ")
	parts := make([]string, len(h.stack))
	for i, v := range h.stack {
		if i == len(h.stack)-1 {
			parts[i] = sb(th.Accent.GetForeground(), v.Title())
		} else {
			parts[i] = st(muted, v.Title())
		}
	}
	return st(muted, " ") + strings.Join(parts, sep)
}

func (h *Host) contextCues(v View) string {
	th := theme.Cur()
	muted := th.Dim.GetForeground()
	var parts []string
	for _, c := range v.Context() {
		if c[1] == "" {
			continue
		}
		parts = append(parts, st(muted, c[0]+": ")+st(th.Val.GetForeground(), c[1]))
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, st(muted, " · ")) + st(muted, " ")
}

func (h *Host) footer(v View) string {
	f := layout.ScreenFrame(h.width)
	full := f.BodyWidth() + 4
	hints := append([][2]string{}, v.Hints()...)
	hints = append(hints, [2]string{"esc", "back"}, [2]string{"ctrl+c", "quit"})
	legend := layout.IndentLines(f.HintLine(hints...), 1)
	bar := theme.StripBlock(full, layout.SpreadBG(theme.StripBg(), h.statusSegments(), "", full))
	return layout.Stack(bar, legend)
}

func (h *Host) statusSegments() string {
	if !h.hasStatus || len(h.status.Services) == 0 {
		return ""
	}
	th := theme.Cur()
	sep := st(th.Dim.GetForeground(), " · ")
	parts := make([]string, 0, len(h.status.Services))
	for _, s := range h.status.Services {
		label := s.Name
		if s.Detail != "" {
			label += " " + s.Detail
		}
		g := s.Glyph
		if g != "" {
			g += " "
		}
		color := s.Color
		if color == nil {
			color = th.Val.GetForeground()
		}
		parts = append(parts, st(color, g)+st(th.Val.GetForeground(), label))
	}
	return st(th.Dim.GetForeground(), " ") + strings.Join(parts, sep)
}

func st(fg lipgloss.TerminalColor, s string) string { return theme.StripText(fg, s) }
func sb(fg lipgloss.TerminalColor, s string) string { return theme.StripBold(fg, s) }
