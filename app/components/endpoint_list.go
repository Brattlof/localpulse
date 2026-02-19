package components

import (
	"strings"

	"github.com/Brattlof/localpulse/monitor"
	"github.com/Brattlof/localpulse/ui"
	"github.com/charmbracelet/lipgloss"
)

type EndpointList struct {
	Endpoints  []*monitor.Endpoint
	Selected   int
	Width      int
	Height     int
	styles     *ui.Styles
	scrollable bool
}

func NewEndpointList(styles *ui.Styles) *EndpointList {
	return &EndpointList{
		Endpoints: make([]*monitor.Endpoint, 0),
		Selected:  0,
		styles:    styles,
	}
}

func (l *EndpointList) SetEndpoints(endpoints []*monitor.Endpoint) {
	l.Endpoints = endpoints
	if l.Selected >= len(l.Endpoints) {
		l.Selected = len(l.Endpoints) - 1
	}
	if l.Selected < 0 {
		l.Selected = 0
	}
}

func (l *EndpointList) AddEndpoint(ep *monitor.Endpoint) {
	for _, existing := range l.Endpoints {
		if existing.URL == ep.URL {
			return
		}
	}
	l.Endpoints = append(l.Endpoints, ep)
}

func (l *EndpointList) RemoveSelected() {
	if len(l.Endpoints) == 0 {
		return
	}
	l.Endpoints = append(l.Endpoints[:l.Selected], l.Endpoints[l.Selected+1:]...)
	if l.Selected >= len(l.Endpoints) {
		l.Selected = len(l.Endpoints) - 1
	}
	if l.Selected < 0 {
		l.Selected = 0
	}
}

func (l *EndpointList) SetSize(width, height int) {
	l.Width = width
	l.Height = height
}

func (l *EndpointList) Up() {
	if l.Selected > 0 {
		l.Selected--
	}
}

func (l *EndpointList) Down() {
	if l.Selected < len(l.Endpoints)-1 {
		l.Selected++
	}
}

func (l *EndpointList) SelectedEndpoint() *monitor.Endpoint {
	if l.Selected < 0 || l.Selected >= len(l.Endpoints) {
		return nil
	}
	return l.Endpoints[l.Selected]
}

func (l *EndpointList) View() string {
	if len(l.Endpoints) == 0 {
		empty := l.styles.Theme.ColorMuted("No endpoints discovered.\nPress 'r' to scan or 'a' to add manually.")
		return l.styles.Panel.Width(l.Width).Height(l.Height).Render(empty)
	}

	start := 0
	end := len(l.Endpoints)

	if l.Height > 0 && len(l.Endpoints) > l.Height {
		l.scrollable = true
		start = l.Selected - l.Height/2
		if start < 0 {
			start = 0
		}
		end = start + l.Height
		if end > len(l.Endpoints) {
			end = len(l.Endpoints)
			start = end - l.Height
		}
	}

	var lines []string
	for i := start; i < end; i++ {
		ep := l.Endpoints[i]
		selected := i == l.Selected

		icon := ep.StatusIcon()
		name := ui.Truncate(ep.Name, l.Width-15)
		latency := ui.FormatLatency(ep.LastLatency.Nanoseconds())

		if selected {
			line := lipgloss.JoinHorizontal(
				lipgloss.Top,
				icon+" ",
				l.styles.EndpointSelected.Render(name),
				" ",
				l.styles.Theme.ColorMuted(latency),
			)
			lines = append(lines, line)
		} else {
			line := lipgloss.JoinHorizontal(
				lipgloss.Top,
				icon+" ",
				name,
				" ",
				l.styles.Theme.ColorMuted(latency),
			)
			lines = append(lines, l.styles.Endpoint.Render(line))
		}
	}

	content := strings.Join(lines, "\n")
	return l.styles.Panel.Width(l.Width).Height(l.Height).Render(content)
}

func (l *EndpointList) Count() int {
	return len(l.Endpoints)
}

func (l *EndpointList) StatusCounts() (healthy, slow, down int) {
	for _, ep := range l.Endpoints {
		switch ep.Status {
		case monitor.StatusHealthy:
			healthy++
		case monitor.StatusSlow:
			slow++
		case monitor.StatusDown:
			down++
		}
	}
	return
}
