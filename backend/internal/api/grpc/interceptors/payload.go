package interceptors

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

const MaxPayloadSize = 4 * 1024 * 1024 // 4MB

// PayloadSizeInterceptor rejects streamed messages larger than 4MB.
func PayloadSizeInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wrapper := &payloadSizeWrapper{
			ServerStream: ss,
		}
		return handler(srv, wrapper)
	}
}

type payloadSizeWrapper struct {
	grpc.ServerStream
}

func (w *payloadSizeWrapper) RecvMsg(m interface{}) error {
	if err := w.ServerStream.RecvMsg(m); err != nil {
		return err
	}

	// Check size of the received message
	if msg, ok := m.(proto.Message); ok {
		size := proto.Size(msg)
		if size > MaxPayloadSize {
			return status.Errorf(codes.ResourceExhausted, "payload size %d exceeds limit of %d", size, MaxPayloadSize)
		}
	}

	return nil
}
