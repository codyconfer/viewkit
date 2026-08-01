# viewkit public API cheat-sheet

Signatures for the exported surface, grouped by package. Keep this in sync with
the source; the `_test.go` files in each package are the best runnable examples.

## layout — `github.com/codyconfer/viewkit/layout`

### Frame (the render context)
```go
type Frame struct { Width, Height int; Focused bool; UI *ui.Scope }

func NewFrame(width int) Frame        // clamps to [theme.MinBodyWidth, …]
func DefaultFrame() Frame             // NewFrame(theme.BodyWidth)
func (f Frame) WithUI(scope *ui.Scope) Frame  // carry the rendering scope
func (f Frame) WithWidth(width int) Frame     // keeps the scope
func (f Frame) Screen() Frame
func (f Frame) Focus() Frame          // returns copy with Focused=true
func (f Frame) WithHeight(h int) Frame
func (f Frame) BodyWidth() int

// Scope accessors — built-in defaults when UI is nil:
func (f Frame) Theme() theme.Theme
func (f Frame) Scheme() keys.Scheme
func (f Frame) Glyphs() glyph.Set
var StrictScope bool                  // test hook: nil-UI reads panic when true
```

### Structural helpers
```go
func (f Frame) Header(title string, detail ...string) string  // title + rule
func (f Frame) Rule() string
func (f Frame) Box(lines ...string) string                    // bordered; focus-aware
func (f Frame) Panel(title string, lines ...string) string    // titled Box
func (f Frame) Row(label, value string) string                // dim label ⟷ value
func (f Frame) Spread(left, right string) string              // left …… right
func (f Frame) Fit(s string) string                           // truncate to width
func (f Frame) Selectable(label string, selected bool) string // "▸ " cursor + label
func (f Frame) HintLine(pairs ...keys.Hint) string            // wrapping key legend
func (f Frame) Cursor(selected bool) string
func (f Frame) CellBox(title string, lines ...string) string
func (f Frame) CellPanel(title string, lines []string, offset int) string

func Stack(sections ...string) string       // join non-empty with "\n\n"
func StackTight(sections ...string) string  // join non-empty with "\n"
func SpreadBGIn(th theme.Theme, bg lipgloss.TerminalColor, left, right string, width int) string
```

### Tiers (responsive height)
```go
type Tier int
const ( TierShort Tier = iota; TierMedium; TierTall )
type Section struct { Content string; MinTier Tier }

func StackFit(tier Tier, sections ...Section) string       // drop sections above tier
func StackTightFit(tier Tier, sections ...Section) string
func TierForHeight(height int) Tier
func BodyBudget(height int) int
func ContentRows(height int) int

type TierRows struct { Short, Medium, Tall int }
func (r TierRows) At(t Tier) int
```

### Panes / layouts / screen
```go
type Pane struct {
    Name, Title string
    Group       string      // section label; only the "sections" layout uses it
    Interactive bool        // joins the focus ring
    MinTier     Tier        // hidden when terminal too short
    Pos         *GridPos    // only for Grid layout
    Slim        bool
    Render      func(Frame) string
}

type Layout interface { Arrange(f Frame, tier Tier, panes []Pane, focusedName string) string }

type Screen struct { Layout Layout; Panes []Pane }
func (s Screen) Ring() Ring
func (s Screen) Render(f Frame, tier Tier, focus int) string

type SingleColumn struct{}                       // stacks panes vertically
type Grid struct { Cols, Rows int }              // explicit tiled grid
type GridPos struct { Col, Row, ColSpan, RowSpan int }
type FlexColumns struct { MinWidth, MaxCols int }// responsive column masonry
type FlexRows struct { MinWidth, MaxCols int }   // responsive row flow
type FlexSections struct { MinWidth, MaxCols int }// panes grouped by Pane.Group under labeled headers, each group flows as flex-columns
func FlexColCount(width, minWidth, maxCols int) int
const ( DefaultFlexMinWidth = 40; DefaultFlexMaxCols = 4 )
```

