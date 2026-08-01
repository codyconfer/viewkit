package forms

import (
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Suggester proposes completions for the token a text field is currently
// building. It receives that token — not the whole field — so a delimited
// field can complete each entry independently.
type Suggester func(prefix string) []string

// Static suggests from a fixed vocabulary.
func Static(vals ...string) Suggester {
	return SuggestFunc(func() []string { return vals })
}

// SuggestFunc suggests from a vocabulary read fresh on every keystroke, for
// values that change while the form is open.
func SuggestFunc(load func() []string) Suggester {
	if load == nil {
		return nil
	}
	return func(prefix string) []string {
		return Match(load(), prefix)
	}
}

// Match filters vals down to those extending prefix, case-insensitively. An
// exact match is dropped: there is nothing left to complete. Results are
// deduplicated and sorted so the candidate list does not reorder as the
// vocabulary is rebuilt.
func Match(vals []string, prefix string) []string {
	want := strings.ToLower(strings.TrimSpace(prefix))
	seen := make(map[string]bool, len(vals))
	out := make([]string, 0, len(vals))
	for _, v := range vals {
		if v == "" || seen[v] {
			continue
		}
		low := strings.ToLower(v)
		if low == want || !strings.HasPrefix(low, want) {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

func splitTail(text, delim string) (head, tail string) {
	if delim == "" {
		return "", text
	}
	i := strings.LastIndex(text, delim)
	if i < 0 {
		return "", text
	}
	return text[:i+len(delim)], strings.TrimLeft(text[i+len(delim):], " ")
}

func joinTail(head, pick, delim string) string {
	if head == "" {
		return pick
	}
	if delim != " " && !strings.HasSuffix(head, " ") {
		head += " "
	}
	return head + pick
}

func ghostOf(pick, tail string) string {
	if pick == "" {
		return ""
	}
	rest := pick
	for _, want := range tail {
		r, size := utf8.DecodeRuneInString(rest)
		if size == 0 || unicode.ToLower(r) != unicode.ToLower(want) {
			return ""
		}
		rest = rest[size:]
	}
	return rest
}
