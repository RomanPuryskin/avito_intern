package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Postgres postgresConfig
	Server   serverConfig
	Logger   loggerConfig
}

type postgresConfig struct {
	Host     string `env:"DB_HOST" envDefault:"localhost"`
	Port     string `env:"DB_PORT" envDefault:"5432"`
	User     string `env:"DB_USER,required"`
	Password string `env:"DB_PASSWORD,required"`
	Name     string `env:"DB_NAME,required"`
	SSLMode  string `env:"DB_SSLMODE" envDefault:"disable"`
}

type serverConfig struct {
	ServerPort string `env:"SERVER_PORT" envDefault:":8080"`
}

type loggerConfig struct {
	LogLevel string `env:"LOG_LEVEL" envDefault:"INFO"`
}

func MustLoad() (*Config, error) {

	var cfg Config

	if err := cleanenv.ReadConfig(".env", &cfg); err == nil {
		return &cfg, nil
	}

	// Если .env файла нет, читаем из environment
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
