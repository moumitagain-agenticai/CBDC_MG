package cache

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/apache/fineract-cbdc-fx/internal/domain/rate"
    "github.com/apache/fineract-cbdc-fx/internal/infrastructure/config"

    "github.com/go-redis/redis/v8"
    "go.uber.org/zap"
)

// RateCache defines the rate cache interface
type RateCache interface {
    Get(ctx context.Context, baseCurrency, quoteCurrency string) (*rate.ExchangeRate, error)
    Set(ctx context.Context, rate *rate.ExchangeRate) error
    Delete(ctx context.Context, baseCurrency, quoteCurrency string) error
    Clear(ctx context.Context) error
    GetStats(ctx context.Context) (*CacheStats, error)
}

// CacheStats represents cache statistics
type CacheStats struct {
    Hits      int64 `json:"hits"`
    Misses    int64 `json:"misses"`
    Size      int64 `json:"size"`
    HitRate   float64 `json:"hitRate"`
}

// RedisRateCache implements RateCache using Redis
type RedisRateCache struct {
    client  *redis.Client
    logger  *zap.Logger
    config  *config.CacheConfig
    hits    int64
    misses  int64
}

// NewRateCache creates a new rate cache
func NewRateCache(client *redis.Client, logger *zap.Logger, config *config.CacheConfig) RateCache {
    return &RedisRateCache{
        client: client,
        logger: logger,
        config: config,
    }
}

// Get retrieves a rate from cache
func (c *RedisRateCache) Get(ctx context.Context, baseCurrency, quoteCurrency string) (*rate.ExchangeRate, error) {
    key := c.getKey(baseCurrency, quoteCurrency)

    data, err := c.client.Get(ctx, key).Bytes()
    if err != nil {
        if err == redis.Nil {
            c.misses++
            return nil, nil
        }
        c.misses++
        return nil, fmt.Errorf("failed to get from cache: %w", err)
    }

    var exchangeRate rate.ExchangeRate
    if err := json.Unmarshal(data, &exchangeRate); err != nil {
        c.misses++
        return nil, fmt.Errorf("failed to unmarshal rate: %w", err)
    }

    c.hits++
    return &exchangeRate, nil
}

// Set stores a rate in cache
func (c *RedisRateCache) Set(ctx context.Context, rate *rate.ExchangeRate) error {
    key := c.getKey(rate.BaseCurrency, rate.QuoteCurrency)

    data, err := json.Marshal(rate)
    if err != nil {
        return fmt.Errorf("failed to marshal rate: %w", err)
    }

    ttl := c.config.TTL
    if ttl == 0 {
        ttl = 30 * time.Second
    }

    if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
        return fmt.Errorf("failed to set cache: %w", err)
    }

    return nil
}

// Delete removes a rate from cache
func (c *RedisRateCache) Delete(ctx context.Context, baseCurrency, quoteCurrency string) error {
    key := c.getKey(baseCurrency, quoteCurrency)
    return c.client.Del(ctx, key).Err()
}

// Clear clears all rates from cache
func (c *RedisRateCache) Clear(ctx context.Context) error {
    return c.client.FlushDB(ctx).Err()
}

// GetStats returns cache statistics
func (c *RedisRateCache) GetStats(ctx context.Context) (*CacheStats, error) {
    // Get size from Redis
    info, err := c.client.Info(ctx, "keyspace").Result()
    if err != nil {
        return nil, err
    }

    // Parse size from info (simplified)
    var size int64
    // In production, use proper parsing

    total := c.hits + c.misses
    hitRate := 0.0
    if total > 0 {
        hitRate = float64(c.hits) / float64(total)
    }

    return &CacheStats{
        Hits:    c.hits,
        Misses:  c.misses,
        Size:    size,
        HitRate: hitRate,
    }, nil
}

// getKey generates a cache key
func (c *RedisRateCache) getKey(baseCurrency, quoteCurrency string) string {
    return fmt.Sprintf("fx:rate:%s:%s", baseCurrency, quoteCurrency)
}