package postgres

import (
	"context"

	"github.com/File-Sharer/file-service/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserSpace interface {
	Create(ctx context.Context, d model.UserSpace) error
	Get(ctx context.Context, userID string) (*model.UserSpace, error)
	GetSize(ctx context.Context, userID string) (int64, error)
	UpdateLevel(ctx context.Context, userID string, newLevel uint8) error
}

type File interface {
	Create(ctx context.Context, file *model.File) error
	FindByID(ctx context.Context, id string) (*model.File, error)
	FindUserFiles(ctx context.Context, userID string) ([]*model.File, error)
	AddPermission(ctx context.Context, fileID string, userID string) error
	HasPermission(ctx context.Context, fileID string, userID string) (bool, error)
	DeletePermission(ctx context.Context, fileID string, userID string) error
	Delete(ctx context.Context, id string) error
	FindPermissionsToFile(ctx context.Context, id, creatorID string) ([]*model.Permission, error)
	TogglePublic(ctx context.Context, id, creatorID string) error
	ClearPermissions(ctx context.Context, id, creatorID string) error
}

type PostgresRepository struct {
	UserSpace
	File
}

func NewPostgresRepo(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		UserSpace: newUserSpaceRepo(db),
		File: NewFileRepo(db),
	}
}
