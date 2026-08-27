package ui

import "github.com/charmbracelet/lipgloss"

var (
	// titleStyle formats the application header.
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			Padding(0, 0, 1, 0)

	// statusStyle formats the normal status message.
	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15"))

	// errorStyle formats the error message in red.
	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("9"))

	// helpStyle formats the bottom help line.
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Padding(1, 0, 0, 0)

	// detailStyle formats the detail lines (API and Slurm versions).
	detailStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7"))

	// warningStyle formats the warnings returned by slurmrestd.
	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("3"))
)
