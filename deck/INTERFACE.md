# viewkit/deck interface

**Invariant:** `bubbletea` only in this package; viewkit core stays tea-free
(no tea imports outside `deck/`). Deck is part of the single `viewkit` module.

## Surfaces

| API | Role |
|---|---|
| `View` | Navigable screen: `Title/Init/Update/Body/Hints/Context` |
| `Model` | Stateful tea root: stack nav + chrome (brand/subtitle injected; no app literals) |
| `RegisterView` / `NamedView` | View registry (plugin views) |
| `RegisterComponent` / `NamedComponent` | Fragment registry |
| `Content` / `Text` | Domain-agnostic job body (apps adapt before crossing) |
| `Job` + `Work.Collect` / `Work.RunInteractive` | A set of concurrent jobs: errgroup headless + tea progressive UI |
| `Confirm` | Yes/no tea prompt |
| `Menu` / `Message` / `Scroll` | Generic views (optional; apps may roll their own) |
| `ItemList` | Lazy Fetch+Bind selectable list (domain → `list.Item` in Bind) |
| `HomeShell` | Dual-pane menu + optional side list (Fetch+Bind; no app domain types) |

## Model + singleton contract

**Stateful session model:** `deck.Model` is the
only tea.Model for a deck session. Views implement `Update(m *Model, msg)` and
navigate with `m.Push` / `m.Pop`. Domain/plugin state stays in the View (or an
app kit), never inside Model fields beyond chrome/stack/size/status.

**Process-global singletons** (install once before `deck.Run`):

| Singleton | API | Notes |
|---|---|---|
| Theme | `theme.Use` / `theme.Cur` | Active palette; `theme.Register` for named palettes |
| Keys | `keys.Use` / `keys.Cur` | Active scheme; `keys.Register` for named schemes |
| Views | `deck.RegisterView` / `NamedView` | Plugin/app screen constructors |
| Glyphs | `glyph.Register` | Nerd/Uni/ASCII variants (core) |

Overlays and plugins must not invent a second tea root — register Views/themes
and let the host `Model` own the program.

## Chrome contract

Apps inject branding via `WithChrome(Chrome{Brand, BrandGlyph, Subtitle, ClockGlyph})`.
Status strip via `WithStatus(StatusFunc)` returning `StatusInfo{Identity, Services}`.
Call `Model.RefreshStatus()` to reload immediately (same path as the 60s ticker).
Deck never hard-codes product names.

## Dual-host panels (core)

Panels that work in both inline shells and deck live in `viewkit/panels` as
`DualHost` (`RenderInline` / `RenderDeck`) — no tea. Deck bodies call
`panels.Render(..., panels.TargetDeck, ...)`. See `panels/host.go`.

## Content boundary

Deck must not import app domain types (e.g. an app's `signals.Section`). Jobs
return `Content`; apps render domain → string/`Content` before
`Work.RunInteractive` / `Work.Collect`.

## Key bindings

Apps may install a scheme (`keys.Use` / `keys.Register`) before `deck.Run`.

**Single source:** generic views resolve input *and* render their footer legend
through the active scheme (`navMap` over `keys.Cur()`). No view matches raw key
literals, so a hint can never advertise a binding the view does not honour.
Paging is `keys.PageUp`/`PageDown`; pane switching is `keys.FocusNext`/`FocusPrev`.
`Scroll` drives the viewport from the scheme rather than the bubbles keymap for
the same reason.

Model quit matching defaults to the scheme's `keys.Quit` binding, so behaviour and
the `quit` legend share one source. `WithQuitCheck` takes an opaque matcher that
cannot be rendered — pair it with `WithQuitHint` or the legend will disagree with
the matcher. Global app hotkeys use `WithKeyHook` (runs after the quit check,
before the top view).

A view's `IsAction` hook *replaces* the scheme map rather than extending it: map
`PageUp`/`PageDown`/`FocusNext` too if paging and pane switching should survive.

## Consumer checklist

- [ ] View / Model surface stable enough for a second consumer
- [ ] Chrome injection covers branding without forks
- [ ] Content boundary keeps domain types out of deck
- [ ] DualHost panels usable from inline shells
- [ ] No tea leakage into viewkit core imports
- [ ] No raw key literals in views; hints derived from `keys.Cur()`

## Package consolidation

Merging small core packages (`browser`, `timefmt`, …) into fewer import paths
needs a re-export / compatibility plan and a tagged bump. Do not fold them
casually — document intent here until a dedicated change lands.
