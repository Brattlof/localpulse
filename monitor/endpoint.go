package monitor

import (
	"net/url"
	"strings"
	"time"
)

type EndpointStatus string

const (
	StatusHealthy EndpointStatus = "healthy"
	StatusSlow    EndpointStatus = "slow"
	StatusDown    EndpointStatus = "down"
	StatusUnknown EndpointStatus = "unknown"
)

type Endpoint struct {
	URL         string         `json:"url"`
	Name        string         `json:"name"`
	Status      EndpointStatus `json:"status"`
	LastCheck   time.Time      `json:"last_check"`
	LastLatency time.Duration  `json:"last_latency"`
	IsHTTPS     bool           `json:"is_https"`
	IsActive    bool           `json:"is_active"`
}

func NewEndpoint(rawURL string) (*Endpoint, error) {
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "http://" + rawURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	if parsed.Port() == "" {
		if parsed.Scheme == "https" {
			parsed.Host = parsed.Host + ":443"
		} else {
			parsed.Host = parsed.Host + ":80"
		}
	}

	name := parsed.Host
	if parsed.Path != "" && parsed.Path != "/" {
		name = parsed.Host + parsed.Path
	}

	return &Endpoint{
		URL:      parsed.String(),
		Name:     name,
		Status:   StatusUnknown,
		IsActive: true,
		IsHTTPS:  parsed.Scheme == "https",
	}, nil
}

func (e *Endpoint) DetermineStatus(latency time.Duration, err error) EndpointStatus {
	if err != nil {
		return StatusDown
	}
	if latency > 500*time.Millisecond {
		return StatusSlow
	}
	return StatusHealthy
}

func (e *Endpoint) Update(latency time.Duration, err error) {
	e.LastCheck = time.Now()
	e.LastLatency = latency
	e.Status = e.DetermineStatus(latency, err)
}

func (e *Endpoint) StatusIcon() string {
	switch e.Status {
	case StatusHealthy:
		return "🟢"
	case StatusSlow:
		return "🟡"
	case StatusDown:
		return "🔥"
	default:
		return "⚪"
	}
}
