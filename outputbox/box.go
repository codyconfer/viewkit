// Package outputbox renders append-only, status-colored command output blocks.
package outputbox

import (
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
)

// Status identifies the outcome shown in a box header and background.
type Status string

const (
	// Pending marks an action still in progress.
	Pending Status = "RUNNING"
	// Success marks a completed action.
	Success Status = "DONE"
	// Warning marks output that needs review.
	Warning Status = "REVIEW"
	// Failure marks a failed action.
	Failure Status = "ERROR"
)

// Render returns a borderless box. Width includes horizontal padding. Without
// color the status label remains visible, including in redirected transcripts.
// Widths below eight cells are clamped to eight. Content may contain trusted ANSI
// styling; callers must sanitize terminal control sequences from untrusted input.
// Render does not write output, move the cursor, or change global color settings.
func Render(title string, lines []string, width int, color bool, status Status) string {
	width = max(8, width)
	inner := width - 4
	r := lipgloss.NewRenderer(io.Discard)
	profile := termenv.Ascii
	if color {
		profile = termenv.ANSI256
	}
	r.SetColorProfile(profile)
	bg, fg := lipgloss.Color("236"), lipgloss.Color("252")
	switch status {
	case Success:
		bg = "22"
	case Warning:
		bg = "58"
	case Failure:
		bg = "52"
	}
	row := r.NewStyle().Background(bg).Foreground(fg).Width(width).Padding(0, 2)
	header := r.NewStyle().Bold(true)
	// Wrap before styling so every physical line receives full-width shading.
	content := []string{header.Render("[" + string(status) + "] " + title), ""}
	content = append(content, lines...)
	result := []string{row.Render("")}
	for _, block := range content {
		for _, line := range strings.Split(ansi.Wrap(block, inner, ""), "\n") {
			result = append(result, row.Render(line))
		}
	}
	result = append(result, row.Render(""))
	return strings.Join(result, "\n")
}
