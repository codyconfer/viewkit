package deck

import (
	"context"

	"github.com/charmbracelet/lipgloss"
)

// StatusLevel aliases theme severity for status chips.
type StatusLevel int

// ServiceStatus is one right-strip status chip.
type ServiceStatus struct {
	Name   string
	Detail string
	Level  StatusLevel
	Glyph  string
	Color  lipgloss.TerminalColor
}

// StatusInfo is optional chrome identity + service chips.
type StatusInfo struct {
	Identity string // e.g. "@user ✓"
	Services []ServiceStatus
}

// StatusFunc loads status asynchronously for the chrome strip.
type StatusFunc func(context.Context) StatusInfo

// Chrome configures brand/header appearance. Apps inject product branding;
// viewkit/deck never hard-codes app names.
type Chrome struct {
	Brand      string // e.g. "MUNIN"
	BrandGlyph string
	Subtitle   string // e.g. "ono-sendai deck"
	ClockGlyph string
}
