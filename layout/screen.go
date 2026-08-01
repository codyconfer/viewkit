package layout

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"

	"github.com/codyconfer/viewkit/theme"
)

// FrameFor sizes a Frame to the terminal behind w when w is a *os.File on a
// tty; anything else (pipes, buffers, size errors) gets DefaultFrame.
func FrameFor(w io.Writer) Frame {
	if f, ok := w.(*os.File); ok {
		if width, _, err := term.GetSize(f.Fd()); err == nil && width > 0 {
			return ScreenFrame(width)
		}
	}
	return DefaultFrame()
}

// FitsScreenWidth reports whether a terminal of this width can host the UI.
// Unknown widths (<= 0) are assumed to fit; below theme.MinScreenWidth render
// TooNarrow instead.
func FitsScreenWidth(screenWidth int) bool {
	return screenWidth <= 0 || screenWidth >= theme.MinScreenWidth
}

// ScreenFrame converts a full terminal width into a body Frame by subtracting
// theme.ScreenPaddingWidth; unknown widths get DefaultFrame. NewFrame's
// minimum-width clamp still applies.
func ScreenFrame(screenWidth int) Frame {
	if screenWidth <= 0 {
		return DefaultFrame()
	}
	return NewFrame(screenWidth - theme.ScreenPaddingWidth)
}

// TooNarrow renders the theme's "terminal too narrow" notice, sized to the
// given width (or the minimum layout width when the width is unknown), showing
// the current and required column counts.
func TooNarrow(screenWidth int) string {
	current := "unknown"
	if screenWidth > 0 {
		current = fmt.Sprintf("%d", screenWidth)
	}

	width := theme.MinScreenWidth - theme.ScreenPaddingWidth
	if screenWidth > 0 {
		width = max(screenWidth-theme.AppMarginX*2, 1)
	}

	t := theme.Cur()
	title := t.Title.Render(ansi.Truncate(t.TooNarrowTitle, width, "…"))
	subtitle := t.Dim.Render(ansi.Truncate(fmt.Sprintf(t.TooNarrowNeed, theme.MinScreenWidth), width, "…"))
	body := lipgloss.NewStyle().Width(width).Render(
		fmt.Sprintf(t.TooNarrowBody, current, theme.MinScreenWidth),
	)
	return lipgloss.JoinVertical(lipgloss.Left, title, subtitle, body)
}
