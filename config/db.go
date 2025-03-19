package config

type Database struct {
	Host     string `env:"DB_HOST,notEmpty,unset"`
	Username string `env:"DB_USERNAME,notEmpty,unset"`
	Password string `env:"DB_PASSWORD,notEmpty,unset"`
	Database string `env:"DB_DATABASE,notEmpty,unset"`
	SSLMode  string `env:"DB_SSL_MODE" envDefault:"require"`
}
