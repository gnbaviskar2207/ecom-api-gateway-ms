package product

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/gnbaviskar2207/ecom-api-gateway-ms/internal/config"
	productsV1 "github.com/gnbaviskar2207/ecom-common/pkg/gen/products"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func Connect(address string, cfg config.Config) (*grpc.ClientConn, productsV1.ProductServiceClient, error) {
	var transport credentials.TransportCredentials
	var conn *grpc.ClientConn
	var err error

	if cfg.ProductConfig.CaFile == "" {
		transport = insecure.NewCredentials()
		conn, err = grpc.NewClient(
			address,
			grpc.WithTransportCredentials(transport),
			grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		)
	} else {
		pem, err := os.ReadFile(cfg.ProductConfig.CaFile)
		if err != nil {
			return nil, nil, fmt.Errorf("could not read the ca file for product repo %w", err)
		}
		roots := x509.NewCertPool()
		if !roots.AppendCertsFromPEM(pem) {
			return nil, nil, fmt.Errorf("could not append the ca file for product repo %w", err)
		}
		transport = credentials.NewTLS(&tls.Config{
			RootCAs:    roots,
			ServerName: cfg.ProductConfig.ServerName,
			MinVersion: tls.VersionTLS12,
		})
		conn, err = grpc.NewClient(
			address,
			grpc.WithTransportCredentials(transport),
			grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		)
	}
	if err != nil {
		return nil, nil, err
	}
	return conn, productsV1.NewProductServiceClient(conn), nil
}
