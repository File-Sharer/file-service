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
	_, err := r.db.Exec(ctx, "INSERT INTO files(id, main_folder_id, folder_id, creator_id, size, url, public, filename, download_name) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)", file.ID, file.MainFolderID, file.FolderID, file.CreatorID, file.Size, file.URL, file.Public, file.Filename, file.DownloadName)
	return err
}

func (r *fileRepo) FindByID(ctx context.Context, id string) (*model.File, error) {
	var file model.File
	if err := r.db.QueryRow(
		ctx,
		"SELECT id, main_folder_id, creator_id, size, url, public, filename, download_name, date_added FROM files WHERE id = $1 AND main_folder_id IS NULL",
		id).Scan(
			&file.ID,
			&file.MainFolderID,
			&file.CreatorID,
			&file.Size,
			&file.URL,
			&file.Public,
			&file.Filename,
			&file.DownloadName,
			&file.DateAdded,
			); err != nil  {
		return nil, err
	}

	return &file, nil
}

func (r *fileRepo) FindUserFiles(ctx context.Context, userID string) ([]*model.File, error) {
	rows, err := r.db.Query(ctx, "SELECT id, creator_id, size, url, public, filename, download_name, date_added FROM files WHERE creator_id = $1 AND main_folder_id IS NULL", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*model.File
	for rows.Next() {
		var f model.File
		if err := rows.Scan(&f.ID, &f.CreatorID, &f.Size, &f.URL, &f.Public, &f.Filename, &f.DownloadName, &f.DateAdded); err != nil {
			return nil, err
		}

		files = append(files, &f)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return files, nil
}

func (r *fileRepo) AddPermission(ctx context.Context, fileID, username string) error {
	_, err := r.db.Exec(ctx, "INSERT INTO file_permissions(file_id, username) VALUES($1, $2)", fileID, username)
	return err
}

func (r *fileRepo) HasPermission(ctx context.Context, fileID, username string) (bool, error) {
	exists := false
	if err := r.db.QueryRow(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM file_permissions p JOIN files f ON f.id = p.file_id WHERE p.file_id = $1 AND p.username = $2 AND f.main_folder_id IS NULL)",
		fileID, username,
		).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *fileRepo) DeletePermission(ctx context.Context, fileID, username string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM file_permissions WHERE file_id = $1 AND username = $2", fileID, username)
	return err
}

func (r *fileRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM files WHERE id = $1", id)
	return err
}

func (r *fileRepo) FindPermissionsToFile(ctx context.Context, id, creatorID string) ([]*string, error) {
	rows, err := r.db.Query(ctx, "SELECT p.username FROM file_permissions p JOIN files f ON f.id = p.file_id WHERE p.file_id = $1 AND f.creator_id = $2", id, creatorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []*string
	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			return nil, err
		}

		permissions = append(permissions, &username)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}

func (r *fileRepo) TogglePublic(ctx context.Context, id, creatorID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var public bool
	if err := tx.QueryRow(ctx, "UPDATE files SET public = NOT public WHERE id = $1 AND creator_id = $2 RETURNING public", id, creatorID).Scan(&public); err != nil {
		return err
	}

	if public {
		if _, err := tx.Exec(ctx, "DELETE FROM file_permissions WHERE file_id = $1", id); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
