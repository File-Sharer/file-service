package redisrepo

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

func SetJSON(r *redis.Client, ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.Set(ctx, key, valueJSON, expiration).Err()
}

func Get[T any](r *redis.Client, ctx context.Context, key string) (*T, error) {
	value, err := r.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	if value == "null" {
		return nil, nil
	}

	var result T
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func GetMany[T any](r *redis.Client, ctx context.Context, key string) ([]*T, error) {
	value, err := r.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	if value == "null" {
		return nil, nil
	}

	var result []*T
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		return nil, err
	}

	return result, nil
}
