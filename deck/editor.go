package deck

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/viewkit/browser"
	"github.com/codyconfer/viewkit/forms"
	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/list"
	"github.com/codyconfer/viewkit/theme"
	"github.com/codyconfer/viewkit/ui"
)

// Results is the lower pane of an Editor: whatever a run produced.
type Results interface {
	// Items renders the result set as list rows sized to f.
	Items(f layout.Frame) []list.Item
	// Count reports how many underlying records the rows represent.
	Count() int
}

// ErrorCounter is an optional Results extension: error records are
// excluded from the collapsed item count.
type ErrorCounter interface {
	Errored() int
}

// EditorDoc is the document an Editor edits. Implementations own all
// domain concerns — parsing, validation, persistence — and hand the Editor
// only strings, form fields and Results.
type EditorDoc interface {
	// Kind names the document type for messages, e.g. "query".
	Kind() string
	// Title heads the form panel.
	Title() string
	// Context supplies the chrome context pairs.
	Context() []keys.Hint
	// SavedName is the persisted name, or "" when never saved. Delete and
	// the delete hint are suppressed while it is empty.
	SavedName() string
	// Fields builds the form, seeded from previously entered values.
	Fields(prev map[string]any) []forms.Field
	// Sync reports whether editing a field should rebuild the whole form,
	// for documents whose field set depends on earlier answers.
	Sync() bool
	// Summary is the one-line description shown when the form is collapsed.
	Summary() string
	// PreviewLines renders the document for the yaml-style preview panel.
	PreviewLines(f layout.Frame) []string
	// ValidateLines checks the document and returns display lines for the
	// validation panel. A non-nil error means the document could not be
	// built at all, and is surfaced as the status line instead.
	ValidateLines(f layout.Frame) ([]string, error)
	// Run executes the document. The returned loader runs off the UI
	// goroutine, so it must not touch the doc.
	Run() (label string, load func() Results, err error)
	// Persist saves the document and returns a confirmation summary.
	Persist() (summary string, err error)
	// Remove deletes the saved document. A non-nil error lands in the
	// status line, like Persist.
	Remove() (summary string, err error)
}

// EditorOutput is an optional EditorDoc extension for documents whose run
// produces text worth taking away — a rendered report, a generated file. An
// Editor offers Copy and Write only when its doc implements this and the host
// bound the matching keys, so documents without an output are unaffected.
type EditorOutput interface {
	// CopyOutput puts the last run's text on the clipboard.
	CopyOutput() (summary string, err error)
	// WriteOutput persists the last run's text and names where it landed.
	WriteOutput() (summary string, err error)
}

// EditorKeys binds an Editor's commands to a host application's key scheme.
// Actions are supplied by the caller because command vocabularies are
// app-specific; only Cancel and the navigation actions come from viewkit.
type EditorKeys struct {
	// Map resolves keys while the form or results pane has focus.
	Map *keys.Map
	// Confirm resolves keys while the delete dialog is open.
	Confirm *keys.Map

	Run      keys.Action
	Save     keys.Action
	Validate keys.Action
	Preview  keys.Action
	Delete   keys.Action
	Focus    keys.Action

	// Copy and Write drive EditorOutput. Leave them empty for documents that
	// produce no text.
	Copy  keys.Action
	Write keys.Action
}

const (
	editorPanelChrome  = 3
	editorMinResults   = 3
	editorRoomyResults = 10
	editorMinForm      = 5
	editorListIndent   = 2
)

type editorCollapse int

const (
	collapseNone editorCollapse = iota
	collapseResults
	collapseForm
	collapseBoth
)

type editorRanMsg struct {
	label   string
	results Results
}

// Editor is a two-pane View: an editable form above a scrollable result
// list. When the terminal is too short it collapses the results, then the
// form, to a one-line summary rather than overflowing.
type Editor struct {
	doc  EditorDoc
	keys EditorKeys

	form   *forms.Form
	sticky map[string]any

	status  string
	notice  []string
	preview bool
	confirm *forms.Confirm

	results       list.Model
	set           Results
	resultLabel   string
	resultsLines  int
	resultsHeight int
	hasResults    bool
	running       bool
	onResults     bool

	bodyHeight int
	bodyWidth  int
	frame      layout.Frame

	bound      bool
	boundWidth int
	boundUI    *ui.Scope
}

