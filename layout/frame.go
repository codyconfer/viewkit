package layout

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/codyconfer/viewkit/glyph"
	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/theme"
	"github.com/codyconfer/viewkit/ui"
)

// Frame is the rendering context handed to every layout helper: the width and
// height (in terminal cells) a block may occupy, and whether the block owns
// focus. A zero Height means "unconstrained"; helpers that need a height fall
// back to natural content height. Frame is a value type — Focus and WithHeight
// return modified copies.
type Frame struct {
	Width   int
	Height  int
	Focused bool
	// UI is the rendering scope (theme, keys, glyphs). Nil falls back to the
	// built-in defaults; derive new frames with WithWidth/Screen/WithUI so the
	// scope is never dropped.
	UI *ui.Scope
}

// WithUI returns a copy of the frame carrying scope.
func (f Frame) WithUI(scope *ui.Scope) Frame {
	f.UI = scope
	return f
}

// WithWidth returns a copy at the given body width (NewFrame clamping rules),
// preserving the scope.
func (f Frame) WithWidth(width int) Frame {
	n := NewFrame(width)
	n.UI = f.UI
	n.Focused = f.Focused
	return n
}

// Screen returns a copy sized as a body frame for a full terminal of the
// frame's width (ScreenFrame rules), preserving the scope.
func (f Frame) Screen() Frame {
	n := ScreenFrame(f.Width)
	n.UI = f.UI
	return n
}

// StrictScope is a test hook — when true, Frame accessors panic on nil UI so
// fallback leaks surface.
var StrictScope bool

func strictScopeCheck() {
	if StrictScope {
		panic("layout: unscoped Frame read (UI is nil) with StrictScope on")
	}
}

// Theme returns the scoped theme, or the built-in default when unscoped.
func (f Frame) Theme() theme.Theme {
	if f.UI != nil {
		return f.UI.Theme
	}
	strictScopeCheck()
	return theme.Default()
}

// Glyphs returns the scoped glyph set, or the default-mode set when unscoped.
func (f Frame) Glyphs() glyph.Set {
	if f.UI != nil {
		return f.UI.Glyphs
	}
	strictScopeCheck()
	return glyph.DefaultSet()
}

// Scheme returns the scoped key scheme, or the built-in default when unscoped.
func (f Frame) Scheme() keys.Scheme {
	if f.UI != nil {
		return f.UI.Keys
	}
	strictScopeCheck()
	return keys.Default()
}

// NewFrame returns a Frame for the given body width. Non-positive widths fall
// back to theme.BodyWidth, and any width below theme.MinBodyWidth is clamped
// *up* to it — so boxes built on NewFrame widths can exceed the space a tiling
// layout allotted; see Frame.CellBox for the clamp-down alternative.
func NewFrame(width int) Frame {
	if width <= 0 {
		width = theme.BodyWidth
	}
	if width < theme.MinBodyWidth {
		width = theme.MinBodyWidth
	}
	return Frame{Width: width}
}

// Focus returns a copy of the frame with Focused set.
func (f Frame) Focus() Frame {
	f.Focused = true
	return f
}

// WithHeight returns a copy of the frame constrained to h rows.
func (f Frame) WithHeight(h int) Frame {
	f.Height = h
	return f
}

// DefaultFrame returns a Frame at the standard body width (theme.BodyWidth).
func DefaultFrame() Frame { return NewFrame(theme.BodyWidth) }

// DocumentFrame returns a Frame at the standard document width
// (theme.BodyWidth) — the frame for scrollback/document output that is not
// sized to a live terminal.
func DocumentFrame() Frame { return NewFrame(theme.BodyWidth) }

// BodyWidth is the frame's width after NewFrame's clamping: non-positive
// widths become theme.BodyWidth and narrow ones are raised to
// theme.MinBodyWidth. Box, Panel, and TitledBox all size their bodies from it.
func (f Frame) BodyWidth() int {
	return NewFrame(f.Width).Width
}

