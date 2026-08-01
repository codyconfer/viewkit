package forms

import (
	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/panels"
	"github.com/codyconfer/viewkit/theme"
)

// Result is the outcome of feeding a key action to a Confirm dialog.
type Result int

const (
	// Pending means the dialog is still open awaiting a decision.
	Pending Result = iota

	// Submitted means the confirm key was pressed; the chosen answer is in
	// Confirm.Yes.
	Submitted

	// Cancelled means the dialog was dismissed with the cancel key.
	Cancelled
)

// Confirm is a modal yes/no dialog. The zero value is usable: YesLabel and
// NoLabel default to "Yes" and "No", and Yes holds the current selection.
type Confirm struct {
	Title    string
	Message  string
	YesLabel string
	NoLabel  string
	Yes      bool
}

func (c Confirm) labels() (yes, no string) {
	yes, no = c.YesLabel, c.NoLabel
	if yes == "" {
		yes = "Yes"
	}
	if no == "" {
		no = "No"
	}
	return yes, no
}

// Handle applies a key action: Left/Dec selects yes, Right/Inc selects no,
// Confirm returns Submitted (with the answer left in c.Yes), and Cancel
// returns Cancelled. Every other action leaves the dialog Pending.
func (c *Confirm) Handle(a keys.Action) Result {
	switch a {
	case keys.Left, keys.Dec:
		c.Yes = true
	case keys.Right, keys.Inc:
		c.Yes = false
	case keys.Confirm:
		return Submitted
	case keys.Cancel:
		return Cancelled
	}
	return Pending
}

// Render draws the dialog as a titled panel sized to f: the optional Message
// followed by the yes/no toggle.
func (c Confirm) Render(f layout.Frame) string {
	yes, no := c.labels()
	lines := []string{}
	if c.Message != "" {
		lines = append(lines, theme.Cur().Val.Render(f.Fit(c.Message)))
		lines = append(lines, "")
	}
	lines = append(lines, panels.Toggle(yes, no, c.Yes))
	return f.Panel(c.Title, lines...)
}

// Overlay renders the dialog and composes it over the background bg with
// layout.Overlay, placed at pos (centered by default).
func (c Confirm) Overlay(bg string, f layout.Frame, pos ...layout.OverlayPos) string {
	return layout.Overlay(bg, c.Render(f), pos...)
}
