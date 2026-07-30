package panels

import (
	"regexp"
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
			out = append(out, f.Rule())
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

func MarkdownPanel(f layout.Frame, title, src string) string {
	return f.Panel(title, strings.Split(Markdown(f, src), "\n")...)
}

const mdMarkers = "*_`["

func mdInline(s string) string {
	if !strings.ContainsAny(s, mdMarkers) {
		return s
	}
	t := theme.Cur()
	s = mdReplace(s, mdCode, func(b *strings.Builder, src string, loc []int) {
		b.WriteString(t.Key.Render(mdGroup(src, loc, 1)))
	})
	s = mdReplace(s, mdBold, func(b *strings.Builder, src string, loc []int) {
		b.WriteString(t.Accent.Bold(true).Render(mdGroup(src, loc, 1)))
	})
	s = mdReplace(s, mdItalic, func(b *strings.Builder, src string, loc []int) {
		text := mdGroup(src, loc, 1)
		if text == "" {
			text = mdGroup(src, loc, 2)
		}
		b.WriteString(t.Val.Italic(true).Render(text))
	})
	s = mdReplace(s, mdLink, func(b *strings.Builder, src string, loc []int) {
		b.WriteString(t.Accent.Render(mdGroup(src, loc, 1)))
		b.WriteString(t.Dim.Render(" (" + mdGroup(src, loc, 2) + ")"))
	})
	return s
}

func mdReplace(s string, re *regexp.Regexp, render func(*strings.Builder, string, []int)) string {
	locs := re.FindAllStringSubmatchIndex(s, -1)
	if len(locs) == 0 {
		return s
	}
	var b strings.Builder
	b.Grow(len(s) + 32*len(locs))
	end := 0
	for _, loc := range locs {
		b.WriteString(s[end:loc[0]])
		render(&b, s, loc)
		end = loc[1]
	}
	b.WriteString(s[end:])
	return b.String()
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
