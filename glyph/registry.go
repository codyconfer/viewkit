package glyph

import (
	"strings"
	"sync"
)

var (
	regMu sync.RWMutex
	reg   = map[string]Variants{}
)

func init() {
	Register("github", github)
	Register("slack", slack)
	Register("google", google)
}

// NormalizeID lowercases and trims a registry id.
func NormalizeID(id string) string {
	return strings.ToLower(strings.TrimSpace(id))
}

// Register associates id with glyph variants for Nerd/Uni/ASCII modes.
// Built-in helpers (GitHub, Slack, …) remain; plugins use Register for
// contribution glyphs. Ids are normalized (trim + lowercase).
func Register(id string, v Variants) {
	id = NormalizeID(id)
	if id == "" {
		return
	}
	regMu.Lock()
	reg[id] = v
	regMu.Unlock()
}

// Named returns registered variants for id (normalized).
func Named(id string) (Variants, bool) {
	id = NormalizeID(id)
	if id == "" {
		return Variants{}, false
	}
	regMu.RLock()
	defer regMu.RUnlock()
	v, ok := reg[id]
	return v, ok
}

// ResolveID returns the mode-appropriate glyph string for a registered id.
func ResolveID(id string) string {
	if v, ok := Named(id); ok {
		return v.String()
	}
	return ""
}

// StatusContribution is the status-strip contract plugins declare.
// tone maps through Severity → theme colors; plugins never touch lipgloss.
type StatusContribution struct {
	BrandGlyph string
	Info       func() string
	Status     func() (glyph string, tone Severity)
}

// StatusChip is one right-strip entry carrying glyph text and its Severity
// so hosts can color via theme.
type StatusChip struct {
	Glyph    string
	Severity Severity
}

// StatusStrip aggregates left (brand/role/context) and right (status) slots.
type StatusStrip struct {
	Left  []string
	Right []StatusChip
}

// BuildStatusStrip assembles a strip from brand/role/context chips and
// plugin status contributions. Right chips retain Severity for coloring.
func BuildStatusStrip(brand, role string, contexts []string, contribs []StatusContribution) StatusStrip {
	var left []string
	if brand != "" {
		left = append(left, brand)
	}
	if role != "" {
		left = append(left, role)
	}
	left = append(left, contexts...)
	var right []StatusChip
	for _, c := range contribs {
		if c.Status == nil {
			continue
		}
		g, sev := c.Status()
		if g != "" {
			right = append(right, StatusChip{Glyph: g, Severity: sev})
		}
	}
	return StatusStrip{Left: left, Right: right}
}
