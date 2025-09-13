package database

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"context"
)

func InitDb(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	conn, err := pgxpool.New(ctx, dsn)
	err = conn.Ping(ctx)
	return conn, err
}