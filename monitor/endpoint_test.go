package monitor

import (
	"testing"
	"time"
)

func TestNewEndpoint(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantURL   string
		wantName  string
		wantHTTPS bool
		wantErr   bool
	}{
		{
			name:      "http URL with port",
			input:     "http://localhost:3000/api",
			wantURL:   "http://localhost:3000/api",
			wantName:  "localhost:3000/api",
			wantHTTPS: false,
			wantErr:   false,
		},
		{
			name:      "https URL",
			input:     "https://localhost:8443/health",
			wantURL:   "https://localhost:8443/health",
			wantName:  "localhost:8443/health",
			wantHTTPS: true,
			wantErr:   false,
		},
		{
			name:      "URL without scheme",
			input:     "localhost:8080",
			wantURL:   "http://localhost:8080",
			wantName:  "localhost:8080",
			wantHTTPS: false,
			wantErr:   false,
		},
		{
			name:      "URL with root path only",
			input:     "http://localhost:3000/",
			wantURL:   "http://localhost:3000/",
			wantName:  "localhost:3000",
			wantHTTPS: false,
			wantErr:   false,
		},
		{
			name:      "empty URL defaults to port 80",
			input:     "http://example.com",
			wantURL:   "http://example.com:80",
			wantName:  "example.com:80",
			wantHTTPS: false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ep, err := NewEndpoint(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewEndpoint(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if ep.URL != tt.wantURL {
				t.Errorf("NewEndpoint(%q).URL = %q, want %q", tt.input, ep.URL, tt.wantURL)
			}
			if ep.Name != tt.wantName {
				t.Errorf("NewEndpoint(%q).Name = %q, want %q", tt.input, ep.Name, tt.wantName)
			}
			if ep.IsHTTPS != tt.wantHTTPS {
				t.Errorf("NewEndpoint(%q).IsHTTPS = %v, want %v", tt.input, ep.IsHTTPS, tt.wantHTTPS)
			}
		})
	}
}

func TestEndpoint_DetermineStatus(t *testing.T) {
	ep := &Endpoint{}

	tests := []struct {
		name    string
		latency time.Duration
		err     error
		want    EndpointStatus
	}{
		{
			name:    "fast response is healthy",
			latency: 50 * time.Millisecond,
			err:     nil,
			want:    StatusHealthy,
		},
		{
			name:    "slow response over 500ms",
			latency: 600 * time.Millisecond,
			err:     nil,
			want:    StatusSlow,
		},
		{
			name:    "exactly at threshold",
			latency: 500 * time.Millisecond,
			err:     nil,
			want:    StatusHealthy,
		},
		{
			name:    "just over threshold",
			latency: 501 * time.Millisecond,
			err:     nil,
			want:    StatusSlow,
		},
		{
			name:    "error means down",
			latency: 0,
			err:     assertAnError,
			want:    StatusDown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ep.DetermineStatus(tt.latency, tt.err)
			if got != tt.want {
				t.Errorf("DetermineStatus(%v, %v) = %v, want %v", tt.latency, tt.err, got, tt.want)
			}
		})
	}
}

func TestEndpoint_StatusIcon(t *testing.T) {
	tests := []struct {
		status EndpointStatus
		want   string
	}{
		{StatusHealthy, "🟢"},
		{StatusSlow, "🟡"},
		{StatusDown, "🔥"},
		{StatusUnknown, "⚪"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			ep := &Endpoint{Status: tt.status}
			if got := ep.StatusIcon(); got != tt.want {
				t.Errorf("StatusIcon() = %q, want %q", got, tt.want)
			}
		})
	}
}

var assertAnError = error(&testError{})

type testError struct{}

func (e *testError) Error() string { return "test error" }
