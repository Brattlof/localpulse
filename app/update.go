package app

import (
	"strings"

	"github.com/Brattlof/localpulse/monitor"
	"github.com/Brattlof/localpulse/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()
		return m, nil

	case TickMsg:
		return m.handleTick(msg)

	case ScanCompleteMsg:
		return m.handleScanComplete(msg)

	case MetricsUpdateMsg:
		return m.handleMetricsUpdate(msg)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.inputForm.IsActive() {
		return m.handleInputFormKeys(msg)
	}

	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "r":
		m.state = StateScanning
		return m, DoScan(m.scanner)

	case "tab":
		m.focus = (m.focus + 1) % 3
		return m, nil

	case "shift+tab":
		m.focus = (m.focus - 1 + 3) % 3
		return m, nil

	case "up", "k":
		if m.focus == FocusEndpoints {
			m.endpointList.Up()
			m.selectedIdx = m.endpointList.Selected
		}
		return m, nil

	case "down", "j":
		if m.focus == FocusEndpoints {
			m.endpointList.Down()
			m.selectedIdx = m.endpointList.Selected
		}
		return m, nil

	case "+", "=":
		m.rps += 5
		if m.loadGenerator.IsRunning() {
			m.loadGenerator.SetRPS(m.rps)
		}
		return m, nil

	case "-":
		m.rps -= 5
		if m.rps < 1 {
			m.rps = 1
		}
		if m.loadGenerator.IsRunning() {
			m.loadGenerator.SetRPS(m.rps)
		}
		return m, nil

	case "enter":
		if m.focus == FocusEndpoints && len(m.endpoints) > 0 {
			return m.toggleLoadTesting()
		}
		return m, nil

	case "a":
		m.state = StateAddingEndpoint
		m.inputForm.Focus()
		return m, m.inputForm.Init()

	case "d":
		if m.focus == FocusEndpoints && len(m.endpoints) > 0 {
			m.removeSelectedEndpoint()
		}
		return m, nil

	case "s":
		if len(m.endpoints) > 0 {
			return m.startLoadTesting()
		}
		return m, nil

	case "x":
		m.stopLoadTesting()
		return m, nil
	}

	return m, nil
}

func (m Model) handleInputFormKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.inputForm, cmd = m.inputForm.Update(msg)

	if m.inputForm.IsSubmitted() {
		url := m.inputForm.Value()
		if url != "" {
			ep, err := monitor.NewEndpoint(url)
			if err == nil {
				m.addEndpoint(ep)
				m.logPanel.AddEntry("Added endpoint: "+ep.URL, false, true)
			}
		}
		m.inputForm.Acknowledge()
		m.state = StateIdle
		return m, nil
	}

	if m.inputForm.IsCancelled() {
		m.inputForm.Acknowledge()
		m.state = StateIdle
		return m, nil
	}

	return m, cmd
}

func (m Model) handleTick(msg TickMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	m.healthy, m.slow, m.down = m.endpointList.StatusCounts()

	sysMetrics := m.sysMonitor.GetMetrics()
	m.summaryPanel.UpdateCPU(sysMetrics.CPUPercent)
	m.summaryPanel.UpdateRAM(sysMetrics.RAMUsed, sysMetrics.RAMTotal)

	if m.state == StateLoadTesting {
		for url, metrics := range m.metricsMap {
			cmds = append(cmds, DoMetricsUpdate(url, metrics))
		}
	}

	cmds = append(cmds, DoTick())
	return m, tea.Batch(cmds...)
}

func (m Model) handleScanComplete(msg ScanCompleteMsg) (tea.Model, tea.Cmd) {
	m.state = StateIdle

	for _, ep := range msg.Endpoints {
		m.addEndpoint(ep)
	}

	m.logPanel.AddEntry(
		"Scan complete: found "+itoa(len(msg.Endpoints))+" endpoints",
		false,
		false,
	)

	return m, DoTick()
}

func (m Model) handleMetricsUpdate(msg MetricsUpdateMsg) (tea.Model, tea.Cmd) {
	m.chartPanel.AddLatencyPoint(float64(msg.Stats.AvgLatency.Milliseconds()))
	m.chartPanel.AddThroughputPoint(msg.Stats.Throughput)

	m.summaryPanel.UpdateMetrics(
		int64(msg.Stats.Throughput),
		msg.Stats.AvgLatency.Nanoseconds(),
		msg.Stats.ErrorRate,
	)

	return m, nil
}