// NewEditor builds an Editor over doc. seed pre-fills the form and is
// carried across rebuilds when doc.Sync reports true.
func NewEditor(doc EditorDoc, km EditorKeys, seed map[string]any) *Editor {
	if seed == nil {
		seed = map[string]any{}
	}
	e := &Editor{doc: doc, keys: km, sticky: seed, results: list.New()}
	e.form = forms.NewForm(doc.Fields(seed)...)
	return e
}

// Form exposes the live form. Its Values() covers only currently rendered
// fields; use Remembered when Sync() rebuilds may have dropped some.
func (e *Editor) Form() *forms.Form { return e.form }

// Value returns the trimmed text currently entered under key.
func (e *Editor) Value(key string) string {
	return forms.Str(e.form.Values(), key)
}

// Remembered returns every value the Editor is carrying, including those of
// fields a later rebuild stopped rendering. Live form values win over
// remembered ones. A document whose field set shrinks reads this rather than
// Form().Values() when it must not silently discard what is still held.
func (e *Editor) Remembered() map[string]any {
	out := make(map[string]any, len(e.sticky)+len(e.form.Fields))
	for k, val := range e.sticky {
		out[k] = val
	}
	for k, val := range e.form.Values() {
		out[k] = val
	}
	return out
}

// SelectedOf returns the chosen option index of the select field under key,
// or -1 when no such field exists.
func (e *Editor) SelectedOf(key string) int {
	for i := range e.form.Fields {
		if e.form.Fields[i].Key == key {
			return e.form.Fields[i].Selected
		}
	}
	return -1
}

// Read-only accessors below exist for host tests in other modules.

// Status is the error line shown beneath the form, or "" when clear.
func (e *Editor) Status() string { return e.status }

// Notice is the current contents of the validation panel.
func (e *Editor) Notice() []string { return e.notice }

// Confirming reports whether the delete dialog is open.
func (e *Editor) Confirming() bool { return e.confirm != nil }

// Running reports whether a run is in flight.
func (e *Editor) Running() bool { return e.running }

// OnResults reports whether focus is on the results pane rather than the form.
func (e *Editor) OnResults() bool { return e.onResults }

// Selected returns the highlighted result row, if any.
func (e *Editor) Selected() (list.Item, bool) { return e.results.Selected() }

// Title implements View, delegating to the document.
func (e *Editor) Title() string { return e.doc.Title() }

// Init implements View; an Editor needs no startup command.
func (e *Editor) Init() tea.Cmd { return nil }

// Context implements View, delegating to the document.
func (e *Editor) Context(scope *ui.Scope) []keys.Hint { return e.doc.Context() }

// Hints implements View. The legend follows the current mode: delete
// confirmation, results pane, suggestion completion, or plain form editing.
func (e *Editor) Hints(scope *ui.Scope) []keys.Hint {
	if e.confirm != nil {
		return []keys.Hint{{Key: "←/→", Label: "choose"}, {Key: "enter", Label: "confirm"}}
	}
	if e.onResults {
		hints := []keys.Hint{
			{Key: "↑/↓", Label: "item"},
			{Key: "pgup/pgdn", Label: "page"},
			{Key: "enter", Label: "open"},
		}
		hints = e.withHint(hints, e.keys.Focus, "edit")
		return e.withHint(hints, e.keys.Run, "rerun")
	}
	if len(e.form.Suggestions()) > 0 {
		hints := []keys.Hint{
			{Key: "tab", Label: "accept"},
			{Key: "ctrl+n/ctrl+p", Label: "suggestion"},
			{Key: "↑/↓", Label: "field"},
		}
		hints = e.withHint(hints, e.keys.Run, "run")
		return e.withHint(hints, e.keys.Save, "save")
	}
	hints := []keys.Hint{
		{Key: "↑/↓", Label: "field"},
		{Key: "←/→", Label: "adjust"},
	}
	hints = e.withHint(hints, e.keys.Run, "run")
	hints = e.withHint(hints, e.keys.Validate, "validate")
	hints = e.withHint(hints, e.keys.Preview, "yaml")
	hints = e.withHint(hints, e.keys.Save, "save")
	hints = append(hints, e.outputHints()...)
	if e.doc.SavedName() != "" {
		hints = e.withHint(hints, e.keys.Delete, "delete")
	}
	if e.hasResults && !e.running {
		hints = e.withHint(hints, e.keys.Focus, "results")
	}
	return hints
}

