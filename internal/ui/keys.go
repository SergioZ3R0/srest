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
	Up      key.Binding
	Down    key.Binding
	PageUp  key.Binding
	PageDn  key.Binding
	HalfUp  key.Binding
	HalfDn  key.Binding
	GoTop   key.Binding
	GoBot   key.Binding
	Select  key.Binding
	Filter  key.Binding
	Refresh key.Binding
}

// ShortHelp returns the bindings shown in the footer by default.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.PrevTab, k.NextTab, k.Home, k.Help, k.Quit}
}

// FullHelp returns the bindings shown when the help is expanded.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.NextTab, k.PrevTab, k.Home, k.Help, k.Quit},
		{k.Up, k.Down, k.PageUp, k.PageDn},
		{k.HalfUp, k.HalfDn, k.GoTop, k.GoBot},
		{k.Select, k.Filter, k.Refresh},
	}
}

var keys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	NextTab: key.NewBinding(
		key.WithKeys("tab", "]"),
		key.WithHelp("tab/]", "next tab"),
	),
	PrevTab: key.NewBinding(
		key.WithKeys("shift+tab", "["),
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
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("pgup", "b"),
		key.WithHelp("pgup/b", "page up"),
	),
	PageDn: key.NewBinding(
		key.WithKeys("pgdown", "f"),
		key.WithHelp("pgdn/f", "page down"),
	),
	HalfUp: key.NewBinding(
		key.WithKeys("ctrl+u"),
		key.WithHelp("ctrl+u", "½ page up"),
	),
	HalfDn: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "½ page down"),
	),
	GoTop: key.NewBinding(
		key.WithKeys("home", "g"),
		key.WithHelp("g/home", "top"),
	),
	GoBot: key.NewBinding(
		key.WithKeys("end", "G"),
		key.WithHelp("G/end", "bottom"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
}
