package ui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
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
	err   error
	jobID uint32
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
// cluster). ping is first so it is selected by default. "get jobs" targets the
// slurmdb API with rich filters. "submit jobs" sends a POST with the form
// fields as body.
var endpoints = []endpoint{
	{
		name:   "ping",
		method: "GET",
		base:   "slurm",
		path:   "/ping",
		params: []param{},
	},
	{
		name:   "get jobs",
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
	{
		name:   "submit jobs",
		method: "POST",
		base:   "slurm",
		path:   "/job/submit",
		params: []param{
			{name: "name"},
			{name: "partition"},
			{name: "qos"},
			{name: "account"},
			{name: "gres"},
			{name: "time_limit", value: "60"},
			{name: "nodes", value: "1"},
			{name: "cpus_per_task", value: "1"},
			{name: "memory_per_node"},
			{name: "standard_output"},
			{name: "standard_error"},
			{name: "current_working_directory", value: "/tmp"},
			{name: "script_path"},
			{name: "script", value: "#!/bin/bash\nhostname"},
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
// parameters). For POST endpoints, only the base path is returned since
// parameters are sent as the request body.
func (c composer) builtPath() string {
	ep := endpoints[c.active]

	if ep.method == "POST" {
		if ep.base == "slurmdb" {
			return "/slurmdb/" + c.version.String() + ep.path
		}
		return ep.path
	}

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

// setGresOptions feeds the GRES gathered from the nodes into the
// "gres" parameter of the submit jobs endpoint.
func (c *composer) setGresOptions(names []string) {
	for i := range endpoints {
		for j := range endpoints[i].params {
			if endpoints[i].params[j].name == "gres" {
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

		// POST endpoints: build body from params.
		if ep.method == "POST" {
			// Submit jobs wraps fields in a "job" object per the OpenAPI spec.
			if ep.path == "/job/submit" {
				jobFields := make(map[string]any)
				for _, p := range ep.params {
					if p.value == "" || p.name == "script" || p.name == "script_path" {
						continue
					}
					switch p.name {
					case "time_limit", "nodes", "cpus_per_task", "memory_per_node":
						if v, err := strconv.Atoi(p.value); err == nil {
							jobFields[p.name] = v
						}
					default:
						jobFields[p.name] = p.value
					}
				}

				// Environment is required for job submission; default to PATH.
				if _, ok := jobFields["environment"]; !ok {
					jobFields["environment"] = []string{"PATH=/usr/bin:/bin"}
				}

				body := map[string]any{"job": jobFields}

				// If script_path is set, read the file and use its contents as script.
				for _, p := range ep.params {
					if p.name == "script_path" && p.value != "" {
						data, err := os.ReadFile(p.value)
						if err != nil {
							return composerRunMsg{err: fmt.Errorf("reading script: %w", err)}
						}
						body["script"] = string(data)
					}
					if p.name == "script" && p.value != "" {
						body["script"] = p.value
					}
				}

				result, err := client.SubmitJob(ctx, body)
				if err != nil {
					return composerRunMsg{err: err}
				}
				return composerRunMsg{jobID: result.JobID}
			}
			return composerRunMsg{err: fmt.Errorf("unsupported POST endpoint: %s", ep.path)}
		}

		return composerRunMsg{err: client.Get(ctx, c.builtPath(), nil)}
	}
}

// fieldLabel returns a short display label for the field.
func fieldLabel(name string) string {
	labels := map[string]string{
		"cpus_per_task":             "cpus/task",
		"memory_per_node":           "mem",
		"standard_output":           "stdout",
		"standard_error":            "stderr",
		"time_limit":                "wall",
		"nodes":                     "nodes",
		"current_working_directory": "cwd",
	}
	if l, ok := labels[name]; ok {
		return l
	}
	return name
}

// setResult updates the status line and response body shown in the output.
func (c *composer) setResult(status, body string) {
	c.status = status
	c.body = body
	c.rebuild()
	c.output.GotoBottom()
}

// openEditorCmd returns a tea.Cmd that opens the default editor on a temp
// file containing the current script. When the editor closes, the file
// contents are read back into the script parameter.
func openEditorCmd(script string) tea.Cmd {
	return tea.ExecProcess(exec.Command(
		os.Getenv("EDITOR"),
		"-c", "set filetype=bash",
		"/tmp/srest-script.sh",
	), func(err error) tea.Msg {
		data, _ := os.ReadFile("/tmp/srest-script.sh")
		return editorDoneMsg{content: string(data), err: err}
	})
}

// editorDoneMsg is sent when the editor closes.
type editorDoneMsg struct {
	content string
	err     error
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
		// Calculate max label width so all labels align.
		maxLabel := 0
		for _, p := range ep.params {
			l := len(fieldLabel(p.name))
			if l > maxLabel {
				maxLabel = l
			}
		}
		nameStyle := lipgloss.NewStyle().Width(maxLabel).Foreground(lipgloss.Color("12"))

		for i, p := range ep.params {
			label := fieldLabel(p.name)
			var row string
			if c.editing && i == c.cursor {
				row = nameStyle.Render(label) + "  " + c.input.View()
			} else {
				row = nameStyle.Render(label) + "  " + composerParamValue.Render(p.value)
			}
			if i == c.cursor && !c.editing {
				row = composerParamCursor.Render(row)
			}
			sb.WriteString(row + "\n")
		}
	}

	sb.WriteString(composerHint.Render("enter=edit e=editor ←/→=option ↑/↓=nav del=clear f=focus 1-3=endpoint r=run") + "\n")
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
	case "e":
		// Open editor for the script parameter.
		ep := endpoints[c.active]
		if len(ep.params) > 0 && c.cursor < len(ep.params) && ep.params[c.cursor].name == "script" {
			return c, openEditorCmd(ep.params[c.cursor].value)
		}
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
	case "3":
		c.selectEndpoint(2)
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
