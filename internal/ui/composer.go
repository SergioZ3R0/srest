package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/SergioZ3R0/srest/internal/api"
)

// param is a single editable query parameter for an endpoint. options holds
// selectable values (static or gathered from the cluster); when present they
// can be cycled with the left/right keys.
type param struct {
	name    string
	value   string
	options []string
}

// composerRunMsg is returned after the composer executes a request.
type composerRunMsg struct {
	err error
}

// endpoint describes a Slurm API endpoint and its supported query parameters.
// base is the API family: "slurm" (slurmctld) or "slurmdb" (accounting).
type endpoint struct {
	name   string
	method string
	base   string
	path   string
	params []param
}

// endpoints defines the Slurm REST endpoints exposed by the composer and the
// query parameters each one accepts (taken from the OpenAPI spec of the
// cluster). ping is first so it is selected by default. The rich job filters
// live under the slurmdb API, so jobs targets /slurmdb/vX/jobs.
var endpoints = []endpoint{
	{
		name:   "ping",
		method: "GET",
		base:   "slurm",
		path:   "/ping",
		params: []param{},
	},
	{
		name:   "jobs",
		method: "GET",
		base:   "slurmdb",
		path:   "/jobs",
		params: []param{
			{name: "state", options: []string{"RUNNING", "PENDING", "COMPLETED", "FAILED", "CANCELLED", "TIMEOUT"}},
			{name: "account"},
			{name: "partition"},
			{name: "qos"},
			{name: "node"},
			{name: "users"},
		},
	},
}

// composer is the visual query builder state. The builder (endpoints, URL,
// params) and the response output live in separate scrollable viewports.
type composer struct {
	active  int
	cursor  int
	editing bool
	input   textinput.Model
	status  string
	body    string
	version api.Version
	builder viewport.Model
	output  viewport.Model
}

func newComposer() composer {
	ti := textinput.New()
	ti.CharLimit = 256
	ti.Focus()

	return composer{
		input:   ti,
		builder: viewport.New(0, 0),
		output:  viewport.New(0, 0),
	}
}

// setVersion records the detected API version used to build request paths.
func (c *composer) setVersion(v api.Version) {
	c.version = v
	c.rebuild()
}

// builtPath returns the full request path (with API family prefix and query
// parameters).
func (c composer) builtPath() string {
	ep := endpoints[c.active]
	parts := []string{}
	for _, p := range ep.params {
		if p.value != "" {
			parts = append(parts, p.name+"="+p.value)
		}
	}
	q := ""
	if len(parts) > 0 {
		q = "?" + strings.Join(parts, "&")
	}
	if ep.base == "slurmdb" {
		return "/slurmdb/" + c.version.String() + ep.path + q
	}
	return ep.path + q
}

// rebuild regenerates both viewports from the current state.
func (c *composer) rebuild() {
	c.builder.SetContent(c.renderBuilder())
	c.output.SetContent(c.renderOutput())
}

// ensureCursorVisible scrolls the builder so the selected parameter is shown.
func (c *composer) ensureCursorVisible() {
	// Layout: line 0 = endpoints, line 1 = URL, params start at line 2.
	line := 2 + c.cursor
	y := line - c.builder.Height/2
	if y < 0 {
		y = 0
	}
	c.builder.SetYOffset(y)
}

// selectEndpoint switches the active endpoint and resets the cursor.
func (c *composer) selectEndpoint(i int) {
	c.active = i
	c.cursor = 0
	c.editing = false
	c.status = ""
	c.rebuild()
}

// move moves the param cursor by delta, respecting bounds.
func (c *composer) move(delta int) {
	ep := endpoints[c.active]
	if len(ep.params) == 0 {
		return
	}
	c.cursor = (c.cursor + delta + len(ep.params)) % len(ep.params)
	c.rebuild()
	c.ensureCursorVisible()
}

// startEdit focuses the text input on the selected parameter.
func (c *composer) startEdit() {
	ep := endpoints[c.active]
	if len(ep.params) == 0 {
		return
	}
	c.input.SetValue(ep.params[c.cursor].value)
	c.editing = true
	c.rebuild()
}

// stopEdit commits the edited value.
func (c *composer) stopEdit() {
	if !c.editing {
		return
	}
	ep := endpoints[c.active]
	if len(ep.params) > 0 {
		ep.params[c.cursor].value = c.input.Value()
	}
	c.editing = false
	c.rebuild()
}

// clearValue empties the selected parameter's value.
func (c *composer) clearValue() {
	ep := endpoints[c.active]
	if len(ep.params) == 0 {
		return
	}
	ep.params[c.cursor].value = ""
	c.rebuild()
}

// cycleOption moves the selected parameter's value through its options.
func (c *composer) cycleOption(delta int) {
	ep := endpoints[c.active]
	if len(ep.params) == 0 {
		return
	}
	p := &ep.params[c.cursor]
	if len(p.options) == 0 {
		return
	}
	idx := 0
	for i, o := range p.options {
		if o == p.value {
			idx = i
			break
		}
	}
	idx = (idx + delta + len(p.options)) % len(p.options)
	p.value = p.options[idx]
	c.rebuild()
}

