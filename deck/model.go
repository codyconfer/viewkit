package deck

import (
	"context"
	"slices"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/codyconfer/viewkit/glyph"
	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/theme"
)

const statusRefreshInterval = 60 * time.Second

type tickMsg time.Time
type statusMsg struct{ info StatusInfo }
type statusRefreshMsg struct{}

// Option configures Model.
type Option func(*Model)

// WithStatus installs an async status loader for the footer strip.
func WithStatus(fn StatusFunc) Option {
	return func(h *Model) { h.statusFn = fn }
}

// WithChrome sets brand chrome.
func WithChrome(c Chrome) Option {
	return func(h *Model) { h.chrome = c }
}

// WithQuitCheck overrides quit key matcher. An opaque matcher cannot be
// rendered, so pair it with WithQuitHint to keep the footer legend truthful.
func WithQuitCheck(fn func(string) bool) Option {
	return func(h *Model) { h.quitCheck = fn }
}

// WithQuitHint overrides the quit glyph in the footer legend. Defaults to the
// active scheme's Quit binding.
func WithQuitHint(glyph string) Option {
	return func(h *Model) { h.quitHint = glyph }
}

// WithoutClock hides the header clock and disables its per-second tick.
func WithoutClock() Option {
	return func(h *Model) { h.noClock = true }
}

// WithKeyMapQuit installs a quit matcher from keys.Cur() Quit binding.
// This is also the default; the option remains for explicit call sites.
func WithKeyMapQuit() Option {
	return func(h *Model) { h.quitCheck = schemeQuit }
}

func schemeQuit(k string) bool {
	return slices.Contains(keys.Cur().Binding(keys.Quit).Keys, k)
}

// KeyHook handles a key before the top view. Return handled=true to skip
// view Update (and optionally a Cmd such as Push).
type KeyHook func(m *Model, key tea.KeyMsg) (cmd tea.Cmd, handled bool)

// WithKeyHook installs a global key interceptor (e.g. app hotkeys).
func WithKeyHook(fn KeyHook) Option {
	return func(h *Model) { h.keyHook = fn }
}

// MsgHook handles a non-key message before the top view. Return handled=true
// to skip view Update (e.g. debounced role-lifecycle settle).
type MsgHook func(m *Model, msg tea.Msg) (cmd tea.Cmd, handled bool)

type ownedMsg interface{ recipient() View }

// WithMsgHook installs a global message interceptor (e.g. debounce settle).
func WithMsgHook(fn MsgHook) Option {
	return func(h *Model) { h.msgHook = fn }
}

// WithInitCmd adds a command to run when the program starts, alongside the root
// view's own Init. Use it to start a watcher that outlives navigation: a tick
// armed here and re-armed from a MsgHook keeps running whatever is on the stack,
// whereas one armed by a view stops as soon as that view is pushed over.
//
// Repeated calls accumulate.
func WithInitCmd(cmd tea.Cmd) Option {
	return func(h *Model) {
		if cmd == nil {
			return
		}
		h.initCmds = append(h.initCmds, cmd)
	}
}

