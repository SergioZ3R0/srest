package ui

import "github.com/charmbracelet/lipgloss"

var (
	// titleStyle da formato al encabezado de la aplicación.
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			Padding(0, 0, 1, 0)

	// statusStyle da formato al mensaje de estado normal.
	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15"))

	// errorStyle da formato al mensaje de error en rojo.
	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("9"))

	// helpStyle da formato a la ayuda inferior.
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Padding(1, 0, 0, 0)
)