// withHint appends the hint for a host-supplied action, taking the glyph from the
// binding the host actually installed rather than assuming one. An action the host
// left unbound contributes no hint, so the footer never advertises a dead key.
func (e *Editor) withHint(hints []keys.Hint, act keys.Action, label string) []keys.Hint {
	if act == "" || e.keys.Map == nil || !e.keys.Map.Has(act) {
		return hints
	}
	return append(hints, e.keys.Map.HintLabeled(act, label))
}

// Update implements View. Keys go to the delete-confirm dialog when it is
// open, otherwise to the focused pane (form or results); it also handles
// resize, run completion (editorRanMsg) and ReloadMsg re-runs.
func (e *Editor) Update(a *Model, msg tea.Msg) tea.Cmd {
	e.frame.UI = a.UI()
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		e.bodyHeight, e.bodyWidth = m.Height, m.Width
		e.relayoutResults()
		return nil
	case editorRanMsg:
		e.running = false
		e.set = m.results
		e.resultLabel = m.label
		e.rebindResults()
		e.setFocus(true)
		return nil
	case ReloadMsg:
		if !e.hasResults || e.running {
			return nil
		}
		return e.run()
	case tea.KeyMsg:
		if e.confirm != nil {
			return e.updateConfirm(a, m)
		}
		return e.handleKey(a, m)
	}
	return nil
}

func (e *Editor) outputHints() []keys.Hint {
	if _, ok := e.doc.(EditorOutput); !ok {
		return nil
	}
	var out []keys.Hint
	for _, act := range []keys.Action{e.keys.Copy, e.keys.Write} {
		if act == "" || !e.keys.Map.Has(act) {
			continue
		}
		out = append(out, e.keys.Map.Hint(act))
	}
	return out
}

func (e *Editor) output(a *Model, act keys.Action) (tea.Cmd, bool) {
	doc, ok := e.doc.(EditorOutput)
	if !ok || act == "" {
		return nil, false
	}
	var (
		run  func() (string, error)
		kind string
	)
	switch act {
	case e.keys.Copy:
		run, kind = doc.CopyOutput, "copy"
	case e.keys.Write:
		run, kind = doc.WriteOutput, "write"
	default:
		return nil, false
	}
	summary, err := run()
	if err != nil {
		e.status = err.Error()
		return nil, true
	}
	e.status = ""
	return a.Push(NewMessage(kind, summary, e.doc.Context())), true
}

func (e *Editor) handleKey(a *Model, key tea.KeyMsg) tea.Cmd {
	act, ok := e.keys.Map.Action(key.String())
	if !ok {
		if e.onResults {
			return nil
		}
		if key.String() == " " {
			e.form.Insert(" ")
		} else if key.Type == tea.KeyRunes {
			e.form.Insert(string(key.Runes))
		}
		return nil
	}
	if cmd, owned := e.output(a, act); owned {
		return cmd
	}
	switch act {
	case keys.Cancel:
		return a.Pop()
	case e.keys.Run:
		return e.run()
	case e.keys.Save:
		return e.save(a)
	case e.keys.Validate:
		e.validate()
	case e.keys.Preview:
		e.preview = !e.preview
	case e.keys.Delete:
		e.askDelete()
	case e.keys.Focus:
		if !e.onResults && e.form.AcceptSuggestion() {
			return nil
		}
		e.setFocus(!e.onResults)
	default:
		if e.onResults {
			return e.scrollResults(act)
		}
		e.form.Handle(act)
		e.syncFields()
	}
	return nil
}

func (e *Editor) scrollResults(act keys.Action) tea.Cmd {
	page := max(e.resultsHeight-1, 1)
	switch act {
	case keys.Up:
		e.results.Move(-1)
	case keys.Down:
		e.results.Move(1)
	case keys.PageUp:
		e.results.Scroll(-page)
	case keys.PageDown:
		e.results.Scroll(page)
	case keys.Confirm:
		return e.openSelected()
	}
	return nil
}

