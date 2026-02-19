package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/localpulse/localpulse/ui"
)

type SparklineChart struct {
	Data   []float64
	Title  string
	Width  int
	Height int
	styles *ui.Styles
}

func NewSparklineChart(title string, styles *ui.Styles) *SparklineChart {
	return &SparklineChart{
		Data:   make([]float64, 0),
		Title:  title,
		styles: styles,
	}
}

func (c *SparklineChart) SetData(data []float64) {
	c.Data = data
}

func (c *SparklineChart) AddPoint(value float64) {
	c.Data = append(c.Data, value)
	maxPoints := c.Width - 4
	if maxPoints < 10 {
		maxPoints = 30
	}
	if len(c.Data) > maxPoints {
		c.Data = c.Data[len(c.Data)-maxPoints:]
	}
}

func (c *SparklineChart) SetSize(width, height int) {
	c.Width = width
	c.Height = height
}

func (c *SparklineChart) View() string {
	titleLine := c.styles.CardTitle.Render(c.Title)
	chartLine := ui.Sparkline(c.Data, c.Height, c.styles.Theme)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		titleLine,
		"",
		chartLine,
	)

	return c.styles.Panel.Width(c.Width).Height(c.Height).Render(content)
}

type ChartPanel struct {
	LatencyChart    *SparklineChart
	ThroughputChart *SparklineChart
	Width           int
	Height          int
	styles          *ui.Styles
}

func NewChartPanel(styles *ui.Styles) *ChartPanel {
	return &ChartPanel{
		LatencyChart:    NewSparklineChart("Latency (ms)", styles),
		ThroughputChart: NewSparklineChart("Throughput (req/s)", styles),
		styles:          styles,
	}
}

func (p *ChartPanel) SetSize(width, height int) {
	p.Width = width
	p.Height = height

	chartHeight := height/2 - 2
	if chartHeight < 4 {
		chartHeight = 4
	}

	p.LatencyChart.SetSize(width, chartHeight)
	p.ThroughputChart.SetSize(width, chartHeight)
}

func (p *ChartPanel) AddLatencyPoint(latencyMs float64) {
	p.LatencyChart.AddPoint(latencyMs)
}

func (p *ChartPanel) AddThroughputPoint(rps float64) {
	p.ThroughputChart.AddPoint(rps)
}

func (p *ChartPanel) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Top,
		p.LatencyChart.View(),
		"",
		p.ThroughputChart.View(),
	)
}
