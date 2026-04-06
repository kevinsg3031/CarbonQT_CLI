package report

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"carbonqt/internal/models"

	"github.com/phpdave11/gofpdf"
)

type PDFOptions struct {
	Title        string
	Subtitle     string
	MaxTableRows int
	TopProcesses int
}

func WritePDF(basePath string, data ReportData, opts PDFOptions) (string, error) {
	base := strings.TrimSpace(basePath)
	if base == "" {
		return "", fmt.Errorf("output path is required")
	}

	path := base
	if strings.HasSuffix(strings.ToLower(path), ".pdf") {
		path = strings.TrimSuffix(path, filepath.Ext(path))
	}
	path = path + ".pdf"

	if err := ensureDir(path); err != nil {
		return "", err
	}

	if opts.Title == "" {
		opts.Title = "CarbonQT Report"
	}
	if opts.Subtitle == "" {
		opts.Subtitle = "Carbon and Energy Impact"
	}
	if opts.MaxTableRows <= 0 {
		opts.MaxTableRows = 15
	}
	if opts.TopProcesses <= 0 {
		opts.TopProcesses = opts.MaxTableRows
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(18, 18, 18)
	pdf.SetAutoPageBreak(true, 18)
	pdf.SetTitle(opts.Title, true)
	pdf.SetCreator("carbonqt", true)
	pdf.AliasNbPages("")
	generatedLabel := data.GeneratedAt.Format("02 Jan 2006")

	pdf.SetFooterFunc(func() {
		pdf.SetY(-14)
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(120, 120, 120)
		left := "CarbonQT Report"
		center := fmt.Sprintf("Generated on %s", generatedLabel)
		right := fmt.Sprintf("Page %d / {nb}", pdf.PageNo())
		pdf.CellFormat(60, 8, left, "", 0, "L", false, 0, "")
		pdf.CellFormat(70, 8, center, "", 0, "C", false, 0, "")
		pdf.CellFormat(0, 8, right, "", 0, "R", false, 0, "")
	})

	addCoverPage(pdf, data, opts)
	addSummaryPage(pdf, data)
	addChartsPage(pdf, data, opts)
	addTablePage(pdf, data, opts)

	if err := pdf.OutputFileAndClose(path); err != nil {
		return "", err
	}

	return path, nil
}

func addCoverPage(pdf *gofpdf.Fpdf, data ReportData, opts PDFOptions) {
	pdf.AddPage()

	pageW, _ := pdf.GetPageSize()
	bannerH := 40.0

	pdf.SetFillColor(20, 83, 45)
	pdf.Rect(0, 0, pageW, bannerH, "F")

	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", 26)
	pdf.SetXY(18, 12)
	pdf.CellFormat(0, 12, opts.Title, "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 14)
	pdf.SetX(18)
	pdf.CellFormat(0, 8, opts.Subtitle, "", 1, "L", false, 0, "")

	pdf.SetTextColor(35, 35, 35)
	pdf.SetFont("Helvetica", "", 11)
	pdf.SetY(bannerH + 10)

	info := [][]string{
		{"Generated", data.GeneratedAt.Format(time.RFC1123)},
		{"Duration", formatDuration(data.Duration)},
		{"Interval", data.Interval.String()},
	}
	if strings.TrimSpace(data.RepoRoot) != "" {
		info = append(info, []string{"Repo Root", data.RepoRoot})
	}

	y := drawKeyValueBlock(pdf, 18, bannerH+14, 170, info)

	pdf.SetY(y + 6)
	pdf.SetFont("Helvetica", "B", 15)
	pdf.CellFormat(0, 8, "Highlights", "", 1, "L", false, 0, "")

	cardX := 18.0
	cardY := pdf.GetY() + 2
	cardW := 56.0
	cardH := 40.0
	gap := 4.0
	cards := []metricCard{
		{
			Title: "Total Carbon Emitted",
			Value: fmt.Sprintf("%s mg", formatCarbonMg(data.TotalCarbonKg)),
		},
		{
			Title: "Top Process",
			Value: topProcessName(data),
		},
		{
			Title: "CPU Model",
			Value: fallbackText(data.System.CPUModel, "Unknown"),
		},
	}
	drawMetricCards(pdf, cardX, cardY, cardW, cardH, gap, cards)

	pdf.SetY(cardY + cardH + 12)
	pdf.SetFont("Helvetica", "B", 13)
	pdf.CellFormat(0, 7, "Executive Summary", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 11)
	pdf.SetTextColor(55, 55, 55)
	pdf.MultiCell(0, 6, buildExecutiveSummary(data), "", "L", false)

	pdf.Ln(3)
	pdf.SetFont("Helvetica", "I", 10)
	pdf.SetTextColor(90, 90, 90)
	pdf.MultiCell(0, 5, "Units: Carbon in mg CO2, power in W, CPU and memory in %.", "", "L", false)
}

func addSummaryPage(pdf *gofpdf.Fpdf, data ReportData) {
	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 18)
	pdf.SetTextColor(30, 30, 30)
	pdf.CellFormat(0, 10, "Summary", "", 1, "L", false, 0, "")
	pdf.Ln(2)

	leftX := 18.0
	rightX := 108.0
	topY := pdf.GetY()

	ramUsage := fmt.Sprintf("%s / %s (%.1f %%)", formatBytes(data.System.MemoryUsedBytes), formatBytes(data.System.MemoryTotalBytes), data.System.MemoryPercent)
	systemRows := [][]string{
		{"CPU", fmt.Sprintf("%s (%d cores)", fallbackText(data.System.CPUModel, "Unknown"), data.System.CPUCores)},
		{"CPU Usage", fmt.Sprintf("%.1f %%", data.System.CPUPercent)},
		{"RAM Usage", ramUsage},
		{"Platform", fallbackText(data.System.Platform, "Unknown")},
		{"Uptime", formatDuration(time.Duration(data.System.UptimeSeconds) * time.Second)},
	}

	reportRows := [][]string{
		{"Total Carbon", fmt.Sprintf("%s mg", formatCarbonMg(data.TotalCarbonKg))},
		{"CPU TDP", fmt.Sprintf("%.2f W", data.Estimator.CPUWatts)},
		{"Emission Factor", fmt.Sprintf("%.8g kg CO2/J", data.Estimator.EmissionFactor)},
		{"Processes", fmt.Sprintf("%d", len(data.Processes))},
		{"Duration", formatDuration(data.Duration)},
		{"Interval", data.Interval.String()},
	}

	drawTitledBox(pdf, leftX, topY, 84, 58, "System", systemRows)
	drawTitledBox(pdf, rightX, topY, 84, 58, "Report", reportRows)
}

func addChartsPage(pdf *gofpdf.Fpdf, data ReportData, opts PDFOptions) {
	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 18)
	pdf.SetTextColor(30, 30, 30)
	pdf.CellFormat(0, 10, "Charts", "", 1, "L", false, 0, "")
	pdf.Ln(2)

	chartX := 18.0
	chartW := 174.0
	chartH := 48.0

	trendMg := scaleKgToMg(data.Trend)
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(0, 8, "Carbon Trend (mg CO2 per interval)", "", 1, "L", false, 0, "")
	drawLineChart(pdf, chartX, pdf.GetY(), chartW, chartH, trendMg, "mg")

	pdf.SetY(pdf.GetY() + chartH + 10)
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(0, 8, "Top Processes by Carbon", "", 1, "L", false, 0, "")
	drawBarChart(pdf, chartX, pdf.GetY(), chartW, chartH, topProcessBars(data, opts.TopProcesses))
}

