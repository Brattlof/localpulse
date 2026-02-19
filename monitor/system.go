package monitor

import (
	"runtime"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type SystemMetrics struct {
	CPUPercent float64
	RAMUsed    uint64
	RAMTotal   uint64
	RAMPercent float64
}

type SystemMonitor struct {
	previousCPU float64
}

func NewSystemMonitor() *SystemMonitor {
	return &SystemMonitor{}
}

func (sm *SystemMonitor) GetMetrics() SystemMetrics {
	metrics := SystemMetrics{}

	cpuPercents, err := cpu.Percent(0, false)
	if err == nil && len(cpuPercents) > 0 {
		metrics.CPUPercent = cpuPercents[0]
	} else {
		metrics.CPUPercent = sm.estimateCPU()
	}

	vmStat, err := mem.VirtualMemory()
	if err == nil {
		metrics.RAMUsed = vmStat.Used
		metrics.RAMTotal = vmStat.Total
		metrics.RAMPercent = vmStat.UsedPercent
	} else {
		metrics.RAMTotal = sm.getSystemRAM()
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		metrics.RAMUsed = m.Sys
		if metrics.RAMTotal > 0 {
			metrics.RAMPercent = float64(metrics.RAMUsed) / float64(metrics.RAMTotal) * 100
		}
	}

	sm.previousCPU = metrics.CPUPercent
	return metrics
}

func (sm *SystemMonitor) estimateCPU() float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return sm.previousCPU
}

func (sm *SystemMonitor) getSystemRAM() uint64 {
	vmStat, err := mem.VirtualMemory()
	if err == nil {
		return vmStat.Total
	}
	return 8 * 1024 * 1024 * 1024
}

func GetCPUCount() int {
	return runtime.NumCPU()
}

func GetGoVersion() string {
	return runtime.Version()
}
