package wal

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"backend/internal/api/grpc/telemetry"
)

const DefaultWALFile = "telemetry.wal"

// WALEntry is a container for different types of telemetry data in the WAL.
type WALEntry struct {
	Metric *telemetry.MetricSample `json:"metric,omitempty"`
	Log    *telemetry.LogEntry    `json:"log,omitempty"`
}

// DiskWAL implements a simple disk-backed Write-Ahead Log for buffering telemetry data.
type DiskWAL struct {
	mu      sync.Mutex
	path    string
	maxSize int64
}

// NewDiskWAL initializes a new WAL in the specified directory.
func NewDiskWAL(dataDir string, maxSize int64) (*DiskWAL, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	return &DiskWAL{
		path:    filepath.Join(dataDir, DefaultWALFile),
		maxSize: maxSize,
	}, nil
}

func (w *DiskWAL) push(entry *WALEntry) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Check size constraints
	fi, err := os.Stat(w.path)
	if err == nil && fi.Size() >= w.maxSize {
		log.Printf("WARNING: WAL size limit reached (%d bytes), evicting all buffered entries", fi.Size())
		if err := os.Truncate(w.path, 0); err != nil {
			return fmt.Errorf("failed to evict WAL: %w", err)
		}
	}

	f, err := os.OpenFile(w.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	// NOTE: encoding/json is used here because the WALEntry wrapper is a plain Go struct.
	// Inner proto message fields are serialized via their json struct tags from protoc-gen-go.
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	_, err = f.Write(append(data, '\n'))
	return err
}

// PushMetric appends a new metric sample to the WAL.
func (w *DiskWAL) PushMetric(sample *telemetry.MetricSample) error {
	return w.push(&WALEntry{Metric: sample})
}

// PushLog appends a new log entry to the WAL.
func (w *DiskWAL) PushLog(entry *telemetry.LogEntry) error {
	return w.push(&WALEntry{Log: entry})
}

// ReadAll returns all entries currently in the WAL.
func (w *DiskWAL) ReadAll() ([]*WALEntry, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	f, err := os.Open(w.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var entries []*WALEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var e WALEntry
		// We can use encoding/json here since WALEntry is a plain Go struct
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			log.Printf("WARNING: skipping corrupt WAL entry: %v", err)
			continue
		}
		entries = append(entries, &e)
	}

	return entries, nil
}

// Clear removes all entries from the WAL.
func (w *DiskWAL) Clear() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return os.Truncate(w.path, 0)
}
