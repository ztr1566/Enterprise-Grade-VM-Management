package telemetry

import (
	"database/sql"
	"io"
	"log"

	"backend/internal/db"
)

// TelemetryHandler implements the TelemetryIngestion gRPC service.
type TelemetryHandler struct {
	UnimplementedTelemetryIngestionServer
	DB *sql.DB
}

// NewTelemetryHandler creates a new instance of TelemetryHandler.
func NewTelemetryHandler(dbInst *sql.DB) *TelemetryHandler {
	return &TelemetryHandler{
		DB: dbInst,
	}
}

// StreamMetricsBatch handles client-side streaming of metric batches.
func (h *TelemetryHandler) StreamMetricsBatch(stream TelemetryIngestion_StreamMetricsBatchServer) error {
	var totalProcessed int32
	for {
		batch, err := stream.Recv()
		if err == io.EOF {
			log.Printf("Metrics stream closed, total processed: %d", totalProcessed)
			return stream.SendAndClose(&IngestionAck{
				ProcessedCount: totalProcessed,
			})
		}
		if err != nil {
			log.Printf("Metrics stream error: %v", err)
			return err
		}

		// Map gRPC samples to DB samples to break circular dependency
		dbSamples := make([]*db.MetricSample, len(batch.Samples))
		for i, s := range batch.Samples {
			dbSamples[i] = &db.MetricSample{
				Timestamp:         s.Timestamp,
				CpuUsagePercent:   s.CpuUsagePercent,
				MemoryUsedBytes:   s.MemoryUsedBytes,
				MemoryTotalBytes:  s.MemoryTotalBytes,
				DiskUsagePercent:  s.DiskUsagePercent,
				NetworkTxBytes:    s.NetworkTxBytes,
				NetworkRxBytes:    s.NetworkRxBytes,
			}
		}

		// Insert metrics into the database
		if err := db.InsertMetrics(h.DB, batch.VmId, dbSamples); err != nil {
			log.Printf("Failed to insert metrics for VM %s: %v", batch.VmId, err)
			return err
		}

		totalProcessed += int32(len(batch.Samples))
	}
}
