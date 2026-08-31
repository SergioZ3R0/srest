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
	"github.com/charmbracelet/bubbles/textinput"
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
	tabs           []string
	active         int
	jobs           table.Model
	nodes          table.Model
	partitions     table.Model
	queries        []api.Query
	queryVP        viewport.Model
	composer       composer
	queryFocus     int
	jobDetail      api.JobDetail
	jobDetailVP    viewport.Model
	selectedJob    uint32
	jobsData       []api.JobInfo
	nodesData      []api.NodeInfo
	partitionsData []api.PartitionInfo
	accountsData   []api.AccountInfo
	searchInput    textinput.Model
	searching      bool
	rawInput       textinput.Model
	spinner        spinner.Model
	help           help.Model
	width          int
	height         int
}

// New returns a Model ready to be used with an API client.
func New(client *api.Client) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return Model{
		client:      client,
		status:      "Connecting to Slurm...",
		tabs:        []string{"Dashboard", "Jobs", "Nodes", "Partitions", "Query"},
		active:      0,
		jobs:        newJobsTable(),
		nodes:       newNodesTable(),
		partitions:  newPartitionsTable(),
		queryVP:     viewport.New(0, 0),
		composer:    newComposer(),
		spinner:     s,
		help:        help.New(),
		jobDetailVP: viewport.New(0, 0),
		searchInput: textinput.New(),
		rawInput:    textinput.New(),
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

// jobsMsg carries the jobs fetched from the cluster.
type jobsMsg struct {
	jobs []api.JobInfo
	err  error
}

// jobsCmd fetches the jobs visible to the user.
func jobsCmd(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		jobs, err := c.Jobs(ctx)
		return jobsMsg{jobs: jobs, err: err}
	}
}

// setJobs fills the Jobs table with the given jobs.
func (m *Model) setJobs(jobs []api.JobInfo) {
	m.jobsData = jobs
	m.jobs.SetRows(jobRows(jobs, m.searchInput.Value()))
	m.refreshQueries()
}

// setNodes fills the Nodes table.
func (m *Model) setNodes(nodes []api.NodeInfo) {
	m.nodesData = nodes
	m.nodes.SetRows(nodeRows(nodes, m.searchInput.Value()))
}

// setPartitions fills the Partitions table.
func (m *Model) setPartitions(parts []api.PartitionInfo) {
	m.partitionsData = parts
	m.partitions.SetRows(partitionRows(parts, m.searchInput.Value()))
}

// rowMatches reports whether a row contains q (case-insensitive).
func rowMatches(row table.Row, q string) bool {
	if q == "" {
		return true
	}
	for _, cell := range row {
		if strings.Contains(strings.ToLower(cell), q) {
			return true
		}
	}
	return false
}

// jobRows builds the job table rows, optionally filtered by q.
func jobRows(data []api.JobInfo, q string) []table.Row {
	rows := make([]table.Row, 0, len(data))
	for _, j := range data {
		row := table.Row{
			fmt.Sprintf("%d", j.JobID),
			j.Name,
			j.User,
			strings.Join(j.State, ","),
			formatDuration(j.RunTime),
			j.Nodes,
			j.Partition,
		}
		if !rowMatches(row, q) {
			continue
		}
		rows = append(rows, row)
	}
	return rows
}

// nodeRows builds the node table rows, optionally filtered by q.
func nodeRows(data []api.NodeInfo, q string) []table.Row {
	rows := make([]table.Row, 0, len(data))
	for _, n := range data {
		row := table.Row{
			n.Name,
			strings.Join(n.State, ","),
			fmt.Sprintf("%d", n.CPUs),
			fmt.Sprintf("%dMB", n.RealMemory),
			strings.Join(n.Partitions, ","),
		}
		if !rowMatches(row, q) {
			continue
		}
		rows = append(rows, row)
	}
	return rows
}

// partitionRows builds the partition table rows, optionally filtered by q.
func partitionRows(data []api.PartitionInfo, q string) []table.Row {
	rows := make([]table.Row, 0, len(data))
	for _, p := range data {
		row := table.Row{
			p.Name,
			p.Nodes.Configured,
			fmt.Sprintf("%d", p.Nodes.Total),
		}
		if !rowMatches(row, q) {
			continue
		}
		rows = append(rows, row)
	}
	return rows
}

