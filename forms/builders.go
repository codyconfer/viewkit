package forms

// Field builders: constructors for each FieldKind plus chainable With* setters,
// so call sites read as a declaration instead of a struct literal with a Kind
// tag. Each With* returns a copy; a builder chain never mutates its receiver.

// Text builds a single-line text field.
func Text(key, label string) Field {
	return Field{Key: key, Label: label, Kind: FieldText}
}

// Multiline builds a multi-line text field.
func Multiline(key, label string) Field {
	return Field{Key: key, Label: label, Kind: FieldMultiline}
}

// Select builds a single-choice field over options.
func Select(key, label string, options ...string) Field {
	return Field{Key: key, Label: label, Kind: FieldSelect, Options: options}
}

// Multiselect builds a multi-choice field over options.
func Multiselect(key, label string, options ...string) Field {
	return Field{Key: key, Label: label, Kind: FieldMultiselect, Options: options}
}

// Radio builds a one-of field over options rendered as radio buttons.
func Radio(key, label string, options ...string) Field {
	return Field{Key: key, Label: label, Kind: FieldRadio, Options: options}
}

// Toggle builds an on/off field starting at on.
func Toggle(key, label string, on bool) Field {
	return Field{Key: key, Label: label, Kind: FieldToggle, On: on}
}

// WithText seeds the field's text value.
func (fd Field) WithText(v string) Field {
	fd.Text = v
	return fd
}

// WithSecret masks the field's input.
func (fd Field) WithSecret() Field {
	fd.Secret = true
	return fd
}

// WithSuggest attaches a completion source (text and multiline fields).
func (fd Field) WithSuggest(s Suggester) Field {
	fd.Suggest = s
	return fd
}

// WithDelim splits the text into independently completed tokens.
func (fd Field) WithDelim(d string) Field {
	fd.Delim = d
	return fd
}

// WithSelected pre-selects the option at index i (select and radio fields).
func (fd Field) WithSelected(i int) Field {
	fd.Selected = i
	return fd
}
