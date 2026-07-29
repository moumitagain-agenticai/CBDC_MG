package api

import (
    "net/http"
    "strings"
    "time"

    "github.com/fineract/cbdc/india-connector/pkg/metrics"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/go-chi/cors"
    "github.com/go-chi/httprate"
    "go.uber.org/zap"
)

// Middleware contains the API middleware
type Middleware struct {
    logger     *zap.Logger
    rateLimit  httprate.RateLimiter
}

// NewMiddleware creates a new middleware handler
func NewMiddleware(logger *zap.Logger, rateLimitEnabled bool, rpm, burst int) *Middleware {
    var rateLimit httprate.RateLimiter
    if rateLimitEnabled {
        rateLimit = httprate.NewRateLimiter(rpm, time.Minute,
            httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
                http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
            }),
        )
    } else {
        rateLimit = httprate.NewRateLimiter(0, time.Minute)
    }

    return &Middleware{
        logger:    logger,
        rateLimit: rateLimit,
    }
}

// Chain returns the middleware chain
func (m *Middleware) Chain() []func(http.Handler) http.Handler {
    return []func(http.Handler) http.Handler{
        m.RequestID,
        m.Recoverer,
        m.Logger,
        m.CORS,
        m.RateLimit,
        m.Metrics,
        m.Timeout,
    }
}

// RequestID adds a request ID to the context
func (m *Middleware) RequestID(next http.Handler) http.Handler {
    return middleware.RequestID(next)
}

// Recoverer recovers from panics
func (m *Middleware) Recoverer(next http.Handler) http.Handler {
    return middleware.Recoverer
}

// Logger logs HTTP requests
func (m *Middleware) Logger(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        wrapped := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

        next.ServeHTTP(wrapped, r)

        m.logger.Info("HTTP request",
            zap.String("method", r.Method),
            zap.String("path", r.URL.Path),
            zap.Int("status", wrapped.Status()),
            zap.Duration("duration", time.Since(start)),
            zap.String("remote_addr", r.RemoteAddr),
            zap.String("user_agent", r.UserAgent()),
        )
    })
}

// CORS handles CORS headers
func (m *Middleware) CORS(next http.Handler) http.Handler {
    c := cors.New(cors.Options{
        AllowedOrigins:   []string{"*"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
        ExposedHeaders:   []string{"Link"},
        AllowCredentials: true,
        MaxAge:           300,
    })
    return c.Handler(next)
}

// RateLimit applies rate limiting
func (m *Middleware) RateLimit(next http.Handler) http.Handler {
    return m.rateLimit.Handler(next)
}

// Metrics records HTTP metrics
func (m *Middleware) Metrics(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        wrapped := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

        next.ServeHTTP(wrapped, r)

        metrics.RecordHTTPRequest(
            strings.TrimPrefix(r.URL.Path, "/api/v1"),
            r.Method,
            wrapped.Status(),
            time.Since(start),
        )
    })
}

// Timeout applies a timeout to requests
func (m *Middleware) Timeout(next http.Handler) http.Handler {
    return middleware.Timeout(30 * time.Second)
}