// jobDetailMsg carries the full detail of a single job.
type jobDetailMsg struct {
	job api.JobDetail
	err error
}

// jobDetailCmd fetches the full detail of a job.
func jobDetailCmd(c *api.Client, id uint32) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		jd, err := c.Job(ctx, id)
		return jobDetailMsg{job: jd, err: err}
	}
}

// setJobDetail fills the job detail panel.
func (m *Model) setJobDetail(jd api.JobDetail) {
	m.jobDetail = jd
	m.jobDetailVP.SetContent(jobDetailView(jd))
	m.jobDetailVP.GotoTop()
}

// cursorJobID returns the job ID of the currently selected row (0 if none).
func (m Model) cursorJobID() uint32 {
	rows := m.jobs.Rows()
	idx := m.jobs.Cursor()
	if idx < 0 || idx >= len(rows) {
		return 0
	}
	var id uint32
	_, _ = fmt.Sscanf(rows[idx][0], "%d", &id)
	return id
}

// nodesMsg carries the nodes fetched from the cluster.
type nodesMsg struct {
	nodes []api.NodeInfo
	err   error
}

func nodesCmd(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		n, err := c.Nodes(ctx)
		return nodesMsg{nodes: n, err: err}
	}
}

// partitionsMsgFull carries detailed partitions.
type partitionsMsgFull struct {
	parts []api.PartitionInfo
	err   error
}

func partitionsCmd(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		p, err := c.PartitionList(ctx)
		return partitionsMsgFull{parts: p, err: err}
	}
}

// accountsMsg carries the accounts fetched from the cluster.
type accountsMsg struct {
	accounts []api.AccountInfo
	err      error
}

func accountsCmd(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		a, err := c.Accounts(ctx)
		return accountsMsg{accounts: a, err: err}
	}
}

// startSearch opens the table filter for the active data tab.
func (m *Model) startSearch() {
	m.searching = true
	m.searchInput.Focus()
	m.searchInput.SetValue("")
}

// handleSearchKey routes keys while a table filter is active.
func (m Model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "esc":
		m.searching = false
		m.searchInput.Blur()
		m.applyFilter("")
		return m, nil
	}
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	m.applyFilter(m.searchInput.Value())
	return m, cmd
}

