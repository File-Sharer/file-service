package postgres

import (
	"context"
	"database/sql"

	"github.com/File-Sharer/file-service/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userSpaceRepo struct {
	db *pgxpool.Pool
}

func newUserSpaceRepo(db *pgxpool.Pool) UserSpace {
	return &userSpaceRepo{
		db: db,
	}
}

func (r *userSpaceRepo) Create(ctx context.Context, d model.UserSpace) error {
	_, err := r.db.Exec(ctx, "INSERT INTO users_spaces(user_id, username) VALUES($1, $2)", d.UserID, d.Username)
	return err
}

func (r *userSpaceRepo) GetByUserID(ctx context.Context, userID string) (*model.UserSpace, error) {
	space := new(model.UserSpace)
	if err := r.db.QueryRow(
		ctx,
		"SELECT user_id, username, level, created_at FROM users_spaces WHERE user_id = $1",
		userID,
	).Scan(&space.UserID, &space.Username, &space.Level, &space.CreatedAt); err != nil && err != pgx.ErrNoRows {
		return nil, err
	}

	return space, nil
}

func (r *userSpaceRepo) GetFull(ctx context.Context, userID string) (*model.FullUserSpace, error) {
	space := new(model.FullUserSpace)
	if err := r.db.QueryRow(
		ctx,
		`
		SELECT s.user_id, s.username, s.level, s.created_at, SUM(f.size)
		FROM users_spaces s
		LEFT JOIN files f ON f.creator_id = s.user_id
		WHERE s.user_id = $1
		GROUP BY s.user_id, s.username, s.level, s.created_at
		`,
		userID,
	).Scan(&space.UserID, &space.Username, &space.Level, &space.CreatedAt, &space.Size); err != nil && err != pgx.ErrNoRows {
		return nil, err
	}

	return space, nil
}

func (r *userSpaceRepo) GetByUsername(ctx context.Context, username string) (*model.UserSpace, error) {
	space := new(model.UserSpace)
	if err := r.db.QueryRow(
		ctx,
		"SELECT user_id, username, level, created_at FROM users_spaces WHERE username = $1",
		username,
	).Scan(&space.UserID, &space.Username, &space.Level, &space.CreatedAt); err != nil && err != pgx.ErrNoRows {
		return nil, err
	}
	
	return space, nil
}

func (r *userSpaceRepo) GetSize(ctx context.Context, userID string) (int64, error) {
	var nullableSize sql.NullInt64
	if err := r.db.QueryRow(
		ctx,
		"SELECT SUM(size) FROM files WHERE creator_id = $1",
		userID,
	).Scan(&nullableSize); err != nil {
		return 0, err
	}

	if nullableSize.Valid {
		return nullableSize.Int64, nil
	}

	return 0, nil
}

func (r *userSpaceRepo) UpdateLevel(ctx context.Context, userID string, newLevel uint8) error {
	_, err := r.db.Exec(ctx, "UPDATE users_spaces SET level = $1 WHERE user_id = $2", newLevel, userID)
	return err
}
