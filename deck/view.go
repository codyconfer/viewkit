package deck

import tea "github.com/charmbracelet/bubbletea"

// View is a navigable screen hosted by Host. Apps implement domain views;
// the tea runtime lives only in this module (tea must not enter viewkit core).
type View interface {
	Title() string
	Init() tea.Cmd
	Update(h *Host, msg tea.Msg) tea.Cmd
	Body(width, height int) string
	Hints() [][2]string
	Context() [][2]string
}