// Spread lays left and right on one line with the gap between them padded so
// the line is exactly width cells (theme.BodyWidth when width <= 0). When both
// sides cannot fit, the right side wins: left is truncated with an ellipsis
// first, and dropped entirely before right is ever shortened.
func Spread(left, right string, width int) string {
	if width <= 0 {
		width = theme.BodyWidth
	}
	leftW, rightW := ansi.StringWidth(left), ansi.StringWidth(right)
	if leftW+rightW+1 > width {
		switch {
		case rightW >= width:
			left, right = "", ansi.Truncate(right, width, "…")
		case width-rightW > 1:
			left = ansi.Truncate(left, width-rightW-1, "…")
		default:
			left = ""
		}
		leftW, rightW = ansi.StringWidth(left), ansi.StringWidth(right)
	}
	line := left + strings.Repeat(" ", max(width-leftW-rightW, 0)) + right
	if ansi.StringWidth(line) > width {
		line = ansi.Truncate(line, width, "")
	}
	return line
}

// Spread is the package-level Spread at the frame's body width.
func (f Frame) Spread(left, right string) string {
	return Spread(left, right, f.BodyWidth())
}

// SpreadBGIn is Spread with the gap painted in the given background color, for
// lines that sit on a filled strip (status bars, headers). The truncation
// ellipsis is styled to match the strip, dimmed with th.
func SpreadBGIn(th theme.Theme, bg lipgloss.TerminalColor, left, right string, width int) string {
	if width <= 0 {
		width = theme.BodyWidth
	}
	ell := lipgloss.NewStyle().Background(bg).Foreground(th.Dim.GetForeground()).Render("…")
	lw, rw := ansi.StringWidth(left), ansi.StringWidth(right)
	if lw+rw+1 > width {
		switch {
		case rw >= width:
			left, right = "", ansi.Truncate(right, width, ell)
		case width-rw > 1:
			left = ansi.Truncate(left, width-rw-1, ell)
		default:
			left = ""
		}
		lw, rw = ansi.StringWidth(left), ansi.StringWidth(right)
	}
	gap := max(width-lw-rw, 0)
	pad := lipgloss.NewStyle().Background(bg).Render(strings.Repeat(" ", gap))
	line := left + pad + right
	if ansi.StringWidth(line) > width {
		line = ansi.Truncate(line, width, "")
	}
	return line
}

// FillHeight pads body with trailing newlines until it spans height rows.
// A body already at or over the height is returned unchanged — nothing is
// trimmed.
func FillHeight(body string, height int) string {
	lines := CountLines(body)
	if lines >= height {
		return body
	}
	return body + strings.Repeat("\n", height-lines)
}

// IndentLines prefixes every line of s (including blank ones) with n spaces.
func IndentLines(s string, n int) string {
	if n <= 0 {
		return s
	}
	pad := strings.Repeat(" ", n)
	return pad + strings.ReplaceAll(s, "\n", "\n"+pad)
}

// Fit truncates a single line to width cells, ANSI-aware, appending an
// ellipsis when it cuts. Unlike Spread it only ever shortens one string;
// widths below 1 return "".
func Fit(s string, width int) string {
	if width < 1 {
		return ""
	}
	return ansi.Truncate(s, width, "…")
}

// Fit truncates s to the frame's body width with an ellipsis.
func (f Frame) Fit(s string) string {
	return Fit(s, f.BodyWidth())
}

// Rule renders a dim horizontal rule spanning the body width plus the 4
// columns of box border/padding, so it lines up with Box and Panel edges.
func (f Frame) Rule() string {
	return f.Theme().Dim.Render(strings.Repeat("─", f.BodyWidth()+4))
}

