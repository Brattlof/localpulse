package monitor

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type LoadTester struct {
	mu sync.RWMutex

	client      *http.Client
	endpoint    *Endpoint
	metrics     *Metrics
	concurrency int
	maxConcur   int
	timeout     time.Duration

	running    atomic.Bool
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	reqChan    chan struct{}
	resultChan chan RequestResult

	requestsSent atomic.Int64
}

type LoadTesterOption func(*LoadTester)

func WithConcurrency(concurrency int) LoadTesterOption {
	return func(lt *LoadTester) {
		if concurrency > 0 && concurrency <= lt.maxConcur {
			lt.concurrency = concurrency
		}
	}
}

func WithMaxConcurrency(max int) LoadTesterOption {
	return func(lt *LoadTester) {
		if max > 0 {
			lt.maxConcur = max
		}
	}
}

func WithClientTimeout(timeout time.Duration) LoadTesterOption {
	return func(lt *LoadTester) {
		lt.timeout = timeout
		lt.client.Timeout = timeout
	}
}

func NewLoadTester(endpoint *Endpoint, metrics *Metrics, opts ...LoadTesterOption) *LoadTester {
	timeout := 10 * time.Second
	lt := &LoadTester{
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		endpoint:    endpoint,
		metrics:     metrics,
		concurrency: 10,
		maxConcur:   100,
		timeout:     timeout,
		reqChan:     make(chan struct{}, 1000),
		resultChan:  make(chan RequestResult, 1000),
	}
	for _, opt := range opts {
		opt(lt)
	}
	return lt
}

func (lt *LoadTester) Start() error {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	if lt.running.Load() {
		return fmt.Errorf("load tester already running")
	}

	lt.ctx, lt.cancel = context.WithCancel(context.Background())
	lt.running.Store(true)
	lt.requestsSent.Store(0)

	for i := 0; i < lt.concurrency; i++ {
		lt.wg.Add(1)
		go lt.worker(i)
	}

	lt.wg.Add(1)
	go lt.resultCollector()

	return nil
}

func (lt *LoadTester) Stop() {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	if !lt.running.Load() {
		return
	}

	lt.running.Store(false)
	if lt.cancel != nil {
		lt.cancel()
	}
	lt.wg.Wait()
}

func (lt *LoadTester) worker(id int) {
	defer lt.wg.Done()

	for {
		select {
		case <-lt.ctx.Done():
			return
		case _, ok := <-lt.reqChan:
			if !ok {
				return
			}
			result := lt.makeRequest()
			select {
			case lt.resultChan <- result:
			case <-lt.ctx.Done():
				return
			}
		}
	}
}

