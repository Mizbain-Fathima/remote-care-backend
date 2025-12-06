package transport

import (
	"context"
	"net/http"

	pb "github.com/Mizbain-Fathima/remote-care-backend/gen/github.com/Mizbain-Fathima/remote-care-backend/gen"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

func NewHTTPGateway(ctx context.Context, grpcEndpoint string) (http.Handler, error) {
	mux := runtime.NewServeMux()

	opts := []grpc.DialOption{grpc.WithInsecure()} // local dev
	if err := pb.RegisterVoucherServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts); err != nil {
		return nil, err
	}

	return mux, nil
}
