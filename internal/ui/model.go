// Package ui contiene el modelo de Bubble Tea de srest.
//
// La interfaz consume el cliente de internal/api de forma asíncrona mediante
// comandos (tea.Cmd), de modo que la llamada HTTP nunca bloquea el renderizado.
package ui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/SergioZ3R0/srest/internal/api"
)

// Model representa el estado completo de la interfaz.
type Model struct {
	client *api.Client

	// status es el texto de estado que se muestra en pantalla.
	status string

	// err guarda el error del último ping, si lo hubo.
	err error

	// done indica que la petición de ping ya ha terminado.
	done bool

	width  int
	height int
}

// New devuelve un Model listo para usarse con un cliente de API.
func New(client *api.Client) Model {
	return Model{
		client: client,
		status: "Conectando a Slurm...",
	}
}

// statusMsg es el mensaje que el comando de ping devuelve al modelo.
// Un valor nil en err significa éxito.
type statusMsg struct {
	err error
}

// pingCmd ejecuta la petición de ping en segundo plano y devuelve el
// resultado como un statusMsg. Es un tea.Cmd, por lo que no bloquea la UI.
func pingCmd(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		return statusMsg{err: c.Ping(ctx)}
	}
}

// Init lanza la petición de ping en cuanto arranca el programa.
func (m Model) Init() tea.Cmd {
	return pingCmd(m.client)
}

// Update reacciona a los mensajes de Bubble Tea (teclado, red, ventana).
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case statusMsg:
		m.done = true
		if msg.err != nil {
			m.err = msg.err
			m.status = "Error de conexión"
		} else {
			m.status = "¡Conexión exitosa!"
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

// View renderiza el estado actual del modelo.
func (m Model) View() string {
	title := titleStyle.Render("srest")
	help := helpStyle.Render("Pulsa 'q' para salir")

	content := statusStyle.Render(m.status)
	if m.done && m.err != nil {
		content = errorStyle.Render(m.status + ": " + m.err.Error())
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		content,
		help,
	)
}
