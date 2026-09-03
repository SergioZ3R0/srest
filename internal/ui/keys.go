package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
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

// renderKey renders a key binding as "key  desc".
func renderKey(b key.Binding, style lipgloss.Style) string {
	return style.Render(b.Help().Key) + " " + b.Help().Desc
}

// helpView renders the full help with section labels.
func (k keyMap) helpView() string {
	section := lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	title := lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true).Underline(true)

	row := func(bindings ...key.Binding) string {
		parts := make([]string, 0, len(bindings))
		for _, b := range bindings {
			parts = append(parts, renderKey(b, dim))
		}
		return strings.Join(parts, "   ")
	}

	var lines []string
	lines = append(lines, title.Render("Global"))
	lines = append(lines, "  "+section.Render("Tabs")+":     "+row(k.NextTab, k.PrevTab, k.Home, k.Quit))
	lines = append(lines, "  "+section.Render("Help")+":      "+renderKey(k.Help, dim))
	lines = append(lines, "")
	lines = append(lines, title.Render("Tables (Jobs / Nodes / Partitions)"))
	lines = append(lines, "  "+section.Render("Move")+":      "+row(k.Up, k.Down, k.PageUp, k.PageDn))
	lines = append(lines, "  "+section.Render("Jump")+":       "+row(k.HalfUp, k.HalfDn, k.GoTop, k.GoBot))
	lines = append(lines, "  "+section.Render("Actions")+":    "+row(k.Select, k.Filter, k.Refresh))
	lines = append(lines, "  "+dim.Render("enter on Partitions → filter jobs by partition"))
	lines = append(lines, "")
	lines = append(lines, title.Render("Query tab"))
	lines = append(lines, "  "+dim.Render("f")+"           cycle panels (Builder / Response / History / Raw)")
	lines = append(lines, "  "+dim.Render("r")+"           run request")
	lines = append(lines, "  "+dim.Render("1 / 2 / 3")+"    select endpoint (ping / get jobs / submit)")
	lines = append(lines, "  "+dim.Render("↑/k  ↓/j")+"    move cursor in builder")
	lines = append(lines, "  "+dim.Render("←/h  →/l")+"    cycle parameter options")
	lines = append(lines, "  "+dim.Render("enter")+"        edit selected parameter")
	lines = append(lines, "  "+dim.Render("e")+"           open $EDITOR for script param")
	lines = append(lines, "  "+dim.Render("del / x")+"     clear parameter value")
	lines = append(lines, "")
	lines = append(lines, title.Render("Search"))
	lines = append(lines, "  "+dim.Render("/")+"            open filter")
	lines = append(lines, "  "+dim.Render("enter")+"        close filter (keeps active)")
	lines = append(lines, "  "+dim.Render("esc")+"          close filter and clear")
	lines = append(lines, "  "+dim.Render("ctrl+u")+"       clear input")

	return strings.Join(lines, "\n")
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
