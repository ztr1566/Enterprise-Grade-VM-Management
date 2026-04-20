package telemetry

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"backend/internal/crypto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	pb "backend/internal/api/grpc/telemetry"
)

type AgentClient struct {
	VMID       string
	ServerAddr string
	KeyPath    string
	CertPath   string
	CACertPath string
}

func NewAgentClient(vmID, serverAddr, keyPath, certPath, caCertPath string) *AgentClient {
	return &AgentClient{
		VMID:       vmID,
		ServerAddr: serverAddr,
		KeyPath:    keyPath,
		CertPath:   certPath,
		CACertPath: caCertPath,
	}
}

// Enroll handles the initial CSR bootstrap using an OTT.
func (c *AgentClient) Enroll(ctx context.Context, ott string) error {
	// 1. Generate or Load ECDSA Key
	key, err := crypto.LoadECDSAKey(c.KeyPath)
	if err != nil {
		key, err = crypto.GenerateECDSAKey(c.KeyPath)
		if err != nil {
			return fmt.Errorf("failed to manage key: %w", err)
		}
	}

	// 2. Create CSR
	template := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   c.VMID,
			Organization: []string{"Enterprise VM Management"},
		},
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, key)
	if err != nil {
		return fmt.Errorf("failed to create CSR: %w", err)
	}

	csrPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})

	// 3. Connect to Server (Unauthenticated for CSR)
	conn, err := grpc.Dial(c.ServerAddr, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})))
	if err != nil {
		return fmt.Errorf("failed to connect for enrollment: %w", err)
	}
	defer conn.Close()

	client := pb.NewAgentIdentityClient(conn)

	// 4. Send CSR with OTT
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+ott)
	resp, err := client.SignCSR(ctx, &pb.CSRRequest{
		VmId:   c.VMID,
		CsrPem: csrPEM,
	})
	if err != nil {
		return fmt.Errorf("CSR signing failed: %w", err)
	}

	// 5. Save Certificate and CA Certificate
	if err := os.WriteFile(c.CertPath, resp.CertificatePem, 0644); err != nil {
		return fmt.Errorf("failed to save certificate: %w", err)
	}
	if err := os.WriteFile(c.CACertPath, resp.CaCertificatePem, 0644); err != nil {
		return fmt.Errorf("failed to save CA certificate: %w", err)
	}

	return nil
}

// GetClientCredentials loads mTLS credentials for telemetry streaming.
func (c *AgentClient) GetClientCredentials() (credentials.TransportCredentials, error) {
	cert, err := tls.LoadX509KeyPair(c.CertPath, c.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load mTLS key pair: %w", err)
	}

	caCert, err := os.ReadFile(c.CACertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caCert)

	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
		ServerName:   "localhost", // Match server certificate
	}), nil
}

// CheckAndRenew verifies certificate validity and triggers renewal if within 24 hours of expiration.
func (c *AgentClient) CheckAndRenew(ctx context.Context) error {
	certPEM, err := os.ReadFile(c.CertPath)
	if err != nil {
		return fmt.Errorf("failed to read certificate for renewal check: %w", err)
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return fmt.Errorf("failed to decode certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	// If certificate expires in less than 24 hours, renew it
	if time.Until(cert.NotAfter) < 24*time.Hour {
		return c.Renew(ctx)
	}

	return nil
}

// Renew performs a certificate renewal using existing mTLS identity.
func (c *AgentClient) Renew(ctx context.Context) error {
	// 1. Generate new CSR with existing key
	key, err := crypto.LoadECDSAKey(c.KeyPath)
	if err != nil {
		return fmt.Errorf("failed to load key for renewal: %w", err)
	}

	template := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   c.VMID,
			Organization: []string{"Enterprise VM Management"},
		},
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, key)
	if err != nil {
		return fmt.Errorf("failed to create CSR for renewal: %w", err)
	}

	csrPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})

	// 2. Connect to Server using current mTLS
	creds, err := c.GetClientCredentials()
	if err != nil {
		return fmt.Errorf("failed to get mTLS credentials for renewal: %w", err)
	}

	conn, err := grpc.Dial(c.ServerAddr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return fmt.Errorf("failed to connect for renewal: %w", err)
	}
	defer conn.Close()

	client := pb.NewAgentIdentityClient(conn)

	// 3. Send CSR
	resp, err := client.SignCSR(ctx, &pb.CSRRequest{
		VmId:   c.VMID,
		CsrPem: csrPEM,
	})
	if err != nil {
		return fmt.Errorf("renewal CSR signing failed: %w", err)
	}

	// 4. Update Certificate
	if err := os.WriteFile(c.CertPath, resp.CertificatePem, 0644); err != nil {
		return fmt.Errorf("failed to update certificate: %w", err)
	}

	return nil
}
