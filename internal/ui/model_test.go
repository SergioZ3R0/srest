package ui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/SergioZ3R0/srest/internal/api"
)

func TestModelViewSuccess(t *testing.T) {
	m := New(nil)

	updated, _ := m.Update(statusMsg{
		info: api.PingInfo{
			API:   api.Version{Major: 0, Minor: 0, Micro: 45},
			Slurm: api.SlurmVersion{Release: "26.05.2"},
			Warnings: []api.Warning{
				{Description: "Ignored field", Source: "test"},
			},
		},
	})
	m = updated.(Model)

	view := m.View()
	for _, want := range []string{
		"Connected",
		"v0.0.45",
		"26.05.2",
		"Ignored field",
		"Dashboard",
		"Jobs",
		"Nodes",
		"Partitions",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("View() does not contain %q:\n%s", want, view)
		}
	}
}

func TestModelViewError(t *testing.T) {
	m := New(nil)

	updated, _ := m.Update(statusMsg{err: errors.New("boom")})
	m = updated.(Model)

	view := m.View()
	for _, want := range []string{"Connection error", "boom"} {
		if !strings.Contains(view, want) {
			t.Errorf("View() does not contain %q:\n%s", want, view)
		}
	}
}

func TestModelQuit(t *testing.T) {
	m := New(nil)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Fatal("expected a quit command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Errorf("expected tea.Quit, got %T", cmd())
	}
}
