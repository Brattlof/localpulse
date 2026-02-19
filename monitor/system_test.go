package monitor

import (
	"runtime"
	"testing"
)

func TestSystemMonitor_GetMetrics(t *testing.T) {
	sm := NewSystemMonitor()
	metrics := sm.GetMetrics()

	if metrics.RAMTotal == 0 {
		t.Error("RAMTotal should not be zero")
	}

	if metrics.CPUPercent < 0 || metrics.CPUPercent > 100 {
		t.Errorf("CPUPercent = %v, should be between 0 and 100", metrics.CPUPercent)
	}
}

func TestGetCPUCount(t *testing.T) {
	count := GetCPUCount()
	if count < 1 {
		t.Errorf("GetCPUCount() = %d, should be at least 1", count)
	}
	if count != runtime.NumCPU() {
		t.Errorf("GetCPUCount() = %d, runtime.NumCPU() = %d", count, runtime.NumCPU())
	}
}

func TestGetGoVersion(t *testing.T) {
	version := GetGoVersion()
	if version == "" {
		t.Error("GetGoVersion() returned empty string")
	}
}
