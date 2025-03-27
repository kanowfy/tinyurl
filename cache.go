package tinyurl

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var (
	ErrNotInCache = errors.New("not in cache")
)

type Cache interface {
	Get(code string) (string, error)
	Set(code string, url string) error
}

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client}
}

func (c *RedisCache) Get(code string) (string, error) {
	res, err := c.client.Get(context.Background(), code).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrNotInCache
		} else {
			return "", fmt.Errorf("get from cache: %w", err)
		}
	}

	return res, nil
}

func (c *RedisCache) Set(code string, url string) error {
	cmd := c.client.Set(context.Background(), code, url, 0)
	if err := cmd.Err(); err != nil {
		return fmt.Errorf("set to cache: %w", err)
	}

	return nil
}
