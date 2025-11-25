package postgres

import (
	"avito_intern/internal/config"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func NewPostgresDB(ctx context.Context, cfg *config.Config) (*pgx.Conn, error) {
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Host, cfg.Postgres.Port,
		cfg.Postgres.Name, cfg.Postgres.SSLMode)

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("[NewPostgresDB]: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("[NewPostgresDB]: %w", err)
	}

	return conn, nil
}

func ClosePostgresDB(ctx context.Context, conn *pgx.Conn) error {
	err := conn.Close(ctx)
	if err != nil {
		return fmt.Errorf("[ClosePostgresDB]: %w", err)
	}

	return nil
}
