package tracing

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Tracer is the global tracer.
var Tracer trace.Tracer

func init() {
	Tracer = otel.Tracer("fineract/cbdc/connector")
}

// Span wraps a trace span with helper methods.
type Span struct {
	trace.Span
}

// StartSpan starts a new span.
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, *Span) {
	ctx, span := Tracer.Start(ctx, name, opts...)
	return ctx, &Span{span}
}

// StartChildSpan starts a new child span.
func StartChildSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, *Span) {
	opts = append(opts, trace.WithNewRoot())
	return StartSpan(ctx, name, opts...)
}

// AddAttribute adds an attribute to the span.
func (s *Span) AddAttribute(key string, value interface{}) {
	if s.Span == nil {
		return
	}

	var attr attribute.KeyValue
	switch v := value.(type) {
	case string:
		attr = attribute.String(key, v)
	case int:
		attr = attribute.Int(key, v)
	case int64:
		attr = attribute.Int64(key, v)
	case float64:
		attr = attribute.Float64(key, v)
	case bool:
		attr = attribute.Bool(key, v)
	case time.Time:
		attr = attribute.String(key, v.Format(time.RFC3339Nano))
	case fmt.Stringer:
		attr = attribute.String(key, v.String())
	default:
		attr = attribute.String(key, fmt.Sprintf("%v", v))
	}
	s.SetAttributes(attr)
}

// AddError records an error on the span.
func (s *Span) AddError(err error) {
	if s.Span == nil || err == nil {
		return
	}
	s.RecordError(err)
	s.SetStatus(codes.Error, err.Error())
}

// AddEvent adds an event to the span.
func (s *Span) AddEvent(name string, attrs ...attribute.KeyValue) {
	if s.Span == nil {
		return
	}
	s.Span.AddEvent(name, trace.WithAttributes(attrs...))
}

// WithTransaction adds transaction attributes to the span.
func (s *Span) WithTransaction(txID, txType, status string) *Span {
	s.AddAttribute("transaction.id", txID)
	s.AddAttribute("transaction.type", txType)
	s.AddAttribute("transaction.status", status)
	return s
}

// WithWallet adds wallet attributes to the span.
func (s *Span) WithWallet(walletID string) *Span {
	s.AddAttribute("wallet.id", walletID)
	return s
}

// WithAmount adds amount attributes to the span.
func (s *Span) WithAmount(amount, currency string) *Span {
	s.AddAttribute("amount.value", amount)
	s.AddAttribute("amount.currency", currency)
	return s
}

// GetSpanFromContext retrieves the span from context.
func GetSpanFromContext(ctx context.Context) *Span {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return nil
	}
	return &Span{span}
}
