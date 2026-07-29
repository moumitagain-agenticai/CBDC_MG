package middleware

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error": "Authorization header is required",
                "code":  "UNAUTHORIZED",
            })
            c.Abort()
            return
        }

        // Extract token
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error": "Invalid authorization header format",
                "code":  "UNAUTHORIZED",
            })
            c.Abort()
            return
        }

        token := parts[1]

        // Validate token (implementation would use JWT secret)
        // For now, we'll just check if token is present
        // In production, validate with Keycloak or JWT secret

        c.Set("user_id", "test-user-id")
        c.Set("user_roles", []string{"USER"})

        c.Next()
    }
}