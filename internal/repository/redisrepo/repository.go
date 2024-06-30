package redisrepo

import (
	"context"
	"time"

	"github.com/File-Sharer/file-service/internal/model"
	"github.com/redis/go-redis/v9"
)

type File interface {
	Create(ctx context.Context, key string, value []byte, expiry time.Duration) error
	Find(ctx context.Context, key string) (*model.File, error)
	FindMany(ctx context.Context, key string) ([]*model.File, error)
	HasPermission(ctx context.Context, key string) (bool, error)
	Delete(ctx context.Context, keys ...string) error
}

type RedisRepository struct {
	File
}

func NewRedisRepo(rdb *redis.Client) *RedisRepository {
	return &RedisRepository{
		File: NewFileRepo(rdb),
	}
}