func (lt *LoadTester) makeRequest() RequestResult {
	result := RequestResult{
		Timestamp: time.Now(),
	}

	if lt.endpoint == nil {
		result.IsError = true
		result.ErrorMessage = "no endpoint configured"
		return result
	}

	start := time.Now()

	req, err := http.NewRequestWithContext(lt.ctx, http.MethodGet, lt.endpoint.URL, nil)
	if err != nil {
		result.IsError = true
		result.ErrorMessage = err.Error()
		result.Latency = time.Since(start)
		return result
	}

	req.Header.Set("User-Agent", "LocalPulse/1.0")
	req.Header.Set("Accept", "*/*")

	resp, err := lt.client.Do(req)
	result.Latency = time.Since(start)

	if err != nil {
		result.IsError = true
		result.ErrorMessage = err.Error()
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	result.Size, _ = io.Copy(io.Discard, resp.Body)

	return result
}

func (lt *LoadTester) resultCollector() {
	defer lt.wg.Done()

	for {
		select {
		case <-lt.ctx.Done():
			return
		case result, ok := <-lt.resultChan:
			if !ok {
				return
			}
			if lt.metrics != nil {
				lt.metrics.Record(result)
			}
			lt.endpoint.Update(result.Latency, func() error {
				if result.IsError {
					return fmt.Errorf("request error: %s", result.ErrorMessage)
				}
				return nil
			}())
		}
	}
}

func (lt *LoadTester) SendRequest() {
	if !lt.running.Load() {
		return
	}
	lt.requestsSent.Add(1)
	select {
	case lt.reqChan <- struct{}{}:
	default:
	}
}

func (lt *LoadTester) SendBurst(count int) {
	for i := 0; i < count; i++ {
		lt.SendRequest()
	}
}

func (lt *LoadTester) SetConcurrency(n int) {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	if n < 1 {
		n = 1
	}
	if n > lt.maxConcur {
		n = lt.maxConcur
	}

	current := lt.concurrency
	lt.concurrency = n

	if lt.running.Load() {
		if n > current {
			for i := current; i < n; i++ {
				lt.wg.Add(1)
				go lt.worker(i)
			}
		}
	}
}

func (lt *LoadTester) IncreaseConcurrency() {
	lt.SetConcurrency(lt.concurrency + 1)
}

func (lt *LoadTester) DecreaseConcurrency() {
	lt.SetConcurrency(lt.concurrency - 1)
}

func (lt *LoadTester) GetConcurrency() int {
	lt.mu.RLock()
	defer lt.mu.RUnlock()
	return lt.concurrency
}

func (lt *LoadTester) IsRunning() bool {
	return lt.running.Load()
}

func (lt *LoadTester) RequestsSent() int64 {
	return lt.requestsSent.Load()
}

type LoadGenerator struct {
	mu sync.RWMutex

	testers map[string]*LoadTester
	rps     int
	ticker  *time.Ticker
	running atomic.Bool
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

func NewLoadGenerator() *LoadGenerator {
	return &LoadGenerator{
		testers: make(map[string]*LoadTester),
		rps:     10,
	}
}

func (lg *LoadGenerator) AddTester(endpoint *Endpoint, metrics *Metrics) {
	lg.mu.Lock()
	defer lg.mu.Unlock()

	if _, exists := lg.testers[endpoint.URL]; !exists {
		lg.testers[endpoint.URL] = NewLoadTester(endpoint, metrics)
	}
}

func (lg *LoadGenerator) RemoveTester(url string) {
	lg.mu.Lock()
	defer lg.mu.Unlock()

	if tester, exists := lg.testers[url]; exists {
		tester.Stop()
		delete(lg.testers, url)
	}
}

func (lg *LoadGenerator) Start(rps int) error {
	lg.mu.Lock()
	defer lg.mu.Unlock()

	if lg.running.Load() {
		return fmt.Errorf("load generator already running")
	}

	if rps < 1 {
		rps = 1
	}
	lg.rps = rps
	lg.ctx, lg.cancel = context.WithCancel(context.Background())
	lg.running.Store(true)

	for _, tester := range lg.testers {
		if err := tester.Start(); err == nil {
			lg.wg.Add(1)
			go lg.runGenerator(tester)
		}
	}

	return nil
}

func (lg *LoadGenerator) runGenerator(tester *LoadTester) {
	defer lg.wg.Done()

	interval := time.Second / time.Duration(lg.rps)
	lg.ticker = time.NewTicker(interval)
	defer lg.ticker.Stop()

	for {
		select {
		case <-lg.ctx.Done():
			return
		case <-lg.ticker.C:
			tester.SendRequest()
		}
	}
}

func (lg *LoadGenerator) Stop() {
	lg.mu.Lock()
	defer lg.mu.Unlock()

	if !lg.running.Load() {
		return
	}

	lg.running.Store(false)
	if lg.cancel != nil {
		lg.cancel()
	}
	lg.wg.Wait()

	for _, tester := range lg.testers {
		tester.Stop()
	}
}

func (lg *LoadGenerator) SetRPS(rps int) {
	lg.mu.Lock()
	defer lg.mu.Unlock()

	if rps < 1 {
		rps = 1
	}
	lg.rps = rps

	if lg.ticker != nil {
		interval := time.Second / time.Duration(rps)
		lg.ticker.Reset(interval)
	}
}

func (lg *LoadGenerator) IncreaseRPS() {
	lg.SetRPS(lg.rps + 5)
}

func (lg *LoadGenerator) DecreaseRPS() {
	lg.SetRPS(lg.rps - 5)
	if lg.rps < 1 {
		lg.SetRPS(1)
	}
}

func (lg *LoadGenerator) GetRPS() int {
	lg.mu.RLock()
	defer lg.mu.RUnlock()
	return lg.rps
}

func (lg *LoadGenerator) IsRunning() bool {
	return lg.running.Load()
}
