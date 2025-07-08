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
	_, err := r.db.Exec(ctx, "insert into files(id, creator_id, url, public, filename, download_filename) values($1, $2, $3, $4, $5, $6)", file.ID, file.CreatorID, file.URL, file.Public, file.Filename, file.DownloadFilename)
	return err
}

func (r *FileRepo) FindByID(ctx context.Context, id string) (*model.File, error) {
	var file model.File
	if err := r.db.QueryRow(ctx, "select id, creator_id, url, public, filename, download_filename, date_added from files where id = $1", id).Scan(&file.ID, &file.CreatorID, &file.URL, &file.Public, &file.Filename, &file.DownloadFilename, &file.DateAdded); err != nil  {
		return nil, err
	}

	return &file, nil
}

func (r *FileRepo) FindUserFiles(ctx context.Context, userID string) ([]*model.File, error) {
	rows, err := r.db.Query(ctx, "select id, creator_id, url, public, filename, download_filename, date_added from files where creator_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*model.File
	for rows.Next() {
		var file model.File
		if err := rows.Scan(&file.ID, &file.CreatorID, &file.URL, &file.Public, &file.Filename, &file.DownloadFilename, &file.DateAdded); err != nil {
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

func (r *FileRepo) DeletePermission(ctx context.Context, fileID string, userID string) error {
	_, err := r.db.Exec(ctx, "delete from permissions where file_id = $1 and user_id = $2", fileID, userID)
	return err
}

func (r *FileRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "delete from files where id = $1", id)
	return err
}

func (r *FileRepo) FindPermissionsToFile(ctx context.Context, id, creatorID string) ([]*model.Permission, error) {
	rows, err := r.db.Query(ctx, "select p.file_id, p.user_id from permissions p join files f on f.id = p.file_id and f.creator_id = $2 where file_id = $1", id, creatorID)
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
	_, err := r.db.Exec(ctx, "update files set public = not public where id = $1 and creator_id = $2", id, creatorID)
	return err
}

func (r *FileRepo) ClearPermissions(ctx context.Context, id, creatorID string) error {
	_, err := r.db.Exec(ctx, "delete from permissions p join files f on f.id = p.file_id and f.creator_id = $2 where p.file_id = $1", id, creatorID)
	return err
}
