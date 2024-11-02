package redisrepo

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type DefaultRepo struct {
	rdb *redis.Client
}

func NewDefaultRedisRepo(rdb *redis.Client) *DefaultRepo {
	return &DefaultRepo{rdb: rdb}
}

func (r *DefaultRepo) Set(ctx context.Context, key string, value interface{}, expiry time.Duration) error {
	return r.rdb.Set(ctx, key, value, expiry).Err()
}

func (r *DefaultRepo) Get(ctx context.Context, key string) *redis.StringCmd {
	return r.rdb.Get(ctx, key)
}

func (r *DefaultRepo) Delete(ctx context.Context, keys ...string) error {
	return r.rdb.Del(ctx, keys...).Err()
}

func (r *DefaultRepo) Incr(ctx context.Context, key string) *redis.IntCmd {
	return r.rdb.Incr(ctx, key)
}

func (r *DefaultRepo) Decr(ctx context.Context, key string) *redis.IntCmd {
	return r.rdb.Decr(ctx, key)
}

func (r *DefaultRepo) TTL(ctx context.Context, key string) time.Duration {
	return r.rdb.TTL(ctx, key).Val()
}
