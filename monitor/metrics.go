package monitor

import (
	"math"
	"sort"
	"sync"
	"time"
)

type RequestResult struct {
	Timestamp    time.Time
	Latency      time.Duration
	StatusCode   int
	Size         int64
	IsError      bool
	ErrorMessage string
}

type Metrics struct {
	mu sync.RWMutex

	TotalRequests int64
	TotalErrors   int64
	TotalBytes    int64
	TotalLatency  time.Duration

	RecentResults    []RequestResult
	maxRecentResults int

	Latencies    []time.Duration
	maxLatencies int

	StatusCodeCounts map[int]int64

	WindowStart time.Time
}

func NewMetrics(windowSize int) *Metrics {
	if windowSize <= 0 {
		windowSize = 1000
	}
	return &Metrics{
		RecentResults:    make([]RequestResult, 0, windowSize),
		Latencies:        make([]time.Duration, 0, windowSize),
		StatusCodeCounts: make(map[int]int64),
		maxRecentResults: windowSize,
		maxLatencies:     windowSize,
		WindowStart:      time.Now(),
	}
}

func (m *Metrics) Record(result RequestResult) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalRequests++
	m.TotalLatency += result.Latency

	if result.IsError {
		m.TotalErrors++
	} else {
		m.StatusCodeCounts[result.StatusCode]++
		m.TotalBytes += result.Size
	}

	m.Latencies = append(m.Latencies, result.Latency)
	if len(m.Latencies) > m.maxLatencies {
		m.Latencies = m.Latencies[1:]
	}

	m.RecentResults = append(m.RecentResults, result)
	if len(m.RecentResults) > m.maxRecentResults {
		m.RecentResults = m.RecentResults[1:]
	}
}

func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalRequests = 0
	m.TotalErrors = 0
	m.TotalBytes = 0
	m.TotalLatency = 0
	m.RecentResults = m.RecentResults[:0]
	m.Latencies = m.Latencies[:0]
	m.StatusCodeCounts = make(map[int]int64)
	m.WindowStart = time.Now()
}

type Stats struct {
	P50           time.Duration
	P95           time.Duration
	P99           time.Duration
	AvgLatency    time.Duration
	MinLatency    time.Duration
	MaxLatency    time.Duration
	Throughput    float64
	ErrorRate     float64
	AvgSize       int64
	TotalRequests int64
	TotalErrors   int64
	StatusCode2xx int64
	StatusCode4xx int64
	StatusCode5xx int64
}

func (m *Metrics) GetStats() Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := Stats{
		TotalRequests: m.TotalRequests,
		TotalErrors:   m.TotalErrors,
	}

	if m.TotalRequests > 0 {
		stats.ErrorRate = float64(m.TotalErrors) / float64(m.TotalRequests) * 100
		stats.AvgLatency = time.Duration(int64(m.TotalLatency) / m.TotalRequests)
		stats.AvgSize = m.TotalBytes / m.TotalRequests
	}

	windowDuration := time.Since(m.WindowStart).Seconds()
	if windowDuration > 0 {
		stats.Throughput = float64(m.TotalRequests) / windowDuration
	}

	for code, count := range m.StatusCodeCounts {
		switch {
		case code >= 200 && code < 300:
			stats.StatusCode2xx += count
		case code >= 400 && code < 500:
			stats.StatusCode4xx += count
		case code >= 500:
			stats.StatusCode5xx += count
		}
	}

	if len(m.Latencies) > 0 {
		sorted := make([]time.Duration, len(m.Latencies))
		copy(sorted, m.Latencies)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i] < sorted[j]
		})

		stats.MinLatency = sorted[0]
		stats.MaxLatency = sorted[len(sorted)-1]
		stats.P50 = percentile(sorted, 50)
		stats.P95 = percentile(sorted, 95)
		stats.P99 = percentile(sorted, 99)
	}

	return stats
}

func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	if len(sorted) == 1 {
		return sorted[0]
	}

	idx := int(math.Ceil(float64(len(sorted)-1)*p/100)) - 1
	if idx < 0 {
		idx = 0
	}
	return sorted[idx]
}

func (m *Metrics) GetRecentLatencies(count int) []time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if count > len(m.Latencies) {
		count = len(m.Latencies)
	}
	if count == 0 {
		return nil
	}

	start := len(m.Latencies) - count
	result := make([]time.Duration, count)
	copy(result, m.Latencies[start:])
	return result
}

func (m *Metrics) GetRecentResults(count int) []RequestResult {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if count > len(m.RecentResults) {
		count = len(m.RecentResults)
	}
	if count == 0 {
		return nil
	}

	start := len(m.RecentResults) - count
	result := make([]RequestResult, count)
	copy(result, m.RecentResults[start:])
	return result
}
