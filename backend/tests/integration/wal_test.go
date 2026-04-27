package integration

import (
	"os"
	"path/filepath"
	"testing"

	"backend/internal/agent/wal"
	"backend/internal/api/grpc/telemetry"
)

func TestWALCorruptionRecovery(t *testing.T) {
	dataDir := t.TempDir()
	walPath := filepath.Join(dataDir, "telemetry.wal")
	
	// 1. Setup WAL and push valid data
	w, err := wal.NewDiskWAL(dataDir, 10*1024*1024)
	if err != nil {
		t.Fatalf("failed to create WAL: %v", err)
	}

	sample := &telemetry.MetricSample{
		Timestamp:        100,
		CpuUsagePercent:  10.0,
		MemoryUsedBytes:  1024,
		MemoryTotalBytes: 2048,
	}

	if err := w.PushMetric(sample); err != nil {
		t.Fatalf("PushMetric failed: %v", err)
	}
	
	// Close so we can modify the file
	w.Close()

	// 2. Simulate Corruption
	// Manually append garbage to the file
	f, err := os.OpenFile(walPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("failed to open wal for corruption: %v", err)
	}
	// Append junk that doesn't follow the 9-byte header + CRC pattern
	f.Write([]byte("GARBAGE_BYTES_THAT_WILL_FAIL_CRC_AND_HEADER_VALIDATION"))
	f.Close()

	// 3. Re-open WAL and try to read
	// The constructor calls ReadAllAndRecover which should detect the garbage and truncate
	w2, err := wal.NewDiskWAL(dataDir, 10*1024*1024)
	if err != nil {
		t.Fatalf("failed to re-open WAL: %v", err)
	}
	defer w2.Close()

	entries, err := w2.ReadAllAndClear()
	if err != nil {
		t.Fatalf("ReadAllAndClear failed after corruption: %v", err)
	}

	// Should have recovered exactly 1 valid entry (the first one)
	if len(entries) != 1 {
		t.Errorf("Expected 1 recovered entry, got %d", len(entries))
	}
	if entries[0].Metric.CpuUsagePercent != 10.0 {
		t.Errorf("Expected recovered metric CPU 10.0, got %f", entries[0].Metric.CpuUsagePercent)
	}

	// 4. Verify file is now clean/empty after ReadAllAndClear
	info, _ := os.Stat(walPath)
	if info.Size() != 0 {
		t.Errorf("Expected WAL file to be empty after ReadAllAndClear, got size %d", info.Size())
	}
}
