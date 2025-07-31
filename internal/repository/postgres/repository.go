package postgres

import (
	"context"

	"github.com/File-Sharer/file-service/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserSpace interface {
	Create(ctx context.Context, d model.UserSpace) error
	GetByID(ctx context.Context, userID string) (*model.UserSpace, error)
	GetByUsername(ctx context.Context, username string) (*model.UserSpace, error)
	GetSize(ctx context.Context, userID string) (int64, error)
	UpdateLevel(ctx context.Context, userID string, newLevel uint8) error
}

type Folder interface {
	Create(ctx context.Context, f model.Folder) error
	FindByID(ctx context.Context, id string) (*model.Folder, error)
	HasPermission(ctx context.Context, id, username string) (bool, error)
	Update(ctx context.Context, id string, fields map[string]interface{}) error
	GetFolderContents(ctx context.Context, id string) ([]*model.File, []*model.Folder, error)
	GetUserFolders(ctx context.Context, userID string) ([]*model.Folder, error)
	AddPermission(ctx context.Context, folderID, username string) error
	DeletePermission(ctx context.Context, folderID, username string) error
	GetPermissions(ctx context.Context, folderID, creatorID string) ([]*string, error)
}

type File interface {
	Create(ctx context.Context, file *model.File) error
	FindByID(ctx context.Context, id string) (*model.File, error)
	FindUserFiles(ctx context.Context, userID string) ([]*model.File, error)
	AddPermission(ctx context.Context, fileID, username string) error
	HasPermission(ctx context.Context, fileID, username string) (bool, error)
	DeletePermission(ctx context.Context, fileID, username string) error
	Delete(ctx context.Context, id string) error
	FindPermissionsToFile(ctx context.Context, id, creatorID string) ([]*string, error)
	TogglePublic(ctx context.Context, id, creatorID string) error
}

type PostgresRepository struct {
	UserSpace
	Folder
	File
}

func NewPostgresRepo(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		UserSpace: newUserSpaceRepo(db),
		Folder: newFolderRepo(db),
		File: newFileRepo(db),
	}
}
