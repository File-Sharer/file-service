package postgres

import (
	"context"

	"github.com/File-Sharer/file-service/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type fileRepo struct {
	db *pgxpool.Pool
}

func newFileRepo(db *pgxpool.Pool) File {
	return &fileRepo{db: db}
}

func (r *fileRepo) Create(ctx context.Context, file *model.File) error {
	_, err := r.db.Exec(ctx, "INSERT INTO files(id, main_folder_id, folder_id, creator_id, size, url, public, filename, download_filename) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)", file.ID, file.MainFolderID, file.FolderID, file.CreatorID, file.Size, file.URL, file.Public, file.Filename, file.DownloadFilename)
	return err
}

func (r *fileRepo) FindByID(ctx context.Context, id string) (*model.File, error) {
	var file model.File
	if err := r.db.QueryRow(ctx, "SELECT id, main_folder_id, creator_id, size, url, public, filename, download_filename, date_added FROM files WHERE id = $1 AND main_folder_id IS NULL", id).Scan(&file.ID, &file.MainFolderID, &file.CreatorID, &file.Size, &file.URL, &file.Public, &file.Filename, &file.DownloadFilename, &file.DateAdded); err != nil  {
		return nil, err
	}

	return &file, nil
}

func (r *fileRepo) FindUserFiles(ctx context.Context, userID string) ([]*model.File, error) {
	rows, err := r.db.Query(ctx, "SELECT id, creator_id, size, url, public, filename, download_filename, date_added FROM files WHERE creator_id = $1 AND main_folder_id IS NULL", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*model.File
	for rows.Next() {
		var f model.File
		if err := rows.Scan(&f.ID, &f.CreatorID, &f.Size, &f.URL, &f.Public, &f.Filename, &f.DownloadFilename, &f.DateAdded); err != nil {
			return nil, err
		}

		files = append(files, &f)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return files, nil
}

func (r *fileRepo) AddPermission(ctx context.Context, fileID string, userID string) error {
	_, err := r.db.Exec(ctx, "INSERT INTO file_permissions(file_id, user_id) VALUES($1, $2)", fileID, userID)
	return err
}

func (r *fileRepo) HasPermission(ctx context.Context, fileID string, userID string) (bool, error) {
	var exists bool
	if err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM file_permissions WHERE file_id = $1 AND user_id = $2)", fileID, userID).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *fileRepo) DeletePermission(ctx context.Context, fileID string, userID string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM file_permissions WHERE file_id = $1 AND user_id = $2", fileID, userID)
	return err
}

func (r *fileRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM files WHERE id = $1", id)
	return err
}

func (r *fileRepo) FindPermissionsToFile(ctx context.Context, id, creatorID string) ([]*model.Permission, error) {
	rows, err := r.db.Query(ctx, "SELECT p.file_id, p.user_id FROM file_permissions p JOIN files f ON f.id = p.file_id AND f.creator_id = $2 WHERE p.file_id = $1", id, creatorID)
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

func (r *fileRepo) TogglePublic(ctx context.Context, id, creatorID string) error {
	_, err := r.db.Exec(ctx, "UPDATE files SET public = NOT public WHERE id = $1 AND creator_id = $2", id, creatorID)
	return err
}

func (r *fileRepo) ClearPermissions(ctx context.Context, id, creatorID string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM file_permissions p USING files f WHERE p.file_id = $1 AND f.id = p.file_id AND f.creator_id = $2", id, creatorID)
	return err
}
