package postgres

import (
	"context"

	"github.com/File-Sharer/file-service/internal/model"
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
	_, err := r.db.Exec(ctx, "insert into users_spaces(user_id) values($1)", d.UserID)
	return err
}

func (r *userSpaceRepo) Get(ctx context.Context, userID string) (*model.UserSpace, error) {
	var userSpace model.UserSpace
	if err := r.db.QueryRow(
		ctx,
		"select level from users_spaces where user_id = $1",
		userID,
	).Scan(&userSpace.Level); err != nil {
		return nil, err
	}

	userSpace.UserID = userID
	return &userSpace, nil
}

func (r *userSpaceRepo) GetSize(ctx context.Context, userID string) (int64, error) {
	var size int64
	if err := r.db.QueryRow(
		ctx,
		"select sum(size) from files where creator_id = $1",
		userID,
	).Scan(&size); err != nil {
		return 0, err
	}

	return size, nil
}

func (r *userSpaceRepo) UpdateLevel(ctx context.Context, userID string, newLevel uint8) error {
	_, err := r.db.Exec(ctx, "update users_spaces set level = $1 where user_id = $2", newLevel, userID)
	return err
}
