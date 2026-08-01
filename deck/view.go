package deck

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/viewkit/keys"
)

// View is a navigable screen hosted by Model. Apps implement domain views;
// the tea runtime lives only in this module (tea must not enter viewkit core).
type View interface {
	// Title names the view in the breadcrumb trail.
	Title() string
	// Init returns the command to run when the view is first shown.
	Init() tea.Cmd
	// Update handles one message. m is the hosting Model, for Push/Pop
	// and size; the returned command is run by the tea loop.
	Update(m *Model, msg tea.Msg) tea.Cmd
	// Body renders the view's content into the given width and height
	// (the area left between header and footer chrome).
	Body(width, height int) string
	// Hints lists key legend entries for the footer.
	Hints() []keys.Hint
	// Context lists key/label cues for the header context strip.
	Context() []keys.Hint
}
