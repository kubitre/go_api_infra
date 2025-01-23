package pgqueue

import (
	"github.com/goreflect/gostructor"
)

type Config struct {
	PostgresDSN string `cf_env:"POSTGRES_DSN"`
}

func LoadConfig() (*Config, error) {
	result, err := gostructor.ConfigureSmart(&Config{})
	if err != nil {
		return nil, err
	}
	return result.(*Config), nil
}
