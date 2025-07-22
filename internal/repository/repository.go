package repository

import (
	"github.com/File-Sharer/file-service/internal/repository/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	Postgres *postgres.PostgresRepository
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository {
		Postgres: postgres.NewPostgresRepo(db),
	}
}
