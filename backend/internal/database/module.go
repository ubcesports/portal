package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ubcesports/memberships/internal/database/db"
	"go.uber.org/fx"
)

var Module = fx.Module("database",
	fx.Provide(providePool),
	fx.Provide(provideDatabase),
	fx.Provide(provideStdlibDB),
)

func providePool(lc fx.Lifecycle) (*pgxpool.Pool, error) {
	pool, err := ConnectDB()
	if err != nil {
		return nil, fmt.Errorf("connect application database pool: %w", err)
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			slog.InfoContext(ctx, "closing database connection pool")
			pool.Close()
			return nil
		},
	})

	slog.Info("database connection established")
	return pool, nil
}

func provideDatabase(pool *pgxpool.Pool) *db.Queries {
	return db.New(pool)
}

func provideStdlibDB(lc fx.Lifecycle) (*sql.DB, error) {
	databaseURL, err := connectionString()
	if err != nil {
		return nil, err
	}

	sqlDB, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open authentication database connection: %w", err)
	}
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			slog.InfoContext(ctx, "closing authentication database connection")
			if err := sqlDB.Close(); err != nil {
				return fmt.Errorf("close authentication database connection: %w", err)
			}
			return nil
		},
	})
	return sqlDB, nil
}
