package wal

import (
	"hash/crc32"
)

const (
	EntryTypeMetric uint8 = 0x01
	EntryTypeLog    uint8 = 0x02
)

// WALEntry represents a single binary entry in the WAL file.
// Format on disk:
// [4 bytes] Length (uint32)
// [4 bytes] Checksum (uint32, CRC32 of payload)
// [1 byte ] Type (uint8, 0x01=Metric, 0x02=Log)
// [N bytes] Payload (serialized protobuf)
type WALEntry struct {
	Type    uint8
	Payload []byte
}

// CalculateChecksum returns the CRC32 checksum of the payload.
func (e *WALEntry) CalculateChecksum() uint32 {
	return crc32.ChecksumIEEE(e.Payload)
}
