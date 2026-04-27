package wal

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
	"os"
)

// WALReader handles reading and validating the binary WAL file.
type WALReader struct {
	path string
}

// NewWALReader creates a new reader for the specified path.
func NewWALReader(path string) *WALReader {
	return &WALReader{path: path}
}

// ReadAllAndRecover reads all valid entries from the WAL.
// If corruption is detected (checksum mismatch or partial write), it truncates the file
// to the last known good position to recover the log.
func (r *WALReader) ReadAllAndRecover() ([]*WALEntry, error) {
	f, err := os.OpenFile(r.path, os.O_RDWR, 0600)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var entries []*WALEntry
	var lastGoodOffset int64

	for {
		// 1. Read Header (9 bytes: 4 len, 4 checksum, 1 type)
		header := make([]byte, 9)
		_, err := io.ReadFull(f, header)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			// io.ErrUnexpectedEOF handles partial reads before EOF
			if errors.Is(err, io.ErrUnexpectedEOF) {
				log.Printf("WAL: detected partial header at offset %d, truncating...", lastGoodOffset)
				return entries, f.Truncate(lastGoodOffset)
			}
			// Other fatal errors
			return nil, err
		}

		length := binary.BigEndian.Uint32(header[0:4])
		expectedChecksum := binary.BigEndian.Uint32(header[4:8])
		entryType := header[8]

		// 2. Read Payload
		payload := make([]byte, length)
		_, err = io.ReadFull(f, payload)
		if err != nil {
			// Partial payload or unexpected EOF: Truncate
			log.Printf("WAL: detected partial payload at offset %d, truncating...", lastGoodOffset)
			return entries, f.Truncate(lastGoodOffset)
		}

		// 3. Verify Checksum
		entry := &WALEntry{Type: entryType, Payload: payload}
		if entry.CalculateChecksum() != expectedChecksum {
			// Checksum mismatch: Truncate
			log.Printf("WAL: checksum mismatch at offset %d, truncating...", lastGoodOffset)
			return entries, f.Truncate(lastGoodOffset)
		}

		entries = append(entries, entry)
		lastGoodOffset, _ = f.Seek(0, io.SeekCurrent)
	}

	return entries, nil
}
