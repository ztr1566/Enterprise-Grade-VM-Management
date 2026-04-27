package telemetry

import (
	"context"

	"backend/internal/api/grpc/telemetry"
	"google.golang.org/protobuf/proto"
)

// SafePayloadSize is set to 3.5MB to leave room for gRPC headers and VMID string.
const SafePayloadSize = 3584 * 1024 

// Sender handles splitting large batches into smaller ones to respect gRPC limits.
type Sender struct {
	VMID string
}

func (s *Sender) StreamMetrics(ctx context.Context, client telemetry.TelemetryIngestionClient, samples []*telemetry.MetricSample) error {
	stream, err := client.StreamMetricsBatch(ctx)
	if err != nil {
		return err
	}

	var currentBatch []*telemetry.MetricSample
	currentSize := 0

	for _, sample := range samples {
		sampleSize := proto.Size(sample)
		
		// If adding this sample exceeds the safe size, send the current batch
		if currentSize+sampleSize > SafePayloadSize && len(currentBatch) > 0 {
			batch := &telemetry.MetricsBatch{
				VmId:    s.VMID,
				Samples: currentBatch,
			}
			if err := stream.Send(batch); err != nil {
				return err
			}
			currentBatch = nil
			currentSize = 0
		}

		currentBatch = append(currentBatch, sample)
		currentSize += sampleSize
	}

	// Send remaining
	if len(currentBatch) > 0 {
		batch := &telemetry.MetricsBatch{
			VmId:    s.VMID,
			Samples: currentBatch,
		}
		if err := stream.Send(batch); err != nil {
			return err
		}
	}

	_, err = stream.CloseAndRecv()
	return err
}

func (s *Sender) StreamLogs(ctx context.Context, client telemetry.TelemetryIngestionClient, logs []*telemetry.LogEntry) error {
	stream, err := client.StreamLogsBatch(ctx)
	if err != nil {
		return err
	}

	var currentBatch []*telemetry.LogEntry
	currentSize := 0

	for _, entry := range logs {
		entrySize := proto.Size(entry)
		
		// If adding this entry exceeds the safe size, send the current batch
		if currentSize+entrySize > SafePayloadSize && len(currentBatch) > 0 {
			batch := &telemetry.LogsBatch{
				VmId:    s.VMID,
				Entries: currentBatch,
			}
			if err := stream.Send(batch); err != nil {
				return err
			}
			currentBatch = nil
			currentSize = 0
		}

		currentBatch = append(currentBatch, entry)
		currentSize += entrySize
	}

	// Send remaining
	if len(currentBatch) > 0 {
		batch := &telemetry.LogsBatch{
			VmId:    s.VMID,
			Entries: currentBatch,
		}
		if err := stream.Send(batch); err != nil {
			return err
		}
	}

	_, err = stream.CloseAndRecv()
	return err
}
