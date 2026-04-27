package integration

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"net"
	"testing"
	"time"

	"backend/internal/agent/bootstrap"
	"backend/internal/api/grpc/interceptors"
	"backend/internal/api/grpc/telemetry"
	"backend/internal/ca"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func TestCSRExpirationsAndRevocation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// 1. Setup Backend CA and gRPC Server with Interceptors
	dataDir := t.TempDir()
	caInst, err := ca.LoadOrCreateCA(dataDir)
	if err != nil {
		t.Fatalf("failed to setup CA: %v", err)
	}

	serverCertPEM, serverKeyPEM, err := caInst.GenerateServerCertificate(dataDir, "localhost")
	if err != nil {
		t.Fatalf("failed to generate server cert: %v", err)
	}
	serverCert, _ := tls.X509KeyPair(serverCertPEM, serverKeyPEM)

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caInst.CertBytes)

	serverTLS := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.VerifyClientCertIfGiven,
		ClientCAs:    certPool,
	}

	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer(
		grpc.Creds(credentials.NewTLS(serverTLS)),
		grpc.UnaryInterceptor(interceptors.CRLInterceptor(db)),
		grpc.StreamInterceptor(interceptors.CRLStreamInterceptor(db)),
	)
	telemetry.RegisterAgentIdentityServer(s, telemetry.NewIdentityHandler(caInst, db))
	telemetry.RegisterTelemetryIngestionServer(s, telemetry.NewTelemetryHandler(db))
	
	go s.Serve(lis)
	defer s.Stop()

	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	// Scenario 1: Expired OTT
	t.Run("ExpiredOTT", func(t *testing.T) {
		vmID := "expired-vm"
		token := "expired-token"
		hash := sha256.Sum256([]byte(token))
		hashStr := hex.EncodeToString(hash[:])
		
		// Insert expired token (created 25 hours ago, expired 1 hour ago)
		expiredAt := time.Now().Add(-1 * time.Hour)
		_, err := db.Exec("INSERT INTO agent_tokens (id, machine_id, token_hash, created_at, expires_at) VALUES (?, ?, ?, ?, ?)",
			"token-1", vmID, hashStr, expiredAt.Add(-24*time.Hour), expiredAt)
		if err != nil {
			t.Fatalf("failed to insert expired token: %v", err)
		}

		key, _ := bootstrap.GenerateKey()
		csrPEM, _ := bootstrap.GenerateCSR(key, vmID)

		conn, _ := grpc.DialContext(context.Background(), "bufnet",
			grpc.WithContextDialer(bufDialer),
			grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})))
		defer conn.Close()

		ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", "Bearer "+token)
		idClient := telemetry.NewAgentIdentityClient(conn)
		_, err = idClient.SignCSR(ctx, &telemetry.CSRRequest{
			VmId:   vmID,
			CsrPem: csrPEM,
		})

		if err == nil {
			t.Error("Expected SignCSR to fail for expired token, but it succeeded")
		} else {
			st, ok := status.FromError(err)
			if !ok || st.Code() != codes.Unauthenticated {
				t.Errorf("Expected Unauthenticated, got %v", st.Code())
			}
		}
	})

	// Scenario 2: Revoked Certificate
	t.Run("RevokedCert", func(t *testing.T) {
		vmID := "revoked-vm"
		token := "valid-token"
		hash := sha256.Sum256([]byte(token))
		hashStr := hex.EncodeToString(hash[:])

		// Insert valid token
		_, err := db.Exec("INSERT INTO agent_tokens (id, machine_id, token_hash, created_at, expires_at) VALUES (?, ?, ?, ?, ?)",
			"token-2", vmID, hashStr, time.Now(), time.Now().Add(1*time.Hour))
		if err != nil {
			t.Fatalf("failed to insert token: %v", err)
		}
		
		key, _ := bootstrap.GenerateKey()
		csrPEM, _ := bootstrap.GenerateCSR(key, vmID)

		conn, _ := grpc.DialContext(context.Background(), "bufnet",
			grpc.WithContextDialer(bufDialer),
			grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})))
		
		ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", "Bearer "+token)
		idClient := telemetry.NewAgentIdentityClient(conn)
		resp, err := idClient.SignCSR(ctx, &telemetry.CSRRequest{
			VmId:   vmID,
			CsrPem: csrPEM,
		})
		if err != nil {
			t.Fatalf("SignCSR failed: %v", err)
		}
		conn.Close()

		// Revoke the certificate
		block, _ := pem.Decode(resp.CertificatePem)
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			t.Fatalf("failed to parse cert: %v", err)
		}
		serialNum := fmt.Sprintf("%x", cert.SerialNumber)

		_, err = db.Exec("UPDATE agent_certificates SET revoked_at = CURRENT_TIMESTAMP WHERE serial_number = ?", serialNum)
		if err != nil {
			t.Fatalf("failed to revoke cert: %v", err)
		}

		// Try to stream telemetry with revoked cert
		keyBytes, _ := x509.MarshalECPrivateKey(key)
		keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})
		clientCert, _ := tls.X509KeyPair(resp.CertificatePem, keyPEM)

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{clientCert},
			RootCAs:      certPool,
			ServerName:   "localhost",
		}

		conn, _ = grpc.DialContext(context.Background(), "bufnet",
			grpc.WithContextDialer(bufDialer),
			grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
		defer conn.Close()

		ingestClient := telemetry.NewTelemetryIngestionClient(conn)
		stream, err := ingestClient.StreamMetricsBatch(context.Background())
		if err != nil {
			// Some gRPC versions fail on Dial, others on first stream action
			st, _ := status.FromError(err)
			if st.Code() != codes.Unauthenticated {
				t.Errorf("Expected Unauthenticated on dial, got %v", err)
			}
			return
		}

		err = stream.Send(&telemetry.MetricsBatch{VmId: vmID})
		if err == nil {
			_, err = stream.CloseAndRecv()
		}

		if err == nil {
			t.Error("Expected Stream to fail for revoked certificate, but it succeeded")
		} else {
			st, ok := status.FromError(err)
			if !ok || st.Code() != codes.Unauthenticated {
				t.Errorf("Expected Unauthenticated (revoked), got %v: %v", st.Code(), err)
			}
		}
	})
}
