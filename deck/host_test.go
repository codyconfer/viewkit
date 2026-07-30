package deck

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/codyconfer/viewkit/theme"
)

type stubView struct{ title string }

func (s stubView) Title() string                  { return s.title }
func (s stubView) Init() tea.Cmd                  { return nil }
func (s stubView) Update(*Model, tea.Msg) tea.Cmd { return nil }
func (s stubView) Body(int, int) string           { return "body" }
func (s stubView) Hints() [][2]string             { return nil }
func (s stubView) Context() [][2]string           { return nil }

type stubComp struct{}

func (stubComp) Render(int, int) string { return "c" }

func TestRegisterView(t *testing.T) {
	RegisterView("test.stub", func() View { return stubView{title: "Stub"} })
	v, ok := LookupView("test.stub")
	if !ok || v.Title() != "Stub" {
		t.Fatalf("lookup = %v ok=%v", v, ok)
	}
}

func TestRegisterComponent(t *testing.T) {
	RegisterComponent("test.comp", func() Component { return stubComp{} })
	c, ok := LookupComponent("test.comp")
	if !ok || c.Render(1, 1) != "c" {
		t.Fatal("component lookup")
	}
}

func TestHostPushPop(t *testing.T) {
	h := New(stubView{title: "Root"})
	_ = h.Push(stubView{title: "Child"})
	if h.top().Title() != "Child" {
		t.Fatal(h.top().Title())
	}
	cmd := h.Pop()
	if cmd == nil || h.top().Title() != "Root" {
		t.Fatalf("after pop title=%s", h.top().Title())
	}
}

func TestHostKeyHook(t *testing.T) {
	var seen string
	h := New(stubView{title: "Root"}, WithKeyHook(func(m *Model, key tea.KeyMsg) (tea.Cmd, bool) {
		seen = key.String()
		return m.Push(stubView{title: "Hot"}), true
	}))
	m, cmd := h.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}, Alt: true})
	h = m.(*Model)
	if seen != "alt+n" {
		t.Fatalf("hook key = %q", seen)
	}
	if cmd != nil {
		_ = cmd()
	}
	if h.top().Title() != "Hot" {
		t.Fatalf("after hook title=%s", h.top().Title())
	}
}

type settleTestMsg struct{ n int }

func TestHostMsgHook(t *testing.T) {
	var got int
	h := New(stubView{title: "Root"}, WithMsgHook(func(m *Model, msg tea.Msg) (tea.Cmd, bool) {
		s, ok := msg.(settleTestMsg)
		if !ok {
			return nil, false
		}
		got = s.n
		return m.Push(stubView{title: "Settled"}), true
	}))
	m, cmd := h.Update(settleTestMsg{n: 7})
	h = m.(*Model)
	if got != 7 {
		t.Fatalf("msg hook = %d", got)
	}
	if cmd != nil {
		_ = cmd()
	}
	if h.top().Title() != "Settled" {
		t.Fatalf("after msg hook title=%s", h.top().Title())
	}
}

func TestHostInitCmd(t *testing.T) {
	h := New(stubView{title: "Root"},
		WithInitCmd(func() tea.Msg { return settleTestMsg{n: 1} }),
		WithInitCmd(func() tea.Msg { return settleTestMsg{n: 2} }),
		WithInitCmd(nil),
	)
	if got := len(h.initCmds); got != 2 {
		t.Fatalf("initCmds = %d, want 2 (nil dropped)", got)
	}
	if h.Init() == nil {
		t.Fatal("Init() returned no command")
	}
}

func TestHostInitCmdSurvivesPush(t *testing.T) {
	var seen []int
	h := New(stubView{title: "Root"},
		WithInitCmd(func() tea.Msg { return settleTestMsg{n: 1} }),
		WithMsgHook(func(m *Model, msg tea.Msg) (tea.Cmd, bool) {
			s, ok := msg.(settleTestMsg)
			if !ok {
				return nil, false
			}
			seen = append(seen, s.n)
			return nil, true
		}),
	)
	_ = h.Push(stubView{title: "Child"})
	if _, cmd := h.Update(settleTestMsg{n: 1}); cmd != nil {
		_ = cmd()
	}
	if _, cmd := h.Update(settleTestMsg{n: 2}); cmd != nil {
		_ = cmd()
	}
	if len(seen) != 2 || seen[0] != 1 || seen[1] != 2 {
		t.Fatalf("hook saw %v while a child view was on top, want [1 2]", seen)
	}
}

type ctxView struct {
	stubView
	ctx [][2]string
}

func (c ctxView) Context() [][2]string { return c.ctx }

