package theme

import (
	"sync/atomic"

	"github.com/charmbracelet/lipgloss"
)

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

// These package-level styles are a snapshot of the default theme taken during
// package initialization. They are written once, before any goroutine can
// observe them, and are never written again — so they are safe to read
// concurrently but they do not follow Use. Cur() is the sanctioned path to the
// active theme; prefer Cur().Dim over DimSty in anything that renders after a
// theme has been applied.
var (
	TitleSty  lipgloss.Style
	AccentSty lipgloss.Style
	DimSty    lipgloss.Style
	ValSty    lipgloss.Style
	KeySty    lipgloss.Style
	CanSty    lipgloss.Style
	CantSty   lipgloss.Style

	AppFrame lipgloss.Style

	PanelSty      lipgloss.Style
	PanelFocusSty lipgloss.Style
	PanelTitleSty lipgloss.Style
	CardSty       lipgloss.Style

	NotifPositiveSty lipgloss.Style
	NotifNeutralSty  lipgloss.Style
	NotifWarningSty  lipgloss.Style
	NotifNegativeSty lipgloss.Style
	NotifIdleSty     lipgloss.Style
	NotifTitleSty    lipgloss.Style

	Series []lipgloss.Style
)

const (
	DefaultTooNarrowTitle = "TERMINAL TOO NARROW"
	DefaultTooNarrowNeed  = "Need at least %d columns."
	DefaultTooNarrowBody  = "Current width: %s columns. Resize the terminal to at least %d characters wide to use this screen."
)

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

func Default() Theme { return New(defaultPalette) }

var current = func() *atomic.Pointer[Theme] {
	p := new(atomic.Pointer[Theme])
	t := withCachedStyles(Default())
	p.Store(&t)
	syncExported(t)
	return p
}()

func Cur() *Theme { return current.Load() }

// Use makes t the active theme. Safe to call while other goroutines render.
func Use(t Theme) {
	t = withCachedStyles(t)
	current.Store(&t)
}

func withCachedStyles(t Theme) Theme {
	t.stripBg = stripBgOf(t)
	t.stripText = lipgloss.NewStyle().Background(t.stripBg)
	t.stripBold = t.stripText.Bold(true)
	return t
}

func syncExported(t Theme) {
	TitleSty = t.Title
	AccentSty = t.Accent
	DimSty = t.Dim
	ValSty = t.Val
	KeySty = t.Key
	CanSty = t.Can
	CantSty = t.Cant

	AppFrame = t.AppFrame

	PanelSty = t.Panel
	PanelFocusSty = t.PanelFocus
	PanelTitleSty = t.PanelTitle
	CardSty = t.Card

	NotifPositiveSty = t.NotifPositive
	NotifNeutralSty = t.NotifNeutral
	NotifWarningSty = t.NotifWarning
	NotifNegativeSty = t.NotifNegative
	NotifIdleSty = t.NotifIdle
	NotifTitleSty = t.NotifTitle

	Series = t.Series
}
