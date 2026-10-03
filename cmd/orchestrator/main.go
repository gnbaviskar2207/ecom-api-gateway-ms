package main

import (
	"context"
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
	"github.com/gnbaviskar2207/ecom-common/pkg/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	if err := run(); err != nil {
		slog.Error("product service is stopped with error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	logLevel := slog.LevelError
	if os.Getenv("ENV") == "development" {
		logLevel = slog.LevelDebug
	}
	// logger with trace correlation
	logger := slog.New(telemetry.NewTracehandler(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})))
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

	// init tracer
	shutDownTracer, err := telemetry.InitTracer(rootCtx, telemetry.Config{
		ServiceName:    cfg.ServiceName,
		ServiceVersion: cfg.ServiceVersion,
		Environment:    cfg.Environment,
		CollectorURL:   cfg.OTelConfig.CollectorURL,
	})

	if err != nil {
		logger.Error("failed to init telemetry", "error", err)
	}
	defer func() {
		if shutDownTracer != nil {
			_ = shutDownTracer(context.Background())
		}
	}()

	// all connections
	productConn, productClient, err := product.Connect(srv.cfg.ProductConfig.Address, *srv.cfg)
	if err != nil {
		return err
	}

	// middlewares

	// TODO(metrics)
	// metrics
	// registry := prometheus.NewRegistry()

	// handlers
	// wrap the mux handler in otel http handler
	otelMux := otelhttp.NewHandler(
		srv.apiMux,
		"api-gateway",
		otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
	)

	ph := httpapi.NewProductHandler(logger, productClient)
	ph.Register(srv.apiMux)

	apiServer := &http.Server{
		Addr:         srv.cfg.HttpApiConfig.Address,
		Handler:      otelMux,
		ReadTimeout:  time.Duration(srv.cfg.HttpApiConfig.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(srv.cfg.HttpApiConfig.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(srv.cfg.HttpApiConfig.IdleTimeout) * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		srv.logger.Info("http server started", "address", srv.cfg.HttpApiConfig.Address)
		errCh <- apiServer.ListenAndServe()
	}()

	select {
	case <-rootCtx.Done():
		srv.logger.Info("server is shutting down gracefully", "error", rootCtx.Err())
		// rootCtx is cancelled here automatically.
	case serveErr := <-errCh:
		srv.logger.Info("server is shutting down due to error", "error", serveErr)
		rootCtxCancel()
	}

	shutDownContext, shutDownCancel := context.WithTimeout(
		context.Background(),
		time.Duration(srv.cfg.HttpApiConfig.ShitDownTimeoutSec)*time.Second,
	)
	defer shutDownCancel()

	// gracefully shutdown
	srv.logger.Info("gracefull shutdown started...")
	if err := apiServer.Shutdown(shutDownContext); err != nil {
		srv.logger.Error("server shutdown error", "error", err)
	}
	// calling after http server shut down
	if err := productConn.Close(); err != nil {
		srv.logger.Error("product connection close error", "error", err)
	}

	close(errCh)
	srv.logger.Info("server is stopped")
	return nil
}
