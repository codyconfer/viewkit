---
name: viewkit-theme
description: >-
  Work with viewkit themes, palettes, colors, and the screen background. Use when
  touching theme.Theme values, theme.Default, theme.Palette / theme.New,
  theme.Named / theme.Keys / theme.DisplayName / theme.Register, Theme.Screen,
  ui.Scope, or the layout-contract constants (BodyWidth, MinScreenWidth, tiers).
  Covers the scoped-value model, adding a palette, painting the background, and
  per-test themes.
---

# viewkit theming

A theme is a **plain value you carry**, not process state. Build one (or take
`theme.Default()`), bundle it into a `ui.Scope`, and hand the scope to whatever
renders.

## Build once, pass explicitly

```go
th := theme.Default()                   // memoized built-in default
scope := &ui.Scope{Theme: th, Keys: keys.Default(), Glyphs: glyph.DefaultSet()}
// or, for all built-in defaults: scope := ui.Default()
```

Hand the scope to the root: `deck.WithScope(scope)` for a deck app, or
`layout.NewFrame(w).WithUI(scope)` for a bare frame. Read styles off the value:

```go
th.Dim.Render("quiet")                  // directly, when you hold a Theme
f.Theme().Accent.Render("loud")         // via a Frame's scope (defaults when UI is nil)
m.UI().Theme.Val.Render("42")           // via a deck Model's scope
```

Anti-pattern: stashing a `Theme` in a package-level var to avoid threading the
scope. Don't — pass the scope down.

## A Theme is built from a Palette

`theme.Palette` is 12 named color roles; `theme.New(p)` maps them to lipgloss
styles. Build a custom look by constructing a palette:

```go
p := theme.Palette{
    Accent:   lipgloss.Color("#6e9fff"),
    Border:   lipgloss.Color("#44474e"),
    Muted:    lipgloss.Color("#9c9fa3"),
    Text:     lipgloss.Color("#ececed"),
    Selected: lipgloss.Color("#ff9900"),
    Success:  lipgloss.Color("#6ccf8e"),
    Warning:  lipgloss.Color("#fbad37"),
    Failure:  lipgloss.Color("#ff5286"),
    Info:     lipgloss.Color("#6e9fff"),
    Series2:  lipgloss.Color("#d4a0ff"),
    Series3:  lipgloss.Color("#5ad2c8"),
    Bg:       lipgloss.Color("#1c1e26"),  // "" ⇒ no background paint
}
th := theme.New(p)
```

You can also start from a preset and override a few fields:

```go
th := theme.Default()
th.Accent = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
th.TooNarrowTitle = "SCREEN TOO SMALL"   // UI copy is overridable too
scope := &ui.Scope{Theme: th, Keys: keys.Default(), Glyphs: glyph.DefaultSet()}
```

## Named palettes

Built-in keys: `default` (Default), `solarized-dark`, `solarized-light`,
`one-dark-vivid`, `monokai`, `classic`, `retro-dark`, `retro-light`.

```go
names := theme.Keys()                    // []string of keys
t, ok := theme.Named("monokai")          // !ok ⇒ t is theme.Default()
label := theme.DisplayName("monokai")    // "Monokai"
```

Register a **new named** palette from outside the package with `theme.Register`.
It becomes resolvable by `theme.Named`, listed in `theme.Keys`, and titled by
`theme.DisplayName`. Re-registering an existing key overwrites it. For a one-off
look that needs no name, just `theme.New(p)`.

```go
theme.Register("dracula", "Dracula", theme.Palette{
    Accent: lipgloss.Color("#bd93f9"),
    // ...remaining roles
})
th, _ := theme.Named("dracula")
```

The registry is safe for concurrent use; registering at startup keeps the
`theme.Keys()` ordering predictable.

## Background

The background color is painted by the `Theme.Screen` method, not by a style you
apply yourself. Call it once at the top of your model's `View()`:

```go
th := m.UI().Theme
return th.Screen(body, m.width, m.height) // fills palette Bg across the screen
```

If the palette's `Bg` is `""`, it's a no-op (transparent).

## Layout-contract constants (not part of Theme)

Structural dimensions are exported **constants**, deliberately outside `Theme`:
`BodyWidth=81`, `MinBodyWidth=24`, `MinScreenWidth=80`, `MinBodyHeight=35`,
`TallBodyHeight=46`, `AppMarginX=2`, `AppMarginY=1`, `ScreenPaddingWidth`,
`RuleWidth`. Set per-view width with `layout.NewFrame(width)`; use the constants
for width/height guards (`layout.FitsScreenWidth`, tiers) rather than hardcoding.

## Persisting a theme choice

The consumer owns persistence. goose stores the chosen key in `~/.goose/theme.json`
and re-applies it on load (reference: `internal/game/theme_config.go`):

```go
th, _ := theme.Named(cfg.Theme)          // unknown key falls back to Default()
scope := &ui.Scope{Theme: th, Keys: keys.Default(), Glyphs: glyph.DefaultSet()}
```

To switch themes at runtime in a deck app, build a fresh scope and swap it —
scopes are immutable, so always swap, never mutate — and run the returned
relayout command:

```go
return m.SetScope(&ui.Scope{Theme: th, Keys: m.UI().Keys, Glyphs: m.UI().Glyphs})
```

## Tests just build values

There is no process state to install or restore. Construct the theme (or scope)
each test needs and pass it; such tests are order-independent and can run in
parallel:

```go
func TestSomething(t *testing.T) {
    t.Parallel()
    th, _ := theme.Named("solarized-light")
    scope := &ui.Scope{Theme: th, Keys: keys.Default(), Glyphs: glyph.DefaultSet()}
    // hand scope to deck.WithScope(...) or layout.NewFrame(w).WithUI(scope)
}
```

See **viewkit-test** for the full testing patterns (color-profile pinning, etc.).

## Verification

`go build ./...`; run the app and confirm colors/background apply and switching
themes (via `SetScope`) restyles everything. Run `go test ./...` from the repo
root.

Full API: see the `viewkit` skill's [references/api.md](../viewkit/references/api.md).
