package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
}

func New(addr string) *Cache {
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	return &Cache{client: rdb}
}

func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *Cache) Set(ctx context.Context, key, val string, ttl time.Duration) error {
	return c.client.Set(ctx, key, val, ttl).Err()
}

func (c *Cache) Close() error {
	return c.client.Close()
}
