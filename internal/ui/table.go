package ui

import (
	"fmt"
	"strings"
	"time"

	"carbonqt/internal/models"

	"github.com/charmbracelet/lipgloss"
)

var (
	accentPrimary   = lipgloss.Color("35")
	accentSecondary = lipgloss.Color("34")
	inkStrong       = lipgloss.Color("231")
	inkMuted        = lipgloss.Color("244")
	panelBorder     = lipgloss.Color("236")
	panelBackdrop   = lipgloss.Color("22")
	selectBg        = lipgloss.Color("120")
	selectFg        = lipgloss.Color("16")

	titleStyle      = lipgloss.NewStyle().Bold(true).Foreground(inkStrong)
	headerStyle     = lipgloss.NewStyle().Bold(true).Foreground(accentPrimary)
	mutedStyle      = lipgloss.NewStyle().Foreground(inkMuted)
	panelStyle      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(panelBorder).Padding(1, 2)
	panelTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(inkStrong).Background(accentSecondary).Padding(0, 1)
	panelBadgeStyle = lipgloss.NewStyle().Bold(true).Foreground(inkStrong).Background(panelBackdrop).Padding(0, 1)
	barFillStyle    = lipgloss.NewStyle().Foreground(accentPrimary)
	barBaseStyle    = lipgloss.NewStyle().Foreground(panelBorder)
	selectStyle     = lipgloss.NewStyle().Background(selectBg).Foreground(selectFg).Bold(true)
	rowEvenStyle    = lipgloss.NewStyle().Foreground(inkMuted)
	rowOddStyle     = lipgloss.NewStyle().Foreground(inkStrong)
	helpKeyStyle    = lipgloss.NewStyle().Bold(true).Foreground(inkStrong).Background(panelBackdrop).Padding(0, 1)
	helpTextStyle   = lipgloss.NewStyle().Foreground(inkMuted)
	helpBarStyle    = lipgloss.NewStyle().Background(accentPrimary).Foreground(inkStrong).Bold(true).Padding(0, 1)
	splashTitle     = lipgloss.NewStyle().Bold(true).Foreground(inkStrong)
	splashSubtitle  = lipgloss.NewStyle().Foreground(inkMuted)
)

func RenderCarbonTrend(values []float64) string {
	chartWidth := 40
	trend := renderSparkline(values, chartWidth)
	if trend == "" {
		return "Carbon Trend\n(n/a)"
	}
	return strings.Join([]string{"Carbon Trend", trend}, "\n")
}

func RenderCarbonTrendWithLabel(values []float64, label string, width int) string {
	chartWidth := width
	if chartWidth <= 0 {
		chartWidth = 40
	}
	trend := renderSparkline(values, chartWidth)
	if trend == "" {
		return label + "\n(n/a)"
	}
	return strings.Join([]string{label, trend}, "\n")
}

func RenderProcessTable(processes []models.ProcessMetrics) string {
	if len(processes) == 0 {
		return "No processes found."
	}

	headers := []string{"PID", "NAME", "CPU %", "MEM %", "POWER (W)", "CARBON (mg)", "RUNTIME", "PATH"}
	rows := make([][]string, 0, len(processes))
	for _, proc := range processes {
		runtime := "n/a"
		if !proc.StartTime.IsZero() {
			runtime = formatDuration(time.Since(proc.StartTime))
		}
		rows = append(rows, []string{
			fmt.Sprintf("%d", proc.PID),
			truncateMiddle(proc.Name, 20),
			fmt.Sprintf("%.2f", proc.CPUPercent),
			fmt.Sprintf("%.2f", proc.MemoryPercent),
			fmt.Sprintf("%.2f", proc.PowerW),
			formatCarbonMg(proc.CarbonKg),
			runtime,
			truncateMiddle(fallback(proc.ExePath, "n/a"), 64),
		})
	}

	return renderTable(headers, rows)
}

func RenderDashboardTable(processes []models.ProcessMetrics) string {
	return RenderDashboardTableSelected(processes, -1, 0)
}

func RenderDashboardTableSelected(processes []models.ProcessMetrics, selected int, maxRows int) string {
	if len(processes) == 0 {
		return "No processes found."
	}

	headers := []string{"PID", "NAME", "CPU %", "MEM %", "POWER (W)", "CARBON (mg)", "RUNTIME", "PATH"}
	rows := make([][]string, 0, len(processes))
	for _, proc := range processes {
		rows = append(rows, dashboardRow(proc))
	}

	return renderTableSelected(headers, rows, selected, maxRows)
}

func dashboardRow(proc models.ProcessMetrics) []string {
	runtime := "n/a"
	if !proc.StartTime.IsZero() {
		runtime = formatDuration(time.Since(proc.StartTime))
	}
	return []string{
		fmt.Sprintf("%d", proc.PID),
		truncateMiddle(proc.Name, 20),
		fmt.Sprintf("%.2f", proc.CPUPercent),
		fmt.Sprintf("%.2f", proc.MemoryPercent),
		fmt.Sprintf("%.2f", proc.PowerW),
		formatCarbonMg(proc.CarbonKg),
		runtime,
		truncateMiddle(fallback(proc.ExePath, "n/a"), 64),
	}
}

