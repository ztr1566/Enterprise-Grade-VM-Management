package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/internal/agent/bootstrap"
	"backend/internal/agent/governance"
	agentTelemetry "backend/internal/agent/telemetry"
	"backend/internal/agent/wal"
	"backend/internal/api/grpc/telemetry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

func main() {
	fmt.Println("vm-agent starting...")

	dataDir := "data"
	backendURL := os.Getenv("BACKEND_URL")
	if backendURL == "" {
		backendURL = "localhost:50051"
	}
	vmID := os.Getenv("VM_ID")
	if vmID == "" {
		vmID = "local-vm-001" // Default for testing
	}

	if !bootstrap.HasIdentity(dataDir) {
		fmt.Println("No identity found. Starting bootstrap flow...")
		ott := os.Getenv("AGENT_OTT")
		if ott == "" {
			log.Fatal("Bootstrap required but AGENT_OTT environment variable is not set.")
		}
		if err := runBootstrap(dataDir, backendURL, vmID, ott); err != nil {
			log.Fatalf("Bootstrap failed: %v", err)
		}
		fmt.Println("Bootstrap successful.")
	} else {
		fmt.Println("Identity found. Ready to stream.")
	}

	// Phase 5: Initialize resource governance
	monitor := governance.NewMonitor()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go monitor.Start(ctx)
	agentTelemetry.StartHealthServer("localhost:8081", monitor)
	fmt.Println("Resource monitor and health server (localhost:8081) started.")

	// Phase 4: Start telemetry streamer
	walInst, err := wal.NewDiskWAL(dataDir, 50*1024*1024) // 50MB
	if err != nil {
		log.Fatalf("Failed to initialize WAL: %v", err)
	}

	streamer := &agentTelemetry.Streamer{
		WAL:        walInst,
		Monitor:    monitor,
		BackendURL: backendURL,
		VMID:       vmID,
		DataDir:    dataDir,
	}
	fmt.Println("Starting telemetry streamer...")
	go func() {
		if err := streamer.Run(ctx); err != nil && err != context.Canceled {
			log.Fatalf("Streamer failed: %v", err)
		}
	}()

	// Phase 5: Start log collector
	logCollector := &agentTelemetry.LogCollector{
		WAL:     walInst,
		Monitor: monitor,
	}
	logPath := os.Getenv("LOG_PATH")
	if logPath == "" {
		logPath = "agent.log"
	}
	// Ensure the log file exists without overwriting existing content
	if f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		f.Close()
	}
	fmt.Printf("Starting log collector on %s...\n", logPath)
	if err := logCollector.Run(ctx, logPath); err != nil && err != context.Canceled {
		log.Fatalf("Log collector failed: %v", err)
	}
}

func runBootstrap(dataDir, backendURL, vmID, ott string) error {
	// 1. Generate private key locally
	key, err := bootstrap.GenerateKey()
	if err != nil {
		return fmt.Errorf("failed to generate key: %w", err)
	}

	// 2. Generate CSR
	csrPEM, err := bootstrap.GenerateCSR(key, vmID)
	if err != nil {
		return fmt.Errorf("failed to generate CSR: %w", err)
	}

	// 3. Connect to backend gRPC server
	// We use InsecureSkipVerify: true for the initial bootstrap because the agent 
	// does not yet have the CA certificate to verify the server's identity.
	log.Println("WARNING: Using InsecureSkipVerify for initial bootstrap (no CA cert yet)")
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	
	conn, err := grpc.NewClient(backendURL, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	if err != nil {
		return fmt.Errorf("failed to connect to backend: %w", err)
	}
	defer conn.Close()

	client := telemetry.NewAgentIdentityClient(conn)
	
	// Add OTT to metadata
	ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", "Bearer "+ott)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 4. Request certificate signing
	resp, err := client.SignCSR(ctx, &telemetry.CSRRequest{
		VmId:   vmID,
		CsrPem: csrPEM,
	})
	if err != nil {
		return fmt.Errorf("CSR signing request failed: %w", err)
	}

	// 5. Securely store the signed certificate and the private key
	if err := bootstrap.SaveIdentity(dataDir, key, resp.CertificatePem, resp.CaCertificatePem); err != nil {
		return fmt.Errorf("failed to save identity: %w", err)
	}

	return nil
}