### Focus ring
```go
type Focusable struct { Name string; Interactive bool }
type Ring []string
func NewRing(all ...Focusable) Ring   // filters to interactive
func PaneRing(panes []Pane) Ring
func (r Ring) At(idx int) string
func (r Ring) Step(idx, delta int) int
```

### Data-driven registry + spec
```go
type PaneFactory[Ctx any] func(ctx Ctx) (Pane, bool)   // bool = VISIBILITY, not success
type LayoutFactory func(params Params) (Layout, error)
type PaneInfo struct { Key, Title string }

type Registry[Ctx any]
func NewRegistry[Ctx any]() *Registry[Ctx]  // pre-registers: single, flex-columns, flex-rows, grid, sections
func (r *Registry[Ctx]) Pane(key, title string, f PaneFactory[Ctx]) *Registry[Ctx]
func (r *Registry[Ctx]) LayoutFn(key string, f LayoutFactory) *Registry[Ctx]
func (r *Registry[Ctx]) PaneKeys() []PaneInfo
func (r *Registry[Ctx]) LayoutKeys() []string

type Params map[string]any
func (p Params) Int(key string, def int) int

type PaneRef struct { Key string; Pos *GridPos; MinTier *Tier; Slim bool } // json-tagged
type ScreenSpec struct { Layout string; LayoutParams Params; Panes []PaneRef } // json-tagged
func BuildScreen[Ctx any](s ScreenSpec, ctx Ctx, r *Registry[Ctx]) (Screen, error)
```

### Scrolling / viewport / screen guards
```go
type ScrollState struct { /* Offset etc. */ }
func (s *ScrollState) Scroll(delta, total, rows int)
func (s *ScrollState) Reveal(index, total, rows int)

func ScrollPanel(title string, lines []string, rows, offset int) string
func (f Frame) ScrollPanel(title string, lines []string, rows, offset int) string
func (f Frame) ScrollPanelWithPrefix(title string, prefix, lines []string, rows, offset int) string
func (f Frame) Viewport(body string, rows, offset int) string // hint styled with f.Theme()
func ViewportContentRows(rows int) int
func ViewportLayout(body string, rows, offset int) string     // hints use the built-in default theme
func ScrollableBody(body string, rows int) string
func SplitStickyFooter(body string) (content, footer string)
func PadLines(body string, rows int) string
func CountLines(s string) int

func FitsScreenWidth(screenWidth int) bool
func ScreenFrame(screenWidth int) Frame
func FrameFor(w io.Writer) Frame
func TooNarrowIn(t theme.Theme, screenWidth int) string

type OverlayPos struct { XFrac, YFrac float64 }
var Center OverlayPos
func Overlay(bg, fg string, pos ...OverlayPos) string
func FitBlock(block string, w, h int) string
```

