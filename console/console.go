// Package console detects output capabilities without writing escape sequences
// or reading terminal input. Font coverage cannot be queried portably, so font
// support is unknown unless explicitly configured by the user.
package console

import (
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/x/term"
)

// Support distinguishes an unverified capability from known lack of support.
type Support uint8

const (
	// Unknown means there is not enough information to determine support.
	Unknown Support = iota
	// Unsupported means the capability is disabled or unavailable.
	Unsupported
	// Supported means the capability is explicitly enabled.
	Supported
)

// Capabilities describes one output destination, not process-wide state.
// Unicode and ANSI are conservative environment-based estimates, not probes.
// Powerline support alone does not imply support for Nerd Font icons.
type Capabilities struct {
	Terminal  bool
	ANSI      bool
	Unicode   bool
	Powerline Support
	NerdFont  Support
}

// Detect checks whether w is a terminal and inspects the current environment.
// Non-file writers, redirected files, and TERM=dumb use plain ASCII output.
// Unicode requires a UTF-8 locale (LC_ALL, LC_CTYPE, then LANG), or an
// explicitly enabled font. VIEWKIT_POWERLINE and VIEWKIT_NERD_FONT accept
// strconv.ParseBool values; unset, "auto", and invalid values mean unknown.
// Nerd Fonts include Powerline symbols, unless Powerline is explicitly disabled.
// Font overrides do not enable capabilities for non-terminal or dumb output.
// Detect neither changes global rendering state nor consumes terminal input.
func Detect(w io.Writer) Capabilities {
	f, ok := w.(*os.File)
	return detect(ok && f != nil && term.IsTerminal(f.Fd()), os.Getenv)
}

func detect(terminal bool, getenv func(string) string) Capabilities {
	c := Capabilities{Terminal: terminal}
	if !terminal || strings.EqualFold(strings.TrimSpace(getenv("TERM")), "dumb") {
		c.Powerline, c.NerdFont = Unsupported, Unsupported
		return c
	}
	c.ANSI = true
	for _, name := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
		if locale := getenv(name); locale != "" {
			locale = strings.ToLower(locale)
			c.Unicode = strings.Contains(locale, "utf-8") || strings.Contains(locale, "utf8")
			break
		}
	}
	c.Powerline = fontSupport(getenv("VIEWKIT_POWERLINE"))
	c.NerdFont = fontSupport(getenv("VIEWKIT_NERD_FONT"))
	if c.NerdFont == Supported && c.Powerline == Unknown {
		c.Powerline = Supported
	}
	if c.Powerline == Supported || c.NerdFont == Supported {
		c.Unicode = true
	}
	return c
}

func fontSupport(value string) Support {
	enabled, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return Unknown
	}
	if enabled {
		return Supported
	}
	return Unsupported
}
