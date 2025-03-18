package db

import (
	"context"
	"fmt"
	"github.com/IndexStorm/common-go/config"
	"github.com/IndexStorm/common-go/telemetry"
	"github.com/jackc/pgx/v5/pgxpool"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"time"
)

func NewPgxPoolWithOtel(
	ctx context.Context,
	dbConfig config.Database,
) (*pgxpool.Pool, error) {
	database, err := NewPgxConnection(ctx,
		fmt.Sprintf(
			"postgresql://%s:%s@%s/%s?sslmode=%s",
			dbConfig.Username,
			dbConfig.Password,
			dbConfig.Host,
			dbConfig.Database,
			dbConfig.SSLMode,
		),
		telemetry.NewPgxTracer(
			semconv.DBSystemNamePostgreSQL,
			semconv.DBNamespace(dbConfig.Host+"/"+dbConfig.Database),
			// semconv.ServerAddress(config.Host),
			// semconv.ServerPort(int(config.Port)),
			// semconv.UserName(config.User),
			// semconv.DBNamespace(config.Database),
		),
		nil,
		time.Second*10)
	if err != nil {
		return nil, fmt.Errorf("open connection: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	if err = database.Ping(ctx); err != nil {
		database.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return database, nil
}

func NewPgxPool(
	ctx context.Context,
	dbConfig config.Database,
) (*pgxpool.Pool, error) {
	database, err := NewPgxConnection(ctx,
		fmt.Sprintf(
			"postgresql://%s:%s@%s/%s?sslmode=%s",
			dbConfig.Username,
			dbConfig.Password,
			dbConfig.Host,
			dbConfig.Database,
			dbConfig.SSLMode,
		),
		nil,
		nil,
		time.Second*10)
	if err != nil {
		return nil, fmt.Errorf("open connection: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	if err = database.Ping(ctx); err != nil {
		database.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return database, nil
}
