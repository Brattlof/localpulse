package app

import tea "github.com/charmbracelet/bubbletea"

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		DoTick(),
		DoScan(m.scanner),
	)
}
