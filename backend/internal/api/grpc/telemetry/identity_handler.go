package telemetry

import (
	"context"
	"backend/internal/ca"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IdentityHandler struct {
	UnimplementedAgentIdentityServer
	CA *ca.CA
}

func NewIdentityHandler(ca *ca.CA) *IdentityHandler {
	return &IdentityHandler{
		CA: ca,
	}
}

func (h *IdentityHandler) SignCSR(ctx context.Context, req *CSRRequest) (*CSRResponse, error) {
	if req.GetVmId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "vm_id is required")
	}
	if len(req.GetCsrPem()) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "csr_pem is required")
	}

	certPEM, err := h.CA.SignCSR(req.CsrPem)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to sign CSR: %v", err)
	}

	return &CSRResponse{
		CertificatePem:   certPEM,
		CaCertificatePem: h.CA.CertBytes,
	}, nil
}
