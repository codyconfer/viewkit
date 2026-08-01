package deck

import (
	"context"

	"github.com/charmbracelet/lipgloss"

	"github.com/codyconfer/viewkit/glyph"
)

// ServiceStatus is one right-strip status chip. Severity drives the default
// glyph and color; Glyph and Color, when set, override that resolution. ID is
// an optional stable identifier apps can use to filter or configure chips.
type ServiceStatus struct {
	ID       string
	Name     string
	Detail   string
	Severity glyph.Severity
	Glyph    string
	Color    lipgloss.TerminalColor
}

// StatusInfo is optional chrome identity + service chips.
type StatusInfo struct {
	Identity string
	Services []ServiceStatus
}

// StatusFunc loads status asynchronously for the chrome strip.
type StatusFunc func(context.Context) StatusInfo

// Chrome configures brand/header appearance. Apps inject product branding;
// viewkit/deck never hard-codes app names.
type Chrome struct {
	Brand      string
	BrandGlyph string
	Subtitle   string
	ClockGlyph string
}
