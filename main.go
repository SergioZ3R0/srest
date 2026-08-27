// srest is a TUI client for interacting remotely with the Slurm REST API.
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

	// If the user pinned an explicit version, use it; otherwise the UI will
	// auto-detect it.
	if cfg.APIVersion != "" {
		v, err := api.ParseVersion(cfg.APIVersion)
		if err != nil {
			fmt.Fprintf(os.Stderr, "srest: %v\n", err)
			os.Exit(1)
		}
		client.SetVersion(v)
	}

	program := tea.NewProgram(
		ui.New(client),
		tea.WithAltScreen(),
	)

	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "srest: %v\n", err)
		os.Exit(1)
	}
}