func (e *Editor) updateConfirm(a *Model, key tea.KeyMsg) tea.Cmd {
	act, ok := e.keys.Confirm.Action(key.String())
	if !ok {
		return nil
	}
	switch e.confirm.Handle(act) {
	case forms.Submitted:
		yes := e.confirm.Yes
		e.confirm = nil
		if !yes {
			return nil
		}
		summary, err := e.doc.Remove()
		if err != nil {
			e.status = err.Error()
			return nil
		}
		e.status = ""
		pop := a.Pop()
		push := a.Push(NewMessage("deleted", summary, e.doc.Context()))
		return tea.Batch(pop, push)
	case forms.Cancelled:
		e.confirm = nil
	}
	return nil
}

func (e *Editor) askDelete() {
	name := e.doc.SavedName()
	if name == "" {
		e.status = "nothing to delete: this " + e.doc.Kind() + " has not been saved yet"
		return
	}
	e.notice = nil
	e.status = ""
	e.confirm = &forms.Confirm{
		Title:    "delete " + name + "?",
		Message:  "This removes the saved document.",
		YesLabel: "Delete",
		NoLabel:  "Keep",
	}
}

func (e *Editor) validate() {
	e.notice = nil
	lines, err := e.doc.ValidateLines(e.frame)
	if err != nil {
		e.status = err.Error()
		return
	}
	e.status = ""
	e.notice = lines
}

func (e *Editor) syncFields() {
	if !e.doc.Sync() {
		return
	}
	e.status = ""
	e.remember()
	focus := e.form.FocusedKey()
	e.form = forms.NewForm(e.doc.Fields(e.sticky)...)
	e.form.FocusKey(focus)
}

func (e *Editor) remember() {
	for k, val := range e.form.Values() {
		e.sticky[k] = val
	}
}

func (e *Editor) run() tea.Cmd {
	label, fetch, err := e.doc.Run()
	if err != nil {
		e.status = err.Error()
		return nil
	}
	e.status = ""
	e.notice = nil
	e.resultLabel = label
	e.hasResults = true
	e.running = true
	e.onResults = false
	e.results.SetFocused(false)
	return func() tea.Msg {
		return editorRanMsg{label: label, results: fetch()}
	}
}

func (e *Editor) save(a *Model) tea.Cmd {
	summary, err := e.doc.Persist()
	if err != nil {
		e.status = err.Error()
		return nil
	}
	e.status = ""
	return a.Push(NewMessage("saved", summary, e.doc.Context()))
}

func (e *Editor) relayoutResults() {
	if e.bound && e.bodyWidth == e.boundWidth && e.boundUI == e.frame.UI {
		return
	}
	e.bindResults(e.results.SetItemsKeepingCursor)
}

func (e *Editor) rebindResults() {
	e.bindResults(e.results.SetItems)
}

func (e *Editor) bindResults(install func([]list.Item)) {
	if !e.hasResults || e.bodyWidth <= 0 || e.set == nil {
		return
	}
	items := e.set.Items(layout.ScreenFrame(e.bodyWidth - editorListIndent).WithUI(e.frame.UI))
	install(items)
	e.bound, e.boundWidth, e.boundUI = true, e.bodyWidth, e.frame.UI
	lines := 0
	for i, it := range items {
		if i > 0 {
			lines += theme.ListItemGapY
		}
		lines += layout.CountLines(it.Block)
	}
	e.resultsLines = lines
}

func (e *Editor) resultsNeed() int {
	return min(max(e.resultsLines, editorMinResults), editorRoomyResults)
}

func (e *Editor) itemCount() int {
	if e.set == nil {
		return 0
	}
	return e.set.Count()
}

func (e *Editor) erroredCount() int {
	if set, ok := e.set.(ErrorCounter); ok {
		return set.Errored()
	}
	return 0
}

func (e *Editor) setFocus(onResults bool) {
	if onResults && (!e.hasResults || e.running) {
		return
	}
	e.onResults = onResults
	e.results.SetFocused(onResults)
}

func (e *Editor) openSelected() tea.Cmd {
	it, ok := e.results.Selected()
	if !ok || it.Key == "" {
		return nil
	}
	url := it.Key
	return func() tea.Msg {
		_ = browser.Open(url)
		return nil
	}
}

