package bootstrap

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"time"
)

// SaveIdentity persists the agent's private key and signed certificate to disk with secure permissions.
func SaveIdentity(dataDir string, key *ecdsa.PrivateKey, certPEM, caCertPEM []byte) error {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}

	keyBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})

	// Save private key with strict 0600 permissions
	if err := os.WriteFile(filepath.Join(dataDir, "agent.key"), keyPEM, 0600); err != nil {
		return err
	}
	// Save certificate
	if err := os.WriteFile(filepath.Join(dataDir, "agent.crt"), certPEM, 0644); err != nil {
		return err
	}
	// Save CA certificate for verification
	if err := os.WriteFile(filepath.Join(dataDir, "ca.crt"), caCertPEM, 0644); err != nil {
		return err
	}

	return nil
}

// HasIdentity checks if the agent already has a valid, non-expired certificate.
func HasIdentity(dataDir string) bool {
	certPath := filepath.Join(dataDir, "agent.crt")
	keyPath := filepath.Join(dataDir, "agent.key")

	// 1. Check if files exist
	if _, err := os.Stat(certPath); err != nil {
		return false
	}
	if _, err := os.Stat(keyPath); err != nil {
		return false
	}

	// 2. Read and parse certificate
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return false
	}

	block, _ := pem.Decode(certPEM)
	if block == nil || block.Type != "CERTIFICATE" {
		return false
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false
	}

	// 3. Check expiry
	if time.Now().After(cert.NotAfter) {
		return false
	}

	return true
}
