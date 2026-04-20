package telemetry

import (
	"context"
	"backend/internal/ca"
	"backend/internal/models"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type IdentityHandler struct {
	UnimplementedAgentIdentityServer
	CA *ca.CA
	DB *sql.DB
}

func NewIdentityHandler(ca *ca.CA, db *sql.DB) *IdentityHandler {
	return &IdentityHandler{
		CA: ca,
		DB: db,
	}
}

func (h *IdentityHandler) SignCSR(ctx context.Context, req *CSRRequest) (*CSRResponse, error) {
	if req.GetVmId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "vm_id is required")
	}
	if len(req.GetCsrPem()) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "csr_pem is required")
	}

	// 1. Check for mTLS Peer (Renewal)
	if p, ok := peer.FromContext(ctx); ok {
		if tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo); ok && len(tlsInfo.State.PeerCertificates) > 0 {
			clientCert := tlsInfo.State.PeerCertificates[0]
			serial := fmt.Sprintf("%x", clientCert.SerialNumber)

			// Verify cert is not revoked
			revoked, err := models.IsCertificateRevoked(h.DB, serial)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to check CRL: %v", err)
			}
			if revoked {
				return nil, status.Error(codes.Unauthenticated, "certificate is revoked")
			}

			// Verify vm_id matches cert CommonName
			if clientCert.Subject.CommonName != req.GetVmId() {
				return nil, status.Error(codes.Unauthenticated, "certificate CommonName mismatch")
			}

			// Authorized for renewal
			return h.signAndStore(req.GetVmId(), req.GetCsrPem())
		}
	}

	// 2. Extract OTT from Metadata (Initial Enrollment)
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata is required")
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return nil, status.Error(codes.Unauthenticated, "authorization header is required")
	}

	ott := strings.TrimPrefix(authHeaders[0], "Bearer ")
	if ott == authHeaders[0] {
		return nil, status.Error(codes.Unauthenticated, "invalid authorization format (expected Bearer <token>)")
	}

	// 2. Validate OTT
	hash := sha256.Sum256([]byte(ott))
	hashStr := hex.EncodeToString(hash[:])

	token, err := models.GetAgentTokenByHash(h.DB, hashStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		return nil, status.Errorf(codes.Internal, "database error: %v", err)
	}

	if token.UsedAt != nil {
		return nil, status.Error(codes.Unauthenticated, "token already used")
	}

	if time.Now().After(token.ExpiresAt) {
		return nil, status.Error(codes.Unauthenticated, "token expired")
	}

	if token.MachineID != req.GetVmId() {
		return nil, status.Error(codes.Unauthenticated, "token machine_id mismatch")
	}

	// 3. Mark token as used before signing (atomic)
	if err := models.MarkTokenUsed(h.DB, token.ID); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to mark token used: %v", err)
	}

	return h.signAndStore(req.GetVmId(), req.GetCsrPem())
}

// signAndStore handles the actual signing and persistence of the certificate record.
func (h *IdentityHandler) signAndStore(vmID string, csrPEM []byte) (*CSRResponse, error) {
	// Sign CSR (30 days TTL)
	certPEM, serial, err := h.CA.SignCSR(csrPEM, 30*24*time.Hour)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to sign CSR: %v", err)
	}

	certRecord := models.AgentCertificate{
		SerialNumber: serial,
		MachineID:    vmID,
		IssuedAt:     time.Now(),
		ExpiresAt:    time.Now().Add(30 * 24 * time.Hour),
	}

	if err := models.CreateAgentCertificate(h.DB, certRecord); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create cert record: %v", err)
	}

	return &CSRResponse{
		CertificatePem:   certPEM,
		CaCertificatePem: h.CA.CertBytes,
	}, nil
}
