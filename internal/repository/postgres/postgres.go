package postgres

import (
	"context"
	"fmt"

	"github.com/File-Sharer/file-service/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPgPool(ctx context.Context, cfg *config.DBConfig) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", cfg.Host, cfg.Username, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode)
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	return db, nil
}
