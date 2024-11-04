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
	fileCache, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var file model.File
	if err := json.Unmarshal([]byte(fileCache), &file); err != nil {
		return nil, err
	}

	return &file, nil
}

func (r *FileRepo) FindMany(ctx context.Context, key string) ([]*model.File, error) {
	filesCache, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var files []*model.File
	if err := json.Unmarshal([]byte(filesCache), &files); err != nil {
		return nil, err
	}

	return files, nil
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

func (r *FileRepo) FindPermissions(ctx context.Context, fileID string) ([]*model.Permission, error) {
	permissionsCache, err := r.rdb.Get(ctx, fileID).Result()
	if err != nil {
		return nil, err
	}

	var permissions []*model.Permission
	if err := json.Unmarshal([]byte(permissionsCache), &permissions); err != nil {
		return nil, err
	}

	return permissions, nil
}
