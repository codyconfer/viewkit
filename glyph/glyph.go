package glyph

import (
	"io"
	"os"
	"strings"
	"sync/atomic"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
)

// Mode selects which Variants field glyphs render as.
type Mode int

// Glyph rendering modes, in order of decreasing symbol richness.
const (
	ModeNerd    Mode = iota // nerd-font icons (the zero value and default)
	ModeUnicode             // plain unicode symbols
	ModeNone                // ASCII only
)

var mode atomic.Int32

// SetMode sets the process-wide glyph mode.
func SetMode(m Mode) { mode.Store(int32(m)) }

// CurrentMode returns the process-wide glyph mode.
func CurrentMode() Mode { return Mode(mode.Load()) }

// ParseMode parses "nerd", "unicode"/"uni", or "none"/"ascii"/"off"
// (case-insensitively, ignoring surrounding space) into a Mode. It returns
// (ModeNerd, false) for anything else, including the empty string.
func ParseMode(s string) (Mode, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "nerd":
		return ModeNerd, true
	case "unicode", "uni":
		return ModeUnicode, true
	case "none", "ascii", "off":
		return ModeNone, true
	}
	return ModeNerd, false
}

// DetectMode sets the process-wide glyph mode: an explicit override wins
// (pass the value of the app's own env var, e.g. os.Getenv("MYAPP_ICONS")),
// otherwise dumb or non-TTY output falls back to ASCII and everything else
// gets nerd glyphs.
func DetectMode(w io.Writer, override string) {
	if m, ok := ParseMode(override); ok {
		SetMode(m)
		return
	}
	if strings.EqualFold(os.Getenv("TERM"), "dumb") || !isTerminal(w) {
		SetMode(ModeNone)
		return
	}
	SetMode(ModeNerd)
}

func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(f.Fd())
}

// Variants holds one glyph's rendering per mode: nerd-font, unicode, and
// ASCII. Fields may be left empty; String degrades to a filled one.
type Variants struct {
	Nerd  string
	Uni   string
	ASCII string
}

// String returns the variant for the active mode, degrading to the nearest
// filled variant when a Variants is only partially populated. A plugin that
// registers a Nerd-only glyph still shows something in unicode/ASCII mode
// instead of silently rendering nothing.
func (v Variants) String() string {
	switch CurrentMode() {
	case ModeUnicode:
		return firstNonEmpty(v.Uni, v.ASCII, v.Nerd)
	case ModeNone:
		return firstNonEmpty(v.ASCII, v.Uni, v.Nerd)
	default:
		return firstNonEmpty(v.Nerd, v.Uni, v.ASCII)
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// Pad appends a single trailing space to s, or returns "" unchanged so an
// absent glyph adds no stray gap.
func Pad(s string) string {
	if s == "" {
		return ""
	}
	return s + " "
}

// LeadWidth is the display-cell width Lead pads glyphs to, including the gap
// before the following text.
const LeadWidth = 3

// Lead pads s with trailing spaces to LeadWidth display cells so leading
// glyphs of different widths align. Glyphs already LeadWidth or wider get a
// single trailing space; "" returns "".
func Lead(s string) string {
	if s == "" {
		return ""
	}
	if gap := LeadWidth - lipgloss.Width(s); gap > 0 {
		return s + strings.Repeat(" ", gap)
	}
	return s + " "
}

var (
	statusOK    = Variants{"", "●", "+"}
	statusWarn  = Variants{"", "▲", "!"}
	statusBad   = Variants{"", "●", "x"}
	statusMuted = Variants{"", "○", "-"}
	check       = Variants{"", "✓", "+"}
	cross       = Variants{"", "✗", "x"}
	warn        = Variants{"", "⚠", "!"}
	arrow       = Variants{"", "→", "->"}
	bullet      = Variants{"", "•", "*"}
	github      = Variants{"", "●", "gh"}
	slack       = Variants{"", "●", "sl"}
	google      = Variants{"", "●", "go"}
	diamond     = Variants{"", "◈", ">"}
	history     = Variants{"", "◷", ">"}
	list        = Variants{"", "≣", ">"}
	database    = Variants{"", "▤", ">"}
	cog         = Variants{"", "⚙", ">"}
	user        = Variants{"", "◆", ">"}
	signOut     = Variants{"", "⏻", ">"}
	clock       = Variants{"", "◰", ">"}
)

// StatusOK returns the healthy status dot for the current mode.
func StatusOK() string { return statusOK.String() }

// StatusWarn returns the warning status marker for the current mode.
func StatusWarn() string { return statusWarn.String() }

// StatusBad returns the failing status dot for the current mode.
func StatusBad() string { return statusBad.String() }

// StatusMuted returns the inactive status dot for the current mode.
func StatusMuted() string { return statusMuted.String() }

// Check returns the check-mark glyph for the current mode.
func Check() string { return check.String() }

// Cross returns the cross/failure glyph for the current mode.
func Cross() string { return cross.String() }

// Warn returns the warning glyph for the current mode.
func Warn() string { return warn.String() }

// Arrow returns the right-arrow glyph for the current mode.
func Arrow() string { return arrow.String() }

// Bullet returns the bullet glyph for the current mode.
func Bullet() string { return bullet.String() }

// GitHub returns the GitHub logo glyph for the current mode.
func GitHub() string { return github.String() }

// Slack returns the Slack logo glyph for the current mode.
func Slack() string { return slack.String() }

// Google returns the Google logo glyph for the current mode.
func Google() string { return google.String() }

// Diamond returns the diamond glyph for the current mode.
func Diamond() string { return diamond.String() }

// History returns the history/recent-activity glyph for the current mode.
func History() string { return history.String() }

// List returns the list glyph for the current mode.
func List() string { return list.String() }

// Database returns the database glyph for the current mode.
func Database() string { return database.String() }

// Cog returns the settings-cog glyph for the current mode.
func Cog() string { return cog.String() }

// User returns the user/account glyph for the current mode.
func User() string { return user.String() }

// SignOut returns the sign-out glyph for the current mode.
func SignOut() string { return signOut.String() }

// Clock returns the clock glyph for the current mode.
func Clock() string { return clock.String() }

// Severity is the shared tone vocabulary for glyphs, notifications, and tray
// state mapping. Apps supply a single kind→Severity classifier; viewkit never
// knows domain kind strings.
type Severity int

// Severity tones. SeverityNeutral is the zero value and the fallback for
// unknown severities in GlyphFor and StatusFor.
const (
	SeverityNeutral  Severity = iota // no tone; renders as Bullet/StatusMuted
	SeverityPositive                 // success; renders as Check/StatusOK
	SeverityWarning                  // caution; renders as Warn/StatusWarn
	SeverityNegative                 // failure; renders as Cross/StatusBad
)

// GlyphFor returns an inline mark for sev (Check/Warn/Cross/Bullet).
func GlyphFor(sev Severity) string { //nolint:revive // name predates the lint rule; renaming would break callers
	switch sev {
	case SeverityPositive:
		return Check()
	case SeverityWarning:
		return Warn()
	case SeverityNegative:
		return Cross()
	default:
		return Bullet()
	}
}

// StatusFor returns the status-strip dot for sev (StatusOK/Warn/Bad/Muted).
// Use GlyphFor for inline marks; StatusFor for status strips and chips.
func StatusFor(sev Severity) string {
	switch sev {
	case SeverityPositive:
		return StatusOK()
	case SeverityNegative:
		return StatusBad()
	case SeverityWarning:
		return StatusWarn()
	default:
		return StatusMuted()
	}
}
