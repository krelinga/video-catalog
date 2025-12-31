package internal

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

var (
	ErrPanicEnvNotSet = errors.New("environment variable not set")
	ErrPanicEnvNotInt = errors.New("environment variable is not an integer")
)

const (
	EnvServerPort       = "VC_SERVER_PORT"
	EnvDatabaseHost     = "VC_DB_HOST"
	EnvDatabasePort     = "VC_DB_PORT"
	EnvDatabaseUser     = "VC_DB_USER"
	EnvDatabasePassword = "VC_DB_PASSWORD"
	EnvDatabaseName     = "VC_DB_NAME"
)

type Config struct {
	ServerPort int
	Database   *DatabaseConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
}

func mustGetenv(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		panic(fmt.Errorf("%w: %q", ErrPanicEnvNotSet, key))
	}
	return value
}

func mustGetenvAtoi(key string) int {
	valueStr := mustGetenv(key)
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		panic(fmt.Errorf("%w: %q", ErrPanicEnvNotInt, key))
	}
	return value
}

func NewConfigFromEnv() *Config {
	return &Config{
		ServerPort: mustGetenvAtoi(EnvServerPort),
		Database: &DatabaseConfig{
			Host:     mustGetenv(EnvDatabaseHost),
			Port:     mustGetenvAtoi(EnvDatabasePort),
			User:     mustGetenv(EnvDatabaseUser),
			Password: mustGetenv(EnvDatabasePassword),
			Name:     mustGetenv(EnvDatabaseName),
		},
	}
}
