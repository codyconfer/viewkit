package panels

import (
	"math/rand"
	"strings"

	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/theme"
)

var glyphSet = []rune(
	"ｦｧｨｩｪｫｬｭｮｯｰｱｲｳｴｵｶｷｸｹｺｻｼｽｾｿﾀﾁﾂﾃﾄﾅﾆﾇﾈﾉﾊﾋﾌﾍﾎﾏﾐﾑﾒﾓﾔﾕﾖﾗﾘﾙﾚﾛﾜﾝ" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789:.=*+-<>",
)

const (
	minTrail    = 3
	maxTrailCap = 12
	maxSpeed    = 3
	spawnProb   = 0.05
)

type column struct {
	head   int
	length int
	speed  int
	tick   int
}

// Rain is the mutable state of a Matrix-style digital rain animation: falling
// glyph columns over a width×rows cell grid, driven by its own seeded RNG so
// runs are reproducible. Advance it with Beat and draw it with Matrix.
type Rain struct {
	width  int
	rows   int
	cols   []column
	glyphs [][]rune
	rng    *rand.Rand
	beats  int
}

// NewRain creates a Rain of width columns by rows cells (each clamped to at
// least 1) seeded with a deterministic RNG. Roughly half the columns start
// with an active trail at a random height; the rest are idle.
func NewRain(width, rows int, seed int64) *Rain {
	r := &Rain{rng: rand.New(rand.NewSource(seed))}
	r.reset(width, rows)
	return r
}

// Resize re-seeds the grid at the new dimensions (in cells), discarding all
// current trail state. It is a no-op when the size is unchanged.
func (r *Rain) Resize(width, rows int) {
	if width == r.width && rows == r.rows {
		return
	}
	r.reset(width, rows)
}

func (r *Rain) reset(width, rows int) {
	if width < 1 {
		width = 1
	}
	if rows < 1 {
		rows = 1
	}
	r.width, r.rows = width, rows
	r.cols = make([]column, width)
	r.glyphs = make([][]rune, rows)
	for y := range r.glyphs {
		r.glyphs[y] = make([]rune, width)
		for x := range r.glyphs[y] {
			r.glyphs[y][x] = r.randGlyph()
		}
	}
	for x := range r.cols {
		if r.rng.Float64() < 0.5 {
			r.cols[x] = column{
				head:   r.rng.Intn(rows),
				length: r.randLength(),
				speed:  1 + r.rng.Intn(maxSpeed),
			}
		} else {
			r.cols[x] = column{head: -1}
		}
	}
}

// Beat advances the animation one tick: idle columns may spawn a new trail
// (5% chance each), active columns move their head down at their per-column
// speed (one cell every 1–3 beats) and retire once fully off-screen, and a
// handful of random cells (width/6) mutate their glyph for shimmer.
func (r *Rain) Beat() {
	r.beats++
	for x := range r.cols {
		c := &r.cols[x]
		if c.head < 0 {
			if r.rng.Float64() < spawnProb {
				c.head = 0
				c.length = r.randLength()
				c.speed = 1 + r.rng.Intn(maxSpeed)
				c.tick = 0
				r.glyphs[0][x] = r.randGlyph()
			}
			continue
		}
		c.tick++
		if c.tick < c.speed {
			continue
		}
		c.tick = 0
		c.head++
		if c.head < r.rows {
			r.glyphs[c.head][x] = r.randGlyph()
		}
		if c.head-c.length >= r.rows {
			c.head = -1
		}
	}

	for i := 0; i < r.width/6; i++ {
		r.glyphs[r.rng.Intn(r.rows)][r.rng.Intn(r.width)] = r.randGlyph()
	}
}

func (r *Rain) randGlyph() rune { return glyphSet[r.rng.Intn(len(glyphSet))] }

func (r *Rain) randLength() int {
	hi := r.rows
	if hi > maxTrailCap {
		hi = maxTrailCap
	}
	if hi <= minTrail {
		return minTrail
	}
	return minTrail + r.rng.Intn(hi-minTrail+1)
}

// Matrix renders r's current frame as a panel: trail heads in the accent
// style, the third of the trail behind them in the positive style, the rest
// dim. Columns beyond the frame body width are clipped; Matrix does not
// advance the animation — call Beat between renders.
func Matrix(f layout.Frame, title string, r *Rain) string {
	lines := r.renderBody(f.BodyWidth())
	for i, line := range lines {
		lines[i] = f.Fit(line)
	}
	return f.Panel(title, lines...)
}

func (r *Rain) renderBody(bw int) []string {
	th := theme.Cur()
	n := min(r.width, bw)
	if n < 1 {
		n = 1
	}
	lines := make([]string, r.rows)
	for y := 0; y < r.rows; y++ {
		var b strings.Builder
		for x := 0; x < n; x++ {
			c := r.cols[x]
			dist := c.head - y
			switch {
			case c.head < 0 || dist < 0 || dist > c.length:
				b.WriteByte(' ')
			case dist == 0:
				b.WriteString(th.Accent.Render(string(r.glyphs[y][x])))
			case dist <= c.length/3:
				b.WriteString(th.Can.Render(string(r.glyphs[y][x])))
			default:
				b.WriteString(th.Dim.Render(string(r.glyphs[y][x])))
			}
		}
		lines[y] = b.String()
	}
	return lines
}
