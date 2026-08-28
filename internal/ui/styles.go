package ui

import "github.com/charmbracelet/lipgloss"

var (
	// titleStyle formats the application header.
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12"))

	// statusStyle formats the transient status message (e.g. connecting).
	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15"))

	// successStyle formats the success indicator.
	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("10"))

	// errorStyle formats the error message in red.
	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("9"))

	// detailStyle formats the detail lines (API and Slurm versions).
	detailStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7"))

	// warningStyle formats the warnings returned by slurmrestd.
	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("3"))

	// tabStyle formats an inactive tab.
	tabStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Padding(0, 2)

	// activeTabStyle formats the currently selected tab.
	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("57")).
			Padding(0, 2)

	// panelTitleStyle formats a panel's title.
	panelTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			Margin(0, 0, 1, 0)

	// cardStyle wraps a dashboard stat card.
	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 2).
			MarginRight(1)

	// cardLabelStyle formats a stat card's label.
	cardLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	// cardValueStyle formats a stat card's value.
	cardValueStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15"))

	// frameStyle draws the outer frame around the whole UI.
	frameStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)

	// dividerStyle renders horizontal separator lines.
	dividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	// queryMethodStyle formats the HTTP method in the query log.
	queryMethodStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("12"))

		// queryDetailStyle formats secondary query log details (duration).
	queryDetailStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8"))

	// composerEndpointActive formats the selected endpoint.
	composerEndpointActive = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("57")).
				Padding(0, 1)

	// composerEndpoint formats an unselected endpoint.
	composerEndpoint = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8")).
				Padding(0, 1)

	// composerParamName formats a parameter name.
	composerParamName = lipgloss.NewStyle().
				Foreground(lipgloss.Color("12")).
				Width(14)

	// composerParamValue formats a parameter value.
	composerParamValue = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15"))

	// composerParamCursor highlights the selected parameter.
	composerParamCursor = lipgloss.NewStyle().
				Foreground(lipgloss.Color("229")).
				Background(lipgloss.Color("57"))

	// composerURL formats the built URL.
	composerURL = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	// composerHint formats the key hints.
	composerHint = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	// composerPanelStyle draws a bordered frame around the composer.
	composerPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("63")).
				Padding(0, 1)

	// outputPanelStyle draws a bordered frame around the response output.
	outputPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("63")).
				Padding(0, 1).
				Margin(0, 0, 1, 1)

	// historyPanelStyle draws a bordered frame around the request history.
	historyPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("63")).
				Padding(0, 1)
)