// setPartitionOptions feeds the partitions gathered from the cluster into the
// "partition" parameter of every endpoint that has one.
func (c *composer) setPartitionOptions(names []string) {
	for i := range endpoints {
		for j := range endpoints[i].params {
			if endpoints[i].params[j].name == "partition" {
				endpoints[i].params[j].options = names
			}
		}
	}
	c.rebuild()
}

// setAccountOptions feeds the accounts gathered from the cluster into the
// "account" parameter of the jobs endpoint.
func (c *composer) setAccountOptions(names []string) {
	for i := range endpoints {
		for j := range endpoints[i].params {
			if endpoints[i].params[j].name == "account" {
				endpoints[i].params[j].options = names
			}
		}
	}
	c.rebuild()
}

// setQoSOptions feeds the QoS gathered from the cluster into the
// "qos" parameter of the jobs endpoint.
func (c *composer) setQoSOptions(names []string) {
	for i := range endpoints {
		for j := range endpoints[i].params {
			if endpoints[i].params[j].name == "qos" {
				endpoints[i].params[j].options = names
			}
		}
	}
	c.rebuild()
}

// run issues the built request and returns a composerRunMsg.
func (c composer) run(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		ep := endpoints[c.active]

		// "node action" is a special endpoint that sends a POST
		// /node/{name} with the chosen state in the body.
		if ep.name == "node action" && ep.method == "POST" {
			nodeName := ""
			action := ""
			for _, p := range ep.params {
				if p.name == "node" {
					nodeName = p.value
				}
				if p.name == "action" {
					action = p.value
				}
			}
			if nodeName == "" || action == "" {
				return composerRunMsg{err: fmt.Errorf("node name and action are required")}
			}
			return composerRunMsg{err: client.NodeState(ctx, nodeName, action, "srest")}
		}

		return composerRunMsg{err: client.Get(ctx, c.builtPath(), nil)}
	}
}

// setResult updates the status line and response body shown in the output.
func (c *composer) setResult(status, body string) {
	c.status = status
	c.body = body
	c.rebuild()
	c.output.GotoBottom()
}

// renderBuilder builds the left panel (endpoints, URL, params, hints).
func (c composer) renderBuilder() string {
	var sb strings.Builder

	var eps []string
	for i, ep := range endpoints {
		if i == c.active {
			eps = append(eps, composerEndpointActive.Render(ep.name))
		} else {
			eps = append(eps, composerEndpoint.Render(ep.name))
		}
	}
	sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, eps...) + "\n")
	sb.WriteString(composerURL.Render(c.method()+" "+c.builtPath()) + "\n")

	ep := endpoints[c.active]
	if len(ep.params) == 0 {
		sb.WriteString(detailStyle.Render("No query parameters for this endpoint.") + "\n")
	} else {
		for i, p := range ep.params {
			var row string
			if c.editing && i == c.cursor {
				row = composerParamName.Render(p.name) + "  " + c.input.View()
			} else {
				row = composerParamName.Render(p.name) + "  " + composerParamValue.Render(p.value)
			}
			if i == c.cursor && !c.editing {
				row = composerParamCursor.Render(row)
			}
			sb.WriteString(row + "\n")
		}
	}

	sb.WriteString(composerHint.Render("enter=edit ←/→=option ↑/↓=nav del=clear f=focus 1-2=endpoint r=run") + "\n")
	return sb.String()
}

// renderOutput builds the right panel (status line + response body).
func (c composer) renderOutput() string {
	var sb strings.Builder
	if c.status != "" {
		sb.WriteString(c.status + "\n")
	} else {
		sb.WriteString(detailStyle.Render("Run a query to see the response.") + "\n")
	}
	if c.body != "" {
		sb.WriteString(c.body + "\n")
	}
	return sb.String()
}

// Update processes keys for the composer.
func (c composer) Update(msg tea.KeyMsg) (composer, tea.Cmd) {
	if c.editing {
		if msg.String() == "enter" {
			c.stopEdit()
			return c, nil
		}
		var cmd tea.Cmd
		c.input, cmd = c.input.Update(msg)
		c.rebuild()
		return c, cmd
	}

	switch msg.String() {
	case "up", "k":
		c.move(-1)
	case "down", "j":
		c.move(1)
	case "enter":
		c.startEdit()
	case "delete", "backspace", "x":
		c.clearValue()
	case "left", "h":
		// Only cycle the selected parameter's options; endpoint switching is
		// done with the number keys.
		if c.hasOptionsAtCursor() {
			c.cycleOption(-1)
		}
	case "right", "l":
		if c.hasOptionsAtCursor() {
			c.cycleOption(1)
		}
	case "pgup", "pgdown", "home", "end":
		c.output, _ = c.output.Update(msg)
	case "1":
		c.selectEndpoint(0)
	case "2":
		c.selectEndpoint(1)
	}
	return c, nil
}

// hasOptionsAtCursor reports whether the selected parameter offers options.
func (c composer) hasOptionsAtCursor() bool {
	ep := endpoints[c.active]
	return len(ep.params) > 0 && len(ep.params[c.cursor].options) > 0
}

func (c composer) method() string {
	return endpoints[c.active].method
}
