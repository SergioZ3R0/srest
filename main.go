// srest es un cliente TUI para interactuar remotamente con la REST API de Slurm.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/SergioZ3R0/srest/internal/api"
	"github.com/SergioZ3R0/srest/internal/config"
	"github.com/SergioZ3R0/srest/internal/ui"
)

func main() {
	cfg := config.Load()

	client := api.New(cfg.URL, cfg.JWT, cfg.Username)

	program := tea.NewProgram(
		ui.New(client),
		tea.WithAltScreen(),
	)

	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "srest: %v\n", err)
		os.Exit(1)
	}
}
