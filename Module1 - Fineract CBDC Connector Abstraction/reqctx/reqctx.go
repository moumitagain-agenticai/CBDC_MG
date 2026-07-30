package reqctx

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type contextKey string

const (
	// RequestIDKey is the key for request ID in context.
	RequestIDKey contextKey = "request_id"

	// ConnectorIDKey is the key for connector ID in context.
	ConnectorIDKey contextKey = "connector_id"

	// TraceIDKey is the key for trace ID in context.
	TraceIDKey contextKey = "trace_id"

	// SpanIDKey is the key for span ID in context.
	SpanIDKey contextKey = "span_id"

	// UserIDKey is the key for user ID in context.
	UserIDKey contextKey = "user_id"

	// ClientIPKey is the key for client IP in context.
	ClientIPKey contextKey = "client_ip"
)

// WithRequestID adds a request ID to the context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// WithConnectorID adds a connector ID to the context.
func WithConnectorID(ctx context.Context, connectorID string) context.Context {
	return context.WithValue(ctx, ConnectorIDKey, connectorID)
}

// WithTraceID adds a trace ID to the context.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// GetRequestID gets the request ID from the context.
func GetRequestID(ctx context.Context) string {
	if v := ctx.Value(RequestIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetConnectorID gets the connector ID from the context.
func GetConnectorID(ctx context.Context) string {
	if v := ctx.Value(ConnectorIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetTraceID gets the trace ID from the context.
func GetTraceID(ctx context.Context) string {
	if v := ctx.Value(TraceIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// NewRequestContext creates a new context with request metadata.
func NewRequestContext(parent context.Context) context.Context {
	ctx := parent
	if ctx == nil {
		ctx = context.Background()
	}

	// Generate request ID if not present
	if GetRequestID(ctx) == "" {
		ctx = WithRequestID(ctx, uuid.New().String())
	}

	// Generate trace ID if not present
	if GetTraceID(ctx) == "" {
		ctx = WithTraceID(ctx, uuid.New().String())
	}

	return ctx
}

// ContextWithTimeout creates a context with timeout.
func ContextWithTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithTimeout(parent, timeout)
}
