// Package ui contains the Bubble Tea model of srest.
//
// The UI consumes the internal/api client asynchronously via commands
// (tea.Cmd), so that HTTP calls never block rendering.
package ui

import (
	"context"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/SergioZ3R0/srest/internal/api"
)

// Model represents the complete UI state.
type Model struct {
	client *api.Client

	// Connection state.
	status   string
	err      error
	done     bool
	version  api.Version
	release  string
	warnings []api.Warning

	// UI state.
	tabs       []string
	active     int
	jobs       table.Model
	nodes      table.Model
	partitions table.Model
	spinner    spinner.Model
	help       help.Model
	width      int
	height     int
}

// New returns a Model ready to be used with an API client.
func New(client *api.Client) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return Model{
		client:     client,
		status:     "Connecting to Slurm...",
		tabs:       []string{"Dashboard", "Jobs", "Nodes", "Partitions"},
		active:     0,
		jobs:       newJobsTable(),
		nodes:      newNodesTable(),
		partitions: newPartitionsTable(),
		spinner:    s,
		help:       help.New(),
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

// Init starts the connection (detection + ping) and the spinner tick.
func (m Model) Init() tea.Cmd {
	return tea.Batch(connectCmd(m.client), m.spinner.Tick)
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
			m.status = "Connected"
			m.version = msg.info.API
			m.release = msg.info.Slurm.Release
			m.warnings = msg.info.Warnings
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		case key.Matches(msg, keys.NextTab):
			m.active = (m.active + 1) % len(m.tabs)
			m.focusTab()
			return m, nil
		case key.Matches(msg, keys.PrevTab):
			m.active = (m.active - 1 + len(m.tabs)) % len(m.tabs)
			m.focusTab()
			return m, nil
		}

		// Delegate remaining keys to the active table (row navigation).
		switch m.active {
		case 1:
			m.jobs, _ = m.jobs.Update(msg)
		case 2:
			m.nodes, _ = m.nodes.Update(msg)
		case 3:
			m.partitions, _ = m.partitions.Update(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Let the tables grow to fill the available vertical space.
		h := m.tableHeight()
		m.jobs.SetHeight(h)
		m.nodes.SetHeight(h)
		m.partitions.SetHeight(h)
	}

	return m, nil
}

// tableHeight returns the height reserved for data tables so the layout fills
// the terminal without overflowing.
func (m Model) tableHeight() int {
	h := m.height - 14
	if h < 3 {
		h = 3
	}
	return h
}

// focusTab blurs every table and focuses the one matching the active tab.
func (m *Model) focusTab() {
	m.jobs.Blur()
	m.nodes.Blur()
	m.partitions.Blur()
	switch m.active {
	case 1:
		m.jobs.Focus()
	case 2:
		m.nodes.Focus()
	case 3:
		m.partitions.Focus()
	}
}

// headerView renders the top bar with the title and connection status, laid
// out to the given width.
func (m Model) headerView(width int) string {
	left := titleStyle.Render("srest")

	var right string
	switch {
	case !m.done:
		right = m.spinner.View() + " " + statusStyle.Render(m.status)
	case m.err != nil:
		right = errorStyle.Render(m.status + ": " + m.err.Error())
	default:
		right = successStyle.Render(m.status) +
			detailStyle.Render(" · Slurm "+m.release+" · API "+m.version.String())
	}

	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		left,
		lipgloss.NewStyle().Width(gap).Render(" "),
		right,
	)
}

// tabsView renders the tab bar.
func (m Model) tabsView() string {
	items := make([]string, 0, len(m.tabs))
	for i, t := range m.tabs {
		if i == m.active {
			items = append(items, activeTabStyle.Render(t))
		} else {
			items = append(items, tabStyle.Render(t))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, items...)
}

// contentView renders the panel corresponding to the active tab.
func (m Model) contentView() string {
	switch m.active {
	case 0:
		return dashboardView()
	case 1:
		return m.jobs.View()
	case 2:
		return m.nodes.View()
	case 3:
		return m.partitions.View()
	default:
		return ""
	}
}

// View renders the whole UI, filling and centered within the terminal.
func (m Model) View() string {
	width := m.width
	if width <= 0 {
		width = 100
	}
	height := m.height
	if height <= 0 {
		height = 30
	}

	// Width available for the inner content (outer frame takes 2 + 2*1).
	innerWidth := width - 4
	if innerWidth < 40 {
		innerWidth = 40
	}

	divider := dividerStyle.Render(strings.Repeat("─", innerWidth))

	lines := []string{
		m.headerView(innerWidth),
		divider,
		m.tabsView(),
		divider,
		m.contentView(),
	}
	for _, w := range m.warnings {
		lines = append(lines, warningStyle.Render("warning: "+w.Description))
	}
	lines = append(lines, divider, m.help.View(keys))

	block := lipgloss.JoinVertical(lipgloss.Left, lines...)
	framed := frameStyle.Render(block)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, framed)
}