// Model is the tea root: the stateful Bubble Tea program root for deck, owning
// the navigable view stack, injectable chrome, and the async status strip. It
// implements tea.Model (Init / Update / View).
//
// Contract:
//   - One Model owns one tea.Program for a deck session.
//   - Views receive *Model in Update so they can Push/Pop and read size.
//   - Domain state lives in View implementations (or app kits), not in Model.
type Model struct {
	stack  []View
	width  int
	height int
	clock  string

	chrome    Chrome
	statusFn  StatusFunc
	status    StatusInfo
	hasStatus bool
	quitCheck func(string) bool
	quitHint  string
	keyHook   KeyHook
	msgHook   MsgHook
	initCmds  []tea.Cmd
	noClock   bool
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
		quitCheck: schemeQuit,
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

func (h *Model) top() View { return h.stack[len(h.stack)-1] }

// Top returns the current view.
func (h *Model) Top() View { return h.top() }

// Width returns the current terminal width.
func (h *Model) Width() int { return h.width }

// SetStatus applies chrome status immediately (tests / non-async hosts).
func (h *Model) SetStatus(info StatusInfo) {
	h.status, h.hasStatus = info, true
}

// RefreshStatus reloads the status strip via the installed StatusFunc.
// Same path as the periodic refresh ticker; no-op when WithStatus was not set.
func (h *Model) RefreshStatus() tea.Cmd {
	if h.statusFn == nil {
		return nil
	}
	return h.fetchStatus()
}

// Height returns the current terminal height.
func (h *Model) Height() int { return h.height }

// Push navigates to v.
func (h *Model) Push(v View) tea.Cmd {
	h.stack = append(h.stack, v)
	return tea.Batch(v.Init(), h.resizeCmd())
}

// Pop leaves the current view (quits on root).
func (h *Model) Pop() tea.Cmd {
	if len(h.stack) <= 1 {
		return tea.Quit
	}
	h.stack = h.stack[:len(h.stack)-1]
	return h.resizeCmd()
}

func (h *Model) resizeCmd() tea.Cmd {
	return func() tea.Msg { return tea.WindowSizeMsg{Width: h.width, Height: h.height} }
}

// Init implements tea.Model: it starts the clock tick, the root view's Init,
// the status fetch when installed, and any WithInitCmd commands.
func (h *Model) Init() tea.Cmd {
	cmds := []tea.Cmd{h.tick(), h.top().Init()}
	if h.statusFn != nil {
		cmds = append(cmds, h.fetchStatus())
	}
	cmds = append(cmds, h.initCmds...)
	return tea.Batch(cmds...)
}

func (h *Model) tick() tea.Cmd {
	if h.noClock {
		return nil
	}
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (h *Model) fetchStatus() tea.Cmd {
	fn := h.statusFn
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		defer cancel()
		return statusMsg{info: fn(ctx)}
	}
}

func (h *Model) scheduleStatusRefresh() tea.Cmd {
	return tea.Tick(statusRefreshInterval, func(time.Time) tea.Msg { return statusRefreshMsg{} })
}

// Update implements tea.Model. Chrome concerns — resize, clock, status,
// quit keys and installed hooks — are handled here; everything else routes
// to the top view, except an ownedMsg, which goes to its recipient even
// when that view is no longer on top.
func (h *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		if h.keyHook != nil {
			if cmd, handled := h.keyHook(h, m); handled {
				return h, cmd
			}
		}
		return h, h.top().Update(h, msg)
	default:
		if h.msgHook != nil {
			if cmd, handled := h.msgHook(h, msg); handled {
				return h, cmd
			}
		}
		if om, ok := msg.(ownedMsg); ok {
			if v := om.recipient(); v != nil {
				return h, v.Update(h, msg)
			}
		}
		return h, h.top().Update(h, msg)
	}
}

// View implements tea.Model, framing the top view's Body between header and
// footer chrome (or a placeholder until the first resize arrives).
func (h *Model) View() string {
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

func (h *Model) header(v View) string {
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
	right := ""
	if !h.noClock {
		clockGlyph := h.chrome.ClockGlyph
		if clockGlyph != "" {
			clockGlyph += " "
		}
		right = sb(th.Accent.GetForeground(), clockGlyph+h.clock)
	}
	if h.hasStatus && h.status.Identity != "" {
		if right == "" {
			right = h.status.Identity
		} else {
			right = h.status.Identity + st(muted, "   ") + right
		}
	}
	return theme.StripBlock(full,
		layout.SpreadBG(theme.StripBg(), brand, right+st(muted, " "), full),
		layout.SpreadBG(theme.StripBg(), h.breadcrumbs(), h.contextCues(v), full),
	)
}

func (h *Model) breadcrumbs() string {
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

func (h *Model) contextCues(v View) string {
	th := theme.Cur()
	muted := th.Dim.GetForeground()
	var parts []string
	for _, c := range v.Context() {
		if c.Label == "" {
			continue
		}
		parts = append(parts, st(muted, c.Key+": ")+st(th.Val.GetForeground(), c.Label))
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, st(muted, " · ")) + st(muted, " ")
}

func (h *Model) footer(v View) string {
	f := layout.ScreenFrame(h.width)
	full := f.BodyWidth() + 4
	hints := append([]keys.Hint{}, v.Hints()...)
	hints = append(hints,
		navMap().HintLabeled(keys.Cancel, "back"),
		keys.Hint{Key: h.quitGlyph(), Label: "quit"},
	)
	legend := layout.IndentLines(f.HintLine(hints...), 1)
	bar := theme.StripBlock(full, layout.SpreadBG(theme.StripBg(), h.statusSegments(), "", full))
	return layout.Stack(bar, legend)
}

func (h *Model) quitGlyph() string {
	if h.quitHint != "" {
		return h.quitHint
	}
	return keys.Cur().Binding(keys.Quit).DisplayGlyph()
}

func (h *Model) statusSegments() string {
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
		if g == "" {
			g = glyph.Lead(glyph.StatusFor(s.Severity))
		}
		g += " "
		color := s.Color
		if color == nil {
			color = theme.SeverityColor(s.Severity)
		}
		parts = append(parts, st(color, g)+st(th.Val.GetForeground(), label))
	}
	return st(th.Dim.GetForeground(), " ") + strings.Join(parts, sep)
}

func st(fg lipgloss.TerminalColor, s string) string { return theme.StripText(fg, s) }
func sb(fg lipgloss.TerminalColor, s string) string { return theme.StripBold(fg, s) }