// applyFilter re-renders the active table's rows filtered by q.
func (m *Model) applyFilter(q string) {
	switch m.active {
	case 1:
		m.jobs.SetRows(jobRows(m.jobsData, q))
	case 2:
		m.nodes.SetRows(nodeRows(m.nodesData, q))
	case 3:
		m.partitions.SetRows(partitionRows(m.partitionsData, q))
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
		return m, tea.Batch(
			gatherPartitionsCmd(m.client),
			jobsCmd(m.client),
			nodesCmd(m.client),
			partitionsCmd(m.client),
			accountsCmd(m.client),
		)

	case partitionsMsg:
		if msg.err == nil {
			m.composer.setPartitionOptions(msg.names)
		}
		return m, nil

	case nodesMsg:
		if msg.err == nil {
			m.setNodes(msg.nodes)
		}
		m.refreshQueries()
		return m, nil

	case partitionsMsgFull:
		if msg.err == nil {
			m.setPartitions(msg.parts)
		}
		m.refreshQueries()
		return m, nil

	case accountsMsg:
		if msg.err == nil {
			m.accountsData = msg.accounts
			names := make([]string, 0, len(msg.accounts))
			for _, a := range msg.accounts {
				if a.Name != "" {
					names = append(names, a.Name)
				}
			}
			m.composer.setAccountOptions(names)
		}
		return m, nil

	case jobsMsg:
		if msg.err == nil {
			m.setJobs(msg.jobs)
			// Auto-select the first job so the detail panel is populated.
			if id := m.cursorJobID(); id != 0 {
				m.selectedJob = id
				return m, jobDetailCmd(m.client, id)
			}
		}
		m.refreshQueries()
		return m, nil

	case jobDetailMsg:
		if msg.err == nil {
			m.setJobDetail(msg.job)
		}
		m.refreshQueries()
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

		// While editing a value in the Query builder, only the composer
		// handles keys (no global shortcuts, so 'r' can be typed).
		if m.active == 4 && m.composer.editing {
			var cmd tea.Cmd
			m.composer, cmd = m.composer.Update(msg)
			return m, cmd
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

		// While a table filter is active, keys go to the search input.
		if m.searching {
			return m.handleSearchKey(msg)
		}

		// Delegate remaining keys to the active panel (table). Arrows here
		// navigate rows; tab switching is handled above.
		switch m.active {
		case 1:
			if msg.String() == "/" {
				m.startSearch()
				return m, nil
			}
			if msg.String() == "r" {
				return m, jobsCmd(m.client)
			}
			var cmd tea.Cmd
			m.jobs, cmd = m.jobs.Update(msg)
			if id := m.cursorJobID(); id != 0 && id != m.selectedJob {
				m.selectedJob = id
				cmd = tea.Batch(cmd, jobDetailCmd(m.client, id))
			}
			return m, cmd
		case 2:
			if msg.String() == "/" {
				m.startSearch()
				return m, nil
			}
			m.nodes, _ = m.nodes.Update(msg)
		case 3:
			if msg.String() == "/" {
				m.startSearch()
				return m, nil
			}
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
	m.queryVP.Height = queryH - composerBudget - 5
	if m.queryVP.Height < 1 {
		m.queryVP.Height = 1
	}

	// Jobs detail viewport.
	tableW := int(float64(panel) * 0.6)
	m.jobDetailVP.Width = panel - tableW
	m.jobDetailVP.Height = h
	if m.jobDetailVP.Height < 2 {
		m.jobDetailVP.Height = 2
	}

	m.refreshQueries()
}

// Query tab focus states.
const (
	focusBuilder = iota
	focusResponse
	focusHistory
	focusRaw
)

// rawRunCmd issues a request with a manually written/typed path and reuses
// the same result handling as the composer.
func rawRunCmd(c *api.Client, path string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		return composerRunMsg{err: c.Get(ctx, path, nil)}
	}
}

// handleQueryTabKey routes keys within the Query tab based on the focused
// panel. 'f' cycles focus; 'r' runs the built request from anywhere.
func (m Model) handleQueryTabKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "f":
		m.queryFocus = (m.queryFocus + 1) % 4
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
	case focusRaw:
		if msg.String() == "enter" {
			path := strings.TrimSpace(m.rawInput.Value())
			if path == "" {
				return m, nil
			}
			return m, rawRunCmd(m.client, path)
		}
		var cmd tea.Cmd
		m.rawInput, cmd = m.rawInput.Update(msg)
		return m, cmd
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

	raw := focusedPanel(m.queryFocus == focusRaw, historyPanelStyle).Width(panel).Render(
		panelTitleStyle.Render("Custom query") + "\n" +
			searchStyle.Render("Path: "+m.rawInput.View()) + " enter=run",
	)

	history := focusedPanel(m.queryFocus == focusHistory, historyPanelStyle).Width(panel).Render(
		panelTitleStyle.Render("Request history") + "\n" + m.queryVP.View(),
	)
	return lipgloss.JoinVertical(lipgloss.Left, top, raw, history)
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
		return m.dashboardView(width)
	case 1:
		return m.jobsView(width)
	case 2:
		return m.nodesView(width)
	case 3:
		return m.partitionsView(width)
	case 4:
		return m.queryTabView(width)
	default:
		return ""
	}
}

