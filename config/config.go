package config

type GlobalConfig struct {
	ServerPort         int    `env:"SERVER_PORT,required"`
	PostgresConnection string `env:"POSTGRES_CONNECTION,required"`
}