## panels — `github.com/codyconfer/viewkit/panels`
```go
type Datum struct { Label string; Value float64 }
type OHLC struct { Open, High, Low, Close float64 }
type LedgerRow struct { Label string; Delta float64 }
type ClockOpts struct { TwentyFour, HideSeconds, ShowDate bool }
type SpectrumOpts struct { Peaks []float64; BarGap, BarWide int }

func Bar(f layout.Frame, title string, data []Datum, width int, fmtNum func(float64) string, empty string) string
func BarScroll(f layout.Frame, title string, data []Datum, width int, fmtNum func(float64) string, empty string, visible, offset int) string
func Line(f layout.Frame, title string, series []float64, width, height int, fmtVal func(float64) string, footer ...string) string
func Candle(f layout.Frame, title string, candles []OHLC, width, height int, fmtVal func(float64) string, footer ...string) string
func Pie(f layout.Frame, title string, data []Datum, barWidth int, fmtNum func(float64) string, empty string) string
func Spectrum(f layout.Frame, title string, levels []float64, height int, empty string, opts ...SpectrumOpts) string // vertical EQ; levels/peaks in [0,1]
func Ledger(f layout.Frame, title string, rows []LedgerRow, unit string, fmtNum func(float64) string, visible, offset int, empty string) string
func Markdown(f layout.Frame, src string) string
func MarkdownPanel(f layout.Frame, title, src string) string
func Clock(f layout.Frame, title string, t time.Time, opts ...ClockOpts) string
func BinaryClock(f layout.Frame, title string, t time.Time) string

// Matrix rain — stateful; Beat() once per tick, then render.
type Rain struct{ /* opaque grid + seeded RNG */ }
func NewRain(width, rows int, seed int64) *Rain
func (r *Rain) Resize(width, rows int)
func (r *Rain) Beat()
func Matrix(f layout.Frame, title string, r *Rain) string

func ProgressBar(frac float64, width int) string
func Meter(frac float64, width int) string
func MeterWidth(frameWidth, desired int) int
func Flash(message string) string
func Toggle(left, right string, leftActive bool) string

func ClampIndex(index, total int) int
func MoveIndex(index, delta, total int) int   // clamps at ends
func StepIndex(index, delta, total int) int   // wraps around

func NotificationToast(f layout.Frame, n notify.Notification) string
func NotificationPanel(f layout.Frame, title string, ns []notify.Notification) string
func NotificationCard(f layout.Frame, n notify.Notification) string
func NotificationOverlay(bg string, f layout.Frame, n notify.Notification, pos ...layout.OverlayPos) string
```

## ui — `github.com/codyconfer/viewkit/ui`
```go
// Scope is one program's rendering context — immutable after construction;
// swap the pointer to change appearance. No process state to install.
type Scope struct { Theme theme.Theme; Keys keys.Scheme; Glyphs glyph.Set }
func Default() *Scope   // built-in default theme, scheme, and glyph set

func (s *Scope) Success(msg string) string  // success-colored check + msg
func (s *Scope) Bullet(msg string) string   // accent-colored bullet + msg
```

## theme — `github.com/codyconfer/viewkit/theme`
```go
type Palette struct { Accent, Border, Muted, Text, Selected, Success, Warning, Failure, Info, Series2, Series3, Bg lipgloss.Color }
type Theme struct { Title, Accent, Dim, Val, Key, Can, Cant lipgloss.Style
                    Panel, PanelFocus, PanelTitle, Card, AppFrame lipgloss.Style
                    NotifPositive, NotifNeutral, NotifWarning, NotifNegative, NotifIdle, NotifTitle lipgloss.Style
                    Series []lipgloss.Style; Bg lipgloss.Color
                    TooNarrowTitle, TooNarrowNeed, TooNarrowBody string }

func New(p Palette) Theme
func Default() Theme          // memoized built-in default

func Keys() []string                       // named-palette keys
func Named(key string) (Theme, bool)       // false ⇒ returns Default()
func DisplayName(key string) string
func Register(key, name string, p Palette) // add/replace a named palette; concurrency-safe

func (t Theme) Screen(body string, width, height int) string  // paints background
func (t Theme) SeverityStyle(s glyph.Severity) lipgloss.Style
func (t Theme) SeverityColor(s glyph.Severity) lipgloss.TerminalColor
func (t Theme) StripBg() lipgloss.TerminalColor
func (t Theme) StripText(fg lipgloss.TerminalColor, s string) string
func (t Theme) StripBold(fg lipgloss.TerminalColor, s string) string
func (t Theme) StripBlock(width int, lines ...string) string
func (t Theme) Icon(icon string, hue int) string
func IconBlank() string

func AppMargin(body string) string
func FillLine(c lipgloss.TerminalColor, width int) string
func PadBlock(c lipgloss.TerminalColor, width, rows int, lines ...string) string

// layout-contract constants (NOT part of Theme):
// BodyWidth=81, MinBodyWidth=24, MinScreenWidth=80, MinBodyHeight=35,
// TallBodyHeight=46, AppMarginX=2, AppMarginY=1, ScreenPaddingWidth, RuleWidth
```