func addTablePage(pdf *gofpdf.Fpdf, data ReportData, opts PDFOptions) {
	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 18)
	pdf.SetTextColor(30, 30, 30)
	pdf.CellFormat(0, 10, "Processes", "", 1, "L", false, 0, "")
	pdf.Ln(2)

	header := []string{"PID", "Name", "CPU (%)", "MEM (%)", "Power (W)", "Carbon (mg)"}
	widths := []float64{14, 70, 14, 14, 22, 20}
	rowH := 6.0
	headerH := 7.0
	rows := topProcesses(data, opts.MaxTableRows)
	pageW, pageH := pdf.GetPageSize()
	bottomMargin := 18.0

	if len(rows) == 0 {
		pdf.SetFont("Helvetica", "", 11)
		pdf.SetTextColor(80, 80, 80)
		pdf.CellFormat(0, 8, "No process data available.", "", 1, "L", false, 0, "")
		return
	}

	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(70, 70, 70)
	pdf.CellFormat(0, 6, fmt.Sprintf("Showing top %d processes by estimated carbon.", len(rows)), "", 1, "L", false, 0, "")
	pdf.Ln(1)

	writeHeader := func() {
		pdf.SetFont("Helvetica", "B", 10)
		pdf.SetFillColor(230, 238, 233)
		for i, label := range header {
			pdf.CellFormat(widths[i], headerH, label, "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)
	}

	writeHeader()
	pdf.SetFont("Helvetica", "", 9)

	for _, proc := range rows {
		if pdf.GetY()+rowH > pageH-bottomMargin {
			pdf.AddPage()
			writeHeader()
			pdf.SetFont("Helvetica", "", 9)
		}
		pdf.CellFormat(widths[0], rowH, fmt.Sprintf("%d", proc.PID), "1", 0, "C", false, 0, "")
		pdf.CellFormat(widths[1], rowH, truncateMiddle(proc.Name, 24), "1", 0, "L", false, 0, "")
		pdf.CellFormat(widths[2], rowH, fmt.Sprintf("%.1f", proc.CPUPercent), "1", 0, "R", false, 0, "")
		pdf.CellFormat(widths[3], rowH, fmt.Sprintf("%.1f", proc.MemoryPercent), "1", 0, "R", false, 0, "")
		pdf.CellFormat(widths[4], rowH, fmt.Sprintf("%.2f", proc.PowerW), "1", 0, "R", false, 0, "")
		pdf.CellFormat(widths[5], rowH, formatCarbonMg(proc.CarbonKg), "1", 0, "R", false, 0, "")
		pdf.Ln(-1)
	}

	_ = pageW
}

func drawKeyValueBlock(pdf *gofpdf.Fpdf, x, y, w float64, rows [][]string) float64 {
	labelW := 28.0
	lineH := 6.0
	pdf.SetXY(x, y)
	for _, row := range rows {
		startX, startY := pdf.GetXY()
		pdf.SetFont("Helvetica", "B", 10)
		pdf.CellFormat(labelW, lineH, row[0]+":", "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 10)
		pdf.MultiCell(w-labelW, lineH, row[1], "", "L", false)
		if pdf.GetY() == startY {
			pdf.SetXY(startX, startY+lineH)
		}
		pdf.SetX(x)
	}

	_, endY := pdf.GetXY()
	return endY
}

func drawTitledBox(pdf *gofpdf.Fpdf, x, y, w, h float64, title string, rows [][]string) {
	pdf.SetDrawColor(200, 200, 200)
	pdf.SetFillColor(246, 248, 246)
	pdf.Rect(x, y, w, h, "DF")

	pdf.SetXY(x+4, y+4)
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(30, 30, 30)
	pdf.CellFormat(0, 6, title, "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(60, 60, 60)
	currentY := y + 12
	for _, row := range rows {
		pdf.SetXY(x+4, currentY)
		pdf.SetFont("Helvetica", "B", 9)
		pdf.CellFormat(24, 5, row[0]+":", "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 9)
		pdf.MultiCell(w-10-24, 5, row[1], "", "L", false)
		currentY = pdf.GetY()
		if currentY < y+h-6 {
			currentY += 1
		}
	}
}

func drawLineChart(pdf *gofpdf.Fpdf, x, y, w, h float64, values []float64, unit string) {
	pdf.SetDrawColor(180, 180, 180)
	pdf.Rect(x, y, w, h, "D")

	if len(values) == 0 {
		pdf.SetXY(x, y+h/2)
		pdf.SetFont("Helvetica", "", 10)
		pdf.SetTextColor(120, 120, 120)
		pdf.CellFormat(w, 6, "No data", "", 0, "C", false, 0, "")
		pdf.SetY(y + h)
		return
	}

	minVal, maxVal := values[0], values[0]
	for _, v := range values {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal == minVal {
		maxVal = minVal + 1
	}

	points := len(values)
	stepX := w / float64(max(points-1, 1))

	pdf.SetDrawColor(25, 115, 59)
	for i := 1; i < points; i++ {
		x1 := x + float64(i-1)*stepX
		x2 := x + float64(i)*stepX
		y1 := y + h - scaleValue(values[i-1], minVal, maxVal, h)
		y2 := y + h - scaleValue(values[i], minVal, maxVal, h)
		pdf.Line(x1, y1, x2, y2)
	}

	pdf.SetTextColor(90, 90, 90)
	pdf.SetFont("Helvetica", "", 8)
	pdf.SetXY(x, y+h+2)
	pdf.CellFormat(w, 4, fmt.Sprintf("Min: %.4f %s   |   Max: %.4f %s", minVal, unit, maxVal, unit), "", 0, "L", false, 0, "")
	pdf.SetY(y + h + 6)
}

func drawBarChart(pdf *gofpdf.Fpdf, x, y, w, h float64, values []processBar) {
	pdf.SetDrawColor(180, 180, 180)
	pdf.Rect(x, y, w, h, "D")

	if len(values) == 0 {
		pdf.SetXY(x, y+h/2)
		pdf.SetFont("Helvetica", "", 10)
		pdf.SetTextColor(120, 120, 120)
		pdf.CellFormat(w, 6, "No data", "", 0, "C", false, 0, "")
		pdf.SetY(y + h)
		return
	}

	maxVal := values[0].Value
	for _, bar := range values {
		if bar.Value > maxVal {
			maxVal = bar.Value
		}
	}
	if maxVal == 0 {
		maxVal = 1
	}

	barGap := 4.0
	barH := (h - barGap*float64(len(values)+1)) / float64(len(values))
	if barH < 4 {
		barH = 4
	}

	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(40, 40, 40)

	currentY := y + barGap
	for _, bar := range values {
		barW := (bar.Value / maxVal) * (w - 50)
		pdf.SetXY(x+2, currentY)
		pdf.CellFormat(52, barH, truncateMiddle(bar.Label, 22), "", 0, "L", false, 0, "")

		pdf.SetFillColor(33, 150, 83)
		pdf.Rect(x+57, currentY, barW, barH, "F")

		pdf.SetXY(x+57+barW+2, currentY)
		pdf.CellFormat(0, barH, fmt.Sprintf("%s mg", formatCarbonMg(bar.Value)), "", 0, "L", false, 0, "")
		currentY += barH + barGap
	}

	pdf.SetY(y + h)
}

type processBar struct {
	Label string
	Value float64
}

func topProcessBars(data ReportData, limit int) []processBar {
	rows := topProcesses(data, limit)
	if len(rows) == 0 {
		return nil
	}

	result := make([]processBar, 0, len(rows))
	for _, proc := range rows {
		result = append(result, processBar{
			Label: fmt.Sprintf("%s (PID %d)", fallbackText(proc.Name, "n/a"), proc.PID),
			Value: proc.CarbonKg,
		})
	}
	return result
}

type metricCard struct {
	Title string
	Value string
}

func drawMetricCards(pdf *gofpdf.Fpdf, x, y, w, h, gap float64, cards []metricCard) {
	for i, card := range cards {
		cardX := x + float64(i)*(w+gap)
		pdf.SetDrawColor(200, 200, 200)
		pdf.SetFillColor(245, 248, 245)
		pdf.Rect(cardX, y, w, h, "DF")

		pdf.SetXY(cardX+4, y+4)
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetTextColor(60, 60, 60)
		pdf.MultiCell(w-8, 4.5, card.Title, "", "L", false)

		valueY := y + h/2
		pdf.SetXY(cardX+4, valueY)
		pdf.SetFont("Helvetica", "B", 12)
		pdf.SetTextColor(25, 25, 25)
		pdf.MultiCell(w-8, 5.5, card.Value, "", "L", false)
	}
}

func buildExecutiveSummary(data ReportData) string {
	primary := topProcessName(data)
	return fmt.Sprintf(
		"During the %s monitoring interval, the system emitted approximately %s mg of CO2, primarily driven by %s. CPU usage averaged %.1f %%, while memory utilization reached %.1f %%.",
		formatDuration(data.Duration),
		formatCarbonMg(data.TotalCarbonKg),
		primary,
		data.System.CPUPercent,
		data.System.MemoryPercent,
	)
}

func scaleKgToMg(values []float64) []float64 {
	if len(values) == 0 {
		return nil
	}
	result := make([]float64, len(values))
	for i, value := range values {
		result[i] = value * 1_000_000
	}
	return result
}

func topProcesses(data ReportData, limit int) []models.ProcessMetrics {
	if len(data.Processes) == 0 {
		return nil
	}

	copySlice := make([]models.ProcessMetrics, len(data.Processes))
	copy(copySlice, data.Processes)
	sort.Slice(copySlice, func(i, j int) bool { return copySlice[i].CarbonKg > copySlice[j].CarbonKg })

	if limit <= 0 || limit >= len(copySlice) {
		return copySlice
	}
	return copySlice[:limit]
}

func topProcessName(data ReportData) string {
	if len(data.Processes) == 0 {
		return "n/a"
	}
	best := data.Processes[0]
	return fallbackText(best.Name, "n/a")
}

func fallbackText(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func scaleValue(value, minVal, maxVal, height float64) float64 {
	ratio := (value - minVal) / (maxVal - minVal)
	return ratio * height
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
