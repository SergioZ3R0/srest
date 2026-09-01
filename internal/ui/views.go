package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"

	"github.com/SergioZ3R0/srest/internal/api"
)

// Base column definitions. Actual widths are recomputed dynamically to fit the
// terminal width (see fitTable).
var (
	jobsColumns = []table.Column{
		{Title: "JobID", Width: 8},
		{Title: "Name", Width: 14},
		{Title: "User", Width: 10},
		{Title: "State", Width: 12},
		{Title: "Time", Width: 10},
		{Title: "Nodes", Width: 12},
		{Title: "Partition", Width: 10},
	}

	nodesColumns = []table.Column{
		{Title: "Name", Width: 12},
		{Title: "State", Width: 14},
		{Title: "CPUs", Width: 8},
		{Title: "Memory", Width: 10},
		{Title: "Partitions", Width: 16},
	}

	partitionsColumns = []table.Column{
		{Title: "Name", Width: 12},
		{Title: "Nodes", Width: 18},
		{Title: "Total", Width: 8},
	}
)

// minColumnWidth is the smallest a column can shrink to when the terminal is
// narrow.
const minColumnWidth = 5

// fitColumns redistributes the given base column widths to fit within total,
// shrinking proportionally with a per-column floor. It never grows columns
// beyond their base width.
func fitColumns(base []int, total int) []int {
	if total <= 0 {
		return append([]int(nil), base...)
	}

	sum := 0
	for _, w := range base {
		sum += w
	}
	if sum <= total {
		return append([]int(nil), base...)
	}

	out := make([]int, len(base))
	// First pass: assign floors.
	remaining := total
	for i, w := range base {
		out[i] = w
		if out[i] > minColumnWidth {
			out[i] = minColumnWidth
		}
		remaining -= out[i]
	}

	// Recompute how much is left to distribute beyond the floors.
	if remaining > 0 {
		// Work out the total "flex" (base width above the floor).
		flex := 0
		for _, w := range base {
			if w > minColumnWidth {
				flex += w - minColumnWidth
			}
		}
		if flex > 0 {
			added := 0
			for i, w := range base {
				if w > minColumnWidth {
					add := int(float64(remaining) * float64(w-minColumnWidth) / float64(flex))
					out[i] += add
					added += add
				}
			}
			// Give any rounding remainder to the last flexible column.
			for i := len(base) - 1; i >= 0; i-- {
				if base[i] > minColumnWidth {
					out[i] += remaining - added
					break
				}
			}
		}
	}
	return out
}

// fitTable sets a table's columns and total width so it fits the given width.
func fitTable(t *table.Model, base []table.Column, width int) {
	baseW := make([]int, len(base))
	for i, c := range base {
		baseW[i] = c.Width
	}
	fitted := fitColumns(baseW, width)

	cols := make([]table.Column, len(base))
	for i, c := range base {
		cols[i] = table.Column{Title: c.Title, Width: fitted[i]}
	}
	t.SetColumns(cols)
	t.SetWidth(width)
}

// focusedPanel returns a copy of base with a highlighted border when active.
func focusedPanel(active bool, base lipgloss.Style) lipgloss.Style {
	if active {
		return base.BorderForeground(lipgloss.Color("45"))
	}
	return base
}

// tableStyles returns the shared style for all data tables.
func tableStyles() table.Styles {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57"))
	return s
}

// newJobsTable builds the Jobs panel. Rows are populated from the API.
func newJobsTable() table.Model {
	t := table.New(
		table.WithColumns(jobsColumns),
		table.WithFocused(true),
		table.WithHeight(8),
	)
	t.SetStyles(tableStyles())
	return t
}

// newNodesTable builds the Nodes panel. Rows are populated from the API.
func newNodesTable() table.Model {
	t := table.New(
		table.WithColumns(nodesColumns),
		table.WithFocused(true),
		table.WithHeight(8),
	)
	t.SetStyles(tableStyles())
	return t
}

// newPartitionsTable builds the Partitions panel. Rows are populated from the API.
func newPartitionsTable() table.Model {
	t := table.New(
		table.WithColumns(partitionsColumns),
		table.WithFocused(true),
		table.WithHeight(8),
	)
	t.SetStyles(tableStyles())
	return t
}

// formatDuration renders a number of seconds as HH:MM:SS (or M:SS).
func formatDuration(sec int64) string {
	if sec <= 0 {
		return "0:00"
	}
	h := sec / 3600
	m := (sec % 3600) / 60
	s := sec % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

// statCard renders a single dashboard stat card.
func statCard(label, value string) string {
	return cardStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			cardLabelStyle.Render(label),
			cardValueStyle.Render(value),
		),
	)
}

