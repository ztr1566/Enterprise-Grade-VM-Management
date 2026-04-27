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

	"backend/internal/agent/governance"
	"backend/internal/agent/wal"
	"backend/internal/api/grpc/telemetry"
)

// Streamer handles the metric collection loop and mTLS streaming to the backend.
type Streamer struct {
	WAL        *wal.DiskWAL
	Monitor    *governance.Monitor
	BackendURL string
	VMID       string
	DataDir    string
	client     *AgentClient
}

// Run starts the collection and streaming loop.
func (s *Streamer) Run(ctx context.Context) error {
	s.client = NewAgentClient(
		s.VMID,
		s.BackendURL,
		filepath.Join(s.DataDir, "agent.key"),
		filepath.Join(s.DataDir, "agent.crt"),
		filepath.Join(s.DataDir, "ca.crt"),
	)

	// 1. Setup dynamic mTLS identity
	caCert, err := os.ReadFile(filepath.Join(s.DataDir, "ca.crt"))
	if err != nil {
		return fmt.Errorf("failed to read CA certificate: %w", err)
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		GetClientCertificate: func(info *tls.CertificateRequestInfo) (*tls.Certificate, error) {
			cert, err := tls.LoadX509KeyPair(
				filepath.Join(s.DataDir, "agent.crt"),
				filepath.Join(s.DataDir, "agent.key"),
			)
			if err != nil {
				return nil, err
			}
			return &cert, nil
		},
		RootCAs: certPool,
	}

	// 2. Establish mTLS gRPC connection with backoff
	conn, err := DialWithBackoff(ctx, s.BackendURL, tlsConfig)
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

	renewalTicker := time.NewTicker(6 * time.Hour)
	defer renewalTicker.Stop()

	// Initial renewal check
	if err := s.client.CheckAndRenew(ctx); err != nil {
		log.Printf("Initial renewal check failed: %v", err)
	}

	sender := &Sender{VMID: s.VMID}
	fmt.Printf("Starting telemetry stream for VM: %s\n", s.VMID)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Always sample to maintain heartbeats, even if throttled.
			// Hardware sampling is lightweight.
			sample, err := CollectSample()
			if err != nil {
				log.Printf("Failed to collect sample: %v", err)
				continue
			}
			if err := s.WAL.PushMetric(sample); err != nil {
				log.Printf("Failed to push to WAL: %v", err)
			}
		case <-flushTicker.C:
			// Apply backpressure to the heavy lifting (Disk I/O, Serialization, Networking)
			if s.Monitor != nil && s.Monitor.IsThrottled() {
				// Send a minimal heartbeat to avoid being marked offline
				if err := s.sendHeartbeat(ctx, client); err != nil {
					log.Printf("Failed to send heartbeat: %v", err)
				}
				continue
			}

			// Attempt to stream WAL contents
			if err := s.flushWAL(ctx, client, sender); err != nil {
				log.Printf("Failed to flush WAL: %v", err)
			}
		case <-renewalTicker.C:
			if err := s.client.CheckAndRenew(ctx); err != nil {
				log.Printf("Periodic renewal check failed: %v", err)
			}
		}
	}
}

func (s *Streamer) sendHeartbeat(ctx context.Context, client telemetry.TelemetryIngestionClient) error {
	stream, err := client.StreamMetricsBatch(ctx)
	if err != nil {
		return err
	}

	batch := &telemetry.MetricsBatch{
		VmId:    s.VMID,
		Samples: []*telemetry.MetricSample{},
	}

	if err := stream.Send(batch); err != nil {
		return err
	}

	_, err = stream.CloseAndRecv()
	return err
}

func (s *Streamer) flushWAL(ctx context.Context, client telemetry.TelemetryIngestionClient, sender *Sender) error {
	entries, err := s.WAL.ReadAllAndClear()
	if err != nil || len(entries) == 0 {
		return err
	}

	var metrics []*telemetry.MetricSample
	var logs []*telemetry.LogEntry

	for _, res := range entries {
		if res.Metric != nil {
			metrics = append(metrics, res.Metric)
		}
		if res.Log != nil {
			logs = append(logs, res.Log)
		}
	}

	// Stream Metrics if present
	if len(metrics) > 0 {
		if err := sender.StreamMetrics(ctx, client, metrics); err != nil {
			return fmt.Errorf("failed to stream metrics: %w", err)
		}
	}

	// Stream Logs if present
	if len(logs) > 0 {
		if err := sender.StreamLogs(ctx, client, logs); err != nil {
			return fmt.Errorf("failed to stream logs: %w", err)
		}
	}

	return nil
}
