package deck

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/codyconfer/viewkit/theme"
)

type stubView struct{ title string }

func (s stubView) Title() string                 { return s.title }
func (s stubView) Init() tea.Cmd                 { return nil }
func (s stubView) Update(*Host, tea.Msg) tea.Cmd { return nil }
func (s stubView) Body(int, int) string          { return "body" }
func (s stubView) Hints() [][2]string            { return nil }
func (s stubView) Context() [][2]string          { return nil }

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
	h = m.(*Host)

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
	h = m.(*Host)
	plain := ansi.Strip(h.View())
	// AppMarginX + one strip pad space, then brand — not an extra blank column.
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
