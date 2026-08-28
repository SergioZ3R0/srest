package ui

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
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
