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
			// Frame.Rule is BodyWidth+4 wide (it underlines a whole panel), so a
			// rule dropped into a BodyWidth body wrapped onto a second line. The
			// fix is intentional and mirrored here: this reference exists to guard
			// the inline passes, not the rule.
			out = append(out, ansi.Truncate(f.Rule(), width, ""))
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

	// The filter is derived from the pass patterns rather than re-typing the
	// guard's alphabet: a line no pattern can match must come back untouched,
	// whether the fast path skipped it or every pass ran and found nothing.
	for _, in := range mdInlineCorpus() {
		matched := false
		for _, p := range mdInlinePasses {
			if p.re.MatchString(in) {
				matched = true
				break
			}
		}
		if matched {
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

// mdInlineEscapeHazards are the inputs mdInlineLegacy corrupts: an earlier pass
// injects a CSI sequence, whose literal '[' the link pattern then matches, so the
// SGR parameters leak into the line as text and the real link keeps its bracket.
// They are deliberately not in mdInlineCorpus — the legacy output is the bug, so
// the byte-for-byte comparison must not treat it as the expectation.
func mdInlineEscapeHazards() []struct{ in, want string } {
	return []struct{ in, want string }{
		{"see `go test` and the [docs](http://x) for more", "see go test and the docs (http://x) for more"},
		{"**bold** then [docs](http://x)", "bold then docs (http://x)"},
		{"*em* then [docs](http://x)", "em then docs (http://x)"},
		{"_em_ and [a](b) and `c` and [d](e)", "em and a (b) and c and d (e)"},
		{"`code` [label](http://x/a_b)", "code label (http://x/a_b)"},
	}
}

func TestMdInlineKeepsLinksIntactAfterStyledSpans(t *testing.T) {
	prev := lipgloss.ColorProfile()
	t.Cleanup(func() { lipgloss.SetColorProfile(prev) })
	lipgloss.SetColorProfile(termenv.TrueColor)

	for _, c := range mdInlineEscapeHazards() {
		got := mdInline(c.in)
		if plain := stripANSI(got); plain != c.want {
			t.Errorf("mdInline(%q)\n plain %q\n  want %q\n   raw %q", c.in, plain, c.want, got)
		}
	}
}

func TestMarkdownKeepsLinksIntactAfterStyledSpans(t *testing.T) {
	prev := lipgloss.ColorProfile()
	t.Cleanup(func() { lipgloss.SetColorProfile(prev) })
	lipgloss.SetColorProfile(termenv.TrueColor)

	f := layout.NewFrame(81)
	for _, c := range mdInlineEscapeHazards() {
		for _, prefix := range []string{"", "- ", "> ", "1. "} {
			out := stripANSI(Markdown(f, prefix+c.in))
			if !strings.Contains(out, c.want) {
				t.Errorf("Markdown(%q) = %q, want it to contain %q", prefix+c.in, out, c.want)
			}
		}
	}
}

// TestMdMarkersCoverEveryInlinePattern is the sufficiency check the hand-kept
// `const mdMarkers = "*_`["` never had: it enumerates short strings over each
// pattern's own literal alphabet and fails if any pattern matches a string the
// fast-path guard would have skipped. Adding a pass whose markers are missing
// from the guard makes this fail instead of silently shipping a dead feature.
func TestMdMarkersCoverEveryInlinePattern(t *testing.T) {
	for _, p := range mdInlinePasses {
		alphabet := append(reLiterals(p.re), 'a', 'b')
		seen := map[rune]bool{}
		uniq := alphabet[:0:0]
		for _, r := range alphabet {
			if !seen[r] {
				seen[r] = true
				uniq = append(uniq, r)
			}
		}
		var walk func(s string, depth int)
		walk = func(s string, depth int) {
			if t.Failed() {
				return
			}
			if s != "" && p.re.MatchString(s) && !strings.ContainsAny(s, mdMarkers) {
				t.Fatalf("pattern %s matches %q, but the mdInline guard %q skips that line",
					p.re, s, mdMarkers)
			}
			if depth == 0 {
				return
			}
			for _, r := range uniq {
				walk(s+string(r), depth-1)
			}
		}
		walk("", 6)
	}
}

const boldPassContract = "every theme viewkit ships already sets Accent.Bold, so dropping Bold(true) from " +
	"the bold pass changes nothing for them and the byte-for-byte tests cannot see it. A host is free to " +
	"register a theme whose Accent is not bold, and then **bold** must still render bold — this is the " +
	"only test that reaches that arm."

func TestMdInlineBoldsEvenWhenTheThemeAccentIsNot(t *testing.T) {
	t.Log(boldPassContract)
	prev := lipgloss.ColorProfile()
	t.Cleanup(func() { lipgloss.SetColorProfile(prev) })
	lipgloss.SetColorProfile(termenv.TrueColor)

	orig := *theme.Cur()
	t.Cleanup(func() { theme.Use(orig) })
	if !orig.Accent.GetBold() {
		t.Skip("the active theme's Accent is already non-bold, so the byte-for-byte tests cover this")
	}

	flat := lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00"))
	th := orig
	th.Accent = flat
	theme.Use(th)

	got := mdInline("**x**")
	if got == flat.Render("x") {
		t.Fatalf("mdInline(\"**x**\") = %q, which is the unbolded Accent render: %s", got, boldPassContract)
	}
	if want := flat.Bold(true).Render("x"); got != want {
		t.Errorf("mdInline(\"**x**\") = %q, want %q", got, want)
	}
}

func TestMdMaskBlanksEscapesInPlace(t *testing.T) {
	cases := []struct{ in, want string }{
		{"plain", "plain"},
		{"a\x1b[1mb", "a\x00\x00\x00\x00b"},
		{"a\x1b]0;title\x07b", "a\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00b"},
		{"a\x1b]0;t\x1b\\b", "a\x00\x00\x00\x00\x00\x00\x00b"},
		{"a\x1bZb", "a\x00\x00b"},
		{"trailing\x1b", "trailing\x00"},
		{"unterminated\x1b[38;2", "unterminated\x00\x00\x00\x00\x00\x00"},
	}
	for _, c := range cases {
		got := mdMask(c.in)
		if got != c.want {
			t.Errorf("mdMask(%q) = %q, want %q", c.in, got, c.want)
		}
		if len(got) != len(c.in) {
			t.Errorf("mdMask(%q) changed length %d -> %d (offsets would shift)", c.in, len(c.in), len(got))
		}
	}
}

func TestMarkdownRuleFitsTheBody(t *testing.T) {
	for _, width := range []int{20, 40, 81} {
		f := layout.NewFrame(width)
		for _, src := range []string{"---", "***", "___"} {
			out := Markdown(f, src)
			if strings.Contains(out, "\n") {
				t.Errorf("width %d: rule %q spans multiple lines: %q", width, src, stripANSI(out))
			}
			if w := ansi.StringWidth(out); w != f.BodyWidth() {
				t.Errorf("width %d: rule %q is %d wide, want the body width %d",
					width, src, w, f.BodyWidth())
			}
		}
	}
}
