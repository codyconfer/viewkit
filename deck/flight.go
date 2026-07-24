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
	"github.com/codyconfer/viewkit/theme"
)

// Task is one unit of work in a flight. Run must be safe for concurrent
// execution; return Content (never domain types).
type Task struct {
	Label string
	Run   func(ctx context.Context) (Content, error)
}

// Execute runs tasks concurrently via errgroup and returns bodies in order.
// This is the headless driver; RunFlight is the tea progressive UI driver.
func Execute(ctx context.Context, tasks []Task) ([]Content, error) {
	out := make([]Content, len(tasks))
	g, ctx := errgroup.WithContext(ctx)
	for i, t := range tasks {
		g.Go(func() error {
			c, err := t.Run(ctx)
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

// RunFlight shows a progressive tea UI while tasks run under errgroup.
// Quit keys follow keys.Cur() (Quit / Cancel).
func RunFlight(ctx context.Context, tasks []Task) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))

	m := &flightModel{
		ctx:    ctx,
		tasks:  tasks,
		panels: make([]flightPanel, len(tasks)),
		spin:   sp,
		left:   len(tasks),
	}
	for i, t := range tasks {
		m.panels[i].label = t.Label
	}

	p := tea.NewProgram(m, tea.WithContext(ctx))
	m.program = p

	if _, err := p.Run(); err != nil {
		return err
	}
	return m.err
}

type flightPanel struct {
	label   string
	done    bool
	content string
}

type flightDoneMsg struct {
	idx     int
	content string
	err     error
}

type flightModel struct {
	ctx     context.Context
	tasks   []Task
	panels  []flightPanel
	spin    spinner.Model
	left    int
	err     error
	program *tea.Program
	once    sync.Once
}

func (m *flightModel) Init() tea.Cmd {
	m.once.Do(func() {
		go m.runWorkers()
	})
	return m.spin.Tick
}

func (m *flightModel) runWorkers() {
	g, ctx := errgroup.WithContext(m.ctx)
	for i, t := range m.tasks {
		g.Go(func() error {
			c, err := t.Run(ctx)
			body := ""
			if c != nil {
				body = c.Render(theme.BodyWidth)
			}
			if m.program != nil {
				m.program.Send(flightDoneMsg{idx: i, content: body, err: err})
			}
			return nil
		})
	}
	_ = g.Wait()
}

func (m *flightModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		sc := keys.Cur()
		if act, ok := keys.NewMap(sc.Binding(keys.Quit), sc.Binding(keys.Cancel)).Action(msg.String()); ok {
			if act == keys.Quit || act == keys.Cancel {
				return m, tea.Quit
			}
		}
	case flightDoneMsg:
		if !m.panels[msg.idx].done {
			m.panels[msg.idx].done = true
			if msg.err != nil {
				m.panels[msg.idx].content = theme.Cur().Cant.Render(msg.err.Error())
				if m.err == nil {
					m.err = msg.err
				}
			} else {
				m.panels[msg.idx].content = msg.content
			}
			m.left--
		}
		if m.left == 0 {
			return m, tea.Quit
		}
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *flightModel) View() string {
	parts := make([]string, len(m.panels))
	f := layout.NewFrame(theme.BodyWidth)
	for i, p := range m.panels {
		if p.done {
			parts[i] = p.content
			continue
		}
		parts[i] = f.TitledBox(p.label, theme.Cur().Dim.Render(m.spin.View()+" loading…"))
	}
	return strings.Join(parts, "\n") + "\n"
}
