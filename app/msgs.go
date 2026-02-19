package app

import (
	"context"
	"time"

	"github.com/Brattlof/localpulse/monitor"
	tea "github.com/charmbracelet/bubbletea"
)

type TickMsg time.Time
type ScanCompleteMsg struct {
	Endpoints []*monitor.Endpoint
}
type MetricsUpdateMsg struct {
	URL   string
	Stats monitor.Stats
}

func DoTick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func DoScan(scanner *monitor.Scanner) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		endpoints := scanner.DiscoverEndpoints(ctx)
		return ScanCompleteMsg{Endpoints: endpoints}
	}
}

func DoMetricsUpdate(url string, metrics *monitor.Metrics) tea.Cmd {
	return func() tea.Msg {
		stats := metrics.GetStats()
		return MetricsUpdateMsg{URL: url, Stats: stats}
	}
}
