package monitor

import (
	"testing"
	"time"
)

func TestMetrics_Record(t *testing.T) {
	m := NewMetrics(100)

	for i := 0; i < 10; i++ {
		m.Record(RequestResult{
			Timestamp:  time.Now(),
			Latency:    time.Duration(i+1) * 10 * time.Millisecond,
			StatusCode: 200,
			Size:       1024,
			IsError:    false,
		})
	}

	m.Record(RequestResult{
		Timestamp:  time.Now(),
		Latency:    100 * time.Millisecond,
		IsError:    true,
		StatusCode: 0,
	})

	stats := m.GetStats()

	if stats.TotalRequests != 11 {
		t.Errorf("TotalRequests = %d, want 11", stats.TotalRequests)
	}
	if stats.TotalErrors != 1 {
		t.Errorf("TotalErrors = %d, want 1", stats.TotalErrors)
	}
	if stats.StatusCode2xx != 10 {
		t.Errorf("StatusCode2xx = %d, want 10", stats.StatusCode2xx)
	}
}

func TestMetrics_Percentiles(t *testing.T) {
	m := NewMetrics(1000)

	latencies := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		40 * time.Millisecond,
		50 * time.Millisecond,
		60 * time.Millisecond,
		70 * time.Millisecond,
		80 * time.Millisecond,
		90 * time.Millisecond,
		100 * time.Millisecond,
	}

	for _, lat := range latencies {
		m.Record(RequestResult{
			Timestamp:  time.Now(),
			Latency:    lat,
			StatusCode: 200,
			Size:       100,
			IsError:    false,
		})
	}

	stats := m.GetStats()

	if stats.P50 < 40*time.Millisecond || stats.P50 > 60*time.Millisecond {
		t.Errorf("P50 = %v, expected around 50ms", stats.P50)
	}
	if stats.P95 < 80*time.Millisecond {
		t.Errorf("P95 = %v, expected >= 80ms", stats.P95)
	}
	if stats.P99 < 90*time.Millisecond {
		t.Errorf("P99 = %v, expected >= 90ms", stats.P99)
	}
}

func TestMetrics_ErrorRate(t *testing.T) {
	tests := []struct {
		name     string
		total    int
		errors   int
		wantRate float64
	}{
		{"no errors", 100, 0, 0.0},
		{"half errors", 100, 50, 50.0},
		{"all errors", 100, 100, 100.0},
		{"single request error", 1, 1, 100.0},
		{"single request success", 1, 0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMetrics(1000)

			for i := 0; i < tt.total; i++ {
				isError := i < tt.errors
				m.Record(RequestResult{
					Timestamp: time.Now(),
					Latency:   10 * time.Millisecond,
					StatusCode: func() int {
						if isError {
							return 0
						}
						return 200
					}(),
					Size:    100,
					IsError: isError,
				})
			}

			stats := m.GetStats()
			if stats.ErrorRate != tt.wantRate {
				t.Errorf("ErrorRate = %v, want %v", stats.ErrorRate, tt.wantRate)
			}
		})
	}
}

func TestMetrics_Reset(t *testing.T) {
	m := NewMetrics(100)

	for i := 0; i < 10; i++ {
		m.Record(RequestResult{
			Timestamp:  time.Now(),
			Latency:    10 * time.Millisecond,
			StatusCode: 200,
			Size:       100,
			IsError:    false,
		})
	}

	m.Reset()
	stats := m.GetStats()

	if stats.TotalRequests != 0 {
		t.Errorf("After Reset, TotalRequests = %d, want 0", stats.TotalRequests)
	}
	if stats.TotalErrors != 0 {
		t.Errorf("After Reset, TotalErrors = %d, want 0", stats.TotalErrors)
	}
}

func TestMetrics_GetRecentLatencies(t *testing.T) {
	m := NewMetrics(100)

	for i := 0; i < 20; i++ {
		m.Record(RequestResult{
			Timestamp:  time.Now(),
			Latency:    time.Duration(i+1) * time.Millisecond,
			StatusCode: 200,
			Size:       100,
			IsError:    false,
		})
	}

	recent := m.GetRecentLatencies(5)
	if len(recent) != 5 {
		t.Errorf("GetRecentLatencies(5) returned %d items, want 5", len(recent))
		return
	}

	for i, lat := range recent {
		expected := time.Duration(16+i) * time.Millisecond
		if lat != expected {
			t.Errorf("recent[%d] = %v, want %v", i, lat, expected)
		}
	}
}

func TestMetrics_WindowOverflow(t *testing.T) {
	m := NewMetrics(10)

	for i := 0; i < 20; i++ {
		m.Record(RequestResult{
			Timestamp:  time.Now(),
			Latency:    time.Duration(i+1) * time.Millisecond,
			StatusCode: 200,
			Size:       100,
			IsError:    false,
		})
	}

	if len(m.Latencies) > 10 {
		t.Errorf("Latencies length = %d, should be capped at 10", len(m.Latencies))
	}
}