func TestHostHeaderStripRowsAlign(t *testing.T) {
	const width = 100
	h := New(ctxView{
		stubView: stubView{title: "main"},
		ctx:      [][2]string{{"role", "triage"}},
	}, WithChrome(Chrome{Brand: "MUNIN", BrandGlyph: "▚▚", Subtitle: "deck"}))
	m, _ := h.Update(tea.WindowSizeMsg{Width: width, Height: 40})
	h = m.(*Model)

	view := h.View()
	var stripRows []string
	for _, ln := range strings.Split(view, "\n") {
		plain := ansi.Strip(ln)
		if strings.Contains(plain, "MUNIN") || (strings.Contains(plain, "main") && strings.Contains(plain, "role")) {
			stripRows = append(stripRows, plain)
		}
	}
	if len(stripRows) < 2 {
		t.Fatalf("expected brand + breadcrumb strip rows, got %d\n%s", len(stripRows), view)
	}
	brand, crumb := stripRows[0], stripRows[1]
	if ansi.StringWidth(brand) != width || ansi.StringWidth(crumb) != width {
		t.Fatalf("strip rows must fill width %d: brand=%d crumb=%d", width, ansi.StringWidth(brand), ansi.StringWidth(crumb))
	}
	brandLead := len(brand) - len(strings.TrimLeft(brand, " "))
	crumbLead := len(crumb) - len(strings.TrimLeft(crumb, " "))
	if brandLead != crumbLead {
		t.Fatalf("header row left inset mismatch: brand=%d crumb=%d\n%q\n%q", brandLead, crumbLead, brand, crumb)
	}
	brandTrail := len(brand) - len(strings.TrimRight(brand, " "))
	crumbTrail := len(crumb) - len(strings.TrimRight(crumb, " "))
	if brandTrail != crumbTrail {
		t.Fatalf("header row right inset mismatch: brand=%d crumb=%d\n%q\n%q", brandTrail, crumbTrail, brand, crumb)
	}
}

func TestHostBrandWithoutGlyphNoExtraPad(t *testing.T) {
	h := New(stubView{title: "root"}, WithChrome(Chrome{Brand: "MUNIN", Subtitle: "ntr"}))
	m, _ := h.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	h = m.(*Model)
	plain := ansi.Strip(h.View())
	wantLead := theme.AppMarginX + 1
	for _, ln := range strings.Split(plain, "\n") {
		if !strings.Contains(ln, "MUNIN") {
			continue
		}
		lead := len(ln) - len(strings.TrimLeft(ln, " "))
		if lead != wantLead {
			t.Fatalf("brand lead inset = %d, want %d (no BrandGlyph pad): %q", lead, wantLead, ln)
		}
		return
	}
	t.Fatal("brand line not found")
}

func TestHostRefreshStatusReloadsStrip(t *testing.T) {
	calls := 0
	marker := "REFRESH-CHIP"
	h := New(stubView{title: "root"}, WithStatus(func(context.Context) StatusInfo {
		calls++
		name := "first"
		if calls > 1 {
			name = marker
		}
		return StatusInfo{Services: []ServiceStatus{{Name: name, Glyph: "·"}}}
	}))
	m, _ := h.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	h = m.(*Model)

	cmd := h.RefreshStatus()
	if cmd == nil {
		t.Fatal("RefreshStatus returned nil with StatusFunc installed")
	}
	msg := cmd()
	m, _ = h.Update(msg)
	h = m.(*Model)
	if !strings.Contains(ansi.Strip(h.View()), "first") {
		t.Fatalf("expected first status load in chrome:\n%s", h.View())
	}

	cmd = h.RefreshStatus()
	msg = cmd()
	m, _ = h.Update(msg)
	h = m.(*Model)
	if !strings.Contains(ansi.Strip(h.View()), marker) {
		t.Fatalf("expected refreshed status chip %q:\n%s", marker, h.View())
	}
	if calls != 2 {
		t.Fatalf("StatusFunc calls = %d, want 2", calls)
	}

	bare := New(stubView{title: "root"})
	if bare.RefreshStatus() != nil {
		t.Fatal("RefreshStatus without WithStatus should be nil")
	}
}

func TestLoadLandingOffTopReachesItsView(t *testing.T) {
	shell := homeShellWithRows(2)
	h := New(shell)
	h = driveHost(h, tea.WindowSizeMsg{Width: 80, Height: 24})

	slow := shell.Init()
	h = driveHost(h, tea.WindowSizeMsg{Width: 80, Height: 24})
	if cmd := h.Push(stubView{title: "detail"}); cmd != nil {
		cmd()
	}

	h = driveHost(h, slow())
	if cmd := h.Pop(); cmd != nil {
		h = driveHost(h, cmd())
	}

	body := ansi.Strip(h.View())
	if strings.Contains(body, "loading") {
		t.Fatalf("side pane stuck on loading after the fetch landed off-top:\n%s", body)
	}
	if !strings.Contains(body, "row-000") {
		t.Fatalf("side pane missing loaded rows:\n%s", body)
	}
}

func TestLoadDeliveryDoesNotCrossViews(t *testing.T) {
	first, second := homeShellWithRows(1), homeShellWithRows(3)
	h := New(first)
	h = driveHost(h, tea.WindowSizeMsg{Width: 80, Height: 24})
	if cmd := h.Push(second); cmd != nil {
		cmd()
	}
	h = driveHost(h, tea.WindowSizeMsg{Width: 80, Height: 24})

	driveHost(h, first.Init()())
	if second.loaded {
		t.Fatal("first view's load marked the second view loaded")
	}
	if !first.loaded {
		t.Fatal("first view's load did not reach it")
	}
}
