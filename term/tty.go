package term

import (
	"io"

	xterm "github.com/charmbracelet/x/term"
)

// IsTerminal reports whether w is attached to a terminal.
//
// It accepts any writer exposing an Fd method, not just *os.File, so callers
// can substitute fakes in tests. Writers without a file descriptor are never
// considered terminals.
func IsTerminal(w io.Writer) bool {
	f, ok := w.(interface{ Fd() uintptr })
	if !ok {
		return false
	}
	return xterm.IsTerminal(f.Fd())
}
