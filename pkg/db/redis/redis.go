package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Get(key string) (string, error)
	Set(key, value string, expiration time.Duration) error
}

type RedisCache struct {
	client *redis.Client
}

func NewRedis(addr string) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("FAILED TO CONNECT TO REDIS: %v", err)
	}

	return &RedisCache{
		client: client,
	}, nil
}

func (c *RedisCache) Get(key string) (string, error) {
	val, err := c.client.Get(context.Background(), key).Result()
	if !errors.Is(err, redis.Nil) && err != nil {
		return "", fmt.Errorf("FAILED TO GET VALUE FROM REDIS: %v", err)
	}
	return val, nil
}

func (c *RedisCache) Set(key, value string, expiration time.Duration) error {
	err := c.client.Set(context.Background(), key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("FAILED TO SET VALUE IN REDIS: %v", err)
	}
	return nil
}

func (c *RedisCache) Delete(key string) error {
	err := c.client.Del(context.Background(), key).Err()
	if err != nil {
		return fmt.Errorf("FAILED TO DELETE VALUE FROM REDIS: %v", err)
	}
	return nil
}

func (c *RedisCache) Clear() error {
	err := c.client.FlushAll(context.Background()).Err()
	if err != nil {
		return fmt.Errorf("FAILED TO FLASH ALL KEYS: %v", err)
	}
	return nil
}
