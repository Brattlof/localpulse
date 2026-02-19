package monitor

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

var DefaultPorts = []int{3000, 3001, 8080, 8000, 5000, 9000, 4000, 4200, 5173, 1313}

var CommonPaths = []string{"/", "/health", "/api", "/status", "/ping", "/api/health", "/healthz", "/ready"}

type ScanResult struct {
	Port    int
	URL     string
	IsHTTPS bool
	Status  EndpointStatus
	Latency time.Duration
	Error   error
}

type Scanner struct {
	ports     []int
	host      string
	timeout   time.Duration
	client    *http.Client
	httpsOnly bool
}

type ScannerOption func(*Scanner)

func WithPorts(ports []int) ScannerOption {
	return func(s *Scanner) {
		s.ports = ports
	}
}

func WithHost(host string) ScannerOption {
	return func(s *Scanner) {
		s.host = host
	}
}

func WithTimeout(timeout time.Duration) ScannerOption {
	return func(s *Scanner) {
		s.timeout = timeout
		s.client.Timeout = timeout
	}
}

func WithHTTPSOnly(httpsOnly bool) ScannerOption {
	return func(s *Scanner) {
		s.httpsOnly = httpsOnly
	}
}

func NewScanner(opts ...ScannerOption) *Scanner {
	timeout := 2 * time.Second
	s := &Scanner{
		ports:   DefaultPorts,
		host:    "localhost",
		timeout: timeout,
		client: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *Scanner) Scan(ctx context.Context) []ScanResult {
	var wg sync.WaitGroup
	results := make(chan ScanResult, len(s.ports)*2)

	for _, port := range s.ports {
		wg.Add(2)

		go func(port int) {
			defer wg.Done()
			if s.httpsOnly {
				return
			}
			result := s.probePort(ctx, port, false)
			if result != nil {
				results <- *result
			}
		}(port)

		go func(port int) {
			defer wg.Done()
			result := s.probePort(ctx, port, true)
			if result != nil {
				results <- *result
			}
		}(port)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var allResults []ScanResult
	for result := range results {
		allResults = append(allResults, result)
	}

	return allResults
}

func (s *Scanner) probePort(ctx context.Context, port int, https bool) *ScanResult {
	scheme := "http"
	if https {
		scheme = "https"
	}

	url := fmt.Sprintf("%s://%s:%d", scheme, s.host, port)

	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil
	}

	req.Header.Set("User-Agent", "LocalPulse/1.0")

	resp, err := s.client.Do(req)
	latency := time.Since(start)

	if err != nil {
		if ctx.Err() == context.Canceled {
			return nil
		}
		if isConnectionRefused(err) {
			return nil
		}
		if isTimeout(err) {
			return nil
		}
		if isTLSError(err) && https {
			return nil
		}
		return nil
	}
	defer resp.Body.Close()

	return &ScanResult{
		Port:    port,
		URL:     url,
		IsHTTPS: https,
		Status:  statusFromResponse(resp, latency),
		Latency: latency,
	}
}

func (s *Scanner) DiscoverEndpoints(ctx context.Context) []*Endpoint {
	results := s.Scan(ctx)

	endpoints := make(map[string]*Endpoint)

	for _, result := range results {
		if result.Status == StatusDown {
			continue
		}

		ep, err := NewEndpoint(result.URL)
		if err != nil {
			continue
		}
		ep.Status = result.Status
		ep.LastLatency = result.Latency
		ep.LastCheck = time.Now()

		for _, path := range CommonPaths {
			fullURL := result.URL + path
			if _, exists := endpoints[fullURL]; !exists {
				probeEp, err := NewEndpoint(fullURL)
				if err != nil {
					continue
				}
				probeResult := s.probePath(ctx, fullURL)
				if probeResult != nil {
					probeEp.Status = probeResult.Status
					probeEp.LastLatency = probeResult.Latency
					probeEp.LastCheck = time.Now()
					endpoints[fullURL] = probeEp
				}
			}
		}

		endpoints[result.URL] = ep
	}

	var list []*Endpoint
	for _, ep := range endpoints {
		list = append(list, ep)
	}

	return list
}

func (s *Scanner) probePath(ctx context.Context, url string) *ScanResult {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil
	}

	req.Header.Set("User-Agent", "LocalPulse/1.0")

	resp, err := s.client.Do(req)
	latency := time.Since(start)

	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	return &ScanResult{
		URL:     url,
		Status:  statusFromResponse(resp, latency),
		Latency: latency,
	}
}

func statusFromResponse(resp *http.Response, latency time.Duration) EndpointStatus {
	if resp.StatusCode >= 500 {
		return StatusDown
	}
	if latency > 500*time.Millisecond {
		return StatusSlow
	}
	return StatusHealthy
}

func isConnectionRefused(err error) bool {
	return strings.Contains(err.Error(), "connection refused") ||
		strings.Contains(err.Error(), "Connection refused")
}

func isTimeout(err error) bool {
	return strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "deadline exceeded") ||
		strings.Contains(err.Error(), "context deadline exceeded")
}

func isTLSError(err error) bool {
	return strings.Contains(err.Error(), "certificate") ||
		strings.Contains(err.Error(), "TLS") ||
		strings.Contains(err.Error(), "x509")
}

func CheckPort(host string, port int, timeout time.Duration) bool {
	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func QuickScan(host string, ports []int) []int {
	var openPorts []int
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, port := range ports {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			if CheckPort(host, p, 500*time.Millisecond) {
				mu.Lock()
				openPorts = append(openPorts, p)
				mu.Unlock()
			}
		}(port)
	}

	wg.Wait()
	return openPorts
}
