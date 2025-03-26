package config

type Server struct {
	ListenAddress string `env:"LISTEN_ADDRESS,notEmpty"`
}
