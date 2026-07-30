package glyph

import (
	"io"
	"os"
	"strings"
	"sync/atomic"

	"github.com/charmbracelet/x/term"
)

type Mode int

const (
	ModeNerd Mode = iota
	ModeUnicode
	ModeNone
)

var mode atomic.Int32

func SetMode(m Mode) { mode.Store(int32(m)) }

func CurrentMode() Mode { return Mode(mode.Load()) }

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

func Detect(w io.Writer, env string) {
	if m, ok := ParseMode(env); ok {
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

func Pad(s string) string {
	if s == "" {
		return ""
	}
	return s + " "
}

func Lead(s string) string {
	if s == "" {
		return ""
	}
	if CurrentMode() == ModeNerd {
		return s + "  "
	}
	return s + " "
}

var (
	statusOK    = Variants{"", "●", "+"}
	statusWarn  = Variants{"", "▲", "!"}
	statusBad   = Variants{"", "●", "x"}
	statusMuted = Variants{"", "○", "-"}
	check       = Variants{"", "✓", "ok"}
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

func StatusOK() string    { return statusOK.String() }
func StatusWarn() string  { return statusWarn.String() }
func StatusBad() string   { return statusBad.String() }
func StatusMuted() string { return statusMuted.String() }
func Check() string       { return check.String() }
func Cross() string       { return cross.String() }
func Warn() string        { return warn.String() }
func Arrow() string       { return arrow.String() }
func Bullet() string      { return bullet.String() }
func GitHub() string      { return github.String() }
func Slack() string       { return slack.String() }
func Google() string      { return google.String() }
func Diamond() string     { return diamond.String() }
func History() string     { return history.String() }
func List() string        { return list.String() }
func Database() string    { return database.String() }
func Cog() string         { return cog.String() }
func User() string        { return user.String() }
func SignOut() string     { return signOut.String() }
func Clock() string       { return clock.String() }

// Severity is the shared tone vocabulary for glyphs, notifications, and tray
// state mapping. Apps supply a single kind→Severity classifier; viewkit never
// knows domain kind strings.
type Severity int

const (
	SeverityNeutral Severity = iota
	SeverityPositive
	SeverityWarning
	SeverityNegative
)

// GlyphFor returns a status glyph for sev.
func GlyphFor(sev Severity) string {
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
