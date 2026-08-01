package theme

import (
	"sync/atomic"

	"github.com/charmbracelet/lipgloss"
)

// Layout metrics shared by every viewkit screen. BodyWidth is the content
// column that panels and cards wrap (they add their own border and padding);
// the Min*/Tall* values are the terminal sizes screens are designed against,
// and RuleWidth is the horizontal-rule span covering body plus panel chrome.
const (
	BodyWidth    = 81
	MinBodyWidth = 24

	MinScreenWidth     = 80
	MinBodyHeight      = 35
	TallBodyHeight     = 46
	AppMarginY         = 1
	AppMarginX         = 2
	ListItemGapY       = 1
	ScreenPaddingWidth = AppMarginX*2 + 4
	MinScreenBodyWidth = MinScreenWidth - ScreenPaddingWidth

	RuleWidth = BodyWidth + 4
)

// Palette is the twelve-color input to New: semantic roles (Accent, Success,
// Failure, ...) plus extra chart series colors and the screen background.
// Registering a Palette is all a custom theme needs; New derives every style.
type Palette struct {
	Accent   lipgloss.Color
	Border   lipgloss.Color
	Muted    lipgloss.Color
	Text     lipgloss.Color
	Selected lipgloss.Color
	Success  lipgloss.Color
	Warning  lipgloss.Color
	Failure  lipgloss.Color
	Info     lipgloss.Color
	Series2  lipgloss.Color
	Series3  lipgloss.Color
	Bg       lipgloss.Color
}

// New derives a full Theme from p: text styles, panel/card borders,
// notification frames, and a seven-entry Series ramp, with the too-narrow
// messages set to their defaults.
func New(p Palette) Theme {
	return Theme{
		Title:  lipgloss.NewStyle().Bold(true).Foreground(p.Accent),
		Accent: lipgloss.NewStyle().Bold(true).Foreground(p.Accent),
		Dim:    lipgloss.NewStyle().Foreground(p.Muted),
		Val:    lipgloss.NewStyle().Foreground(p.Text),
		Key:    lipgloss.NewStyle().Bold(true).Foreground(p.Accent),
		Can:    lipgloss.NewStyle().Foreground(p.Success),
		Cant:   lipgloss.NewStyle().Foreground(p.Failure),

		Panel:      lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.Border).Padding(0, 1).Width(BodyWidth + 2),
		PanelFocus: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.Selected).Padding(0, 1).Width(BodyWidth + 2),
		PanelTitle: lipgloss.NewStyle().Bold(true).Foreground(p.Accent),
		Card:       lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.Accent).Padding(0, 1).Width(BodyWidth + 2).Align(lipgloss.Center),
		AppFrame:   lipgloss.NewStyle().Margin(AppMarginY, AppMarginX),

		NotifPositive: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.Success).Foreground(p.Success).Padding(0, 1),
		NotifNeutral:  lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.Info).Foreground(p.Info).Padding(0, 1),
		NotifWarning:  lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.Warning).Foreground(p.Warning).Padding(0, 1),
		NotifNegative: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.Failure).Foreground(p.Failure).Padding(0, 1),
		NotifIdle:     lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.Muted).Foreground(p.Muted).Padding(0, 1),
		NotifTitle:    lipgloss.NewStyle().Bold(true),

		Series: []lipgloss.Style{
			lipgloss.NewStyle().Foreground(p.Accent),
			lipgloss.NewStyle().Foreground(p.Success),
			lipgloss.NewStyle().Foreground(p.Warning),
			lipgloss.NewStyle().Foreground(p.Failure),
			lipgloss.NewStyle().Foreground(p.Series2),
			lipgloss.NewStyle().Foreground(p.Muted),
			lipgloss.NewStyle().Foreground(p.Series3),
		},

		Bg: p.Bg,

		TooNarrowTitle: DefaultTooNarrowTitle,
		TooNarrowNeed:  DefaultTooNarrowNeed,
		TooNarrowBody:  DefaultTooNarrowBody,
	}
}

// Default text for the too-narrow-terminal screen. TooNarrowNeed takes the
// required column count; TooNarrowBody takes the current width (as a string)
// and the required column count.
const (
	DefaultTooNarrowTitle = "TERMINAL TOO NARROW"
	DefaultTooNarrowNeed  = "Need at least %d columns."
	DefaultTooNarrowBody  = "Current width: %s columns. Resize the terminal to at least %d characters wide to use this screen."
)

// Theme is a bundle of ready-to-use lipgloss styles: text roles, panel and
// notification frames, chart Series styles, the screen background color, and
// the too-narrow-screen message templates. Build one with New and install it
// with Use; read the active theme with Cur.
type Theme struct {
	Title  lipgloss.Style
	Accent lipgloss.Style
	Dim    lipgloss.Style
	Val    lipgloss.Style
	Key    lipgloss.Style
	Can    lipgloss.Style
	Cant   lipgloss.Style

	Panel      lipgloss.Style
	PanelFocus lipgloss.Style
	PanelTitle lipgloss.Style
	Card       lipgloss.Style
	AppFrame   lipgloss.Style

	NotifPositive lipgloss.Style
	NotifNeutral  lipgloss.Style
	NotifWarning  lipgloss.Style
	NotifNegative lipgloss.Style
	NotifIdle     lipgloss.Style
	NotifTitle    lipgloss.Style

	Series []lipgloss.Style

	Bg lipgloss.Color

	TooNarrowTitle string
	TooNarrowNeed  string
	TooNarrowBody  string

	stripBg   lipgloss.TerminalColor
	stripText lipgloss.Style
	stripBold lipgloss.Style
}

// Default returns the theme built from the built-in default palette.
func Default() Theme { return New(defaultPalette) }

var current = func() *atomic.Pointer[Theme] {
	p := new(atomic.Pointer[Theme])
	t := withCachedStyles(Default())
	p.Store(&t)
	return p
}()

var generation atomic.Uint64

// Cur returns the active theme by value (matching keys.Cur).
func Cur() Theme { return *current.Load() }

// Generation returns a counter incremented by every Use. Views that cache
// theme-derived render state compare generations instead of theme identity
// to detect a theme change.
func Generation() uint64 { return generation.Load() }

// Use makes t the active theme. Safe to call while other goroutines render.
func Use(t Theme) {
	t = withCachedStyles(t)
	current.Store(&t)
	generation.Add(1)
}

func withCachedStyles(t Theme) Theme {
	t.stripBg = stripBgOf(t)
	t.stripText = lipgloss.NewStyle().Background(t.stripBg)
	t.stripBold = t.stripText.Bold(true)
	return t
}
