package deck

import (
	"context"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/sync/errgroup"

	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/ui"
)

// Job is one unit of work. Run must be safe for concurrent execution; return
// Content (never domain types).
type Job struct {
	Label string
	Run   func(ctx context.Context) (Content, error)
}

// Work is a set of jobs run together: Collect is the headless driver,
// RunInteractive the tea progressive UI driver.
type Work []Job

// Collect runs the jobs concurrently via errgroup and returns bodies in order,
// with no UI — the headless driver.
func (w Work) Collect(ctx context.Context) ([]Content, error) {
	out := make([]Content, len(w))
	g, ctx := errgroup.WithContext(ctx)
	for i, j := range w {
		g.Go(func() error {
			c, err := j.Run(ctx)
			if err != nil {
				return err
			}
			out[i] = c
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return out, nil
}

// RunInteractive is RunInteractiveIn on the process-default scope.
func (w Work) RunInteractive(ctx context.Context) error {
	return w.RunInteractiveIn(ctx, nil)
}

// RunInteractiveIn shows a progressive tea UI while the jobs run under
// errgroup, rendering with scope (nil snapshots the process defaults).
// Finished results print to scrollback in job order; quit keys follow the
// scope's scheme.
func (w Work) RunInteractiveIn(ctx context.Context, scope *ui.Scope) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if scope == nil {
		scope = ui.Default()
	}

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))

	m := &workModel{
		ctx:    ctx,
		jobs:   w,
		panels: make([]jobPanel, len(w)),
		spin:   sp,
		left:   len(w),
		ui:     scope,
	}
	for i, j := range w {
		m.panels[i].label = j.Label
	}

	p := tea.NewProgram(m, tea.WithContext(ctx))
	m.program = p

	if _, err := p.Run(); err != nil {
		return err
	}
	return m.err
}

type jobPanel struct {
	label   string
	done    bool
	printed bool
	content string
}

type jobDoneMsg struct {
	idx     int
	content string
	err     error
}

type workModel struct {
	ctx     context.Context
	jobs    Work
	panels  []jobPanel
	spin    spinner.Model
	left    int
	next    int
	err     error
	program *tea.Program
	ui      *ui.Scope
	once    sync.Once
}

func (m *workModel) Init() tea.Cmd {
	m.once.Do(func() {
		go m.runWorkers()
	})
	return m.spin.Tick
}

func (m *workModel) runWorkers() {
	g, ctx := errgroup.WithContext(m.ctx)
	for i, j := range m.jobs {
		g.Go(func() error {
			c, err := j.Run(ctx)
			body := ""
			if c != nil {
				body = c.Render(layout.DocumentFrame().WithUI(m.ui))
			}
			if m.program != nil {
				m.program.Send(jobDoneMsg{idx: i, content: body, err: err})
			}
			return nil
		})
	}
	_ = g.Wait()
}

func (m *workModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		sc := m.ui.Keys
		if act, ok := keys.NewMap(sc.Binding(keys.Quit), sc.Binding(keys.Cancel)).Action(msg.String()); ok {
			if act == keys.Quit || act == keys.Cancel {
				return m, tea.Quit
			}
		}
	case jobDoneMsg:
		if !m.panels[msg.idx].done {
			m.panels[msg.idx].done = true
			if msg.err != nil {
				m.panels[msg.idx].content = m.ui.Theme.Cant.Render(msg.err.Error())
				if m.err == nil {
					m.err = msg.err
				}
			} else {
				m.panels[msg.idx].content = msg.content
			}
			m.left--
		}
		var cmd tea.Cmd
		if body, ok := m.drain(); ok {
			cmd = tea.Println(body)
		}
		if m.left == 0 {
			if cmd != nil {
				return m, tea.Sequence(cmd, tea.Quit)
			}
			return m, tea.Quit
		}
		return m, cmd
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *workModel) drain() (string, bool) {
	var out []string
	for m.next < len(m.panels) && m.panels[m.next].done {
		out = append(out, m.panels[m.next].content)
		m.panels[m.next].printed = true
		m.next++
	}
	if len(out) == 0 {
		return "", false
	}
	return strings.Join(out, "\n"), true
}

func (m *workModel) View() string {
	parts := make([]string, 0, len(m.panels))
	f := layout.DocumentFrame().WithUI(m.ui)
	for _, p := range m.panels {
		if p.printed {
			continue
		}
		status := m.spin.View() + " loading…"
		if p.done {
			status = "queued…"
		}
		parts = append(parts, f.TitledBox(p.label, m.ui.Theme.Dim.Render(status)))
	}
	return strings.Join(parts, "\n") + "\n"
}
