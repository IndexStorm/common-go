package config

type Service struct {
	ListenAddress string `env:"LISTEN_ADDRESS,notEmpty"`
}
