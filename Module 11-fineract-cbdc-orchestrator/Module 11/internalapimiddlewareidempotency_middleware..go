package middleware

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

// IdempotencyMiddleware ensures requests are idempotent
// In core banking, idempotency keys are MANDATORY for state-changing operations
func IdempotencyMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Only apply to state-changing methods
        if c.Request.Method != "POST" && c.Request.Method != "PUT" && c.Request.Method != "PATCH" && c.Request.Method != "DELETE" {
            c.Next()
            return
        }

        // For idempotent methods (GET, HEAD, OPTIONS), skip validation
        if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
            c.Next()
            return
        }

        // Validate idempotency key - MANDATORY for state-changing operations
        idempotencyKey := c.GetHeader("Idempotency-Key")

        // In core banking, we REJECT requests without an idempotency key
        if idempotencyKey == "" {
            c.JSON(http.StatusBadRequest, gin.H{
                "error": "Idempotency-Key header is required for this operation",
                "code":  "MISSING_IDEMPOTENCY_KEY",
                "message": "Please provide an Idempotency-Key header to prevent duplicate processing",
            })
            c.Abort()
            return
        }

        // Validate idempotency key format (must be a valid UUID)
        if err := validateIdempotencyKey(idempotencyKey); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "error":   "Invalid Idempotency-Key format",
                "code":    "INVALID_IDEMPOTENCY_KEY",
                "message": "Idempotency-Key must be a valid UUID v4 format: " + err.Error(),
            })
            c.Abort()
            return
        }

        // Store in context for handlers to use
        c.Set("idempotency_key", idempotencyKey)
        c.Header("Idempotency-Key", idempotencyKey)

        c.Next()
    }
}

// validateIdempotencyKey validates the idempotency key format
func validateIdempotencyKey(key string) error {
    if key == "" {
        return nil
    }

    if _, err := uuid.Parse(key); err != nil {
        return err
    }

    return nil
}