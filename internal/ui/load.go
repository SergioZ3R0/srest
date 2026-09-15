package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/SergioZ3R0/srest/internal/api"
)

// barColors defines the color gradient for load bars (low → high).
var (
	barEmpty = lipgloss.Color("240")
	barLow   = lipgloss.Color("10") // green
	barMid   = lipgloss.Color("11") // yellow
	barHigh  = lipgloss.Color("9")  // red
)

// bar renders a btop-style horizontal bar for the given percentage (0–100).
// width is the total character width of the bar (including brackets).
func bar(pct float64, width int) string {
	if width < 4 {
		width = 4
	}
	filled := int(pct / 100 * float64(width-2))
	if filled > width-2 {
		filled = width - 2
	}
	empty := width - 2 - filled

	// Pick color based on utilization.
	var color lipgloss.Color
	switch {
	case pct >= 90:
		color = barHigh
	case pct >= 70:
		color = barMid
	default:
		color = barLow
	}

	barStyle := lipgloss.NewStyle().Foreground(color)
	emptyStyle := lipgloss.NewStyle().Foreground(barEmpty)

	return barStyle.Render(strings.Repeat("█", filled)) +
		emptyStyle.Render(strings.Repeat("░", empty))
}

// loadBar renders a single resource line: label, bar, percentage, and values.
func loadBar(label string, pct float64, used, total int64, unit string, barWidth int) string {
	pctLabel := fmt.Sprintf("%3.0f%%", pct)
	var valLabel string
	if total > 0 {
		if unit == "GB" {
			valLabel = fmt.Sprintf("%d/%dGB", used/1024, total/1024)
		} else {
			valLabel = fmt.Sprintf("%d/%d%s", used, total, unit)
		}
	} else {
		valLabel = "n/a"
	}

	return fmt.Sprintf("  %s  %s  %s  %s",
		loaderLabelStyle.Render(label),
		bar(pct, barWidth),
		loaderPctStyle.Render(pctLabel),
		loaderValueStyle.Render(valLabel),
	)
}

// loaderLabelStyle labels the resource type (cpu, mem, gpu).
var loaderLabelStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("12")).
	Width(4).
	Bold(true)

// loaderPctStyle formats the percentage column.
var loaderPctStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("15")).
	Width(5).
	Align(lipgloss.Right)

// loaderValueStyle formats the used/total value column.
var loaderValueStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("8"))

// loadPartitionCard renders a full partition load card with the partition name
// as header and three resource bars below.
func loadPartitionCard(pl api.PartitionLoad, barWidth int) string {
	title := loadCardTitleStyle.Render(pl.Name) + loadCardNodeStyle.Render(
		fmt.Sprintf("  %d nodes", pl.TotalNodes))

	bars := []string{
		loadBar("cpu", pl.CPUAllocPercent(), int64(pl.UsedCPUs), int64(pl.TotalCPUs), "", barWidth),
		loadBar("mem", pl.MemAllocPercent(), pl.UsedMemMB, pl.TotalMemMB, "GB", barWidth),
	}
	if pl.TotalGPUs > 0 {
		bars = append(bars, loadBar("gpu", pl.GPUAllocPercent(), int64(pl.UsedGPUs), int64(pl.TotalGPUs), "", barWidth))
	}

	return loadCardStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, title, ""),
	) + "\n" + strings.Join(bars, "\n")
}

// loadCardStyle frames a partition load card.
var loadCardStyle = lipgloss.NewStyle()

// loadCardTitleStyle formats the partition name.
var loadCardTitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("15"))

// loadCardNodeStyle formats the node count next to the partition name.
var loadCardNodeStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("8"))