// Body implements View. When both panes exceed the height budget it
// collapses the unfocused pane — then both — to one-line summaries, and it
// overlays the delete-confirm dialog when that is open.
func (e *Editor) Body(frame layout.Frame) string {
	e.frame = frame
	e.results.SetTheme(frame.Theme())
	width, height := frame.Width, frame.Height
	if height > 0 {
		e.bodyHeight = height
	}
	f := frame.WithWidth(width)

	body := e.compose(f, width, collapseNone)
	if e.hasResults && (e.overflows(body) || (e.onResults && e.resultsHeight < e.resultsNeed())) {
		mode := collapseResults
		if e.onResults {
			mode = collapseForm
		}
		body = e.compose(f, width, mode)
		if e.overflows(body) {
			body = e.compose(f, width, collapseBoth)
		}
	}
	if e.confirm != nil {
		return e.confirm.Overlay(body, frame.WithWidth(layout.DialogWidth(width)))
	}
	return body
}

func (e *Editor) overflows(body string) bool {
	return e.bodyHeight > 0 && layout.CountLines(body) > e.bodyHeight
}

func (e *Editor) compose(f layout.Frame, width int, mode editorCollapse) string {
	var tail []string
	if e.preview {
		tail = append(tail, f.Panel("yaml", e.doc.PreviewLines(f)...))
	}
	if len(e.notice) > 0 {
		tail = append(tail, f.Panel("validation", e.notice...))
	}
	if e.status != "" {
		tail = append(tail, f.Theme().Cant.Render(e.status))
	}

	var head string
	if mode == collapseForm || mode == collapseBoth {
		head = e.collapsedForm(f)
	} else {
		head = e.form.RenderWindow(f, e.Title(), e.formBudget(tail))
	}

	parts := append([]string{head}, tail...)
	switch {
	case !e.hasResults:
	case mode == collapseResults || mode == collapseBoth:
		parts = append(parts, e.collapsedResults(f))
	default:
		parts = append(parts, e.resultsPanel(f, width, parts))
	}
	return layout.StackTight(parts...)
}

func (e *Editor) formBudget(tail []string) int {
	if e.bodyHeight <= 0 {
		return 0
	}
	used := editorPanelChrome
	for _, part := range tail {
		used += layout.CountLines(part)
	}
	if e.hasResults {
		reserve := e.resultsNeed() + editorPanelChrome
		if room := e.bodyHeight - used - editorMinForm; reserve > room {
			reserve = room
		}
		used += reserve
	}
	return max(e.bodyHeight-used, editorMinForm)
}

func (e *Editor) collapsedResults(f layout.Frame) string {
	th := f.Theme()
	if e.running {
		return f.Panel(e.resultsTitle(), th.Dim.Render("running…"))
	}
	count := e.itemCount()
	items := count - e.erroredCount()
	summary := fmt.Sprintf("%d item(s)  ·  tab to view", items)
	switch {
	case count == 0:
		summary = "no items"
	case items <= 0:
		summary = "errors  ·  tab to view"
	}
	return f.Panel(e.resultsTitle(), th.Dim.Render(summary))
}

func (e *Editor) resultsTitle() string { return "results: " + e.resultLabel }

func (e *Editor) collapsedForm(f layout.Frame) string {
	th := f.Theme()
	return f.Panel(e.Title(), th.Dim.Render(e.doc.Summary()+"  ·  tab to edit"))
}

func (e *Editor) resultsPanel(f layout.Frame, width int, above []string) string {
	th := f.Theme()
	title := e.resultsTitle()
	switch {
	case e.running:
		return f.Panel(title, th.Dim.Render("running…"))
	case e.itemCount() == 0:
		return f.Panel(title, th.Dim.Render("no items"))
	}

	used := 0
	for _, part := range above {
		used += layout.CountLines(part)
	}
	avail := max(e.bodyHeight-used-editorPanelChrome, editorMinResults)
	e.resultsHeight = avail
	e.results.SetSize(width-editorListIndent, avail)

	hint := th.Dim.Render("tab to edit")
	if !e.onResults {
		hint = th.Dim.Render("tab to scroll")
	}
	return f.Panel(title+"  "+hint, layout.Lines(e.results.View())...)
}
