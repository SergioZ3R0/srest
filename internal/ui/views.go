package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"

	"github.com/SergioZ3R0/srest/internal/api"
)

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

// newJobsTable builds the Jobs panel with placeholder rows.
func newJobsTable() table.Model {
	columns := []table.Column{
		{Title: "JobID", Width: 10},
		{Title: "Name", Width: 16},
		{Title: "User", Width: 10},
		{Title: "State", Width: 12},
		{Title: "Time", Width: 12},
		{Title: "Nodes", Width: 10},
	}
	rows := []table.Row{
		{"42", "sim-train", "alice", "RUNNING", "01:23:45", "c[1-2]"},
		{"43", "preprocess", "bob", "PENDING", "0:00", ""},
		{"44", "inference", "alice", "COMPLETED", "00:12:30", "c1"},
		{"45", "etl", "carol", "RUNNING", "00:45:10", "c2"},
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(8),
	)
	t.SetStyles(tableStyles())
	return t
}

// newNodesTable builds the Nodes panel with placeholder rows.
func newNodesTable() table.Model {
	columns := []table.Column{
		{Title: "Name", Width: 12},
		{Title: "State", Width: 10},
		{Title: "CPUs", Width: 8},
		{Title: "Memory", Width: 12},
		{Title: "Partition", Width: 14},
	}
	rows := []table.Row{
		{"c1", "idle", "32", "64GB", "normal"},
		{"c2", "alloc", "32", "64GB", "normal"},
		{"c3", "mixed", "32", "128GB", "gpu"},
		{"c4", "down", "32", "64GB", "normal"},
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(8),
	)
	t.SetStyles(tableStyles())
	return t
}

// newPartitionsTable builds the Partitions panel with placeholder rows.
func newPartitionsTable() table.Model {
	columns := []table.Column{
		{Title: "Name", Width: 12},
		{Title: "Avail", Width: 8},
		{Title: "Nodes", Width: 8},
		{Title: "TimeLimit", Width: 12},
		{Title: "State", Width: 10},
	}
	rows := []table.Row{
		{"normal", "up", "3", "01:00:00", "idle"},
		{"gpu", "up", "1", "04:00:00", "idle"},
		{"debug", "up", "2", "00:30:00", "idle"},
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(8),
	)
	t.SetStyles(tableStyles())
	return t
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

// dashboardView renders the Dashboard panel with placeholder stat cards.
func dashboardView() string {
	title := panelTitleStyle.Render("Cluster overview")

	cards := lipgloss.JoinHorizontal(
		lipgloss.Top,
		statCard("Nodes", "4"),
		statCard("Jobs", "12"),
		statCard("Partitions", "3"),
		statCard("CPU Load", "37%"),
	)

	return lipgloss.JoinVertical(lipgloss.Left, title, cards)
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
func queryLog(queries []api.Query) string {
	if len(queries) == 0 {
		return detailStyle.Render("No queries yet. Connect to a cluster to see traffic.")
	}

	var sb strings.Builder
	for i, q := range queries {
		if i > 0 {
			sb.WriteString("\n")
		}
		status := "—"
		if q.StatusCode > 0 {
			status = fmt.Sprintf("%d", q.StatusCode)
		}
		method := queryMethodStyle.Render(q.Method)
		code := statusColor(q.StatusCode).Render(status)
		dur := queryDetailStyle.Render(q.Duration.Round(time.Millisecond).String())

		sb.WriteString(fmt.Sprintf("%s  %s  %-8s  %s",
			method, code, dur, q.URL))

		if q.Error != nil {
			sb.WriteString("\n  " + errorStyle.Render("error: "+q.Error.Error()))
		}
		for _, w := range q.Warnings {
			sb.WriteString("\n  " + warningStyle.Render("warning: "+w.Description))
		}
		for _, e := range q.Errors {
			sb.WriteString("\n  " + warningStyle.Render("api error: "+e.Description))
		}
	}
	return sb.String()
}