// Header renders a title line followed by a Rule. Detail parts are appended
// dim after a "·" separator; blank parts are skipped, and the whole line is
// truncated with an ellipsis at the rule's width.
func (f Frame) Header(title string, detail ...string) string {
	var head strings.Builder
	head.WriteString(f.Theme().Title.Render(title))
	for _, part := range detail {
		if strings.TrimSpace(part) == "" {
			continue
		}
		head.WriteString(f.Theme().Dim.Render("   ·   " + part))
	}
	return ansi.Truncate(head.String(), f.BodyWidth()+4, "…") + "\n" + f.Rule()
}

// Stack joins non-empty sections with a blank line between them; empty
// sections vanish rather than leaving double gaps.
func Stack(sections ...string) string {
	out := make([]string, 0, len(sections))
	for _, section := range sections {
		if section != "" {
			out = append(out, section)
		}
	}
	return strings.Join(out, "\n\n")
}

// StackTight joins non-empty sections with single newlines (no blank line
// between them).
func StackTight(sections ...string) string {
	out := make([]string, 0, len(sections))
	for _, section := range sections {
		if section != "" {
			out = append(out, section)
		}
	}
	return strings.Join(out, "\n")
}

// Box renders lines inside the theme panel border at the frame's BodyWidth.
// Because BodyWidth clamps narrow frames *up* to theme.MinBodyWidth, a Box can
// be wider than f.Width; inside a tiling layout use Frame.CellBox instead,
// which clamps down to fit the cell. A focused frame uses the focus border
// style.
func (f Frame) Box(lines ...string) string {
	return f.boxAt(f.BodyWidth(), lines...)
}

func (f Frame) boxAt(inner int, lines ...string) string {
	sty := f.Theme().Panel
	if f.Focused {
		sty = f.Theme().PanelFocus
	}
	return sty.Width(inner + 2).Render(strings.Join(lines, "\n"))
}

// Panel is Box with a styled title as its first line (truncated to the body
// width). It shares Box's clamp-up behavior on narrow frames; the tiling-safe
// variant is Frame.CellBox / Frame.CellPanel.
func (f Frame) Panel(title string, lines ...string) string {
	return f.panelAt(f.BodyWidth(), title, lines...)
}

func (f Frame) panelAt(inner int, title string, lines ...string) string {
	head := f.Theme().PanelTitle.Render(ansi.Truncate(title, inner, "…"))
	return f.boxAt(inner, append([]string{head}, lines...)...)
}

// Row renders a dim label on the left and value on the right, spread across
// the body width.
func (f Frame) Row(label, value string) string {
	return f.Spread(f.Theme().Dim.Render(label), value)
}

// HintLine renders key hints as "key label" pairs separated by dim dots,
// wrapping onto additional lines whenever the next pair would push past the
// body width.
func (f Frame) HintLine(pairs ...keys.Hint) string {
	parts := make([]string, len(pairs))
	for i, p := range pairs {
		parts[i] = f.Theme().Key.Render(p.Key) + f.Theme().Dim.Render(" "+p.Label)
	}
	sep := f.Theme().Dim.Render("   ·   ")
	var lines []string
	var line string
	for _, part := range parts {
		if line == "" {
			line = part
			continue
		}
		next := line + sep + part
		if ansi.StringWidth(next) <= f.BodyWidth() {
			line = next
			continue
		}
		lines = append(lines, line)
		line = part
	}
	if line != "" {
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// Cursor returns the two-cell selection marker: "▸ " when selected, otherwise
// two spaces so unselected rows keep the same indent.
func (f Frame) Cursor(selected bool) string {
	if selected {
		return f.Theme().Title.Render("▸ ")
	}
	return "  "
}

// Selectable renders a list row: Cursor marker plus the label, accent-styled
// when selected and truncated to leave room for the marker.
func (f Frame) Selectable(label string, selected bool) string {
	sty := f.Theme().Val
	if selected {
		sty = f.Theme().Accent
	}
	return f.Cursor(selected) + sty.Render(ansi.Truncate(label, f.BodyWidth()-2, "…"))
}
