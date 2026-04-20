package telemetry

import (
	"io"
	"log"

	"backend/internal/db"
)

// StreamLogsBatch handles client-side streaming of log batches.
func (h *TelemetryHandler) StreamLogsBatch(stream TelemetryIngestion_StreamLogsBatchServer) error {
	var totalProcessed int32
	for {
		batch, err := stream.Recv()
		if err == io.EOF {
			log.Printf("Logs stream closed, total processed: %d", totalProcessed)
			return stream.SendAndClose(&IngestionAck{
				ProcessedCount: totalProcessed,
			})
		}
		if err != nil {
			log.Printf("Logs stream error: %v", err)
			return err
		}

		// Map gRPC logs to DB logs
		dbLogs := make([]*db.LogEntry, len(batch.Entries))
		for i, e := range batch.Entries {
			dbLogs[i] = &db.LogEntry{
				Timestamp: e.Timestamp,
				Severity:  e.Severity,
				Component: e.Component,
				Message:   e.Message,
			}
		}

		if err := db.InsertLogs(h.DB, batch.VmId, dbLogs); err != nil {
			log.Printf("Failed to insert logs for VM %s: %v", batch.VmId, err)
			return err
		}

		totalProcessed += int32(len(batch.Entries))
	}
}