// dashboardView renders the Dashboard with real cluster statistics.
func (m Model) dashboardView(width int) string {
	title := panelTitleStyle.Render("Cluster overview")

	nodesUp, nodesDown := 0, 0
	for _, n := range m.nodesData {
		if hasState(n.State, "DOWN", "ERROR", "DRAIN", "UNKNOWN") {
			nodesDown++
		} else {
			nodesUp++
		}
	}
	jobsRunning, jobsPending, jobsCompleted, jobsFailed := 0, 0, 0, 0
	for _, j := range m.jobsData {
		for _, s := range j.State {
			switch s {
			case "RUNNING":
				jobsRunning++
			case "PENDING":
				jobsPending++
			case "COMPLETED":
				jobsCompleted++
			case "FAILED", "CANCELLED", "TIMEOUT":
				jobsFailed++
			}
		}
	}

	cards := []string{
		statCard("Nodes", fmt.Sprintf("%d/%d", nodesUp, len(m.nodesData))),
		statCard("Jobs running", fmt.Sprintf("%d", jobsRunning)),
		statCard("Jobs", fmt.Sprintf("%d", len(m.jobsData))),
		statCard("Partitions", fmt.Sprintf("%d", len(m.partitionsData))),
		statCard("Accounts", fmt.Sprintf("%d", len(m.accountsData))),
	}

	stats := detailStyle.Render(fmt.Sprintf(
		"Jobs: %d running · %d pending · %d completed · %d failed    Nodes: %d down",
		jobsRunning, jobsPending, jobsCompleted, jobsFailed, nodesDown))

	if width < 90 {
		return lipgloss.JoinVertical(lipgloss.Left, title, lipgloss.JoinVertical(lipgloss.Left, cards...), stats)
	}
	return lipgloss.JoinVertical(lipgloss.Left, title, lipgloss.JoinHorizontal(lipgloss.Top, cards...), stats)
}

// nodesView renders the Nodes tab: table on the left, selected node detail on
// the right.
func (m Model) nodesView(width int) string {
	panel := width - 4
	if panel < 30 {
		panel = 30
	}
	tableW := int(float64(panel) * 0.6)
	detailW := panel - tableW

	tablePanel := composerPanelStyle.Width(tableW).Render(
		panelTitleStyle.Render("Nodes (/ search)") + "\n" + m.searchBar() + m.nodes.View(),
	)
	detail := m.nodeDetailView()
	detailPanel := outputPanelStyle.Width(detailW).Render(
		panelTitleStyle.Render("Node detail") + "\n" + detail,
	)
	return lipgloss.JoinHorizontal(lipgloss.Top, tablePanel, detailPanel)
}

// searchBar renders the filter input when search is active.
func (m Model) searchBar() string {
	if !m.searching {
		return ""
	}
	return searchStyle.Render("Filter: "+m.searchInput.View()) + "\n"
}

// nodeDetailView returns the detail viewport of the selected node.
func (m Model) nodeDetailView() string {
	idx := m.nodes.Cursor()
	if idx >= 0 && idx < len(m.nodesData) {
		return nodeDetailView(m.nodesData[idx])
	}
	return detailStyle.Render("Select a node to see its details.")
}

// partitionsView renders the Partitions tab: table on the left, selected
// partition detail on the right.
func (m Model) partitionsView(width int) string {
	panel := width - 4
	if panel < 30 {
		panel = 30
	}
	tableW := int(float64(panel) * 0.6)
	detailW := panel - tableW

	tablePanel := composerPanelStyle.Width(tableW).Render(
		panelTitleStyle.Render("Partitions (/ search)") + "\n" + m.searchBar() + m.partitions.View(),
	)
	idx := m.partitions.Cursor()
	detail := detailStyle.Render("Select a partition to see its details.")
	if idx >= 0 && idx < len(m.partitionsData) {
		detail = partitionDetailView(m.partitionsData[idx])
	}
	detailPanel := outputPanelStyle.Width(detailW).Render(
		panelTitleStyle.Render("Partition detail") + "\n" + detail,
	)
	return lipgloss.JoinHorizontal(lipgloss.Top, tablePanel, detailPanel)
}

// jobsView renders the Jobs tab: the jobs table on the left and the selected
// job's detail on the right.
func (m Model) jobsView(width int) string {
	panel := width - 4
	if panel < 30 {
		panel = 30
	}
	tableW := int(float64(panel) * 0.6)
	detailW := panel - tableW

	tablePanel := composerPanelStyle.Width(tableW).Render(
		panelTitleStyle.Render("Jobs (/ search · r refresh)") + "\n" + m.searchBar() + m.jobs.View(),
	)
	detailPanel := outputPanelStyle.Width(detailW).Render(
		panelTitleStyle.Render("Job detail") + "\n" + m.jobDetailVP.View(),
	)
	return lipgloss.JoinHorizontal(lipgloss.Top, tablePanel, detailPanel)
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
