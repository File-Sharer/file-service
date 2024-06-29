package postgres

import (
	"context"

	"github.com/File-Sharer/file-service/internal/model"
	"github.com/jackc/pgx/v5"
)

type FileRepo struct {
	db *pgx.Conn
}

func NewFileRepo(db *pgx.Conn) *FileRepo {
	return &FileRepo{db: db}
}

func (r *FileRepo) Create(ctx context.Context, file *model.File) error {
	_, err := r.db.Exec(ctx, "insert into files(id, creator_id, is_public) values($1, $2, $3)", file.ID, file.CreatorID, file.IsPublic)
	return err
}

func (r *FileRepo) FindByID(ctx context.Context, id string) (*model.File, error) {
	var file model.File
	if err := r.db.QueryRow(ctx, "select id, creator_id, is_public, date_added from files where id = $1", id).Scan(&file.ID, &file.CreatorID, &file.IsPublic, &file.DateAdded); err != nil  {
		return nil, err
	}

	return &file, nil
}

func (r *FileRepo) AddPermission(ctx context.Context, fileID string, userID string) error {
	_, err := r.db.Exec(ctx, "insert into permissions(file_id, user_id) values($1, $2)", fileID, userID)
	return err
}

func (r *FileRepo) HasPermission(ctx context.Context, fileID string, userID string) (bool, error) {
	var exists bool
	if err := r.db.QueryRow(ctx, "select exists(select 1 from permissions where file_id = $1 and user_id = $2)", fileID, userID).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}
