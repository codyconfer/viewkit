# viewkit/deck interface

**Invariant:** `bubbletea` only in this module; viewkit core stays tea-free.

## Why a nested module

Core viewkit must stay importable by non-TUI tools and by sisyphus-adjacent code
paths without pulling a tea runtime. Deck is the only place that owns
`tea.NewProgram`.

Versioning: tag `deck/vX.Y.Z` in lockstep with core `vX.Y.Z` on the same commit
when publishing. Local/dev uses uncommitted `go.work` (no committed `replace`).

## Surfaces

| API | Role |
|---|---|
| `View` | Navigable screen: `Title/Init/Update/Body/Hints/Context` |
| `Model` (`Host` alias) | Stateful tea root: stack nav + chrome (brand/subtitle injected; no app literals) |
| `RegisterView` / `LookupView` | View registry (plugin views) |
| `RegisterComponent` / `LookupComponent` | Fragment registry |
| `Content` / `Text` | Domain-agnostic flight body (apps adapt before crossing) |
| `Task` + `Execute` / `RunFlight` | Unified flight: errgroup headless + tea progressive UI |
| `Confirm` | Yes/no tea prompt |
| `Menu` / `Message` / `Scroll` | Generic views (optional; apps may roll their own) |
| `ItemList` | Lazy Fetch+Bind selectable list (domain → `list.Item` in Bind) |
| `HomeShell` | Dual-pane menu + optional side list (Fetch+Bind; no app domain types) |

## Model + singleton contract

**Stateful session model:** `deck.Model` (compatibility alias `deck.Host`) is the
only tea.Model for a deck session. Views implement `Update(m *Model, msg)` and
navigate with `m.Push` / `m.Pop`. Domain/plugin state stays in the View (or an
app kit), never inside Model fields beyond chrome/stack/size/status.

**Process-global singletons** (install once before `deck.Run`):

| Singleton | API | Notes |
|---|---|---|
| Theme | `theme.Use` / `theme.Cur` | Active palette; `theme.Register` for named palettes |
| Keys | `keys.Use` / `keys.Cur` | Active scheme; `keys.Register` for named schemes |
| Views | `deck.RegisterView` / `LookupView` | Plugin/app screen constructors |
| Glyphs | `glyph.Register` | Nerd/Uni/ASCII variants (core) |

Overlays and plugins must not invent a second tea root — register Views/themes
and let the host `Model` own the program.

## Chrome contract

Apps inject branding via `WithChrome(Chrome{Brand, BrandGlyph, Subtitle, ClockGlyph})`.
Status strip via `WithStatus(StatusFunc)` returning `StatusInfo{Identity, Services}`.
Deck never hard-codes product names.

## Dual-host panels (core)

Panels that work in both inline shells and deck live in `viewkit/panels` as
`DualHost` (`RenderInline` / `RenderDeck`) — no tea. Deck bodies call
`panels.Render(..., panels.Deck, ...)`. See `panels/host.go`.

## Content boundary

Deck must not import app domain types (e.g. munin `signals.Section`). Flight
tasks return `Content`; apps render domain → string/`Content` before
`RunFlight` / `Execute`.

## Key bindings

Host quit matching is injectable (`WithQuitCheck`). Generic views use
`keys.Cur()`. Apps may install a scheme (`keys.Use` / `keys.Register`) before
`deck.Run`.

## Consumer checklist

- [ ] View / Host surface stable enough for a second consumer
- [ ] Chrome injection covers branding without forks
- [ ] Content boundary keeps domain types out of deck
- [ ] DualHost panels usable from inline shells
- [ ] No tea leakage into viewkit core imports

## Package consolidation

Merging small core packages (`browser`, `timefmt`, …) into fewer import paths
needs a re-export / compatibility plan and a tagged bump. Do not fold them
casually — document intent here until a dedicated change lands.
