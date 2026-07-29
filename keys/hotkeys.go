package keys

import "strings"

// Normalize canonicalizes a key name for comparison, folding case and trimming
// surrounding whitespace.
func Normalize(key string) string {
	return strings.ToLower(strings.TrimSpace(key))
}

// Resolve looks key up in a user-supplied hotkey table and returns the target
// it is bound to. Lookup is case- and whitespace-insensitive on both sides, so
// a table written by hand still matches. Keys bound to a blank target report
// ok=false.
func Resolve(binds map[string]string, key string) (target string, ok bool) {
	if len(binds) == 0 {
		return "", false
	}
	want := Normalize(key)
	if want == "" {
		return "", false
	}
	if t, hit := binds[want]; hit {
		t = strings.TrimSpace(t)
		return t, t != ""
	}
	for k, t := range binds {
		if Normalize(k) == want {
			t = strings.TrimSpace(t)
			return t, t != ""
		}
	}
	return "", false
}

// ControlKeys filters a binding's keys down to the multi-rune ones ("tab",
// "ctrl+s", "esc"), dropping single characters. Text-entry contexts use it so
// ordinary typing is not swallowed by navigation bindings.
func ControlKeys(in []string) []string {
	out := make([]string, 0, len(in))
	for _, k := range in {
		if len([]rune(k)) > 1 {
			out = append(out, k)
		}
	}
	return out
}

// EditorBindings resolves actions against s and strips their single-character
// keys, yielding the subset safe to bind while a text field has focus.
// Actions left with no keys are omitted.
func (s Scheme) EditorBindings(actions ...Action) []Binding {
	out := make([]Binding, 0, len(actions))
	for _, a := range actions {
		b := s.Binding(a)
		b.Keys = ControlKeys(b.Keys)
		if len(b.Keys) > 0 {
			out = append(out, b)
		}
	}
	return out
}
