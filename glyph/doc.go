// Package glyph renders icons as strings with mode-dependent variants.
//
// Each glyph is a Variants triple (nerd-font, unicode, ASCII); the
// process-wide mode — set explicitly via SetMode, parsed from a string via
// ParseMode, or auto-selected by DetectMode — picks which variant the glyph
// helpers (Check, Arrow, StatusOK, ...) return. Partially populated Variants
// degrade to the nearest filled variant so a glyph never renders as nothing.
// Pad and Lead handle the spacing quirks of variable-width icons, and
// Severity maps app-defined tones onto inline marks (GlyphFor) and status
// dots (StatusFor).
package glyph
