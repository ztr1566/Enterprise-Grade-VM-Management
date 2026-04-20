package telemetry

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"backend/internal/agent/wal"
	"backend/internal/api/grpc/telemetry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Streamer handles the metric collection loop and mTLS streaming to the backend.
type Streamer struct {
	WAL        *wal.DiskWAL
	BackendURL string
	VMID       string
	DataDir    string
}

// Run starts the collection and streaming loop.
func (s *Streamer) Run(ctx context.Context) error {
	// 1. Load mTLS identity
	cert, err := tls.LoadX509KeyPair(
		filepath.Join(s.DataDir, "agent.crt"),
		filepath.Join(s.DataDir, "agent.key"),
	)
	if err != nil {
		return fmt.Errorf("failed to load agent key pair: %w", err)
	}

	caCert, err := os.ReadFile(filepath.Join(s.DataDir, "ca.crt"))
	if err != nil {
		return fmt.Errorf("failed to read CA certificate: %w", err)
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
	}

	// 2. Establish mTLS gRPC connection
	conn, err := grpc.NewClient(s.BackendURL, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	if err != nil {
		return fmt.Errorf("failed to connect to backend: %w", err)
	}
	defer conn.Close()

	client := telemetry.NewTelemetryIngestionClient(conn)

	// 3. Main loop
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	flushTicker := time.NewTicker(5 * time.Second)
	defer flushTicker.Stop()

	fmt.Printf("Starting telemetry stream for VM: %s\n", s.VMID)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Sample and Push to WAL
			sample, err := CollectSample()
			if err != nil {
				log.Printf("Failed to collect sample: %v", err)
				continue
			}
			if err := s.WAL.PushMetric(sample); err != nil {
				log.Printf("Failed to push to WAL: %v", err)
			}
		case <-flushTicker.C:
			// Attempt to stream WAL contents
			if err := s.flushWAL(ctx, client); err != nil {
				log.Printf("Failed to flush WAL: %v", err)
			}
		}
	}
}

func (s *Streamer) flushWAL(ctx context.Context, client telemetry.TelemetryIngestionClient) error {
	entries, err := s.WAL.ReadAll()
	if err != nil || len(entries) == 0 {
		return err
	}

	var metrics []*telemetry.MetricSample
	var logs []*telemetry.LogEntry

	for _, e := range entries {
		if e.Metric != nil {
			metrics = append(metrics, e.Metric)
		}
		if e.Log != nil {
			logs = append(logs, e.Log)
		}
	}

	// Stream Metrics if present
	if len(metrics) > 0 {
		if err := s.streamMetrics(ctx, client, metrics); err != nil {
			return fmt.Errorf("failed to stream metrics: %w", err)
		}
	}

	// Stream Logs if present
	if len(logs) > 0 {
		if err := s.streamLogs(ctx, client, logs); err != nil {
			return fmt.Errorf("failed to stream logs: %w", err)
		}
	}

	// Only clear WAL if all present streams succeeded
	return s.WAL.Clear()
}

func (s *Streamer) streamMetrics(ctx context.Context, client telemetry.TelemetryIngestionClient, samples []*telemetry.MetricSample) error {
	stream, err := client.StreamMetricsBatch(ctx)
	if err != nil {
		return err
	}

	for i := 0; i < len(samples); i += 100 {
		end := i + 100
		if end > len(samples) {
			end = len(samples)
		}

		batch := &telemetry.MetricsBatch{
			VmId:    s.VMID,
			Samples: samples[i:end],
		}

		if err := stream.Send(batch); err != nil {
			return err
		}
	}

	_, err = stream.CloseAndRecv()
	return err
}

func (s *Streamer) streamLogs(ctx context.Context, client telemetry.TelemetryIngestionClient, logs []*telemetry.LogEntry) error {
	stream, err := client.StreamLogsBatch(ctx)
	if err != nil {
		return err
	}

	for i := 0; i < len(logs); i += 100 {
		end := i + 100
		if end > len(logs) {
			end = len(logs)
		}

		batch := &telemetry.LogsBatch{
			VmId:    s.VMID,
			Entries: logs[i:end],
		}

		if err := stream.Send(batch); err != nil {
			return err
		}
	}

	_, err = stream.CloseAndRecv()
	return err
}
