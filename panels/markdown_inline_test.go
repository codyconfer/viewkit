package panels

import (
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"

	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/theme"
)

func mdInlineLegacy(s string) string {
	t := theme.Cur()
	s = mdCode.ReplaceAllStringFunc(s, func(m string) string {
		return t.Key.Render(mdCode.FindStringSubmatch(m)[1])
	})
	s = mdBold.ReplaceAllStringFunc(s, func(m string) string {
		return t.Accent.Bold(true).Render(mdBold.FindStringSubmatch(m)[1])
	})
	s = mdItalic.ReplaceAllStringFunc(s, func(m string) string {
		sub := mdItalic.FindStringSubmatch(m)
		text := sub[1]
		if text == "" {
			text = sub[2]
		}
		return t.Val.Italic(true).Render(text)
	})
	s = mdLink.ReplaceAllStringFunc(s, func(m string) string {
		sub := mdLink.FindStringSubmatch(m)
		return t.Accent.Render(sub[1]) + t.Dim.Render(" ("+sub[2]+")")
	})
	return s
}

func mdInlineCorpus() []string {
	return []string{
		"",
		" ",
		"a plain sentence of prose with no markdown syntax at all",
		"snake_case_identifier and another_one here",
		"_leading underscore",
		"trailing underscore_",
		"__dunder__ name",
		"a_b_c_d_e",
		"https://example.invalid/some_path_with_underscores/and_more",
		"see https://example.invalid/a_b for _details_",
		"2 * 3 * 4 = 24",
		"a * b",
		"*",
		"**",
		"***",
		"****",
		"*****",
		"unbalanced *emphasis",
		"emphasis* unbalanced",
		"**bold**",
		"**bold** and *italic* and `code`",
		"***bold italic***",
		"**nested *inner* outer**",
		"*outer **inner** outer*",
		"**a** **b** **c**",
		"*a* *b* *c*",
		"_a_ _b_ _c_",
		"mixed *a* and _b_ together",
		"`code with *stars* inside`",
		"`code with _score_ inside`",
		"`code` then **bold** then `more code`",
		"``",
		"`",
		"`unclosed code",
		"a ` b ` c ` d",
		"[link](http://x)",
		"[link](http://x) and [other](http://y)",
		"[**bold label**](http://x)",
		"[label](http://x/a_b_c)",
		"[label with *italic*](http://x)",
		"bare [brackets] with no target",
		"[unclosed bracket",
		"unclosed] bracket",
		"[](http://x)",
		"[label]()",
		"[label] (http://x)",
		"![image](http://x/img.png)",
		"a [b](c) *d* _e_ `f` **g** all at once",
		"trailing marker *",
		"* leading marker",
		"_",
		"__",
		"[",
		"]",
		"()",
		"emoji ✨ with *italic* ✨",
		"wide 日本語 **太字** text",
		"tab\tseparated *italic*",
		"a very long line of prose that also contains **bold** near the end so wrapping and styling interact",
		"already \x1b[1mescaped\x1b[0m text with *italic*",
		"*a**b*",
		"**a*b**",
		"_a__b_",
		"*_both_*",
		"_*both*_",
		"`*`",
		"`_`",
		"`[`",
		"[`code`](http://x)",
	}
}

func TestMdInlineMatchesLegacyByteForByte(t *testing.T) {
	prev := lipgloss.ColorProfile()
	t.Cleanup(func() { lipgloss.SetColorProfile(prev) })

	for _, prof := range []struct {
		name string
		p    termenv.Profile
	}{
		{"Ascii", termenv.Ascii},
		{"TrueColor", termenv.TrueColor},
	} {
		t.Run(prof.name, func(t *testing.T) {
			lipgloss.SetColorProfile(prof.p)
			for _, in := range mdInlineCorpus() {
				want := mdInlineLegacy(in)
				if got := mdInline(in); got != want {
					t.Errorf("mdInline(%q):\n got %q\nwant %q", in, got, want)
				}
			}
		})
	}
}

func TestMarkdownMatchesLegacyInlineByteForByte(t *testing.T) {
	prev := lipgloss.ColorProfile()
	t.Cleanup(func() { lipgloss.SetColorProfile(prev) })

	corpus := mdInlineCorpus()
	docs := []string{
		strings.Join(corpus, "\n"),
		benchMarkdownDoc20Lines(),
	}
	for _, line := range corpus {
		docs = append(docs,
			line,
			"- "+line,
			"* "+line,
			"> "+line,
			"1. "+line,
			"# "+line,
			"```\n"+line+"\n```",
		)
	}

	for _, prof := range []struct {
		name string
		p    termenv.Profile
	}{
		{"Ascii", termenv.Ascii},
		{"TrueColor", termenv.TrueColor},
	} {
		t.Run(prof.name, func(t *testing.T) {
			lipgloss.SetColorProfile(prof.p)
			for _, width := range []int{20, 40, 81} {
				f := layout.NewFrame(width)
				for _, src := range docs {
					if got, want := Markdown(f, src), markdownLegacy(f, src); got != want {
						t.Errorf("width %d, Markdown(%q):\n got %q\nwant %q", width, src, got, want)
					}
				}
			}
		})
	}
}

func markdownLegacy(f layout.Frame, src string) string {
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
			body := mdInlineLegacy(strings.TrimPrefix(trimmed, "> "))
			out = append(out, wrapInline(t.Dim.Render("┃ "), body, width)...)
		case strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* "):
			body := mdInlineLegacy(trimmed[2:])
			out = append(out, wrapInline(t.Accent.Render("• "), body, width)...)
		case mdOrdered.MatchString(trimmed):
			mset := mdOrdered.FindStringSubmatch(trimmed)
			body := mdInlineLegacy(mset[2])
			out = append(out, wrapInline(t.Accent.Render(mset[1]+". "), body, width)...)
		default:
			out = append(out, wrapInline("", mdInlineLegacy(trimmed), width)...)
		}
	}
	return strings.Join(out, "\n")
}

func TestMdInlinePlainLineSkipsTheRegexPasses(t *testing.T) {
	prev := lipgloss.ColorProfile()
	t.Cleanup(func() { lipgloss.SetColorProfile(prev) })
	lipgloss.SetColorProfile(termenv.TrueColor)

	plain := "a plain sentence of prose with no markdown syntax at all"
	mdInline(plain)
	mdInlineLegacy(plain)

	guarded := testing.AllocsPerRun(200, func() { benchSink = mdInline(plain) })
	legacy := testing.AllocsPerRun(200, func() { benchSink = mdInlineLegacy(plain) })
	t.Logf("plain line allocations: guarded=%v legacy=%v", guarded, legacy)

	if guarded != 0 {
		t.Errorf("a line with no inline markup should cost 0 allocations, got %v", guarded)
	}
	if legacy <= guarded {
		t.Errorf("the unguarded regex passes were expected to allocate: legacy=%v guarded=%v", legacy, guarded)
	}
}

var mdInlineEscapeRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func TestMdInlinePlainLineEmitsNoEscapes(t *testing.T) {
	prev := lipgloss.ColorProfile()
	t.Cleanup(func() { lipgloss.SetColorProfile(prev) })
	lipgloss.SetColorProfile(termenv.TrueColor)

	for _, in := range mdInlineCorpus() {
		if strings.ContainsAny(in, "*_`[") {
			continue
		}
		if got := mdInline(in); got != in {
			t.Errorf("mdInline(%q) = %q, want the input unchanged", in, got)
		}
		if mdInlineEscapeRe.MatchString(mdInline(in)) {
			t.Errorf("mdInline(%q) emitted escapes", in)
		}
	}
}
