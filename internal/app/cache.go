package app

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"strings"
	"time"

	"hrbackend/config"
	"hrbackend/internal/domain"

	"github.com/goccy/go-json"
	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	client    *redis.Client
	keyPrefix string
}

type noopCache struct{}

func buildCache(cfg config.Config) (domain.Cache, error) {
	if !cfg.CacheEnabled {
		return noopCache{}, nil
	}

	if strings.TrimSpace(cfg.RedisHost) == "" {
		return nil, fmt.Errorf("cache enabled but REDIS_HOST is empty")
	}

	var tlsConfig *tls.Config
	if cfg.Remote {
		tlsConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	client := redis.NewClient(&redis.Options{
		Addr:      cfg.RedisHost,
		Password:  cfg.RedisPassword,
		TLSConfig: tlsConfig,
	})

	return &redisCache{client: client, keyPrefix: cfg.CacheKeyPrefix}, nil
}

func (c *redisCache) Get(ctx context.Context, key string, dest any) (bool, error) {
	if c == nil || c.client == nil {
		return false, nil
	}

	value, err := c.client.Get(ctx, c.cacheKey(key)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, fmt.Errorf("get cache key: %w", err)
	}

	if err := json.Unmarshal(value, dest); err != nil {
		return false, fmt.Errorf("unmarshal cache value: %w", err)
	}

	return true, nil
}

func (c *redisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if c == nil || c.client == nil || ttl <= 0 {
		return nil
	}

	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal cache value: %w", err)
	}

	if err := c.client.Set(ctx, c.cacheKey(key), payload, ttl).Err(); err != nil {
		return fmt.Errorf("set cache key: %w", err)
	}
	return nil
}

func (c *redisCache) Delete(ctx context.Context, keys ...string) error {
	if c == nil || c.client == nil || len(keys) == 0 {
		return nil
	}

	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = c.cacheKey(key)
	}

	if err := c.client.Del(ctx, prefixedKeys...).Err(); err != nil {
		return fmt.Errorf("delete cache keys: %w", err)
	}
	return nil
}

func (c *redisCache) DeleteByPrefix(ctx context.Context, prefix string) error {
	if c == nil || c.client == nil {
		return nil
	}

	iter := c.client.Scan(ctx, 0, c.cacheKey(prefix)+"*", 100).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return fmt.Errorf("delete cache key by prefix: %w", err)
		}
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("scan cache keys by prefix: %w", err)
	}
	return nil
}

func (c *redisCache) Close() error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Close()
}

func (c *redisCache) cacheKey(key string) string {
	return c.keyPrefix + key
}

func (noopCache) Get(context.Context, string, any) (bool, error) { return false, nil }

func (noopCache) Set(context.Context, string, any, time.Duration) error { return nil }

func (noopCache) Delete(context.Context, ...string) error { return nil }

func (noopCache) DeleteByPrefix(context.Context, string) error { return nil }

func (noopCache) Close() error { return nil }
