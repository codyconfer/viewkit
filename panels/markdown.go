package panels

import (
	"regexp"
	"regexp/syntax"
	"strings"

	"github.com/charmbracelet/x/ansi"

	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/theme"
)

var (
	mdBold    = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	mdItalic  = regexp.MustCompile(`(?:\*([^*]+)\*|_([^_]+)_)`)
	mdCode    = regexp.MustCompile("`([^`]+)`")
	mdLink    = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	mdOrdered = regexp.MustCompile(`^(\d+)\. +(.*)$`)
)

// Markdown renders a small markdown subset as themed terminal text wrapped to
// the frame body width, without a surrounding panel. Supported: #/##/###
// headings, ---/***/___ rules, > quotes, -/* and numbered lists, ``` fenced
// code (rendered dim, verbatim, and truncated rather than wrapped; the fence
// lines themselves are dropped), and inline `code`, **bold**, *italic* or
// _italic_, and [text](url) links, which render as "text (url)". Quote, list,
// and plain lines word-wrap with continuation lines indented under the
// marker; headings and rules truncate with an ellipsis. Anything else passes
// through as plain text.
func Markdown(f layout.Frame, src string) string {
	width := f.BodyWidth()
	t := theme.Cur()

	var out []string
	inCode := false
	for _, raw := range strings.Split(src, "\n") {
		line := strings.TrimRight(raw, " ")
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			inCode = !inCode
			continue
		}
		if inCode {
			out = append(out, t.Dim.Render(ansi.Truncate(line, width, "…")))
			continue
		}

		switch {
		case trimmed == "":
			out = append(out, "")
		case trimmed == "---" || trimmed == "***" || trimmed == "___":
			out = append(out, ansi.Truncate(f.Rule(), width, ""))
		case strings.HasPrefix(trimmed, "### "):
			out = append(out, t.Accent.Render(ansi.Truncate(strings.TrimPrefix(trimmed, "### "), width, "…")))
		case strings.HasPrefix(trimmed, "## "):
			out = append(out, t.Title.Render(ansi.Truncate(strings.TrimPrefix(trimmed, "## "), width, "…")))
		case strings.HasPrefix(trimmed, "# "):
			out = append(out, t.Title.Bold(true).Render(ansi.Truncate(strings.TrimPrefix(trimmed, "# "), width, "…")))
		case strings.HasPrefix(trimmed, "> "):
			body := mdInline(strings.TrimPrefix(trimmed, "> "))
			out = append(out, wrapInline(t.Dim.Render("┃ "), body, width)...)
		case strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* "):
			body := mdInline(trimmed[2:])
			out = append(out, wrapInline(t.Accent.Render("• "), body, width)...)
		case mdOrdered.MatchString(trimmed):
			mset := mdOrdered.FindStringSubmatch(trimmed)
			body := mdInline(mset[2])
			out = append(out, wrapInline(t.Accent.Render(mset[1]+". "), body, width)...)
		default:
			out = append(out, wrapInline("", mdInline(trimmed), width)...)
		}
	}
	return strings.Join(out, "\n")
}

// MarkdownPanel wraps Markdown's output in a titled panel.
func MarkdownPanel(f layout.Frame, title, src string) string {
	return f.Panel(title, strings.Split(Markdown(f, src), "\n")...)
}

// mdPass is one inline styling pass. The list below is the single source of
// truth: mdInline runs it in order and mdMarkers is derived from it, so a new
// pattern cannot be dead on arrival behind a stale fast-path guard.
type mdPass struct {
	re     *regexp.Regexp
	render func(b *strings.Builder, t theme.Theme, src string, loc []int)
}

var mdInlinePasses = []mdPass{
	{mdCode, func(b *strings.Builder, t theme.Theme, src string, loc []int) {
		b.WriteString(t.Key.Render(mdGroup(src, loc, 1)))
	}},
	{mdBold, func(b *strings.Builder, t theme.Theme, src string, loc []int) {
		b.WriteString(t.Accent.Bold(true).Render(mdGroup(src, loc, 1)))
	}},
	{mdItalic, func(b *strings.Builder, t theme.Theme, src string, loc []int) {
		text := mdGroup(src, loc, 1)
		if text == "" {
			text = mdGroup(src, loc, 2)
		}
		b.WriteString(t.Val.Italic(true).Render(text))
	}},
	{mdLink, func(b *strings.Builder, t theme.Theme, src string, loc []int) {
		b.WriteString(t.Accent.Render(mdGroup(src, loc, 1)))
		b.WriteString(t.Dim.Render(" (" + mdGroup(src, loc, 2) + ")"))
	}},
}

