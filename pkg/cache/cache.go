package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache wraps a Redis client with context.
type Cache struct {
	ctx    context.Context
	client *redis.Client
}

// NewCache initializes and returns a Cache instance.
func NewCache(ctx context.Context, host, password string) *Cache {
	client := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password,
		DB:       0,
	})

	return &Cache{
		ctx:    ctx,
		client: client,
	}
}

// Get retrieves a value by key from the cache.
func (c *Cache) Get(key string) (string, bool, error) {
	val, err := c.client.Get(c.ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", false, nil // Key does not exist
	} else if err != nil {
		return "", false, err // Redis error
	}
	return val, true, nil
}

// Set stores a key-value pair in the cache with an optional TTL.
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) error {
	return c.client.Set(c.ctx, key, value, ttl).Err()
}
