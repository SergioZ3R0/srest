package ui

import (
	"errors"
	"strings"
	"testing"

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
		"Connection successful!",
		"v0.0.45",
		"26.05.2",
		"Ignored field",
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
	_, cmd := m.Update(statusMsg{})
	if cmd != nil {
		t.Errorf("expected nil cmd, got %v", cmd)
	}
}
