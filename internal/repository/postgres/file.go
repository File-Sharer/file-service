package postgres

import (
	"context"

	"github.com/File-Sharer/file-service/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FileRepo struct {
	db *pgxpool.Pool
}

func NewFileRepo(db *pgxpool.Pool) *FileRepo {
	return &FileRepo{db: db}
}

func (r *FileRepo) Create(ctx context.Context, file *model.File) error {
	_, err := r.db.Exec(ctx, "INSERT INTO files(id, creator_id, size, url, public, filename, download_filename) VALUES($1, $2, $3, $4, $5, $6, $7)", file.ID, file.CreatorID, file.Size, file.URL, file.Public, file.Filename, file.DownloadFilename)
	return err
}

func (r *FileRepo) FindByID(ctx context.Context, id string) (*model.File, error) {
	var file model.File
	if err := r.db.QueryRow(ctx, "SELECT id, creator_id, size, url, public, filename, download_filename, date_added FROM files WHERE id = $1", id).Scan(&file.ID, &file.CreatorID, &file.Size, &file.URL, &file.Public, &file.Filename, &file.DownloadFilename, &file.DateAdded); err != nil  {
		return nil, err
	}

	return &file, nil
}

func (r *FileRepo) FindUserFiles(ctx context.Context, userID string) ([]*model.File, error) {
	rows, err := r.db.Query(ctx, "SELECT id, creator_id, size, url, public, filename, download_filename, date_added FROM files WHERE creator_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*model.File
	for rows.Next() {
		var file model.File
		if err := rows.Scan(&file.ID, &file.CreatorID, &file.Size, &file.URL, &file.Public, &file.Filename, &file.DownloadFilename, &file.DateAdded); err != nil {
			return nil, err
		}

		files = append(files, &file)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return files, nil
}

func (r *FileRepo) AddPermission(ctx context.Context, fileID string, userID string) error {
	_, err := r.db.Exec(ctx, "INSERT INTO permissions(file_id, user_id) VALUES($1, $2)", fileID, userID)
	return err
}

func (r *FileRepo) HasPermission(ctx context.Context, fileID string, userID string) (bool, error) {
	var exists bool
	if err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM permissions WHERE file_id = $1 AND user_id = $2)", fileID, userID).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *FileRepo) DeletePermission(ctx context.Context, fileID string, userID string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM permissions WHERE file_id = $1 AND user_id = $2", fileID, userID)
	return err
}

func (r *FileRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM files WHERE id = $1", id)
	return err
}

func (r *FileRepo) FindPermissionsToFile(ctx context.Context, id, creatorID string) ([]*model.Permission, error) {
	rows, err := r.db.Query(ctx, "SELECT p.file_id, p.user_id FROM permissions p JOIN files f ON f.id = p.file_id AND f.creator_id = $2 WHERE p.file_id = $1", id, creatorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []*model.Permission
	for rows.Next() {
		var permission model.Permission
		if err := rows.Scan(&permission.FileID, &permission.UserID); err != nil {
			return nil, err
		}

		permissions = append(permissions, &permission)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}

func (r *FileRepo) TogglePublic(ctx context.Context, id, creatorID string) error {
	_, err := r.db.Exec(ctx, "UPDATE files SET public = NOT public WHERE id = $1 AND creator_id = $2", id, creatorID)
	return err
}

func (r *FileRepo) ClearPermissions(ctx context.Context, id, creatorID string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM permissions p USING files f WHERE p.file_id = $1 AND f.id = p.file_id AND f.creator_id = $2", id, creatorID)
	return err
}