func renderTable(headers []string, rows [][]string) string {
	return renderTableSelected(headers, rows, -1, 0)
}

func renderTableSelected(headers []string, rows [][]string, selected int, maxRows int) string {
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = len(header)
	}

	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	start := 0
	end := len(rows)
	if maxRows > 0 && len(rows) > maxRows {
		if selected < 0 {
			selected = 0
		}
		if selected >= len(rows) {
			selected = len(rows) - 1
		}
		start = selected - maxRows + 1
		if start < 0 {
			start = 0
		}
		end = start + maxRows
		if end > len(rows) {
			end = len(rows)
			start = end - maxRows
			if start < 0 {
				start = 0
			}
		}
	}

	lines := make([]string, 0, len(rows)+1)
	lines = append(lines, formatRow(headers, widths, headerStyle))
	totalWidth := tableWidth(widths, 2)
	for i := start; i < end; i++ {
		row := rows[i]
		rowText := formatRowPlain(row, widths)
		rowText = padRight(rowText, totalWidth)
		if i == selected {
			lines = append(lines, selectStyle.Width(totalWidth).Render(rowText))
			continue
		}
		if i%2 == 0 {
			lines = append(lines, rowEvenStyle.Render(rowText))
			continue
		}
		lines = append(lines, rowOddStyle.Render(rowText))
	}

	return strings.Join(lines, "\n")
}

func formatRowPlain(values []string, widths []int) string {
	styled := make([]string, len(values))
	for i, value := range values {
		styled[i] = lipgloss.NewStyle().Width(widths[i]).Render(value)
	}
	return strings.Join(styled, "  ")
}

func tableWidth(widths []int, gap int) int {
	total := 0
	for _, width := range widths {
		total += width
	}
	if len(widths) > 1 {
		total += gap * (len(widths) - 1)
	}
	return total
}

func padRight(value string, width int) string {
	if width <= 0 {
		return value
	}
	pad := width - lipgloss.Width(value)
	if pad <= 0 {
		return value
	}
	return value + strings.Repeat(" ", pad)
}

func formatRow(values []string, widths []int, style lipgloss.Style) string {
	cells := make([]string, len(values))
	for i, value := range values {
		cell := lipgloss.NewStyle().Width(widths[i]).Render(value)
		cells[i] = style.Render(cell)
	}
	return strings.Join(cells, "  ")
}

func renderBadge(value string) string {
	return panelBadgeStyle.Render(value)
}

func progressBar(width int, percent float64) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	filled := int(float64(width) * (percent / 100.0))
	empty := width - filled

	return " " + barFillStyle.Render(strings.Repeat("#", filled)) + barBaseStyle.Render(strings.Repeat("-", empty))
}

func formatCarbonMg(carbonKg float64) string {
	carbonMg := carbonKg * 1_000_000
	if carbonMg < 0.01 {
		return fmt.Sprintf("%.4f", carbonMg)
	}
	if carbonMg < 1 {
		return fmt.Sprintf("%.3f", carbonMg)
	}
	if carbonMg < 100 {
		return fmt.Sprintf("%.2f", carbonMg)
	}
	return fmt.Sprintf("%.1f", carbonMg)
}

func formatBytes(value uint64) string {
	const unit = 1024
	if value < unit {
		return fmt.Sprintf("%d B", value)
	}

	div, exp := uint64(unit), 0
	for n := value / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	prefixes := []string{"KB", "MB", "GB", "TB"}
	if exp >= len(prefixes) {
		exp = len(prefixes) - 1
	}

	return fmt.Sprintf("%.1f %s", float64(value)/float64(div), prefixes[exp])
}

func formatDuration(duration time.Duration) string {
	if duration < 0 {
		duration = 0
	}

	seconds := int64(duration.Seconds())
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	if hours > 0 {
		return fmt.Sprintf("%dh%dm", hours, minutes)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm%ds", minutes, secs)
	}
	return fmt.Sprintf("%ds", secs)
}

func fallback(value, defaultValue string) string {
	if strings.TrimSpace(value) == "" {
		return defaultValue
	}
	return value
}

func renderSparkline(values []float64, width int) string {
	if len(values) == 0 || width <= 0 {
		return ""
	}

	min := values[0]
	max := values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	levels := []rune(".:-=+*#%@")
	span := max - min
	if span == 0 {
		span = 1
	}

	step := float64(len(values)) / float64(width)
	var builder strings.Builder
	for i := 0; i < width; i++ {
		idx := int(float64(i) * step)
		if idx >= len(values) {
			idx = len(values) - 1
		}
		value := values[idx]
		levelIndex := int(((value - min) / span) * float64(len(levels)-1))
		if levelIndex < 0 {
			levelIndex = 0
		}
		if levelIndex >= len(levels) {
			levelIndex = len(levels) - 1
		}
		builder.WriteRune(levels[levelIndex])
	}

	return builder.String()
}

func truncateMiddle(value string, max int) string {
	value = strings.TrimSpace(value)
	if max <= 0 || len(value) <= max {
		return value
	}
	if max <= 3 {
		return value[:max]
	}
	keep := max - 3
	left := keep / 2
	right := keep - left
	if left < 1 {
		left = 1
	}
	if right < 1 {
		right = 1
	}
	return value[:left] + "..." + value[len(value)-right:]
}
