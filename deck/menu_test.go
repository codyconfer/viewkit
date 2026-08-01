package deck

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"

	"github.com/codyconfer/viewkit/glyph"
	"github.com/codyconfer/viewkit/theme"
)

func rowIndex(lines []string, want string) int {
	for i, ln := range lines {
		if strings.Contains(ansi.Strip(ln), want) {
			return i
		}
	}
	return -1
}

func menuInEveryProfile(t *testing.T, body func(t *testing.T)) {
	t.Helper()
	prev := lipgloss.ColorProfile()
	t.Cleanup(func() { lipgloss.SetColorProfile(prev) })
	for _, prof := range []struct {
		name string
		p    termenv.Profile
	}{
		{"Ascii", termenv.Ascii},
		{"TrueColor", termenv.TrueColor},
	} {
		t.Run(prof.name, func(t *testing.T) {
			lipgloss.SetColorProfile(prof.p)
			body(t)
		})
	}
}

func TestMenuStylesOnlyTheCursorRow(t *testing.T) {
	prev := lipgloss.ColorProfile()
	t.Cleanup(func() { lipgloss.SetColorProfile(prev) })
	m := NewMenu("queries", nil, MenuItem{Label: "first"}, MenuItem{Label: "second"})

	lipgloss.SetColorProfile(termenv.TrueColor)
	th := theme.Cur()
	keySGR, valSGR := sgrPrefix(t, th.Key), sgrPrefix(t, th.Val)
	if keySGR == valSGR {
		t.Fatalf("theme cannot distinguish the cursor row: key and value both render %q", keySGR)
	}

	lines := strings.Split(m.Body(60, 30), "\n")
	cursorAt, otherAt := rowIndex(lines, "first"), rowIndex(lines, "second")
	if cursorAt < 0 || otherAt < 0 {
		t.Fatalf("rows not found by plain text:\n%q", lines)
	}
	if !strings.Contains(lines[cursorAt], keySGR) {
		t.Errorf("cursor row lacks the key emphasis %q:\n%q", keySGR, lines[cursorAt])
	}
	if strings.Contains(lines[otherAt], keySGR) {
		t.Errorf("a row away from the cursor got the cursor emphasis:\n%q", lines[otherAt])
	}
	if !strings.Contains(lines[otherAt], valSGR) {
		t.Errorf("non-cursor row lacks the value style %q:\n%q", valSGR, lines[otherAt])
	}
	if got := ansi.Strip(lines[cursorAt]); !strings.Contains(got, "▸ first") {
		t.Errorf("cursor row plain text = %q, want it to contain %q", got, "▸ first")
	}
	if raw := lines[cursorAt]; strings.Contains(raw, "▸ first") {
		t.Errorf("glyph and label are styled as one segment, so rowIndex need not strip: %q", raw)
	}

	lipgloss.SetColorProfile(termenv.Ascii)
	if plain := m.Body(60, 30); strings.Contains(plain, "\x1b[") {
		t.Errorf("the escape-free profile still emitted SGR:\n%q", plain)
	}
}

func TestMenuRowsStayTightWhereverTheCursorSits(t *testing.T) {
	menuInEveryProfile(t, func(t *testing.T) {
		m := NewMenu("queries", nil,
			MenuItem{Label: "first"},
			MenuItem{Label: "second"},
			MenuItem{Label: "third"},
		)
		for cursor := range m.items {
			m.cursor = cursor
			lines := strings.Split(m.Body(60, 30), "\n")

			firstAt := rowIndex(lines, "first")
			secondAt := rowIndex(lines, "second")
			thirdAt := rowIndex(lines, "third")
			if firstAt < 0 || secondAt < 0 || thirdAt < 0 {
				t.Fatalf("items missing:\n%s", strings.Join(lines, "\n"))
			}
			if secondAt-firstAt != 1 || thirdAt-secondAt != 1 {
				t.Errorf("cursor %d shifted the rows:\n%s", cursor, strings.Join(lines, "\n"))
			}
		}
	})
}

func TestMenuHeightDoesNotChangeRowSpacing(t *testing.T) {
	menuInEveryProfile(t, func(t *testing.T) {
		items := make([]MenuItem, 12)
		for i := range items {
			items[i] = MenuItem{Label: string(rune('a' + i))}
		}
		m := NewMenu("queries", nil, items...)

		tight := m.Body(60, len(items)+menuBoxChrome)
		if got := strings.Count(tight, "\n") + 1; got != len(items)+menuBoxChrome {
			t.Errorf("menu rendered %d lines, want %d:\n%s", got, len(items)+menuBoxChrome, tight)
		}
		if roomy := m.Body(60, 40); roomy != tight {
			t.Errorf("a tall terminal changed the menu body:\n%s", roomy)
		}
	})
}

func labelColumn(t *testing.T, lines []string, label string) int {
	t.Helper()
	i := rowIndex(lines, label)
	if i < 0 {
		t.Fatalf("row %q missing:\n%s", label, strings.Join(lines, "\n"))
	}
	plain := ansi.Strip(lines[i])
	return lipgloss.Width(plain[:strings.Index(plain, label)])
}

func TestMenuAlignsLabelsWhenOnlySomeItemsHaveIcons(t *testing.T) {
	prev := glyph.CurrentMode()
	t.Cleanup(func() { glyph.SetMode(prev) })

	for _, mode := range []glyph.Mode{glyph.ModeNerd, glyph.ModeUnicode, glyph.ModeNone} {
		glyph.SetMode(mode)
		menuInEveryProfile(t, func(t *testing.T) {
			m := NewMenu("queries", nil,
				MenuItem{Label: "withicon", Icon: glyph.Bullet()},
				MenuItem{Label: "noicon"},
				MenuItem{Label: "widericon", Icon: "gh"},
			)
			lines := strings.Split(m.Body(60, 30), "\n")
			with := labelColumn(t, lines, "withicon")
			without := labelColumn(t, lines, "noicon")
			wider := labelColumn(t, lines, "widericon")
			if with != without || with != wider {
				t.Errorf("mode %v: labels start at columns %d/%d/%d, want all equal:\n%s",
					mode, with, without, wider, strings.Join(lines, "\n"))
			}
		})
	}
}

func TestMenuWithoutAnyIconsKeepsLabelsFlush(t *testing.T) {
	prev := glyph.CurrentMode()
	t.Cleanup(func() { glyph.SetMode(prev) })
	glyph.SetMode(glyph.ModeNone)

	menuInEveryProfile(t, func(t *testing.T) {
		bare := NewMenu("queries", nil, MenuItem{Label: "alpha"}, MenuItem{Label: "beta"})
		iconed := NewMenu("queries", nil, MenuItem{Label: "alpha", Icon: glyph.Bullet()}, MenuItem{Label: "beta"})

		bareAt := labelColumn(t, strings.Split(bare.Body(60, 30), "\n"), "alpha")
		iconedAt := labelColumn(t, strings.Split(iconed.Body(60, 30), "\n"), "alpha")
		if bareAt+glyph.LeadWidth != iconedAt {
			t.Errorf("icon-free labels at column %d and iconed at %d, want a %d-column difference",
				bareAt, iconedAt, glyph.LeadWidth)
		}
	})
}
