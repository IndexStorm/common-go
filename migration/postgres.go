package migration

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/rs/zerolog"
)

type postgresMigrator struct {
	logger zerolog.Logger
}

func NewPostgresMigrator(logger zerolog.Logger) Migrator {
	return &postgresMigrator{
		logger: logger,
	}
}

func (p *postgresMigrator) Migrate(ctx context.Context, config Config) error {
	err := p.runMigrations(config)
	if errors.Is(err, migrate.ErrNoChange) {
		p.logger.Info().Msg("no changes detected")
		return nil
	}
	return err
}

func (p *postgresMigrator) runMigrations(config Config) error {
	connStr := fmt.Sprintf(
		"postgresql://%s:%s@%s/%s?sslmode=%s",
		config.Database.Username,
		config.Database.Password,
		config.Database.Host,
		config.Database.Database,
		config.Database.SSLMode,
	)
	if config.SchemaName != "" {
		connStr += fmt.Sprintf("&currentSchema=%s", config.SchemaName)
	}
	if config.MigrationsTableQuoted != "" {
		connStr += fmt.Sprintf("&x-migrations-table=%s&x-migrations-table-quoted=1", config.MigrationsTableQuoted)
	}
	migrator, err := migrate.New(config.SqlSchemaDir, connStr)
	if err != nil {
		return fmt.Errorf("init migrate: %w", err)
	}
	if config.ForceVersion >= -1 {
		err = migrator.Force(config.ForceVersion)
		if err != nil {
			return fmt.Errorf("force version %d: %w", config.ForceVersion, err)
		}
		err = migrator.Down()
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("run down migrations: %w", err)
		}
	}
	if err = migrator.Up(); err != nil {
		return fmt.Errorf("run up migrations: %w", err)
	}
	return nil
}
