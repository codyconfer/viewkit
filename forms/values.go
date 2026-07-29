package forms

import (
	"strconv"
	"strings"
)

// Str returns the string value stored under key, with surrounding whitespace
// trimmed. Missing keys and non-string values yield "".
func Str(vals map[string]any, key string) string {
	s, _ := vals[key].(string)
	return strings.TrimSpace(s)
}

// Raw returns the string value stored under key verbatim. Seeding a rebuilt
// form uses it rather than Str, so a trailing delimiter or space the user
// deliberately typed survives the rebuild.
func Raw(vals map[string]any, key string) string {
	s, _ := vals[key].(string)
	return s
}

// Bool returns the boolean value stored under key. Missing keys and non-bool
// values yield false.
func Bool(vals map[string]any, key string) bool {
	b, _ := vals[key].(bool)
	return b
}

// Int returns the value stored under key as an integer. Text fields hold
// strings, so the common case parses base-10 text, but already-numeric values
// are accepted too. Missing keys and unparsable text yield 0.
func Int(vals map[string]any, key string) int {
	switch v := vals[key].(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	}
	n, err := strconv.Atoi(Str(vals, key))
	if err != nil {
		return 0
	}
	return n
}

// Strings returns the multi-select value stored under key. Missing keys and
// values of other types yield nil.
func Strings(vals map[string]any, key string) []string {
	ss, _ := vals[key].([]string)
	return ss
}

// SelectIndex returns the position of want in options, for seeding
// Field.Selected. Unknown and empty values select the first option.
func SelectIndex(options []string, want string) int {
	if want == "" {
		return 0
	}
	for i, opt := range options {
		if opt == want {
			return i
		}
	}
	return 0
}

// SelectFirst returns options reordered so cur leads, for select fields that
// should open on their current value. The relative order of the remaining
// options is preserved, and duplicates of cur beyond the first are kept. The
// result is always a fresh slice; options is never mutated.
func SelectFirst(options []string, cur string) []string {
	out := make([]string, 0, len(options))
	found := false
	for _, opt := range options {
		if opt == cur && !found {
			found = true
			continue
		}
		out = append(out, opt)
	}
	if !found {
		return out
	}
	return append([]string{cur}, out...)
}
