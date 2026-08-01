package glyph

// Set renders glyphs for one fixed Mode, the scoped alternative to the
// package-level funcs (which read the process default mode).
type Set struct {
	mode Mode
}

// SetFor returns a Set rendering in mode m.
func SetFor(m Mode) Set { return Set{mode: m} }

// DefaultSet returns a Set rendering in the process default mode.
func DefaultSet() Set { return Set{mode: CurrentMode()} }

// Mode returns the mode this Set renders in.
func (g Set) Mode() Mode { return g.mode }

func (g Set) of(v Variants) string { return v.stringIn(g.mode) }

// Resolve returns the mode-appropriate glyph for a registered id, or "".
func (g Set) Resolve(id string) string {
	if v, ok := Named(id); ok {
		return g.of(v)
	}
	return ""
}

// GlyphFor returns an inline mark for sev (Check/Warn/Cross/Bullet).
func (g Set) GlyphFor(sev Severity) string {
	switch sev {
	case SeverityPositive:
		return g.Check()
	case SeverityWarning:
		return g.Warn()
	case SeverityNegative:
		return g.Cross()
	default:
		return g.Bullet()
	}
}

// StatusFor returns the status-strip dot for sev (StatusOK/Warn/Bad/Muted).
func (g Set) StatusFor(sev Severity) string {
	switch sev {
	case SeverityPositive:
		return g.StatusOK()
	case SeverityNegative:
		return g.StatusBad()
	case SeverityWarning:
		return g.StatusWarn()
	default:
		return g.StatusMuted()
	}
}

// StatusOK renders the healthy status dot in this set's mode.
func (g Set) StatusOK() string { return g.of(statusOK) }

// StatusWarn renders the warning status marker in this set's mode.
func (g Set) StatusWarn() string { return g.of(statusWarn) }

// StatusBad renders the failing status dot in this set's mode.
func (g Set) StatusBad() string { return g.of(statusBad) }

// StatusMuted renders the inactive status dot in this set's mode.
func (g Set) StatusMuted() string { return g.of(statusMuted) }

// Check renders the check mark in this set's mode.
func (g Set) Check() string { return g.of(check) }

// Cross renders the cross mark in this set's mode.
func (g Set) Cross() string { return g.of(cross) }

// Warn renders the warning mark in this set's mode.
func (g Set) Warn() string { return g.of(warn) }

// Arrow renders the arrow in this set's mode.
func (g Set) Arrow() string { return g.of(arrow) }

// Bullet renders the bullet in this set's mode.
func (g Set) Bullet() string { return g.of(bullet) }

// GitHub renders the GitHub logo in this set's mode.
func (g Set) GitHub() string { return g.of(github) }

// Slack renders the Slack logo in this set's mode.
func (g Set) Slack() string { return g.of(slack) }

// Google renders the Google logo in this set's mode.
func (g Set) Google() string { return g.of(google) }

// Diamond renders the diamond in this set's mode.
func (g Set) Diamond() string { return g.of(diamond) }

// History renders the history icon in this set's mode.
func (g Set) History() string { return g.of(history) }

// List renders the list icon in this set's mode.
func (g Set) List() string { return g.of(list) }

// Database renders the database icon in this set's mode.
func (g Set) Database() string { return g.of(database) }

// Cog renders the settings cog in this set's mode.
func (g Set) Cog() string { return g.of(cog) }

// User renders the user icon in this set's mode.
func (g Set) User() string { return g.of(user) }

// SignOut renders the sign-out icon in this set's mode.
func (g Set) SignOut() string { return g.of(signOut) }

// Clock renders the clock icon in this set's mode.
func (g Set) Clock() string { return g.of(clock) }
