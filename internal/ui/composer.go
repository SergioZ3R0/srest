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

// composerRunMsg is returned after the composer executes a request.
type composerRunMsg struct {
	err   error
	jobID uint32
}

// composer is the visual query builder state with sidebar navigation.
type composer struct {
	categoryIdx int
	endpointIdx int
	cursor      int
	editing     bool
	input       textinput.Model
	status      string
	body        string
	version     api.Version
	sidebar     viewport.Model
	builder     viewport.Model
	output      viewport.Model
}

func newComposer() composer {
	ti := textinput.New()
	ti.CharLimit = 256
	ti.Focus()

	return composer{
		input:   ti,
		sidebar: viewport.New(0, 0),
		builder: viewport.New(0, 0),
		output:  viewport.New(0, 0),
	}
}

// currentCategory returns the active category.
func (c composer) currentCategory() category {
	return categories[c.categoryIdx]
}

// currentEndpoint returns the active endpoint.
func (c composer) currentEndpoint() endpoint {
	return c.currentCategory().endpoints[c.endpointIdx]
}

// setVersion records the detected API version used to build request paths.
func (c *composer) setVersion(v api.Version) {
	c.version = v
	c.rebuild()
}

// builtPath returns the full request path with path params substituted and
// query params appended.
func (c composer) builtPath() string {
	ep := c.currentEndpoint()

	// Substitute path parameters.
	path := ep.path
	for _, p := range ep.pathParams() {
		if p.value != "" {
			path = strings.Replace(path, "{"+p.name+"}", p.value, 1)
		}
	}

	if ep.method == "POST" || ep.method == "DELETE" {
		if ep.base == "slurmdb" {
			return "/slurmdb/" + c.version.String() + path
		}
		return path
	}

	// Build query string from query params.
	parts := []string{}
	for _, p := range ep.getParams() {
		if p.value != "" {
			parts = append(parts, p.name+"="+p.value)
		}
	}
	q := ""
	if len(parts) > 0 {
		q = "?" + strings.Join(parts, "&")
	}
	if ep.base == "slurmdb" {
		return "/slurmdb/" + c.version.String() + path + q
	}
	return path + q
}

// rebuild regenerates all viewports from the current state.
func (c *composer) rebuild() {
	c.sidebar.SetContent(c.renderSidebar())
	c.builder.SetContent(c.renderBuilder())
	c.output.SetContent(c.renderOutput())
}

// ensureCursorVisible scrolls the builder so the selected parameter is shown.
func (c *composer) ensureCursorVisible() {
	ep := c.currentEndpoint()
	// Layout: line 0 = endpoint name, line 1 = URL, params start at line 2.
	line := 2 + c.cursor
	_ = ep
	y := line - c.builder.Height/2
	if y < 0 {
		y = 0
	}
	c.builder.SetYOffset(y)
}

// selectEndpoint switches the active endpoint within the current category.
func (c *composer) selectEndpoint(i int) {
	eps := c.currentCategory().endpoints
	if i >= 0 && i < len(eps) {
		c.endpointIdx = i
		c.cursor = 0
		c.editing = false
		c.status = ""
		c.rebuild()
	}
}

// selectCategory switches the active category and resets endpoint selection.
func (c *composer) selectCategory(i int) {
	if i >= 0 && i < len(categories) {
		c.categoryIdx = i
		c.endpointIdx = 0
		c.cursor = 0
		c.editing = false
		c.status = ""
		c.rebuild()
	}
}

// move moves the param cursor by delta, respecting bounds.
func (c *composer) move(delta int) {
	ep := c.currentEndpoint()
	params := ep.getParams()
	if len(params) == 0 {
		return
	}
	c.cursor = (c.cursor + delta + len(params)) % len(params)
	c.rebuild()
	c.ensureCursorVisible()
}

// startEdit focuses the text input on the selected parameter.
func (c *composer) startEdit() {
	ep := c.currentEndpoint()
	params := ep.getParams()
	if len(params) == 0 || c.cursor >= len(params) {
		return
	}
	c.input.SetValue(params[c.cursor].value)
	c.editing = true
	c.rebuild()
}

