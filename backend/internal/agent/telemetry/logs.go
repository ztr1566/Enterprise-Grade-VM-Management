package telemetry

import (
	"bufio"
	"context"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"backend/internal/agent/wal"
	"backend/internal/api/grpc/telemetry"
)

// LogCollector monitors a log file and pushes new entries to the WAL.
type LogCollector struct {
	WAL *wal.DiskWAL
}

// Run starts tailing the specified log file.
func (c *LogCollector) Run(ctx context.Context, logPath string) error {
	f, err := os.Open(logPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Start from the end of the file
	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		return err
	}

	reader := bufio.NewReader(f)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					if err == io.EOF {
						break
					}
					log.Printf("Error reading log: %v", err)
					break
				}

				// Basic severity detection from log line content
				severity := "INFO"
				upperLine := strings.ToUpper(line)
				switch {
				case strings.Contains(upperLine, "ERROR") || strings.Contains(upperLine, "FATAL"):
					severity = "ERROR"
				case strings.Contains(upperLine, "WARN"):
					severity = "WARN"
				case strings.Contains(upperLine, "DEBUG"):
					severity = "DEBUG"
				}

				entry := &telemetry.LogEntry{
					Timestamp: time.Now().UnixMilli(),
					Severity:  severity,
					Component: "system",
					Message:   strings.TrimRight(line, "\n"),
				}

				if err := c.WAL.PushLog(entry); err != nil {
					log.Printf("Failed to push log to WAL: %v", err)
				}
			}
		}
	}
}
