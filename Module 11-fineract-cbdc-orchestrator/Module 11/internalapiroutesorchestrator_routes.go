package routes

import (
    "github.com/apache/fineract-cbdc-orchestrator/internal/api/handlers"
    "github.com/apache/fineract-cbdc-orchestrator/internal/api/middleware"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

// SetupRouter configures all HTTP routes
func SetupRouter(
    transactionService transaction.Service,
    logger *zap.Logger,
) *gin.Engine {
    router := gin.New()

    // Global middleware
    router.Use(gin.Recovery())
    router.Use(middleware.LoggingMiddleware(logger))
    router.Use(middleware.CORSMiddleware())

    // Health endpoints
    router.GET("/health/live", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "UP"})
    })
    router.GET("/health/ready", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "READY"})
    })

    // Create handler
    orchestratorHandler := handlers.NewOrchestratorHandler(transactionService, logger)

    // API v1 routes
    api := router.Group("/api/v1")
    {
        // Protected routes
        protected := api.Group("")
        protected.Use(middleware.AuthMiddleware())
        protected.Use(middleware.RateLimitMiddleware(100, 60))

        // Payment routes
        payments := protected.Group("/payments")
        {
            // POST requires idempotency
            payments.POST("/",
                middleware.IdempotencyMiddleware(),
                orchestratorHandler.InitiatePayment,
            )

            // GET operations
            payments.GET("/",
                orchestratorHandler.ListTransactions,
            )
            payments.GET("/:id",
                orchestratorHandler.GetTransaction,
            )

            // DELETE requires idempotency
            payments.DELETE("/:id",
                middleware.IdempotencyMiddleware(),
                orchestratorHandler.CancelTransaction,
            )

            // Retry
            payments.POST("/:id/retry",
                orchestratorHandler.RetryTransaction,
            )
        }
    }

    return router
}