// stopEdit commits the edited value.
func (c *composer) stopEdit() {
	if !c.editing {
		return
	}
	ep := c.currentEndpoint()
	params := ep.getParams()
	if c.cursor < len(params) {
		params[c.cursor].value = c.input.Value()
	}
	c.editing = false
	c.rebuild()
}

// clearValue empties the selected parameter's value.
func (c *composer) clearValue() {
	ep := c.currentEndpoint()
	params := ep.getParams()
	if c.cursor < len(params) {
		params[c.cursor].value = ""
		c.rebuild()
	}
}

// cycleOption moves the selected parameter's value through its options.
func (c *composer) cycleOption(delta int) {
	ep := c.currentEndpoint()
	params := ep.getParams()
	if c.cursor >= len(params) {
		return
	}
	p := &params[c.cursor]
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

// hasOptionsAtCursor reports whether the selected parameter offers options.
func (c composer) hasOptionsAtCursor() bool {
	ep := c.currentEndpoint()
	params := ep.getParams()
	return c.cursor < len(params) && len(params[c.cursor].options) > 0
}

// setPartitionOptions feeds the partitions gathered from the cluster into the
// "partition" parameter of every endpoint that has one.
func (c *composer) setPartitionOptions(names []string) {
	for ci := range categories {
		for ei := range categories[ci].endpoints {
			for j := range categories[ci].endpoints[ei].params {
				if categories[ci].endpoints[ei].params[j].name == "partition" {
					categories[ci].endpoints[ei].params[j].options = names
				}
			}
		}
	}
	c.rebuild()
}

// setAccountOptions feeds the accounts gathered from the cluster into the
// "account" parameter of every endpoint that has one.
func (c *composer) setAccountOptions(names []string) {
	for ci := range categories {
		for ei := range categories[ci].endpoints {
			for j := range categories[ci].endpoints[ei].params {
				if categories[ci].endpoints[ei].params[j].name == "account" {
					categories[ci].endpoints[ei].params[j].options = names
				}
			}
		}
	}
	c.rebuild()
}

// setQoSOptions feeds the QoS gathered from the cluster into the
// "qos" parameter of every endpoint that has one.
func (c *composer) setQoSOptions(names []string) {
	for ci := range categories {
		for ei := range categories[ci].endpoints {
			for j := range categories[ci].endpoints[ei].params {
				if categories[ci].endpoints[ei].params[j].name == "qos" {
					categories[ci].endpoints[ei].params[j].options = names
				}
			}
		}
	}
	c.rebuild()
}

// setGresOptions feeds the GRES gathered from the nodes into the
// "gres" parameter of the submit jobs endpoint.
func (c *composer) setGresOptions(names []string) {
	for ci := range categories {
		for ei := range categories[ci].endpoints {
			for j := range categories[ci].endpoints[ei].params {
				if categories[ci].endpoints[ei].params[j].name == "gres" {
					categories[ci].endpoints[ei].params[j].options = names
				}
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
		ep := c.currentEndpoint()

		// Handle submit job specially.
		if ep.path == "/job/submit" && ep.method == "POST" {
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
			if _, ok := jobFields["environment"]; !ok {
				jobFields["environment"] = []string{"PATH=/usr/bin:/bin"}
			}
			body := map[string]any{"job": jobFields}
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

		// Handle job action (cancel, requeue, hold, release).
		if ep.method == "POST" && strings.Contains(ep.path, "/job/{job_id}") {
			var jobID uint32
			var action string
			for _, p := range ep.params {
				if p.name == "job_id" && p.value != "" {
					_, _ = fmt.Sscanf(p.value, "%d", &jobID)
				}
				if p.name == "action" && p.value != "" {
					action = p.value
				}
			}
			if jobID == 0 || action == "" {
				return composerRunMsg{err: fmt.Errorf("job_id and action are required")}
			}
			var err error
			switch action {
			case "CANCEL":
				err = client.CancelJob(ctx, jobID)
			case "REQUEUE":
				err = client.RequeueJob(ctx, jobID)
			default:
				err = fmt.Errorf("unsupported action: %s", action)
			}
			if err != nil {
				return composerRunMsg{err: err}
			}
			return composerRunMsg{}
		}

		// Handle delete job.
		if ep.method == "DELETE" && strings.Contains(ep.path, "/job/{job_id}") {
			var jobID uint32
			for _, p := range ep.params {
				if p.name == "job_id" && p.value != "" {
					_, _ = fmt.Sscanf(p.value, "%d", &jobID)
				}
			}
			if jobID == 0 {
				return composerRunMsg{err: fmt.Errorf("job_id is required")}
			}
			err := client.CancelJob(ctx, jobID)
			if err != nil {
				return composerRunMsg{err: err}
			}
			return composerRunMsg{}
		}

		// Generic GET request.
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
		"account_name":              "account",
		"cluster_name":              "cluster",
		"node_name":                 "node",
		"parent_account":            "parent",
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
// file containing the current script.
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

// renderSidebar builds the left sidebar with categories and their endpoints.
func (c composer) renderSidebar() string {
	var sb strings.Builder

	catStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true).Padding(0, 1)
	catActive := lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color("57")).Bold(true).Padding(0, 1)
	epStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Padding(0, 0, 0, 2)
	epActive := lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color("57")).Padding(0, 0, 0, 2)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	for ci, cat := range categories {
		if ci == c.categoryIdx {
			sb.WriteString(catActive.Render("> "+cat.name) + "\n")
		} else {
			sb.WriteString(catStyle.Render("  "+cat.name) + "\n")
		}
		for ei, ep := range cat.endpoints {
			if ci == c.categoryIdx && ei == c.endpointIdx {
				sb.WriteString(epActive.Render("• "+ep.name) + "\n")
			} else if ci == c.categoryIdx {
				sb.WriteString(epStyle.Render("  "+ep.name) + "\n")
			} else {
				sb.WriteString(dimStyle.Render("    "+ep.name) + "\n")
			}
		}
	}
	return sb.String()
}

// renderBuilder builds the params panel for the selected endpoint.
func (c composer) renderBuilder() string {
	var sb strings.Builder

	ep := c.currentEndpoint()
	methodStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	sb.WriteString(methodStyle.Render(ep.method) + " " + pathStyle.Render(c.builtPath()) + "\n\n")

	params := ep.getParams()
	if len(params) == 0 {
		sb.WriteString(detailStyle.Render("No parameters for this endpoint.") + "\n")
	} else {
		maxLabel := 0
		for _, p := range params {
			l := len(fieldLabel(p.name))
			if l > maxLabel {
				maxLabel = l
			}
		}
		nameStyle := lipgloss.NewStyle().Width(maxLabel).Foreground(lipgloss.Color("12"))

		for i, p := range params {
			label := fieldLabel(p.name)
			var row string
			if c.editing && i == c.cursor {
				row = nameStyle.Render(label) + "  " + c.input.View()
			} else {
				valueStyle := composerParamValue
				if p.kind == paramPath {
					valueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
				}
				row = nameStyle.Render(label) + "  " + valueStyle.Render(p.value)
			}
			if i == c.cursor && !c.editing {
				row = composerParamCursor.Render(row)
			}
			sb.WriteString(row + "\n")
		}
	}

	hint := "↑/↓=nav enter=edit ←/→=option del=clear r=run"
	if c.categoryIdx == 0 {
		hint += "  tab/]=next category"
	}
	sb.WriteString(composerHint.Render(hint) + "\n")
	return sb.String()
}

// renderOutput builds the response panel.
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
		ep := c.currentEndpoint()
		params := ep.getParams()
		if len(params) > 0 && c.cursor < len(params) && params[c.cursor].name == "script" {
			return c, openEditorCmd(params[c.cursor].value)
		}
	case "left", "h":
		if c.hasOptionsAtCursor() {
			c.cycleOption(-1)
		}
	case "right", "l":
		if c.hasOptionsAtCursor() {
			c.cycleOption(1)
		}
	case "pgup", "pgdown", "home", "end":
		c.output, _ = c.output.Update(msg)
	case "tab", "]":
		c.selectCategory((c.categoryIdx + 1) % len(categories))
	case "shift+tab", "[":
		c.selectCategory((c.categoryIdx - 1 + len(categories)) % len(categories))
	}
	return c, nil
}