// mdMarkers is every literal byte the passes can need. A line without one of
// them cannot match any pattern, so mdInline returns it untouched.
var mdMarkers = mdMarkerSet(mdInlinePasses)

func mdMarkerSet(passes []mdPass) string {
	var b strings.Builder
	seen := map[rune]bool{}
	for _, p := range passes {
		for _, r := range reLiterals(p.re) {
			if !seen[r] {
				seen[r] = true
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

// reLiterals is the set of literal runes a pattern can match, taken from the
// parsed pattern rather than from a hand-kept copy of it.
func reLiterals(re *regexp.Regexp) []rune {
	parsed, err := syntax.Parse(re.String(), syntax.Perl)
	if err != nil {
		return []rune(re.String())
	}
	var out []rune
	var walk func(*syntax.Regexp)
	walk = func(r *syntax.Regexp) {
		if r.Op == syntax.OpLiteral {
			out = append(out, r.Rune...)
		}
		for _, sub := range r.Sub {
			walk(sub)
		}
	}
	walk(parsed)
	return out
}

func mdInline(s string) string {
	if !strings.ContainsAny(s, mdMarkers) {
		return s
	}
	t := theme.Cur()
	for _, p := range mdInlinePasses {
		s = mdReplace(s, p.re, t, p.render)
	}
	return s
}

func mdReplace(s string, re *regexp.Regexp, t theme.Theme,
	render func(*strings.Builder, theme.Theme, string, []int)) string {
	locs := re.FindAllStringSubmatchIndex(mdMask(s), -1)
	if len(locs) == 0 {
		return s
	}
	var b strings.Builder
	b.Grow(len(s) + 32*len(locs))
	end := 0
	for _, loc := range locs {
		b.WriteString(s[end:loc[0]])
		render(&b, t, s, loc)
		end = loc[1]
	}
	b.WriteString(s[end:])
	return b.String()
}

// mdMask blanks every escape sequence in place, keeping byte offsets, so the
// patterns match against text only: an earlier pass injects CSI sequences, and
// the '[' inside one used to satisfy mdLink, which swallowed the styled span and
// leaked the SGR parameters into the line as literal text.
func mdMask(s string) string {
	if strings.IndexByte(s, 0x1b) < 0 {
		return s
	}
	b := []byte(s)
	for i := 0; i < len(b); {
		if b[i] != 0x1b {
			i++
			continue
		}
		end := mdEscapeEnd(b, i)
		for ; i < end; i++ {
			b[i] = 0
		}
	}
	return string(b)
}

func mdEscapeEnd(b []byte, i int) int {
	if i+1 >= len(b) {
		return len(b)
	}
	switch b[i+1] {
	case '[':
		for j := i + 2; j < len(b); j++ {
			if b[j] >= 0x40 && b[j] <= 0x7e {
				return j + 1
			}
		}
		return len(b)
	case ']':
		for j := i + 2; j < len(b); j++ {
			if b[j] == 0x07 {
				return j + 1
			}
			if b[j] == 0x1b && j+1 < len(b) && b[j+1] == '\\' {
				return j + 2
			}
		}
		return len(b)
	default:
		return i + 2
	}
}

func mdGroup(s string, loc []int, n int) string {
	if 2*n+1 >= len(loc) || loc[2*n] < 0 {
		return ""
	}
	return s[loc[2*n]:loc[2*n+1]]
}

func wrapInline(prefix, styled string, width int) []string {
	indent := strings.Repeat(" ", ansi.StringWidth(prefix))
	avail := width - ansi.StringWidth(prefix)
	if avail < 1 {
		avail = 1
	}
	wrapped := ansi.Wordwrap(styled, avail, "")
	lines := strings.Split(wrapped, "\n")
	for i, l := range lines {
		if i == 0 {
			lines[i] = prefix + l
		} else {
			lines[i] = indent + l
		}
	}
	return lines
}
