package keys

const (
	Up        Action = "nav.up"
	Down      Action = "nav.down"
	Left      Action = "nav.left"
	Right     Action = "nav.right"
	Confirm   Action = "confirm"
	Cancel    Action = "cancel"
	Quit      Action = "quit"
	FocusNext Action = "focus.next"
	FocusPrev Action = "focus.prev"
	Inc       Action = "inc"
	Dec       Action = "dec"
	Erase     Action = "erase"
	PageUp    Action = "page.up"
	PageDown  Action = "page.down"
	Open      Action = "open"
	Reload    Action = "reload"

	// Complete accepts the suggestion a text field is showing. It shares tab
	// with FocusNext: a host that binds both resolves the conflict by build
	// order, and one that binds Complete alone gets tab-to-accept for free.
	Complete     Action = "complete"
	CompleteNext Action = "complete.next"
	CompletePrev Action = "complete.prev"
)

type Scheme struct {
	bindings map[Action]Binding
}

func (s Scheme) Binding(a Action) Binding {
	return s.bindings[a]
}

func (s Scheme) With(overrides ...Binding) Scheme {
	next := Scheme{bindings: make(map[Action]Binding, len(s.bindings))}
	for k, v := range s.bindings {
		next.bindings[k] = v
	}
	for _, b := range overrides {
		next.bindings[b.Action] = b
	}
	return next
}

func Default() Scheme {
	return Scheme{bindings: map[Action]Binding{
		Up:        {Keys: []string{"up", "k"}, Action: Up, Glyph: "↑/↓/j/k"},
		Down:      {Keys: []string{"down", "j"}, Action: Down},
		Left:      {Keys: []string{"left", "h"}, Action: Left, Glyph: "←/→/h/l"},
		Right:     {Keys: []string{"right", "l"}, Action: Right},
		Confirm:   {Keys: []string{"enter", " ", "spacebar"}, Action: Confirm, Glyph: "enter/space"},
		Cancel:    {Keys: []string{"esc"}, Action: Cancel},
		Quit:      {Keys: []string{"ctrl+c"}, Action: Quit},
		FocusNext: {Keys: []string{"tab"}, Action: FocusNext, Glyph: "tab/⇧tab", Label: "focus panel"},
		FocusPrev: {Keys: []string{"shift+tab"}, Action: FocusPrev},
		Inc:       {Keys: []string{"]", "+", "="}, Action: Inc, Glyph: "[ ]/-/+"},
		Dec:       {Keys: []string{"[", "-", "_"}, Action: Dec},
		Erase:     {Keys: []string{"backspace", "ctrl+h"}, Action: Erase, Glyph: "backspace"},
		PageUp:    {Keys: []string{"pgup"}, Action: PageUp, Glyph: "pgup/pgdn"},
		PageDown:  {Keys: []string{"pgdown"}, Action: PageDown},
		Open:      {Keys: []string{"o"}, Action: Open, Glyph: "o", Label: "open"},
		Reload:    {Keys: []string{"r", "f5"}, Action: Reload, Glyph: "r", Label: "reload"},

		Complete:     {Keys: []string{"tab"}, Action: Complete, Glyph: "tab", Label: "accept"},
		CompleteNext: {Keys: []string{"ctrl+n"}, Action: CompleteNext, Glyph: "ctrl+n/ctrl+p", Label: "suggestion"},
		CompletePrev: {Keys: []string{"ctrl+p"}, Action: CompletePrev, Glyph: "ctrl+p"},
	}}
}

var current = Default()

func Cur() Scheme { return current }

func Use(s Scheme) { current = s }

type registryEntry struct {
	key    string
	name   string
	scheme Scheme
}

var registry = []registryEntry{
	{key: "default", name: "Default", scheme: Default()},
}

func Register(key, name string, s Scheme) {
	for i := range registry {
		if registry[i].key == key {
			registry[i].name = name
			registry[i].scheme = s
			return
		}
	}
	registry = append(registry, registryEntry{key: key, name: name, scheme: s})
}

func Keys() []string {
	out := make([]string, len(registry))
	for i, e := range registry {
		out[i] = e.key
	}
	return out
}

func Named(key string) (Scheme, bool) {
	for _, e := range registry {
		if e.key == key {
			return e.scheme, true
		}
	}
	return Default(), false
}

func DisplayName(key string) string {
	for _, e := range registry {
		if e.key == key {
			return e.name
		}
	}
	return key
}
