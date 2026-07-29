package main

import (
    "context"
    "fmt"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/fineract/cbdc/india-connector/internal/adapters/api"
    "github.com/fineract/cbdc/india-connector/internal/adapters/client"
    "github.com/fineract/cbdc/india-connector/internal/adapters/repository"
    "github.com/fineract/cbdc/india-connector/internal/config"
    "github.com/fineract/cbdc/india-connector/internal/service"
    "github.com/fineract/cbdc/india-connector/pkg/logger"
    "github.com/fineract/cbdc/india-connector/pkg/metrics"
    "go.uber.org/zap"
)

// @title Fineract CBDC India Connector API
// @version 1.0.0
// @description Connector service for Indian CBDC (Digital Rupee/e₹) integration
// @contact.name Fineract Team
// @license.name Apache 2.0
// @host localhost:8080
// @BasePath /api/v1
func main() {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        panic(fmt.Sprintf("failed to load config: %v", err))
    }

    // Initialize logger
    log, err := logger.New(cfg.Log)
    if err != nil {
        panic(fmt.Sprintf("failed to initialize logger: %v", err))
    }
    defer log.Sync()

    log.Info("starting Fineract CBDC India Connector",
        zap.String("version", cfg.Version),
        zap.String("environment", cfg.Environment),
    )

    // Initialize metrics
    metrics.Init(cfg.Metrics)

    // Initialize database
    db, err := repository.NewDB(cfg.Database)
    if err != nil {
        log.Fatal("failed to connect to database", zap.Error(err))
    }
    defer db.Close()

    // Auto-migrate schemas
    if err := repository.Migrate(db); err != nil {
        log.Fatal("failed to migrate database", zap.Error(err))
    }

    // Initialize repositories
    txRepo := repository.NewTransactionRepository(db)

    // Initialize clients
    cbdcClient, err := client.NewCBDCClient(cfg.CBDC, log)
    if err != nil {
        log.Fatal("failed to create CBDC client", zap.Error(err))
    }

    fineractClient, err := client.NewFineractClient(cfg.Fineract, log)
    if err != nil {
        log.Fatal("failed to create Fineract client", zap.Error(err))
    }

    // Initialize services
    connectorService := service.NewConnectorService(
        cbdcClient,
        fineractClient,
        txRepo,
        log,
        cfg,
    )

    transactionService := service.NewTransactionService(
        connectorService,
        txRepo,
        log,
        cfg,
    )

    healthService := service.NewHealthService(
        cbdcClient,
        fineractClient,
        db,
        cfg,
    )

    // Initialize HTTP server
    router := api.NewRouter(
        cfg.API,
        connectorService,
        transactionService,
        healthService,
        log,
    )

    srv := &http.Server{
        Addr:         fmt.Sprintf(":%d", cfg.API.Port),
        Handler:      router,
        ReadTimeout:  cfg.API.ReadTimeout,
        WriteTimeout: cfg.API.WriteTimeout,
        IdleTimeout:  cfg.API.IdleTimeout,
    }

    // Start server in goroutine
    go func() {
        log.Info("starting HTTP server", zap.Int("port", cfg.API.Port))
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal("server failed", zap.Error(err))
        }
    }()

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Info("shutting down server...")

    ctx, cancel := context.WithTimeout(context.Background(), cfg.API.ShutdownTimeout)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("server forced to shutdown", zap.Error(err))
    }

    log.Info("server exited properly")
}