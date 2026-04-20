package interceptors

import (
	"context"
	"database/sql"
	"fmt"

	"backend/internal/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// CRLInterceptor returns a UnaryServerInterceptor that validates client certificates against the CRL.
func CRLInterceptor(db *sql.DB) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip CRL check for EnrollAgent (CSR bootstrap)
		if info.FullMethod == "/telemetry.AgentIdentity/SignCSR" {
			return handler(ctx, req)
		}

		p, ok := peer.FromContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "no peer found")
		}

		tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "no TLS info found")
		}

		if len(tlsInfo.State.PeerCertificates) == 0 {
			return nil, status.Error(codes.Unauthenticated, "no client certificate provided")
		}

		clientCert := tlsInfo.State.PeerCertificates[0]
		serial := fmt.Sprintf("%x", clientCert.SerialNumber)

		revoked, err := models.IsCertificateRevoked(db, serial)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to check CRL: %v", err)
		}

		if revoked {
			return nil, status.Error(codes.Unauthenticated, "certificate is revoked")
		}

		return handler(ctx, req)
	}
}

// CRLStreamInterceptor returns a StreamServerInterceptor that validates client certificates against the CRL.
func CRLStreamInterceptor(db *sql.DB) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		p, ok := peer.FromContext(ss.Context())
		if !ok {
			return status.Error(codes.Unauthenticated, "no peer found")
		}

		tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo)
		if !ok {
			return status.Error(codes.Unauthenticated, "no TLS info found")
		}

		if len(tlsInfo.State.PeerCertificates) == 0 {
			return status.Error(codes.Unauthenticated, "no client certificate provided")
		}

		clientCert := tlsInfo.State.PeerCertificates[0]
		serial := fmt.Sprintf("%x", clientCert.SerialNumber)

		revoked, err := models.IsCertificateRevoked(db, serial)
		if err != nil {
			return status.Errorf(codes.Internal, "failed to check CRL: %v", err)
		}

		if revoked {
			return status.Error(codes.Unauthenticated, "certificate is revoked")
		}

		return handler(srv, ss)
	}
}
