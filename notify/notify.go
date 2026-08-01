package notify

import "github.com/codyconfer/viewkit/glyph"

// Notification is a transient user-facing message. Severity uses the shared
// glyph.Severity vocabulary; theme supplies the colors when rendered.
type Notification struct {
	Title    string
	Message  string
	Severity glyph.Severity
}

// Note builds a Notification with an explicit severity.
func Note(sev glyph.Severity, title, message string) Notification {
	return Notification{Title: title, Message: message, Severity: sev}
}

// Positive builds a positive-severity notification.
func Positive(title, message string) Notification {
	return Note(glyph.SeverityPositive, title, message)
}

// Neutral builds a neutral-severity notification.
func Neutral(title, message string) Notification {
	return Note(glyph.SeverityNeutral, title, message)
}

// Warning builds a warning-severity notification.
func Warning(title, message string) Notification {
	return Note(glyph.SeverityWarning, title, message)
}

// Negative builds a negative-severity notification.
func Negative(title, message string) Notification {
	return Note(glyph.SeverityNegative, title, message)
}
