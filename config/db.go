package config

type Database struct {
	Host     string `env:"HOST,notEmpty,unset"`
	Username string `env:"USERNAME,notEmpty,unset"`
	Password string `env:"PASSWORD,notEmpty,unset"`
	Database string `env:"DATABASE,notEmpty,unset"`
	SSLMode  string `env:"SSL_MODE" envDefault:"require"`
}
