package metrics

import (
    "fmt"
    "time"
    
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // RequestDuration measures the duration of connector requests.
    RequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "connector_request_duration_seconds",
            Help:    "Duration of connector requests in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"connector", "operation", "status"},
    )
    
    // RequestTotal counts the total number of connector requests.
    RequestTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "connector_request_total",
            Help: "Total number of connector requests",
        },
        []string{"connector", "operation", "status"},
    )
    
    // RequestInFlight counts the number of in-flight connector requests.
    RequestInFlight = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "connector_request_in_flight",
            Help: "Number of in-flight connector requests",
        },
        []string{"connector", "operation"},
    )
    
    // ConnectorStatusGauge tracks the status of connectors.
    ConnectorStatusGauge = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "connector_status",
            Help: "Status of connectors (0=uninitialized, 1=initializing, 2=ready, 3=degraded, 4=unavailable, 5=error)",
        },
        []string{"connector", "name"},
    )
    
    // BalanceGauge tracks the balance of wallets.
    BalanceGauge = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "connector_balance",
            Help: "Balance of a wallet in base units",
        },
        []string{"connector", "wallet_id", "currency"},
    )
)

// RequestMetrics represents metrics for a connector request.
type RequestMetrics struct {
    Connector string
    Operation string
    Status    string
    Started   time.Time
}

// NewRequestMetrics creates a new RequestMetrics.
func NewRequestMetrics(connector, operation string) *RequestMetrics {
    return &RequestMetrics{
        Connector: connector,
        Operation: operation,
        Status:    "success",
        Started:   time.Now(),
    }
}

// Record records the metrics.
func (m *RequestMetrics) Record() {
    duration := time.Since(m.Started).Seconds()
    
    RequestDuration.WithLabelValues(m.Connector, m.Operation, m.Status).Observe(duration)
    RequestTotal.WithLabelValues(m.Connector, m.Operation, m.Status).Inc()
    RequestInFlight.WithLabelValues(m.Connector, m.Operation).Dec()
}

// SetStatus sets the status of the request.
func (m *RequestMetrics) SetStatus(status string) *RequestMetrics {
    m.Status = status
    return m
}

// RecordRequest records a connector request.
func RecordRequest(connector, operation, status string, duration time.Duration) {
    RequestDuration.WithLabelValues(connector, operation, status).Observe(duration.Seconds())
    RequestTotal.WithLabelValues(connector, operation, status).Inc()
}

// IncrementInFlight increments the in-flight count.
func IncrementInFlight(connector, operation string) {
    RequestInFlight.WithLabelValues(connector, operation).Inc()
}

// DecrementInFlight decrements the in-flight count.
func DecrementInFlight(connector, operation string) {
    RequestInFlight.WithLabelValues(connector, operation).Dec()
}

// SetConnectorStatus sets the connector status.
func SetConnectorStatus(connector, name, status string) {
    var value float64
    switch status {
    case "uninitialized":
        value = 0
    case "initializing":
        value = 1
    case "ready":
        value = 2
    case "degraded":
        value = 3
    case "unavailable":
        value = 4
    case "error":
        value = 5
    default:
        value = 0
    }
    ConnectorStatusGauge.WithLabelValues(connector, name).Set(value)
}

// UpdateBalance updates the balance metric.
func UpdateBalance(connector, walletID, currency string, amount float64) {
    BalanceGauge.WithLabelValues(connector, walletID, currency).Set(amount)
}