---
name: viewkit-test
description: >-
  Test code that consumes viewkit. Use when writing Go tests that render a screen
  and assert on output, verify keymaps/hint legends, check theme/background, or
  sandbox a config file. Covers the render-and-grep idiom, color-profile pinning
  for deterministic ANSI, HOME-sandboxed config tests, per-test ui.Scope
  construction, and the layout.StrictScope hook.
---

# Testing viewkit consumers

viewkit renders to strings, so tests are mostly **render then assert on substrings**.
The only sharp edge left is ANSI determinism (pin the color profile) — theme and
keys are plain values, so there is no global state to restore.

## Scope per test — no restore, free `t.Parallel()`

Theme and keybindings travel in a `ui.Scope` (or a bare `theme.Theme`). A test
constructs what it needs and passes it; nothing leaks between tests:

```go
func TestThing(t *testing.T) {
    t.Parallel()
    th, _ := theme.Named("solarized-light")
    scope := &ui.Scope{Theme: th, Keys: keys.Default(), Glyphs: glyph.DefaultSet()}
    // hand scope to the model (deck.WithScope) or frame (layout.NewFrame(w).WithUI(scope))
    ...
}
```

To catch code that silently falls back to the built-in defaults, set
`layout.StrictScope = true` in `TestMain` — any `Frame` theme/scheme/glyph read
with a nil `UI` then panics instead of defaulting.

## Render-and-grep idiom

Force a known width/height, render, assert substrings. A tiny helper keeps tests
terse (reference: `internal/game/screen_hints_test.go`):

```go
func renderForHints(m Model) string {
    m.width = theme.MinScreenWidth   // ≥ min so it doesn't hit the TooNarrow path
    m.height = 80
    return m.View()
}

func TestFooterShowsHints(t *testing.T) {
    m := New(...)
    view := renderForHints(m)
    for _, want := range []string{"enter/space", "↑/↓/j/k", "esc/q"} {
        if !strings.Contains(view, want) {
            t.Fatalf("view missing %q:\n%s", want, view)
        }
    }
}
```

Asserting on the glyph strings is how you verify a **keymap is wired** without
simulating input — the footer legend is generated from the bindings.

## Exercise a specific screen in isolation

Construct the model, then assign the screen struct directly and drive
`handleKey`:

```go
m := New(...)
s := &mySettingsScreen{}
m.screen = s
s.handleKey(&m, tea.KeyMsg{Type: tea.KeyRight})   // simulate a keypress
```

## Deterministic ANSI: pin the color profile

To assert on raw escape codes (e.g. a background fill), pin lipgloss to TrueColor
and restore it (reference: `internal/game/theme_picker_test.go:59`):

```go
prev := lipgloss.ColorProfile()
lipgloss.SetColorProfile(termenv.TrueColor)
defer lipgloss.SetColorProfile(prev)

th, _ := theme.Named("solarized-light")
m := New(..., deck.WithScope(&ui.Scope{Theme: th, Keys: keys.Default(), Glyphs: glyph.DefaultSet()}))
if !strings.Contains(renderForHints(m), "48;2;253;246;227") { // the bg RGB
    t.Fatal("View() missing themed background fill")
}
```

Compare theme colors via `th.Accent.GetForeground()` on the `Theme` value under
test (`theme.Named(key)` to build it) rather than hardcoding hex.

## Sandbox config files with a temp HOME

If the code reads/writes `~/.<app>/…`, redirect HOME to a temp dir so tests don't
touch the real home and can assert on what was written:

```go
home := t.TempDir()
t.Setenv("HOME", home)
// ... drive the code that persists ...
data, err := os.ReadFile(filepath.Join(home, ".goose", "theme.json"))
```

## Verification

Run `go test ./...` from the repo root. Scoped theme/keys have no cross-test
state, so `-count=1 -shuffle=on` and `t.Parallel()` are safe — if shuffling
changes results, look for remaining process state (e.g. `glyph.SetMode` called
outside `TestMain`; the glyph default mode is deliberately write-once process
state).

Full API: see the `viewkit` skill's [references/api.md](../viewkit/references/api.md).
