package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"carbonqt/internal/energy"
	"carbonqt/internal/models"
	"carbonqt/internal/monitor"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DashboardConfig struct {
	Estimator energy.Estimator
	RepoRoot  string
	Refresh   time.Duration
}

type dashboardModel struct {
	config      DashboardConfig
	system      models.SystemMetrics
	processes   []models.ProcessMetrics
	totalCarbon float64
	width       int
	height      int
	selected    int
	armed       bool
	armedPID    int32
	status      string
	err         error
	loading     bool
	splashIndex int
}

type dashboardTickMsg time.Time
type splashTickMsg time.Time
type splashDoneMsg struct{}

const (
	splashDuration = 2 * time.Second
	splashInterval = 120 * time.Millisecond
)

func StartDashboard(config DashboardConfig) error {
	model := dashboardModel{config: config, loading: true}
	program := tea.NewProgram(model, tea.WithAltScreen())
	_, err := program.Run()
	return err
}

func (m dashboardModel) Init() tea.Cmd {
	return tea.Batch(
		tea.Tick(splashInterval, func(t time.Time) tea.Msg {
			return splashTickMsg(t)
		}),
		tea.Tick(splashDuration, func(time.Time) tea.Msg {
			return splashDoneMsg{}
		}),
	)
}

func (m dashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case dashboardTickMsg:
		if m.loading {
			return m, nil
		}
		system, err := monitor.GetSystemMetrics()
		if err != nil {
			m.err = err
			return m, nil
		}

		processes, err := monitor.ListProcesses(m.config.RepoRoot)
		if err != nil {
			m.err = err
			return m, nil
		}

		total, processes := m.config.Estimator.ApplyCarbon(processes, m.config.Refresh)
		sort.Slice(processes, func(i, j int) bool { return processes[i].CarbonKg > processes[j].CarbonKg })

		if m.selected >= len(processes) {
			m.selected = len(processes) - 1
		}
		if m.selected < 0 {
			m.selected = 0
		}
		if m.armed {
			if findPIDIndex(processes, m.armedPID) == -1 {
				m.armed = false
				m.armedPID = 0
			}
		}

		m.system = system
		m.totalCarbon = total
		m.processes = processes
		return m, tea.Tick(m.config.Refresh, func(t time.Time) tea.Msg {
			return dashboardTickMsg(t)
		})
	case splashTickMsg:
		if !m.loading {
			return m, nil
		}
		m.splashIndex++
		return m, tea.Tick(splashInterval, func(t time.Time) tea.Msg {
			return splashTickMsg(t)
		})
	case splashDoneMsg:
		if !m.loading {
			return m, nil
		}
		m.loading = false
		return m, tea.Tick(m.config.Refresh, func(t time.Time) tea.Msg {
			return dashboardTickMsg(t)
		})
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up":
			if m.selected > 0 {
				m.selected--
				if m.armed && m.selected < len(m.processes) && m.processes[m.selected].PID != m.armedPID {
					m.armed = false
					m.armedPID = 0
				}
			}
		case "down":
			if m.selected < len(m.processes)-1 {
				m.selected++
				if m.armed && m.selected < len(m.processes) && m.processes[m.selected].PID != m.armedPID {
					m.armed = false
					m.armedPID = 0
				}
			}
		case " ", "space":
			if len(m.processes) == 0 || m.selected < 0 || m.selected >= len(m.processes) {
				m.status = "No process to select."
				return m, nil
			}
			pid := m.processes[m.selected].PID
			name := m.processes[m.selected].Name
			if m.armed && m.armedPID == pid {
				m.armed = false
				m.armedPID = 0
				m.status = "Selection cleared."
				return m, nil
			}
			m.armed = true
			m.armedPID = pid
			m.status = fmt.Sprintf("Selected %s (PID %d). Press K to kill.", name, pid)
		case "k":
			if len(m.processes) == 0 || m.selected < 0 || m.selected >= len(m.processes) {
				m.status = "No process selected."
				return m, nil
			}
			if !m.armed {
				m.status = "Press Space to select a process before killing."
				return m, nil
			}
			pid := m.processes[m.selected].PID
			name := m.processes[m.selected].Name
			if pid != m.armedPID {
				m.status = "Selection changed. Press Space to select again."
				return m, nil
			}
			if err := monitor.KillProcess(pid); err != nil {
				m.status = fmt.Sprintf("Failed to kill %s (PID %d).", name, pid)
				return m, nil
			}
			m.armed = false
			m.armedPID = 0
			m.status = fmt.Sprintf("Killed %s (PID %d).", name, pid)
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m dashboardModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}

	width := m.width
	if width <= 0 {
		width = 120
	}
	if m.loading {
		return m.splashView(width)
	}

	systemInfoBlock := strings.Join([]string{
		fmt.Sprintf("CPU: %s (%d cores)", fallback(m.system.CPUModel, "Unknown"), m.system.CPUCores),
		fmt.Sprintf("RAM: %s total", formatBytes(m.system.MemoryTotalBytes)),
		fmt.Sprintf("Platform: %s", fallback(m.system.Platform, "Unknown")),
		fmt.Sprintf("Uptime: %s", formatDuration(time.Duration(m.system.UptimeSeconds)*time.Second)),
	}, "\n")

	systemBlock := strings.Join([]string{
		"CPU Usage " + progressBar(24, m.system.CPUPercent) + fmt.Sprintf(" %.1f%%", m.system.CPUPercent),
		"RAM Usage " + progressBar(24, m.system.MemoryPercent) + fmt.Sprintf(" %.1f%%", m.system.MemoryPercent),
	}, "\n")

	carbonBlock := fmt.Sprintf("Estimated Emissions: %s mg", formatCarbonMg(m.totalCarbon))

	var topProcessBlock string
	if len(m.processes) > 0 {
		top := m.processes[0]
		topProcessBlock = strings.Join([]string{
			"Highest Carbon Process",
			fallback(top.Name, "Unknown"),
			fmt.Sprintf("Power: %.2f W", top.PowerW),
			fmt.Sprintf("Carbon: %s mg", formatCarbonMg(top.CarbonKg)),
		}, "\n")
	} else {
		topProcessBlock = "Highest Carbon Process\n(n/a)"
	}

	stackPanels := width < 120
	panelWidth := width
	if !stackPanels {
		panelWidth = (width - 6) / 2
		if panelWidth < 50 {
			panelWidth = 50
		}
	}
	panelBox := panelStyle.Width(panelWidth)
	badgeRow := strings.Join([]string{
		renderBadge(fmt.Sprintf("CPU %.1f%%", m.system.CPUPercent)),
		renderBadge(fmt.Sprintf("RAM %.1f%%", m.system.MemoryPercent)),
		renderBadge(fmt.Sprintf("Total %s mg", formatCarbonMg(m.totalCarbon))),
	}, " ")

	systemPanel := panelBox.Render(panelTitleStyle.Render("System") + "\n" + systemInfoBlock + "\n\n" + systemBlock + "\n\n" + badgeRow)
	carbonPanel := panelBox.Render(panelTitleStyle.Render("Carbon") + "\n" + carbonBlock + "\n\n" + panelTitleStyle.Render("Top Process") + "\n" + topProcessBlock)

	var upperRow string
	if stackPanels {
		upperRow = lipgloss.JoinVertical(lipgloss.Left, systemPanel, carbonPanel)
	} else {
		upperRow = lipgloss.JoinHorizontal(lipgloss.Top, systemPanel, carbonPanel)
	}

	processBlock := RenderDashboardTableSelected(m.processes, m.selected, 12)

	status := strings.TrimSpace(m.status)
	ribbon := buildHelpBar(width, status)

	title := titleStyle.Width(width).Align(lipgloss.Center).Render("CarbonQT Dashboard")
	processPanel := panelStyle.Width(width).Render(processBlock)
	sections := []string{
		title,
		upperRow,
		panelTitleStyle.Render("Processes"),
		processPanel,
		ribbon,
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(strings.Join(sections, "\n\n"))
}

