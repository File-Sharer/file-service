package redisrepo

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/File-Sharer/file-service/internal/model"
	"github.com/redis/go-redis/v9"
)

type FileRepo struct {
	rdb *redis.Client
}

func NewFileRepo(rdb *redis.Client) *FileRepo {
	return &FileRepo{rdb: rdb}
}

func (r *FileRepo) Create(ctx context.Context, key string, value []byte, expiry time.Duration) error {
	err := r.rdb.Set(ctx, key, value, expiry).Err()
	return err
}

func (r *FileRepo) Find(ctx context.Context, key string) (*model.File, error) {
	file, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var fileDB model.File
	if err := json.Unmarshal([]byte(file), &fileDB); err != nil {
		return nil, err
	}

	return &fileDB, nil
}

func (r *FileRepo) FindMany(ctx context.Context, key string) ([]*model.File, error) {
	files, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var filesDB []*model.File
	if err := json.Unmarshal([]byte(files), &filesDB); err != nil {
		return nil, err
	}

	return filesDB, nil
}

func (r *FileRepo) HasPermission(ctx context.Context, key string) (bool, error) {
	permission, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		return false, err
	}

	p, _ := strconv.ParseBool(permission)

	return p, nil
}

func (r *FileRepo) Delete(ctx context.Context, keys ...string) error {
	err := r.rdb.Del(ctx, keys...).Err()
	return err
}
