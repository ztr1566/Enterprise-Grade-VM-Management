package telemetry

import (
	"time"

	"backend/internal/api/grpc/telemetry"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// CollectSample gathers hardware metrics from the system and returns a gRPC MetricSample.
func CollectSample() (*telemetry.MetricSample, error) {
	sample := &telemetry.MetricSample{
		Timestamp: time.Now().UnixMilli(),
	}

	// 1. CPU Usage
	cpuPercents, err := cpu.Percent(0, false)
	if err == nil && len(cpuPercents) > 0 {
		sample.CpuUsagePercent = cpuPercents[0]
	}

	// 2. Memory Usage
	vMem, err := mem.VirtualMemory()
	if err == nil {
		sample.MemoryUsedBytes = vMem.Used
		sample.MemoryTotalBytes = vMem.Total
	}

	// 3. Disk Usage
	usage, err := disk.Usage("/")
	if err == nil {
		sample.DiskUsagePercent = usage.UsedPercent
	}

	// 4. Network Usage (Cumulative)
	netIO, err := net.IOCounters(false)
	if err == nil && len(netIO) > 0 {
		sample.NetworkTxBytes = netIO[0].BytesSent
		sample.NetworkRxBytes = netIO[0].BytesRecv
	}

	return sample, nil
}
