package ui

import (
	"github.com/charmbracelet/bubbles/key"
)

// keyMap defines the key bindings for srest's UI.
type keyMap struct {
	Quit    key.Binding
	NextTab key.Binding
	PrevTab key.Binding
	Home    key.Binding
	Help    key.Binding
}

// ShortHelp returns the bindings shown in the footer by default.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.PrevTab, k.NextTab, k.Home, k.Help, k.Quit}
}

// FullHelp returns the bindings shown when the help is expanded.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.PrevTab, k.NextTab, k.Home},
		{k.Help, k.Quit},
	}
}

var keys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	NextTab: key.NewBinding(
		key.WithKeys("tab", "]", "right", "l"),
		key.WithHelp("tab/]", "next tab"),
	),
	PrevTab: key.NewBinding(
		key.WithKeys("shift+tab", "[", "left", "h"),
		key.WithHelp("shift+tab/[", "prev tab"),
	),
	Home: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "dashboard"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
}
