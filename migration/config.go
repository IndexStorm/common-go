package migration

import "github.com/IndexStorm/common-go/config"

type Config struct {
	Database              config.Database
	ForceVersion          int    `env:"FORCE_VERSION"`
	SqlSchemaDir          string `env:"SQL_SCHEMA_DIR,notEmpty"`
	MigrationsTableQuoted string `env:"MIGRATION_TABLE_QUOTED"`
}