## keys — `github.com/codyconfer/viewkit/keys`
```go
type Action string
// predefined: Up, Down, Left, Right, Confirm, Cancel, Quit, FocusNext,
// FocusPrev, Inc, Dec, Erase, PageUp, PageDown

type Binding struct { Keys []string; Action Action; Glyph, Label string }
func (b Binding) DisplayGlyph() string   // Glyph, else strings.Join(Keys,"/")
func (b Binding) WithGlyph(g string) Binding
func (b Binding) WithLabel(l string) Binding

type Scheme
func Default() Scheme             // memoized built-in default
func (s Scheme) Binding(a Action) Binding
func (s Scheme) With(overrides ...Binding) Scheme
func (s Scheme) WithDefaults(defaults ...Binding) Scheme
func (s Scheme) MapFor(actions ...Action) *Map

func Register(key, name string, s Scheme)  // add/replace a named scheme; concurrency-safe
func Keys() []string                       // named-scheme keys
func Named(key string) (Scheme, bool)      // false ⇒ returns Default()
func DisplayName(key string) string

type Hint struct { Key, Label string }     // one footer-legend entry

type Map
func NewMap(bindings ...Binding) *Map
func (m *Map) Action(input string) (Action, bool)
func (m *Map) Has(a Action) bool
func (m *Map) Hint(a Action) Hint
func (m *Map) HintLabeled(a Action, label string) Hint
func (m *Map) Hints(actions ...Action) []Hint  // only bindings with a Glyph
```

## forms — `github.com/codyconfer/viewkit/forms`
```go
type FieldKind int
const ( FieldText FieldKind = iota; FieldMultiline; FieldSelect; FieldMultiselect; FieldRadio; FieldToggle )
type Field struct { Key, Label string; Kind FieldKind; Options []string
                    Text string; On bool; Selected int; Checked map[int]bool }
func (fd *Field) Value() any

type Form
func NewForm(fields ...Field) *Form
func (fm *Form) Focused() *Field
func (fm *Form) Handle(a keys.Action) bool     // true = consumed
func (fm *Form) Insert(s string)               // append typed runes to focused text field
func (fm *Form) Values() map[string]any
func (fm *Form) Render(f layout.Frame, title string) string
func (fm *Form) Overlay(bg string, f layout.Frame, title string, pos ...layout.OverlayPos) string

type Result int
const ( Pending Result = iota; Submitted; Cancelled )
type Confirm struct { Title, Message, YesLabel, NoLabel string; Yes bool }
func (c *Confirm) Handle(a keys.Action) Result
func (c Confirm) Render(f layout.Frame) string
func (c Confirm) Overlay(bg string, f layout.Frame, pos ...layout.OverlayPos) string
```

## notify — `github.com/codyconfer/viewkit/notify`
```go
type Tone int
const ( TonePositive Tone = iota; ToneNeutral; ToneWarning; ToneNegative )
type Notification struct { Title, Message string; Tone Tone }
func Note(tone Tone, title, message string) Notification
func Positive(title, message string) Notification
func Neutral(title, message string) Notification
func Warning(title, message string) Notification
func Negative(title, message string) Notification

type Queue
func NewQueue(cap int) *Queue
func (q *Queue) Push(n Notification, ttl int)
func (q *Queue) Beat()                       // tick one TTL step
func (q *Queue) Current() (Notification, bool)
func (q *Queue) Active() bool
func (q *Queue) Len() int
```

## timefmt — `github.com/codyconfer/viewkit/timefmt`
```go
func Rel(t time.Time) string              // relative to time.Now(): "just now", "5m ago", "in 2h", "3d ago"
func RelAt(t, now time.Time) string       // relative to an explicit now (deterministic / testable)
```
