package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gnbaviskar2207/ecom-api-gateway-ms/internal/clients/product"
	"github.com/gnbaviskar2207/ecom-api-gateway-ms/internal/config"
	"github.com/gnbaviskar2207/ecom-api-gateway-ms/internal/httpapi"
)

func main() {
	if err := run(); err != nil {
		slog.Error("product service is stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)
	configPath := flag.String("config", "./configs/dev/config.yaml", "optional YAML configuration file")
	flag.Parse()
	logger.Info("config path", "configPath", *configPath)
	cfg, err := config.Load(*configPath, logger)
	if err != nil {
		return err
	}
	srv := New(cfg, logger)

	rootCtx, rootCtxCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer rootCtxCancel()
	// all connections
	// conn, productClient, err := product.Connect(cfg.Product.Address, cfg.Product.TLSCAFile, cfg.Product.ServerName, !production)

	// all conns
	productConn, productClient, err := product.Connect("0.0.0.0:50051")
	if err != nil {
		return err
	}
	defer productConn.Close()

	// middlewares

	srv.apiMux = http.NewServeMux()

	// metrics
	// registry := prometheus.NewRegistry()

	ph := httpapi.NewProductHandler(logger, productClient)
	ph.Register(srv.apiMux)

	apiServer := &http.Server{
		Addr:    srv.cfg.HttpApiConfig.Address,
		Handler: srv.apiMux,
	}

	errCh := make(chan error, 1)
	go func() {
		srv.logger.Info("http server started", "address", srv.cfg.HttpApiConfig.Address)
		errCh <- apiServer.ListenAndServe()
	}()

	select {
	case <-rootCtx.Done():
		srv.logger.Info("server is shutting down gracefully", "error", rootCtx.Err())

	case serveErr := <-errCh:
		if !errors.Is(serveErr, http.ErrServerClosed) {
			return serveErr
		}
		srv.logger.Info("server is shutting down due to error", "error", serveErr)
	}

	// todo: timeout from config
	shutDownContext, shutDownCancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer shutDownCancel()
	// stoppedCh := make(chan struct{})

	// gracefully shutdown
	srv.logger.Info("gracefull shutdown started...")
	if err := apiServer.Shutdown(shutDownContext); err != nil {
		srv.logger.Error("server shutdown error", "error", err)
	}

	// close(errCh)
	srv.logger.Info("server is stopped")
	return nil
}
