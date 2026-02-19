package monitor

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewLoadTester(t *testing.T) {
	ep, _ := NewEndpoint("http://localhost:8080")
	metrics := NewMetrics(100)

	lt := NewLoadTester(ep, metrics)

	if lt == nil {
		t.Fatal("NewLoadTester returned nil")
	}
	if lt.concurrency != 10 {
		t.Errorf("default concurrency = %d, want 10", lt.concurrency)
	}
	if lt.maxConcur != 100 {
		t.Errorf("default maxConcur = %d, want 100", lt.maxConcur)
	}
}

func TestLoadTester_WithConcurrency(t *testing.T) {
	ep, _ := NewEndpoint("http://localhost:8080")
	metrics := NewMetrics(100)

	lt := NewLoadTester(ep, metrics, WithConcurrency(50))

	if lt.concurrency != 50 {
		t.Errorf("concurrency = %d, want 50", lt.concurrency)
	}
}

func TestLoadTester_StartStop(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ep, _ := NewEndpoint(srv.URL)
	metrics := NewMetrics(100)

	lt := NewLoadTester(ep, metrics, WithConcurrency(2))

	if err := lt.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	if !lt.IsRunning() {
		t.Error("IsRunning() = false after Start()")
	}

	err := lt.Start()
	if err == nil {
		t.Error("second Start() should return error")
	}

	lt.Stop()

	if lt.IsRunning() {
		t.Error("IsRunning() = true after Stop()")
	}
}

func TestLoadTester_SendRequest(t *testing.T) {
	var requestCount atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer srv.Close()

	ep, _ := NewEndpoint(srv.URL)
	metrics := NewMetrics(1000)

	lt := NewLoadTester(ep, metrics, WithConcurrency(2))

	if err := lt.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	lt.SendBurst(10)

	time.Sleep(100 * time.Millisecond)

	lt.Stop()

	if requestCount.Load() == 0 {
		t.Error("no requests were made")
	}

	stats := metrics.GetStats()
	if stats.TotalRequests == 0 {
		t.Error("no requests recorded in metrics")
	}
}

func TestLoadTester_SetConcurrency(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ep, _ := NewEndpoint(srv.URL)
	metrics := NewMetrics(100)

	lt := NewLoadTester(ep, metrics, WithMaxConcurrency(50))

	lt.SetConcurrency(30)
	if lt.GetConcurrency() != 30 {
		t.Errorf("concurrency = %d, want 30", lt.GetConcurrency())
	}

	lt.SetConcurrency(0)
	if lt.GetConcurrency() != 1 {
		t.Errorf("concurrency = %d, want 1 (minimum)", lt.GetConcurrency())
	}

	lt.SetConcurrency(100)
	if lt.GetConcurrency() != 50 {
		t.Errorf("concurrency = %d, want 50 (max)", lt.GetConcurrency())
	}
}

func TestLoadTester_IncreaseDecrease(t *testing.T) {
	ep, _ := NewEndpoint("http://localhost:8080")
	metrics := NewMetrics(100)

	lt := NewLoadTester(ep, metrics, WithConcurrency(10))

	lt.IncreaseConcurrency()
	if lt.GetConcurrency() != 11 {
		t.Errorf("after Increase, concurrency = %d, want 11", lt.GetConcurrency())
	}

	lt.DecreaseConcurrency()
	if lt.GetConcurrency() != 10 {
		t.Errorf("after Decrease, concurrency = %d, want 10", lt.GetConcurrency())
	}
}

func TestLoadGenerator_AddRemoveTester(t *testing.T) {
	lg := NewLoadGenerator()

	ep1, _ := NewEndpoint("http://localhost:8080")
	ep2, _ := NewEndpoint("http://localhost:3000")
	metrics1 := NewMetrics(100)
	metrics2 := NewMetrics(100)

	lg.AddTester(ep1, metrics1)
	lg.AddTester(ep2, metrics2)

	if len(lg.testers) != 2 {
		t.Errorf("testers count = %d, want 2", len(lg.testers))
	}

	lg.AddTester(ep1, metrics1)
	if len(lg.testers) != 2 {
		t.Errorf("adding duplicate should not increase count, got %d", len(lg.testers))
	}

	lg.RemoveTester("http://localhost:8080")
	if len(lg.testers) != 1 {
		t.Errorf("testers count = %d, want 1 after removal", len(lg.testers))
	}
}

func TestLoadGenerator_StartStop(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	lg := NewLoadGenerator()

	ep, _ := NewEndpoint(srv.URL)
	metrics := NewMetrics(1000)

	lg.AddTester(ep, metrics)

	if err := lg.Start(10); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	if !lg.IsRunning() {
		t.Error("IsRunning() = false after Start()")
	}

	err := lg.Start(10)
	if err == nil {
		t.Error("second Start() should return error")
	}

	time.Sleep(200 * time.Millisecond)

	lg.Stop()

	if lg.IsRunning() {
		t.Error("IsRunning() = true after Stop()")
	}
}

func TestLoadGenerator_SetRPS(t *testing.T) {
	lg := NewLoadGenerator()

	lg.SetRPS(50)
	if lg.GetRPS() != 50 {
		t.Errorf("RPS = %d, want 50", lg.GetRPS())
	}

	lg.SetRPS(0)
	if lg.GetRPS() != 1 {
		t.Errorf("RPS = %d, want 1 (minimum)", lg.GetRPS())
	}
}

func TestLoadGenerator_IncreaseDecreaseRPS(t *testing.T) {
	lg := NewLoadGenerator()
	lg.SetRPS(20)

	lg.IncreaseRPS()
	if lg.GetRPS() != 25 {
		t.Errorf("after Increase, RPS = %d, want 25", lg.GetRPS())
	}

	lg.DecreaseRPS()
	if lg.GetRPS() != 20 {
		t.Errorf("after Decrease, RPS = %d, want 20", lg.GetRPS())
	}
}

func TestLoadTester_ErrorHandling(t *testing.T) {
	ep, _ := NewEndpoint("http://localhost:59999")
	metrics := NewMetrics(100)

	lt := NewLoadTester(ep, metrics, WithConcurrency(1), WithClientTimeout(100*time.Millisecond))

	if err := lt.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	lt.SendBurst(1)

	time.Sleep(200 * time.Millisecond)

	lt.Stop()

	stats := metrics.GetStats()
	if stats.TotalRequests == 0 {
		t.Error("request should have been recorded even on error")
	}
	if stats.TotalErrors == 0 {
		t.Error("error should have been recorded")
	}
}
