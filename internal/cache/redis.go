package cache

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
	ttl    time.Duration
}

func New(addr, password string, useTLS bool, defaultTTL time.Duration) *Cache {
	opt := &redis.Options{
		Addr:     addr,
		Password: password,
	}

	// Enable TLS for Upstash
	if useTLS {
		opt.TLSConfig = &tls.Config{
			InsecureSkipVerify: true, // Upstash allows this
		}
	}

	client := redis.NewClient(opt)

	if defaultTTL == 0 {
		defaultTTL = 5 * time.Minute
	}

	return &Cache{
		client: client,
		ttl:    defaultTTL,
	}
}

func (c *Cache) Close() error {
	if c.client == nil {
		return nil
	}
	return c.client.Close()
}

func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *Cache) Set(ctx context.Context, key, val string, ttl time.Duration) error {
	if ttl == 0 {
		ttl = c.ttl
	}
	return c.client.Set(ctx, key, val, ttl).Err()
}

func (c *Cache) Del(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *Cache) GetObj(ctx context.Context, key string, dest interface{}) (bool, error) {
	s, err := c.Get(ctx, key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}

	if err := json.Unmarshal([]byte(s), dest); err != nil {
		return false, err
	}

	return true, nil
}

func (c *Cache) WriteThrough(ctx context.Context, key string, obj interface{}, ttl time.Duration) error {
	b, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	return c.Set(ctx, key, string(b), ttl)
}

func NewWithOptions(opt *redis.Options, ttl time.Duration) *Cache {
	client := redis.NewClient(opt)
	if ttl == 0 {
		ttl = 5 * time.Minute
	}
	return &Cache{client: client, ttl: ttl}
}
