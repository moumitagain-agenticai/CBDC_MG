package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/apache/fineract-cbdc-orchestrator/internal/api/routes"
    "github.com/apache/fineract-cbdc-orchestrator/internal/infrastructure/config"
    "github.com/apache/fineract-cbdc-orchestrator/internal/infrastructure/repository"
    "github.com/apache/fineract-cbdc-orchestrator/internal/service"
    "github.com/apache/fineract-cbdc-orchestrator/pkg/logger"
    "github.com/apache/fineract-cbdc-orchestrator/pkg/metrics"

    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "go.uber.org/zap"
    "gorm.io/gorm"
)

func main() {
    // Load configuration
    cfg, err := config.LoadConfig("configs/config.yaml")
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Initialize logger
    logger, err := logger.NewLogger(cfg.LogLevel)
    if err != nil {
        log.Fatalf("Failed to initialize logger: %v", err)
    }
    defer logger.Sync()

    // Initialize metrics
    metrics.InitMetrics()

    // Initialize database
    db, err := repository.NewPostgresConnection(&cfg.Database)
    if err != nil {
        logger.Fatal("Failed to connect to database", zap.Error(err))
    }

    // Initialize Redis
    redis, err := repository.NewRedisConnection(&cfg.Redis)
    if err != nil {
        logger.Warn("Failed to connect to Redis", zap.Error(err))
    }

    // Initialize repositories
    transactionRepo := repository.NewTransactionRepository(db)

    // Initialize service
    orchestratorService := service.NewOrchestratorService(
        transactionRepo,
        logger,
        cfg.Orchestrator,
        redis,
    )

    // Setup router
    router := routes.SetupRouter(orchestratorService, logger)

    // Prometheus metrics endpoint
    router.GET("/metrics", gin.WrapH(promhttp.Handler()))

    // Health check endpoints
    router.GET("/health/live", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "UP"})
    })

    router.GET("/health/ready", func(c *gin.Context) {
        sqlDB, _ := db.DB()
        if err := sqlDB.Ping(); err != nil {
            c.JSON(http.StatusServiceUnavailable, gin.H{"status": "DOWN", "error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{"status": "READY"})
    })

    // HTTP server
    srv := &http.Server{
        Addr:         fmt.Sprintf(":%d", cfg.Port),
        Handler:      router,
        ReadTimeout:  30 * time.Second,
        WriteTimeout: 30 * time.Second,
    }

    // Start server
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Fatal("Failed to start server", zap.Error(err))
        }
    }()

    logger.Info("Orchestrator started", zap.Int("port", cfg.Port))

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    logger.Info("Shutting down orchestrator...")

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        logger.Error("Server forced to shutdown", zap.Error(err))
    }

    logger.Info("Orchestrator stopped")
}