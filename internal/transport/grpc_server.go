package transport

import (
	pb "github.com/Mizbain-Fathima/remote-care-backend/gen/github.com/Mizbain-Fathima/remote-care-backend/gen"
	"github.com/Mizbain-Fathima/remote-care-backend/internal/interceptors"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewGRPCServer(voucherSvc pb.VoucherServiceServer, logger *zap.Logger) *grpc.Server {

	skipAuth := map[string]bool{
		"/voucher.VoucherService/SearchVouchers":   true,
		"/voucher.VoucherService/GetBalance":       true,
		"/voucher.VoucherService/ListTransactions": true,
	}

	unaryInterceptors := []grpc.UnaryServerInterceptor{
		interceptors.UnaryTimeoutInterceptor(interceptors.DefaultUnaryTimeout),
		interceptors.UnaryLoggingInterceptor(logger),
		interceptors.SimpleAuthInterceptor(logger, skipAuth),

		// OpenTelemetry interceptor
		otelgrpc.UnaryServerInterceptor(),
	}

	streamInterceptors := []grpc.StreamServerInterceptor{
		interceptors.StreamTimeoutInterceptor(interceptors.DefaultUnaryTimeout),
		otelgrpc.StreamServerInterceptor(),
	}

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
	}

	srv := grpc.NewServer(opts...)
	pb.RegisterVoucherServiceServer(srv, voucherSvc)

	return srv
}

func DialGRPCLocal() (*grpc.ClientConn, error) {
	return grpc.Dial(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}
