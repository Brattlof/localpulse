package components

import (
	"github.com/Brattlof/localpulse/ui"
	"github.com/charmbracelet/lipgloss"
)

type SummaryCard struct {
	Title  string
	Value  string
	Width  int
	styles *ui.Styles
}

func NewSummaryCard(title, value string, width int, styles *ui.Styles) SummaryCard {
	return SummaryCard{
		Title:  title,
		Value:  value,
		Width:  width,
		styles: styles,
	}
}

func (c SummaryCard) View() string {
	return c.styles.CardView(c.Title, c.Value, c.Width)
}

type SummaryPanel struct {
	CPU        string
	RAM        string
	ReqPerSec  string
	AvgLatency string
	ErrorRate  string
	Width      int
	styles     *ui.Styles
}

func NewSummaryPanel(styles *ui.Styles) *SummaryPanel {
	return &SummaryPanel{
		CPU:        "--",
		RAM:        "--",
		ReqPerSec:  "0",
		AvgLatency: "--",
		ErrorRate:  "0%",
		styles:     styles,
	}
}

func (p *SummaryPanel) SetWidth(width int) {
	p.Width = width
}

func (p *SummaryPanel) View() string {
	cardWidth := (p.Width - 4) / 5
	if cardWidth < 12 {
		cardWidth = 12
	}

	cards := []string{
		NewSummaryCard("CPU", p.CPU, cardWidth, p.styles).View(),
		NewSummaryCard("RAM", p.RAM, cardWidth, p.styles).View(),
		NewSummaryCard("Req/s", p.ReqPerSec, cardWidth, p.styles).View(),
		NewSummaryCard("Latency", p.AvgLatency, cardWidth, p.styles).View(),
		NewSummaryCard("Errors", p.ErrorRate, cardWidth, p.styles).View(),
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, cards...)
}

func (p *SummaryPanel) UpdateCPU(cpu float64) {
	p.CPU = ui.FormatPercent(cpu)
}

func (p *SummaryPanel) UpdateRAM(used, total uint64) {
	p.RAM = ui.Truncate(ui.FormatBytes(int64(used)), 8) + "/" + ui.Truncate(ui.FormatBytes(int64(total)), 8)
}

func (p *SummaryPanel) UpdateMetrics(rps, avgLatency int64, errorRate float64) {
	p.ReqPerSec = ui.FormatRPS(float64(rps))
	p.AvgLatency = ui.FormatLatency(avgLatency)
	p.ErrorRate = ui.FormatPercent(errorRate)
}
