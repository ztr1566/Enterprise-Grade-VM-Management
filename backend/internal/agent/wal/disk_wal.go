package wal

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"backend/internal/api/grpc/telemetry"
	"google.golang.org/protobuf/proto"
)

const DefaultWALFile = "telemetry.wal"

// DiskWAL implements a hardened disk-backed Write-Ahead Log for buffering telemetry data.
// It wraps binary WALWriter and WALReader to provide high-level metric/log access.
type DiskWAL struct {
	mu     *sync.Mutex
	path   string
	writer *WALWriter
	reader *WALReader
}

// NewDiskWAL initializes a new WAL in the specified directory.
func NewDiskWAL(dataDir string, maxSize int64) (*DiskWAL, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	path := filepath.Join(dataDir, DefaultWALFile)
	
	// Create common mutex for both DiskWAL operations and binary WALWriter
	mu := &sync.Mutex{}
	
	// Spec: STARTUP → VERIFY_CHECKSUMS → TRUNCATE_CORRUPTED → READY
	reader := NewWALReader(path)
	recovered, err := reader.ReadAllAndRecover()
	if err != nil {
		log.Printf("WAL: recovery failed during startup: %v", err)
	} else if len(recovered) > 0 {
		log.Printf("WAL: successfully recovered %d entries during startup", len(recovered))
	}

	writer, err := NewWALWriter(path, maxSize, mu)
	if err != nil {
		return nil, err
	}

	return &DiskWAL{
		mu:     mu,
		path:   path,
		writer: writer,
		reader: reader,
	}, nil
}

// PushMetric appends a new metric sample to the WAL.
func (w *DiskWAL) PushMetric(sample *telemetry.MetricSample) error {
	data, err := proto.Marshal(sample)
	if err != nil {
		return fmt.Errorf("failed to marshal metric: %w", err)
	}
	// Note: w.writer.Write already uses w.mu (since it was passed to NewWALWriter)
	return w.writer.Write(&WALEntry{Type: EntryTypeMetric, Payload: data})
}

// PushLog appends a new log entry to the WAL.
func (w *DiskWAL) PushLog(sample *telemetry.LogEntry) error {
	data, err := proto.Marshal(sample)
	if err != nil {
		return fmt.Errorf("failed to marshal log: %w", err)
	}
	// Note: w.writer.Write already uses w.mu
	return w.writer.Write(&WALEntry{Type: EntryTypeLog, Payload: data})
}

// ReadAllResult is a high-level container for entries read from the WAL.
type ReadAllResult struct {
	Metric *telemetry.MetricSample
	Log    *telemetry.LogEntry
}

// ReadAll returns all valid entries currently in the WAL, recovering from corruption if needed.
func (w *DiskWAL) ReadAll() ([]*ReadAllResult, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	entries, err := w.reader.ReadAllAndRecover()
	if err != nil {
		return nil, err
	}

	var results []*ReadAllResult
	for _, e := range entries {
		res := &ReadAllResult{}
		switch e.Type {
		case EntryTypeMetric:
			m := &telemetry.MetricSample{}
			if err := proto.Unmarshal(e.Payload, m); err != nil {
				log.Printf("WAL: failed to unmarshal metric entry: %v", err)
				continue
			}
			res.Metric = m
		case EntryTypeLog:
			l := &telemetry.LogEntry{}
			if err := proto.Unmarshal(e.Payload, l); err != nil {
				log.Printf("WAL: failed to unmarshal log entry: %v", err)
				continue
			}
			res.Log = l
		default:
			log.Printf("WAL: unknown entry type 0x%02x", e.Type)
			continue
		}
		results = append(results, res)
	}

	return results, nil
}

// ReadAllAndClear atomically reads all entries and clears the WAL under a single lock.
// This prevents data loss from writes that sneak in between a separate ReadAll + Clear.
func (w *DiskWAL) ReadAllAndClear() ([]*ReadAllResult, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	entries, err := w.reader.ReadAllAndRecover()
	if err != nil {
		return nil, err
	}

	var results []*ReadAllResult
	for _, e := range entries {
		res := &ReadAllResult{}
		switch e.Type {
		case EntryTypeMetric:
			m := &telemetry.MetricSample{}
			if err := proto.Unmarshal(e.Payload, m); err != nil {
				log.Printf("WAL: failed to unmarshal metric entry: %v", err)
				continue
			}
			res.Metric = m
		case EntryTypeLog:
			l := &telemetry.LogEntry{}
			if err := proto.Unmarshal(e.Payload, l); err != nil {
				log.Printf("WAL: failed to unmarshal log entry: %v", err)
				continue
			}
			res.Log = l
		default:
			log.Printf("WAL: unknown entry type 0x%02x", e.Type)
			continue
		}
		results = append(results, res)
	}

	// Clear while still holding the lock — no writes can sneak in
	if err := w.writer.clearUnlocked(); err != nil {
		return results, fmt.Errorf("WAL clear after read failed: %w", err)
	}

	return results, nil
}

// Clear removes all entries from the WAL.
func (w *DiskWAL) Clear() error {
	// writer.Clear already uses w.mu
	return w.writer.Clear()
}

// Close closes the underlying WAL file.
func (w *DiskWAL) Close() error {
	// writer.Close already uses w.mu
	return w.writer.Close()
}