func (m *Model) addEndpoint(ep *monitor.Endpoint) {
	for _, existing := range m.endpoints {
		if existing.URL == ep.URL {
			return
		}
	}

	m.endpoints = append(m.endpoints, ep)
	m.metricsMap[ep.URL] = monitor.NewMetrics(1000)
	m.endpointList.SetEndpoints(m.endpoints)

	m.config.AddEndpoint(ep.URL, ep.Name)
}

func (m *Model) removeSelectedEndpoint() {
	if m.selectedIdx < 0 || m.selectedIdx >= len(m.endpoints) {
		return
	}

	ep := m.endpoints[m.selectedIdx]
	delete(m.metricsMap, ep.URL)
	m.config.RemoveEndpoint(ep.URL)

	m.endpointList.RemoveSelected()
	m.endpoints = m.endpointList.Endpoints

	if m.selectedIdx >= len(m.endpoints) {
		m.selectedIdx = len(m.endpoints) - 1
	}
}

func (m Model) toggleLoadTesting() (tea.Model, tea.Cmd) {
	if m.loadGenerator.IsRunning() {
		m.stopLoadTesting()
		return m, nil
	}
	return m.startLoadTesting()
}

func (m Model) startLoadTesting() (tea.Model, tea.Cmd) {
	m.state = StateLoadTesting

	for _, ep := range m.endpoints {
		metrics := m.metricsMap[ep.URL]
		if metrics == nil {
			metrics = monitor.NewMetrics(1000)
			m.metricsMap[ep.URL] = metrics
		}
		m.loadGenerator.AddTester(ep, metrics)
	}

	m.loadGenerator.Start(m.rps)
	m.logPanel.AddEntry("Load testing started at "+itoa(m.rps)+" req/s", false, true)

	return m, DoTick()
}

func (m *Model) stopLoadTesting() {
	m.loadGenerator.Stop()
	m.state = StateIdle
	m.logPanel.AddEntry("Load testing stopped", false, false)
}

func (m *Model) updateLayout() {
	summaryHeight := 4
	logHeight := 8
	helpHeight := 2
	remainingHeight := m.height - summaryHeight - logHeight - helpHeight

	leftWidth := m.width / 2
	rightWidth := m.width - leftWidth

	m.summaryPanel.SetWidth(m.width)
	m.endpointList.SetSize(leftWidth-2, remainingHeight-2)
	m.chartPanel.SetSize(rightWidth-2, remainingHeight-2)
	m.logPanel.SetSize(m.width-2, logHeight-2)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var result string
	isNegative := n < 0
	if isNegative {
		n = -n
	}
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	if isNegative {
		result = "-" + result
	}
	return result
}

func (m Model) View() string {
	if m.quitting {
		return "\n  LocalPulse stopped. Config saved.\n\n"
	}

	var b strings.Builder

	b.WriteString(m.summaryPanel.View())
	b.WriteString("\n")

	leftPanel := m.endpointList.View()
	rightPanel := m.chartPanel.View()

	panels := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPanel,
		" ",
		rightPanel,
	)
	b.WriteString(panels)
	b.WriteString("\n")

	b.WriteString(m.logPanel.View())
	b.WriteString("\n")

	if m.inputForm.IsActive() {
		b.WriteString(m.inputForm.View())
		b.WriteString("\n")
	}

	b.WriteString(m.renderHelpBar())
	b.WriteString("\n")

	return b.String()
}

func (m Model) renderHelpBar() string {
	var keys []ui.HelpKey

	if m.state == StateLoadTesting {
		keys = []ui.HelpKey{
			{Key: "x", Desc: "stop load"},
			{Key: "+/-", Desc: "adjust rps"},
			{Key: "q", Desc: "quit"},
		}
	} else if m.inputForm.IsActive() {
		keys = []ui.HelpKey{
			{Key: "enter", Desc: "submit"},
			{Key: "esc", Desc: "cancel"},
		}
	} else {
		keys = []ui.HelpKey{
			{Key: "r", Desc: "scan"},
			{Key: "s", Desc: "start load"},
			{Key: "a", Desc: "add"},
			{Key: "d", Desc: "delete"},
			{Key: "q", Desc: "quit"},
		}
	}

	rpsInfo := " [RPS: " + itoa(m.rps) + "]"
	return m.styles.HelpBar(keys, m.width-len(rpsInfo)-2) + m.styles.Theme.ColorAccent(rpsInfo)
}
