package deck

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/ui"
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
	// Body renders the view's content into f (the area left between header
	// and footer chrome).
	Body(f layout.Frame) string
	// Hints lists key legend entries for the footer.
	Hints(scope *ui.Scope) []keys.Hint
	// Context lists key/label cues for the header context strip.
	Context(scope *ui.Scope) []keys.Hint
}
