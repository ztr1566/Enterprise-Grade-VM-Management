package wal

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync"
)

const (
	FsyncBatchCount = 10
)

// WALWriter handles thread-safe binary writes to the WAL file with fsync batching and size limits.
type WALWriter struct {
	mu         *sync.Mutex
	path       string
	file       *os.File
	maxSize    int64
	writeCount int
}

// NewWALWriter creates a new writer for the specified path with a given max size.
func NewWALWriter(path string, maxSize int64, mu *sync.Mutex) (*WALWriter, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	if mu == nil {
		mu = &sync.Mutex{}
	}
	return &WALWriter{
		path:    path,
		file:    f,
		maxSize: maxSize,
		mu:      mu,
	}, nil
}

// Write appends an entry to the WAL, enforcing size limits and batching fsyncs.
func (w *WALWriter) Write(entry *WALEntry) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 1. Check size limit before writing
	fi, err := w.file.Stat()
	if err == nil && fi.Size() >= w.maxSize {
		// Eviction policy: Truncate and start over
		if err := w.file.Truncate(0); err != nil {
			return fmt.Errorf("failed to evict WAL: %w", err)
		}
		if _, err := w.file.Seek(0, io.SeekStart); err != nil {
			return fmt.Errorf("failed to seek after eviction: %w", err)
		}
	}

	// 2. Prepare binary buffer
	// [4 bytes] Length
	// [4 bytes] Checksum
	// [1 byte ] Type
	// [N bytes] Payload
	payloadLen := uint32(len(entry.Payload))
	buf := make([]byte, 9+payloadLen)
	binary.BigEndian.PutUint32(buf[0:4], payloadLen)
	binary.BigEndian.PutUint32(buf[4:8], entry.CalculateChecksum())
	buf[8] = entry.Type
	copy(buf[9:], entry.Payload)

	// 3. Write to file
	if _, err := w.file.Write(buf); err != nil {
		return fmt.Errorf("wal write failed: %w", err)
	}

	// 4. Batch fsync
	w.writeCount++
	if w.writeCount >= FsyncBatchCount {
		if err := w.file.Sync(); err != nil {
			return fmt.Errorf("wal fsync failed: %w", err)
		}
		w.writeCount = 0
	}

	return nil
}

// Close closes the underlying WAL file.
func (w *WALWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// Clear truncates the WAL file.
func (w *WALWriter) Clear() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.clearUnlocked()
}

// clearUnlocked truncates and resets the WAL file without locking.
// Caller must hold w.mu.
func (w *WALWriter) clearUnlocked() error {
	if err := w.file.Truncate(0); err != nil {
		return err
	}
	_, err := w.file.Seek(0, io.SeekStart)
	return err
}
