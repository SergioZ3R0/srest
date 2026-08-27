// Package ui contains the Bubble Tea model of srest.
//
// The UI consumes the internal/api client asynchronously via commands
// (tea.Cmd), so that HTTP calls never block rendering.
package ui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/SergioZ3R0/srest/internal/api"
)

// Model represents the complete UI state.
type Model struct {
	client *api.Client

	// status is the status text shown on screen.
	status string

	// err holds the connection error, if any.
	err error

	// done indicates the connection attempt has finished.
	done bool

	// version is the API version we talked to.
	version api.Version

	// release is the Slurm version serving the API.
	release string

	// warnings are the warnings reported by slurmrestd.
	warnings []api.Warning

	width  int
	height int
}

// New returns a Model ready to be used with an API client.
func New(client *api.Client) Model {
	return Model{
		client: client,
		status: "Connecting to Slurm...",
	}
}

// statusMsg is the message the connection command returns to the model.
// A nil err means success.
type statusMsg struct {
	info api.PingInfo
	err  error
}

// connectCmd auto-detects the API version (if not pinned), performs the ping
// request and returns the result as a statusMsg. It is a tea.Cmd, so it does
// not block the UI.
func connectCmd(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		if _, ok := c.Version(); !ok {
			if _, err := c.Detect(ctx); err != nil {
				return statusMsg{err: err}
			}
		}

		info, err := c.Ping(ctx)
		return statusMsg{info: info, err: err}
	}
}

// Init starts the connection (detection + ping) as soon as the program starts.
func (m Model) Init() tea.Cmd {
	return connectCmd(m.client)
}

// Update reacts to Bubble Tea messages (keyboard, network, window).
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case statusMsg:
		m.done = true
		if msg.err != nil {
			m.err = msg.err
			m.status = "Connection error"
		} else {
			m.status = "Connection successful!"
			m.version = msg.info.API
			m.release = msg.info.Slurm.Release
			m.warnings = msg.info.Warnings
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// View renders the current model state.
func (m Model) View() string {
	title := titleStyle.Render("srest")
	help := helpStyle.Render("Press 'q' to quit")

	content := statusStyle.Render(m.status)
	if m.done && m.err != nil {
		content = errorStyle.Render(m.status + ": " + m.err.Error())
	}

	lines := []string{title, content}

	if m.done && m.err == nil {
		lines = append(lines,
			detailStyle.Render(fmt.Sprintf("API:   %s", m.version)),
			detailStyle.Render(fmt.Sprintf("Slurm: %s", m.release)),
		)
		for _, w := range m.warnings {
			lines = append(lines, warningStyle.Render("warning: "+w.Description))
		}
	}

	lines = append(lines, help)

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