// hasState reports whether any of the wanted states is present.
func hasState(states api.StringList, wanted ...string) bool {
	for _, s := range states {
		for _, w := range wanted {
			if s == w {
				return true
			}
		}
	}
	return false
}

// nodeDetailView renders a node's details as key/value lines.
func nodeDetailView(n api.NodeInfo) string {
	f := func(name, val string) string {
		if val == "" {
			val = "—"
		}
		return composerParamName.Render(name) + "  " + composerParamValue.Render(val)
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		f("Name", n.Name),
		f("State", strings.Join(n.State, ",")),
		f("CPUs", fmt.Sprintf("%d", n.CPUs)),
		f("Memory", fmt.Sprintf("%dMB", n.RealMemory)),
		f("Allocated", fmt.Sprintf("%dMB", n.AllocMemory)),
		f("Partitions", strings.Join(n.Partitions, ",")),
	)
}

// partitionDetailView renders a partition's details as key/value lines.
func partitionDetailView(p api.PartitionInfo) string {
	f := func(name, val string) string {
		if val == "" {
			val = "—"
		}
		return composerParamName.Render(name) + "  " + composerParamValue.Render(val)
	}
	maxWall := "∞"
	if !p.Maximums.Time.Infinite && p.Maximums.Time.Number > 0 {
		maxWall = formatDuration(p.Maximums.Time.Number * 60)
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		f("Name", p.Name),
		f("Nodes", p.Nodes.Configured),
		f("Total", fmt.Sprintf("%d", p.Nodes.Total)),
		f("Max wall", maxWall),
	)
}

// jobDetailView renders the details of a single job as key/value lines.
func jobDetailView(j api.JobDetail) string {
	f := func(name, val string) string {
		if val == "" {
			val = "—"
		}
		return composerParamName.Render(name) + "  " + composerParamValue.Render(val)
	}

	timeLimit := "∞"
	if !j.TimeLimit.Infinite && j.TimeLimit.Number > 0 {
		timeLimit = formatDuration(j.TimeLimit.Number * 60)
	}
	startT := ""
	if j.StartTime.Number > 0 {
		startT = time.Unix(j.StartTime.Number, 0).Format("2006-01-02 15:04")
	}
	endT := ""
	if j.EndTime.Number > 0 {
		endT = time.Unix(j.EndTime.Number, 0).Format("2006-01-02 15:04")
	}

	lines := []string{
		f("Job ID", fmt.Sprintf("%d", j.JobID)),
		f("Name", j.Name),
		f("User", j.User),
		f("Account", j.Account),
		f("Partition", j.Partition),
		f("State", strings.Join(j.State, ",")),
		f("Time limit", timeLimit),
		f("Run time", formatDuration(j.RunTime)),
		f("Start", startT),
		f("End", endT),
		f("Nodes", j.Nodes),
		f("Node count", fmt.Sprintf("%d", j.NodeCount.Number)),
		f("CPUs", fmt.Sprintf("%d", j.CPUs.Number)),
		f("Stdout", j.StandardOutput),
		f("Stderr", j.StandardError),
	}

	// Only show exit info when the job has finished. While pending/running the
	// "exit_code" field holds a misleading default, and the real state is
	// already shown above.
	if jobFinished(j) {
		lines = append(lines,
			f("Exit code", strings.Join(j.ExitCode.Status, ",")),
			f("Return code", fmt.Sprintf("%d", j.ExitCode.ReturnCode.Number)),
		)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// jobFinished reports whether a job has reached a terminal state.
func jobFinished(j api.JobDetail) bool {
	for _, s := range j.State {
		switch s {
		case "COMPLETED", "FAILED", "CANCELLED", "TIMEOUT", "OUT_OF_MEMORY", "PREEMPTED", "NODE_FAIL", "DEADLINE":
			return true
		}
	}
	return false
}

// statusColor returns the style used to render an HTTP status code.
func statusColor(code int) lipgloss.Style {
	switch {
	case code >= 500:
		return errorStyle
	case code >= 400:
		return warningStyle
	case code >= 200 && code < 300:
		return successStyle
	default:
		return statusStyle
	}
}

// queryLog renders the recorded HTTP queries as a scrollable log.
// queryLog formats live client queries for display in the viewport.
// Kept as a fallback; the persistent historyView is used in production.