func buildHelpBar(width int, status string) string {
	items := []string{
		helpKeyStyle.Render("Up/Down"),
		helpTextStyle.Render("Navigate"),
		helpKeyStyle.Render("Space"),
		helpTextStyle.Render("Select"),
		helpKeyStyle.Render("K"),
		helpTextStyle.Render("Kill"),
		helpKeyStyle.Render("Q"),
		helpTextStyle.Render("Quit"),
	}
	help := strings.Join(items, " ")
	if status != "" {
		help = status + "  |  " + help
	}
	return helpBarStyle.Width(width).Align(lipgloss.Center).Render(help)
}

func findPIDIndex(processes []models.ProcessMetrics, pid int32) int {
	for i, proc := range processes {
		if proc.PID == pid {
			return i
		}
	}
	return -1
}

func (m dashboardModel) splashView(width int) string {
	frames := []string{"-", "\\", "|", "/"}
	frame := frames[m.splashIndex%len(frames)]
	art := []string{
		"  ____                 _______                 ",
		" / ___|_ __ ___  ___  |_   _| __ __ _  ___ ___ ",
		"| |  _| '__/ _ \\/ _ \\   | || '__/ _` |/ __/ _ \\",
		"| |_| | | |  __/  __/   | || | | (_| | (_|  __/",
		" \\____|_|  \\___|\\___|   |_||_|  \\__,_|\\___\\___|",
	}
	statusText := fmt.Sprintf("Analyzing system metrics... %s", frame)
	body := strings.Join(art, "\n")

	content := strings.Join([]string{
		splashTitle.Width(width).Align(lipgloss.Center).Render(body),
		"",
		splashSubtitle.Width(width).Align(lipgloss.Center).Render(statusText),
	}, "\n")
	return lipgloss.NewStyle().Padding(2, 2).Render(content)
}
