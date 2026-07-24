# viewkit/deck interface (goose review notes)

**Status:** M5 draft for consumer review (goose + munin)  
**ADR:** ADR-2 — deck → nested module after genericization  
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
| `Host` | Stack nav + chrome (brand/subtitle injected; no app literals) |
| `RegisterView` / `LookupView` | View registry (plugin views) |
| `RegisterComponent` / `LookupComponent` | Fragment registry |
| `Content` / `Text` | Domain-agnostic flight body (apps adapt before crossing) |
| `Task` + `Execute` / `RunFlight` | Unified flight: errgroup headless + tea progressive UI |
| `Confirm` | Yes/no tea prompt |
| `Menu` / `Message` / `Scroll` | Generic views (optional; apps may roll their own) |
| `ItemList` | Lazy Fetch+Bind selectable list (domain → `list.Item` in Bind) |
| `HomeShell` | Dual-pane menu + optional side list (Fetch+Bind; no app domain types) |

## Chrome contract

Apps inject branding via `WithChrome(Chrome{Brand, BrandGlyph, Subtitle, ClockGlyph})`.
Status strip via `WithStatus(StatusFunc)` returning `StatusInfo{Identity, Services}`.
Deck never hard-codes product names.

## Dual-host panels (core)

Panels that work in both inline shells and deck live in `viewkit/panels` as
`DualHost` (`RenderInline` / `RenderDeck`) — no tea. Deck bodies call
`panels.Render(..., panels.Deck, ...)`. See `panels/host.go`.

## Content boundary (D13)

Deck must not import app domain types (e.g. munin `signals.Section`). Flight
tasks return `Content`; munin (or goose) renders domain → string/`Content`
before `RunFlight` / `Execute`.

## Key bindings

Host quit matching is injectable (`WithQuitCheck`). Generic views use
`keys.Cur()`. Apps may install a scheme (`keys.Use` / `keys.Register`) before
`deck.Run`.

## Goose review checklist

- [ ] View / Host surface stable enough for a second consumer
- [ ] Chrome injection covers goose branding without forks
- [ ] Content boundary keeps game/domain types out of deck
- [ ] DualHost panels usable from goose inline shells
- [ ] No tea leakage into viewkit core imports

*(Goose repo may be absent during M5; this doc is the review artifact.)*


## Package consolidation (deferred)

Merging small core packages (`browser`, `timefmt`, …) into fewer import paths is
**deferred** past M7. Call sites already import them; a real merge needs a
re-export / compatibility plan and a tagged bump. Do not fold them casually —
document intent here until a dedicated slice lands.
