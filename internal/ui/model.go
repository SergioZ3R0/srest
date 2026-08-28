// Package ui contains the Bubble Tea model of srest.
//
// The UI consumes the internal/api client asynchronously via commands
// (tea.Cmd), so that HTTP calls never block rendering.
package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
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
	queries    []api.Query
	queryVP    viewport.Model
	composer   composer
	queryFocus int
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
		tabs:       []string{"Dashboard", "Jobs", "Nodes", "Partitions", "Query"},
		active:     0,
		jobs:       newJobsTable(),
		nodes:      newNodesTable(),
		partitions: newPartitionsTable(),
		queryVP:    viewport.New(0, 0),
		composer:   newComposer(),
		spinner:    s,
		help:       help.New(),
	}
}

// refreshQueries pulls the latest recorded queries from the client into the
// query inspector viewport.
func (m *Model) refreshQueries() {
	if m.client == nil {
		return
	}
	m.queries = m.client.Queries()
	m.queryVP.SetContent(queryLog(m.queries))
	m.queryVP.GotoBottom()
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

// partitionsMsg carries the partitions gathered from the cluster.
type partitionsMsg struct {
	names []string
	err   error
}

// gatherPartitionsCmd fetches the cluster partitions so the builder can offer
// them as options.
func gatherPartitionsCmd(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		names, err := c.Partitions(ctx)
		return partitionsMsg{names: names, err: err}
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
			m.refreshQueries()
			return m, nil
		}
		m.status = "Connected"
		m.version = msg.info.API
		m.release = msg.info.Slurm.Release
		m.warnings = msg.info.Warnings
		m.composer.setVersion(msg.info.API)
		m.refreshQueries()
		return m, gatherPartitionsCmd(m.client)

	case partitionsMsg:
		if msg.err == nil {
			m.composer.setPartitionOptions(msg.names)
		}
		return m, nil

	case composerRunMsg:
		m.refreshQueries()
		if len(m.queries) > 0 {
			last := m.queries[len(m.queries)-1]
			status := "—"
			if last.StatusCode > 0 {
				status = fmt.Sprintf("%d", last.StatusCode)
			}
			m.composer.setResult(
				statusColor(last.StatusCode).Render(status)+
					detailStyle.Render(" · "+last.Duration.Round(time.Millisecond).String()),
				string(last.Body),
			)
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
		}

		// Tab navigation available from every tab (including Query).
		switch msg.String() {
		case "esc":
			m.active = 0
			m.focusTab()
			return m, nil
		case "[", "shift+tab":
			m.active = (m.active - 1 + len(m.tabs)) % len(m.tabs)
			m.focusTab()
			return m, nil
		case "]", "tab":
			m.active = (m.active + 1) % len(m.tabs)
			m.focusTab()
			return m, nil
		}

		// Query tab: remaining keys (arrows, numbers, letters) drive the
		// composer and its panels.
		if m.active == 4 {
			return m.handleQueryTabKey(msg)
		}

		// Data tabs: arrows switch tabs.
		switch {
		case key.Matches(msg, keys.NextTab):
			m.active = (m.active + 1) % len(m.tabs)
			m.focusTab()
			return m, nil
		case key.Matches(msg, keys.PrevTab):
			m.active = (m.active - 1 + len(m.tabs)) % len(m.tabs)
			m.focusTab()
			return m, nil
		}

		// Delegate remaining keys to the active panel (table).
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
		m.layout()
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

// innerWidth returns the content width available inside the outer frame.
func (m Model) innerWidth() int {
	w := m.width - 4
	if w < 40 {
		w = 40
	}
	return w
}

// layout recomputes all responsive sizes based on the current terminal size.
func (m *Model) layout() {
	inner := m.innerWidth()
	h := m.tableHeight()

	fitTable(&m.jobs, jobsColumns, inner)
	fitTable(&m.nodes, nodesColumns, inner)
	fitTable(&m.partitions, partitionsColumns, inner)

	m.jobs.SetHeight(h)
	m.nodes.SetHeight(h)
	m.partitions.SetHeight(h)

	panel := inner - 4
	if panel < 30 {
		panel = 30
	}

	// Split the Query tab vertical space between composer (60%) and history.
	queryH := h - 2
	if queryH < 4 {
		queryH = 4
	}
	composerBudget := int(float64(queryH) * 0.6)
	if composerBudget < 2 {
		composerBudget = 2
	}
	leftW := int(float64(panel) * 0.55)
	rightW := panel - leftW
	m.composer.builder.Width = leftW
	m.composer.builder.Height = composerBudget
	m.composer.output.Width = rightW
	m.composer.output.Height = composerBudget
	m.composer.rebuild()

	m.queryVP.Width = panel
	m.queryVP.Height = queryH - composerBudget - 3
	if m.queryVP.Height < 1 {
		m.queryVP.Height = 1
	}

	m.refreshQueries()
}

// Query tab focus states.
const (
	focusBuilder = iota
	focusResponse
	focusHistory
)

// handleQueryTabKey routes keys within the Query tab based on the focused
// panel. 'f' cycles focus; 'r' runs the built request from anywhere.
func (m Model) handleQueryTabKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "f":
		m.queryFocus = (m.queryFocus + 1) % 3
		return m, nil
	case "r":
		return m, m.composer.run(m.client)
	}

	switch m.queryFocus {
	case focusBuilder:
		var cmd tea.Cmd
		m.composer, cmd = m.composer.Update(msg)
		return m, cmd
	case focusResponse:
		m.composer.output, _ = m.composer.output.Update(msg)
		return m, nil
	case focusHistory:
		m.queryVP, _ = m.queryVP.Update(msg)
		return m, nil
	}
	return m, nil
}

// queryTabView renders the Query tab: builder and output side by side, with
// the request history below. The focused panel is highlighted.
func (m Model) queryTabView(width int) string {
	panel := width - 4
	if panel < 30 {
		panel = 30
	}
	leftW := int(float64(panel) * 0.55)
	rightW := panel - leftW

	builder := focusedPanel(m.queryFocus == focusBuilder, composerPanelStyle).Width(leftW).Render(
		panelTitleStyle.Render("Builder") + "\n" + m.composer.builder.View(),
	)
	output := focusedPanel(m.queryFocus == focusResponse, outputPanelStyle).Width(rightW).Render(
		panelTitleStyle.Render("Response") + "\n" + m.composer.output.View(),
	)
	top := lipgloss.JoinHorizontal(lipgloss.Top, builder, output)

	history := focusedPanel(m.queryFocus == focusHistory, historyPanelStyle).Width(panel).Render(
		panelTitleStyle.Render("Request history") + "\n" + m.queryVP.View(),
	)
	return lipgloss.JoinVertical(lipgloss.Left, top, history)
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
func (m Model) contentView(width int) string {
	switch m.active {
	case 0:
		return dashboardView(width)
	case 1:
		return m.jobs.View()
	case 2:
		return m.nodes.View()
	case 3:
		return m.partitions.View()
	case 4:
		return m.queryTabView(width)
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

	innerWidth := m.innerWidth()
	divider := dividerStyle.Render(strings.Repeat("─", innerWidth))

	lines := []string{
		m.headerView(innerWidth),
		divider,
		m.tabsView(),
		divider,
		m.contentView(innerWidth),
	}
	for _, w := range m.warnings {
		lines = append(lines, warningStyle.Render("warning: "+w.Description))
	}
	lines = append(lines, divider, m.help.View(keys))

	block := lipgloss.JoinVertical(lipgloss.Left, lines...)
	framed := frameStyle.Render(block)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, framed)
}
