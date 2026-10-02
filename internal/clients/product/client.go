package product

import (
	productsV1 "github.com/gnbaviskar2207/ecom-common/pkg/gen/products"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func Connect(address string) (*grpc.ClientConn, productsV1.ProductServiceClient, error) {
	var transport credentials.TransportCredentials
	transport = insecure.NewCredentials()
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(transport),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return nil, nil, err
	}
	return conn, productsV1.NewProductServiceClient(conn), nil
}
