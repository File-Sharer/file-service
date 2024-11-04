package redisrepo

import (
	"context"
	"time"

	"github.com/File-Sharer/file-service/internal/model"
	"github.com/redis/go-redis/v9"
)

type Default interface {
	Set(ctx context.Context, key string, value interface{}, expiry time.Duration) error
	Get(ctx context.Context, key string) *redis.StringCmd
	Delete(ctx context.Context, keys ...string) error
	Incr(ctx context.Context, key string) *redis.IntCmd
	Decr(ctx context.Context, key string) *redis.IntCmd
	TTL(ctx context.Context, key string) time.Duration
}

type File interface {
	Create(ctx context.Context, key string, value []byte, expiry time.Duration) error
	Find(ctx context.Context, key string) (*model.File, error)
	FindMany(ctx context.Context, key string) ([]*model.File, error)
	HasPermission(ctx context.Context, key string) (bool, error)
	Delete(ctx context.Context, keys ...string) error
	FindPermissions(ctx context.Context, fileID string) ([]*model.Permission, error)
}

type RedisRepository struct {
	Default
	File
}

func NewRedisRepo(rdb *redis.Client) *RedisRepository {
	return &RedisRepository{
		Default: NewDefaultRedisRepo(rdb),
		File: NewFileRepo(rdb),
	}
}
