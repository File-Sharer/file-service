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

func (r *userSpaceRepo) Find(ctx context.Context, userID string) (int64, error) {
	var space int64
	if err := r.db.QueryRow(
		ctx,
		"select space from users_spaces where user_id = $1",
		userID,
	).Scan(&space); err != nil {
		return 0, err
	}

	return space, nil
}
