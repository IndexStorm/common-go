package main

import (
	"context"
	"github.com/IndexStorm/common-go/log"
	"github.com/IndexStorm/common-go/migration"
	"github.com/caarlos0/env/v11"
	"github.com/rs/zerolog"
)

type MigratorType string

const (
	MigratorTypePostgres MigratorType = "postgres"
)

type appConfig struct {
	MigrationConfig migration.Config
	MigratorType    MigratorType `env:"MIGRATOR_TYPE,notEmpty"`
}

func main() {
	log.SetupCallerRootRewrite()
	logger := log.NewZerologWithLevel(zerolog.DebugLevel)
	config, err := env.ParseAs[appConfig]()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to parse config")
	}
	var migrator migration.Migrator
	switch config.MigratorType {
	case MigratorTypePostgres:
		migrator = migration.NewPostgresMigrator(logger)
	default:
		logger.Fatal().Str("migrator_type", string(config.MigratorType)).Msg("unsupported migrator")
		return
	}
	ctx := context.Background()
	if err = migrator.Migrate(ctx, config.MigrationConfig); err != nil {
		logger.Fatal().Err(err).Msg("Failed to migrate")
	}
}
