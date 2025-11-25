package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type contextKey string

const txKey contextKey = "pgx_tx"

// WithTx добавляет pgx.Tx в контекст
func WithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey, tx)
}

// GetTx извлекает pgx.Tx из контекста
func GetTx(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey).(pgx.Tx)
	return tx, ok
}
