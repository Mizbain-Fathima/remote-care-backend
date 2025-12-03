package transport

import (
	"context"
	"net"

	pb "github.com/Mizbain-Fathima/remote-care-backend/gen"
	"github.com/Mizbain-Fathima/remote-care-backend/internal/service"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type GRPCServer struct {
	server *grpc.Server
}

func NewGRPCServer(svc *service.VoucherService, logger *zap.Logger) *GRPCServer {

	unaryInterceptor := func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		logger.Info("grpc request", zap.String("method", info.FullMethod), zap.Any("md", md))
		return handler(ctx, req)
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(unaryInterceptor),        // your logger
		grpc.StatsHandler(otelgrpc.NewServerHandler()), // OpenTelemetry
	)

	pb.RegisterVoucherServiceServer(server, svc)

	return &GRPCServer{server: server}
}

func (g *GRPCServer) Serve(lis net.Listener) error {
	return g.server.Serve(lis)
}

func (g *GRPCServer) GracefulStop() {
	g.server.GracefulStop()
}
