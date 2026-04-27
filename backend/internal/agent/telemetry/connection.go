package telemetry

import (
	"context"
	"crypto/tls"
	"log"
	"math/rand"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	MinBackoff = 1 * time.Second
	MaxBackoff = 60 * time.Second
)

// DialWithBackoff establishes a gRPC connection with exponential backoff and jitter.
// It will retry until successful or the context is cancelled.
func DialWithBackoff(ctx context.Context, target string, tlsConfig *tls.Config) (*grpc.ClientConn, error) {
	backoff := MinBackoff
	rand.Seed(time.Now().UnixNano())

	for {
		// Use a short timeout for each individual dial attempt
		dialCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		
		// Note: grpc.NewClient is the modern way, but DialContext with WithBlock 
		// is used here to explicitly manage the retry loop as per requirements.
		conn, err := grpc.DialContext(dialCtx, target,
			grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
			grpc.WithBlock(),
		)
		cancel()

		if err == nil {
			log.Printf("Successfully connected to %s", target)
			return conn, nil
		}

		log.Printf("Connection to %s failed: %v. Retrying in %v...", target, err, backoff)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
			// Exponential backoff
			backoff = backoff * 2
			if backoff > MaxBackoff {
				backoff = MaxBackoff
			}
			
			// Add jitter (+/- 20%) to prevent thundering herd
			jitterRange := int64(backoff) / 5
			if jitterRange > 0 {
				jitter := time.Duration(rand.Int63n(jitterRange))
				if rand.Intn(2) == 0 {
					backoff += jitter
				} else {
					backoff -= jitter
				}
			}
		}
	}
}
