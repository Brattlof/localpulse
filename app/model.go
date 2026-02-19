package app

import (
	"github.com/Brattlof/localpulse/app/components"
	"github.com/Brattlof/localpulse/config"
	"github.com/Brattlof/localpulse/monitor"
	"github.com/Brattlof/localpulse/ui"
)

type FocusPanel int

const (
	FocusEndpoints FocusPanel = iota
	FocusCharts
	FocusLogs
)

type ModelState int

const (
	StateIdle ModelState = iota
	StateScanning
	StateLoadTesting
	StateAddingEndpoint
)

type Model struct {
	state    ModelState
	focus    FocusPanel
	width    int
	height   int
	quitting bool

	config *config.Config
	theme  ui.Theme
	styles *ui.Styles

	scanner       *monitor.Scanner
	loadGenerator *monitor.LoadGenerator
	sysMonitor    *monitor.SystemMonitor

	endpoints    []*monitor.Endpoint
	endpointList *components.EndpointList
	selectedIdx  int

	metricsMap map[string]*monitor.Metrics

	summaryPanel *components.SummaryPanel
	chartPanel   *components.ChartPanel
	logPanel     *components.LogPanel
	inputForm    *components.InputForm

	healthy int
	slow    int
	down    int

	rps int
}

func NewModel(cfg *config.Config) Model {
	theme := ui.DetectTheme()
	styles := ui.NewStyles(theme)

	summaryPanel := components.NewSummaryPanel(styles)
	chartPanel := components.NewChartPanel(styles)
	logPanel := components.NewLogPanel(100, styles)
	inputForm := components.NewInputForm(styles)
	endpointList := components.NewEndpointList(styles)

	return Model{
		state:         StateIdle,
		focus:         FocusEndpoints,
		config:        cfg,
		theme:         theme,
		styles:        styles,
		scanner:       monitor.NewScanner(monitor.WithPorts(cfg.DefaultPorts)),
		loadGenerator: monitor.NewLoadGenerator(),
		sysMonitor:    monitor.NewSystemMonitor(),
		endpointList:  endpointList,
		metricsMap:    make(map[string]*monitor.Metrics),
		summaryPanel:  summaryPanel,
		chartPanel:    chartPanel,
		logPanel:      logPanel,
		inputForm:     inputForm,
		rps:           cfg.LoadTestRPS,
	}
}

func (m Model) GetEndpoints() []*monitor.Endpoint {
	return m.endpoints
}

func (m Model) GetMetrics(url string) *monitor.Metrics {
	return m.metricsMap[url]
}

func (m Model) IsQuitting() bool {
	return m.quitting
}
