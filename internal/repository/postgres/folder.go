package postgres

import (
	"context"
	"strconv"

	"github.com/File-Sharer/file-service/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type folderRepo struct {
	db *pgxpool.Pool
}

func newFolderRepo(db *pgxpool.Pool) Folder {
	return &folderRepo{db: db}
}

func (r *folderRepo) Create(ctx context.Context, f model.Folder) error {
	_, err := r.db.Exec(
		ctx,
		"INSERT INTO folders(id, main_folder_id, folder_id, creator_id, url, name, public) VALUES($1, $2, $3, $4, $5, $6, $7)",
		f.ID, f.MainFolderID, f.FolderID, f.CreatorID, f.URL, f.Name, f.Public,
	)
	return err
}

func (r *folderRepo) FindByID(ctx context.Context, id string) (*model.Folder, error) {
	var f model.Folder
	if err := r.db.QueryRow(
		ctx,
		"SELECT id, main_folder_id, folder_id, creator_id, url, name, public, created_at FROM folders WHERE id = $1",
		id,
	).Scan(&f.ID, &f.MainFolderID, &f.FolderID, &f.CreatorID, &f.URL, &f.Name, &f.Public, &f.CreatedAt); err != nil {
		return nil, err
	}

	return &f, nil
}

func (r *folderRepo) HasPermission(ctx context.Context, id, username string) (bool, error) {
	exists := false
	if err := r.db.QueryRow(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM folder_permissions p JOIN folders f ON p.folder_id = f.id WHERE p.folder_id = $1 AND p.username = $2 AND f.main_folder_id IS NULL)",
		id, username,
	).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *folderRepo) Update(ctx context.Context, id string, fields map[string]interface{}) error {
	allowedFields := []string{"name", "public"}

	updates := map[string]any{}
	for _, allowedField := range allowedFields {
		for field, value := range fields {
			if field == allowedField {
				updates[field] = value
			}
		}
	}

	if len(updates) == 0 {
		return nil
	}

	query := "UPDATE folders SET "
	args := []interface{}{}
	i := 1

	for column, value := range updates {
		query += (column + " = $" + strconv.Itoa(i) + ", ")
		args = append(args, value)
		i += 1
	}

	query = query[:len(query)-2] + " WHERE id = $" + strconv.Itoa(i)
	args = append(args, id)

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

func (r *folderRepo) GetFolderContents(ctx context.Context, id string) ([]*model.File, []*model.Folder, error) {
	fileRows, err := r.db.Query(
		ctx,
		"SELECT id, main_folder_id, creator_id, size, url, filename, date_added from files WHERE folder_id = $1",
		id,
	)
	if err != nil {
		return nil, nil, err
	}
	defer fileRows.Close()

	var files []*model.File
	for fileRows.Next() {
		var f model.File
		if err := fileRows.Scan(&f.ID, &f.MainFolderID, &f.CreatorID, &f.Size, &f.URL, &f.Filename, &f.DateAdded); err != nil {
			return nil, nil, err
		}
		f.FolderID = &id
		files = append(files, &f)
	}

	if err := fileRows.Err(); err != nil {
		return nil, nil, err
	}

	folderRows, err := r.db.Query(
		ctx,
		"SELECT id, main_folder_id, creator_id, url, name, created_at FROM folders WHERE folder_id = $1",
		id,
	)
	if err != nil {
		return nil, nil, err
	}
	defer folderRows.Close()

	var folders []*model.Folder
	for folderRows.Next() {
		var f model.Folder
		if err := folderRows.Scan(&f.ID, &f.MainFolderID, &f.CreatorID, &f.URL, &f.Name, &f.CreatedAt); err != nil {
			return nil, nil, err
		}
		f.FolderID = &id
		folders = append(folders, &f)
	}

	if err := folderRows.Err(); err != nil {
		return nil, nil, err
	}

	return files, folders, nil
}

func (r *folderRepo) GetUserFolders(ctx context.Context, userID string) ([]*model.Folder, error) {
	rows, err := r.db.Query(
		ctx,
		"SELECT id, url, name, public, created_at FROM folders WHERE creator_id = $1 AND main_folder_id IS NULL",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var folders []*model.Folder
	for rows.Next() {
		var f model.Folder
		if err := rows.Scan(&f.ID, &f.URL, &f.Name, &f.Public, &f.CreatedAt); err != nil {
			return nil, err
		}
		f.CreatorID = userID
		folders = append(folders, &f)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return folders, nil
}

func (r *folderRepo) AddPermission(ctx context.Context, folderID, username string) error {
	_, err := r.db.Exec(ctx, "INSERT INTO folder_permissions(folder_id, username) VALUES($1, $2)", folderID, username)
	return err
}

func (r *folderRepo) DeletePermission(ctx context.Context, folderID, username string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM folder_permissions WHERE folder_id = $1 AND username = $2", folderID, username)
	return err
}

func (r *folderRepo) Delete(ctx context.Context, folderID, userID string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM folders WHERE id = $1 AND creator_id = $2", folderID, userID)
	return err
}

func (r *folderRepo) GetPermissions(ctx context.Context, folderID, creatorID string) ([]*string, error) {
	rows, err := r.db.Query(
		ctx,
		"SELECT p.username FROM folder_permissions p JOIN folders f ON f.id = p.folder_id WHERE p.folder_id = $1 AND f.creator_id = $2",
		folderID, creatorID,
	)
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

func (r *folderRepo) HasFile(ctx context.Context, folderID, filename string) (bool, error) {
	var exists bool
	if err := r.db.QueryRow(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM files WHERE folder_id = $1 AND download_name = $2)",
		folderID, filename,
	).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *folderRepo) HasFolder(ctx context.Context, userID string, folderName string) (bool, error) {
	var exists bool
	if err := r.db.QueryRow(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM folders WHERE creator_id = $1 AND name = $2)",
		userID, folderName,
	).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *folderRepo) HasFolderInFolder(ctx context.Context, folderName, folderID string) (bool, error) {
	var exists bool
	if err := r.db.QueryRow(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM folders WHERE folder_id = $1 AND name = $2)",
		folderID, folderName,
	).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}
