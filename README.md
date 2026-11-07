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
| `console` | Output capability detection: TTY, ANSI, Unicode, Powerline/Nerd Fonts |
| `outputbox` | Borderless, status-shaded command output blocks for scrollback |
| `keys` | Keybinding tables |
| `forms` | Field builders |
| `list` / `browser` | List/browser helpers |
| `notify` | Notifications (`glyph.Severity`) + TTL queue |
| `timefmt` | Time formatting |
| `term` | Terminal launcher helpers |
| `deck` | Tea `Model` / screens (`Menu`, `Scroll`, `ItemList`, `HomeShell`, `Work`) — **only** package that imports tea |

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

Rendering context is scoped, not global: bundle a theme, key scheme, and glyph
set into a `ui.Scope` and hand it to frames/views. The only process state left
is glyph's write-once default mode (`glyph.SetMode`), by design.

```go
th := theme.Default()
th.Accent = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))

scope := ui.Default()
scope.Theme = th

frame := layout.NewFrame(80).WithUI(scope)
```

Structural dimensions (`theme.BodyWidth`, …) are exported constants — set
per-view width via `layout.NewFrame(width)`.

### Console capabilities

```go
caps := console.Detect(os.Stdout)
if caps.Powerline == console.Supported {
    // Use Powerline separators; otherwise use a plain-text fallback.
}
glyph.DetectMode(os.Stdout, os.Getenv("MYAPP_ICONS"))
```

Import `github.com/codyconfer/viewkit/console` (and `glyph` for the mode helper).
Detection is passive: it never prints probes or reads input. Non-TTY and
`TERM=dumb` output falls back to ASCII. Unicode is estimated from the first
non-empty locale variable (`LC_ALL`, `LC_CTYPE`, `LANG`) containing UTF-8.
Terminal names and character widths cannot prove that a font contains a glyph,
so font support defaults to `console.Unknown`, not a claim of support.

Set `VIEWKIT_POWERLINE=true` and/or `VIEWKIT_NERD_FONT=true` to declare installed
font support, or `false` to disable it; unset/`auto` keeps it unknown. Nerd Fonts
imply Powerline support unless explicitly disabled; Powerline alone does not
imply Nerd Font icons. Font overrides do not enable fancy output in pipes.

`glyph.DetectMode` now chooses Unicode for UTF-8 terminals with unknown fonts,
rather than assuming Nerd Fonts. Its explicit `nerd`/`unicode`/`ascii` override
still wins, including for redirected output. `glyph.SetMode` and the default
mode for applications that never call `DetectMode` are unchanged. Call detection
before constructing a `ui.Scope` so it captures the selected glyph mode.

### Deck Model + scope

Interactive apps use `deck.Model` as the session tea root. Pass the scope with
`deck.WithScope` (swap at runtime via `Model.SetScope`) and register named
themes/keys/views before `deck.Run`. Full contract:
[`deck/INTERFACE.md`](deck/INTERFACE.md).

### Command output boxes

`outputbox.Render` renders borderless command output blocks with full-width
status backgrounds: neutral running, green success, yellow review, red failure.
Explicit status labels remain readable with color disabled. Rendering is pure:
append each returned box to keep earlier actions in scrollback; no cursor moves,
terminal detection, or global renderer changes are performed.

```go
import "github.com/codyconfer/viewkit/outputbox"

box := outputbox.Render("go test ./...", []string{"ok example/package"},
    80, true, outputbox.Success)
fmt.Fprintln(os.Stdout, box)
```

Width includes padding and clamps to at least eight cells. Long text wraps without
truncation, including wide Unicode. Pass the color capability resolved by your
application (for example from `console.Detect`); use `false` for plain transcripts.
Content may include trusted ANSI styles; sanitize untrusted terminal controls
before rendering. The package has no Bubble Tea dependency.

### Inline CLI workflows

`deck.Inline` provides sequential setup/onboarding panels without entering a
full-screen deck session. Each panel replaces the previous panel on an ANSI
terminal; redirected streams, `TERM=dumb`, and `Plain: true` keep an append-only
transcript. No step numbers or application-specific checks are imposed.

```go
func setup(ctx context.Context) error {
    flow := deck.NewInline(deck.InlineSpec{
        Input: os.Stdin, Output: os.Stdout, UI: ui.Default(),
    })
    if err := flow.Panel("Workspace", "Choose a local directory."); err != nil {
        return err
    }
    path, err := flow.Ask(ctx, "Workspace", "Directory", ".")
    if err != nil {
        return flow.Report(err) // stays inside the current panel
    }
    if path == "" {
        return flow.Report(errors.New("a directory is required"))
    }
    return flow.Panel("Ready", "Directory: "+path)
}
```

A matching `Ask` retains the preceding panel's instructions alongside its text
field and key hints. Defaults can be edited with typing, backspace, and clear;
scoped key bindings drive both dispatch and hints. Cancel returns
`deck.ErrCancelled`. Text responses are visible, not password prompts. Caller
strings are treated as text, with terminal controls stripped; redact secrets
before reporting errors.

`Report` preserves error identity (`errors.Is`/`errors.As`) and marks successfully
rendered errors. At the CLI boundary, use `deck.ErrorReported(err)` to suppress
duplicate stderr output **without** suppressing a failing exit status. A render
failure returns both errors and does not mark them reported.

Call `flow.Release()` before any external command or other output. Pass
`flow.Reader()` to a child command's stdin: it preserves scripted bytes already
buffered by an earlier `Ask`. Never write concurrently or nest an `Inline` inside
a running deck. A width change preserves the previous view as transcript rather
than risking erasure of unrelated reflowed scrollback. Interactive prompts honor
context cancellation; generic line readers can only check it before/after reads.
See the executable `ExampleInline` and terminal tests in `deck/inline*_test.go`.

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
