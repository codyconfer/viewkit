# viewkit

[![GitHub release](https://img.shields.io/github/v/tag/codyconfer/viewkit?logo=github&label=latest)](https://github.com/codyconfer/viewkit/tags)
[![CI](https://github.com/codyconfer/viewkit/actions/workflows/ci.yml/badge.svg)](https://github.com/codyconfer/viewkit/actions/workflows/ci.yml)

A small toolkit for building **terminal UIs** with
[Lip Gloss](https://github.com/charmbracelet/lipgloss). Core packages stay
**Bubble Tea-free** (no tea imports); interactive hosts live in the **`deck`**
package ([Bubble Tea](https://github.com/charmbracelet/bubbletea) allowed there
only — see [`deck/INTERFACE.md`](deck/INTERFACE.md)). One module version covers
core + deck.

## Packages

| Package | Role |
|---|---|
| `layout` | `Frame`, panels/sections, sticky footer, scroll, focus `Ring` |
| `panels` | Charts/widgets from neutral structs (`Bar`, `Line`, `Meter`, …) |
| `theme` | `Theme` + `Use` / `Cur` palettes and status helpers |
| `glyph` | Nerd/Uni/ASCII variants, status strip, severity vocabulary |
| `keys` | Keybinding tables |
| `forms` | Field builders |
| `list` / `browser` | List/browser helpers |
| `notify` | Notification tone helpers |
| `timefmt` | Time formatting |
| `term` | Terminal launcher helpers |
| `deck` | Tea `Model` (+ `Host` alias) / screens (`Menu`, `Scroll`, `ItemList`, `HomeShell`, flight) — **only** package that imports tea |

Longer API notes: [`skills/viewkit/references/api.md`](skills/viewkit/references/api.md).

Core packages depend on `charmbracelet/lipgloss`, `charmbracelet/x/ansi`, and the
standard library. Typical flow: `panels → layout → theme`. Deck adds Bubble Tea.

## Install

```sh
go get github.com/codyconfer/viewkit@latest
```

```go
import (
    "github.com/codyconfer/viewkit/deck"
    "github.com/codyconfer/viewkit/layout"
    "github.com/codyconfer/viewkit/panels"
    "github.com/codyconfer/viewkit/theme"
)
```

## Design contract

viewkit is domain-agnostic: data crosses the boundary as neutral structs
(`panels.Datum`, `panels.OHLC`, `panels.LedgerRow`) and formatter callbacks —
never application domain types. A `layout.Frame` carries render width and focus;
construct with `layout.NewFrame(width)`.

```go
frame := layout.NewFrame(80)

body := frame.Panel("STATUS", frame.Row("tokens", "1.2M"))

chart := panels.Bar(frame, "GPUs", []panels.Datum{
    {Label: "gpu", Value: 12},
    {Label: "cloud", Value: 30},
}, 40, fmtNum, "no data")
```

### Theming

```go
th := theme.Default()
th.Accent = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
theme.Use(th)
```

Structural dimensions (`theme.BodyWidth`, …) are exported constants — set
per-view width via `layout.NewFrame(width)`.

### Deck Model + singletons

Interactive apps use `deck.Model` (alias `deck.Host`) as the session tea root.
Install process-global theme/keys (and optional `RegisterView`) before
`deck.Run`. Full contract: [`deck/INTERFACE.md`](deck/INTERFACE.md).

## Status

API may still shift before a published `v1`.

## Development

```sh
make build          # go build ./...
make check          # build + fmt-check + lint + govulncheck + test (CI gate is `make ci`)
make test           # go test ./...
```

Linters live in the nested `tools/` module (`go tool -modfile=tools/go.mod`).

### Local multi-repo development (`go.work`)

When editing viewkit alongside munin/sisyphus, use an **uncommitted** `go.work`
in the consumer that `use`s sibling checkouts (e.g. `../viewkit`).
Do not commit `go.work` / `go.work.sum` and do not add committed `replace`
directives — CI builds against tagged pins.

`deck` used to be a nested module (`deck/go.mod`). It is now a normal package
in this module. Consumers that previously required `github.com/codyconfer/viewkit/deck`
should require only `github.com/codyconfer/viewkit` and exclude any published
nested `viewkit/deck` versions (Go prefers the longer module path otherwise).

## License

[MIT](LICENSE) © Cody Confer
