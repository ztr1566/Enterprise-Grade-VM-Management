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

		// Update VM status to online
		res, err := h.DB.Exec(`UPDATE vms SET status='online', last_seen_at=CURRENT_TIMESTAMP WHERE id=?`, batch.VmId)
		if err != nil {
			log.Printf("Failed to update status for VM %s: %v", batch.VmId, err)
		} else {
			rows, _ := res.RowsAffected()
			if rows == 0 {
				log.Printf("WARNING: Log batch received for UNKNOWN VM ID: %s. Ensure agent VM_ID matches dashboard ID.", batch.VmId)
			} else {
				log.Printf("Heartbeat (log) received: VM %s is now ONLINE", batch.VmId)
			}
		}

		totalProcessed += int32(len(batch.Entries))
	}
}
