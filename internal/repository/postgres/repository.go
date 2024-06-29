package postgres

import (
	"context"

	"github.com/File-Sharer/file-service/internal/model"
	"github.com/jackc/pgx/v5"
)

type File interface {
	Create(ctx context.Context, file *model.File) error
	FindByID(ctx context.Context, id string) (*model.File, error)
	AddPermission(ctx context.Context, fileID string, userID string) error
	HasPermission(ctx context.Context, fileID string, userID string) (bool, error)
}

type PostgresRepository struct {
	File
}

func NewPostgresRepo(db *pgx.Conn) *PostgresRepository {
	return &PostgresRepository{
		File: NewFileRepo(db),
	}
}
