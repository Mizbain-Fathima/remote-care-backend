package interceptors

import (
	"context"
	"strconv"
	"time"

	rcAuth "github.com/Mizbain-Fathima/remote-care-backend/internal/auth"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const DefaultUnaryTimeout = 8 * time.Second

// Extract metadata fields
func extractMeta(ctx context.Context) (reqID, auth, ua string) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {

		if v := md.Get("x-request-id"); len(v) > 0 {
			reqID = v[0]
		}
		if v := md.Get("authorization"); len(v) > 0 {
			auth = v[0]
		}
		if v := md.Get("user-agent"); len(v) > 0 {
			ua = v[0]
		}
	}
	return
}

// Timeout interceptor
func UnaryTimeoutInterceptor(defaultDuration time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if _, ok := ctx.Deadline(); !ok && defaultDuration > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, defaultDuration)
			defer cancel()
		}
		return handler(ctx, req)
	}
}

// Logging interceptor
func UnaryLoggingInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		reqID, auth, ua := extractMeta(ctx)

		peerAddr := ""
		if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
			peerAddr = p.Addr.String()
		}

		logger.Info("gRPC request started",
			zap.String("method", info.FullMethod),
			zap.String("request_id", reqID),
			zap.String("user_agent", ua),
			zap.String("peer", peerAddr),
		)

		resp, err := handler(ctx, req)

		duration := time.Since(start)

		if err != nil {
			logger.Error("gRPC request failed",
				zap.String("method", info.FullMethod),
				zap.Duration("latency_ms", duration),
				zap.Error(err),
				zap.String("request_id", reqID),
				zap.String("auth_present", strconv.FormatBool(auth != "")),
			)
		} else {
			logger.Info("gRPC request finished",
				zap.String("method", info.FullMethod),
				zap.Duration("latency_ms", duration),
				zap.String("request_id", reqID),
			)
		}

		return resp, err
	}
}

// Simple authentication interceptor
func SimpleAuthInterceptor(logger *zap.Logger, skip map[string]bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		if skip[info.FullMethod] {
			return handler(ctx, req)
		}

		_, authHeader, _ := extractMeta(ctx)

		if authHeader == "" {
			logger.Warn("unauthenticated request", zap.String("method", info.FullMethod))
			return nil, grpc.Errorf(codes.Unauthenticated, "missing authorization header")
		}

		token, err := jwt.Parse(authHeader, func(token *jwt.Token) (interface{}, error) {
			return rcAuth.JWTSecret, nil
		})

		if err != nil || !token.Valid {
			logger.Warn("invalid jwt", zap.Error(err))
			return nil, grpc.Errorf(codes.Unauthenticated, "invalid token")
		}

		return handler(ctx, req)
	}
}

// Streaming timeout interceptor
func StreamTimeoutInterceptor(defaultTimeout time.Duration) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()

		if _, ok := ctx.Deadline(); !ok && defaultTimeout > 0 {
			ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
			defer cancel()

			wrapped := &wrappedStream{ServerStream: ss, ctx: ctx}
			return handler(srv, wrapped)
		}

		return handler(srv, ss)
	}
}

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (ws *wrappedStream) Context() context.Context { return ws.ctx }
