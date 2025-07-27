package config

import (
	"os"
	"strconv"
)

type Config struct {
	Db         string
	DbPassword string
	DbUsername string
	DbPort     string
	DbHost     string
	Port       int
}

// New returns a config object from the env and a non-nil error if validation errors occurred
func New() (*Config, error) {
	// database configs
	db := getEnv("DB_DATABASE", "chat")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbUsername := getEnv("DB_USERNAME", "postgres")
	dbPort := getEnv("DB_PORT", "5432")
	dbHost := getEnv("DB_HOST", "localhost")

	// server configs
	port := getEnvInt("PORT", 8000)

	return &Config{
		Db:         db,
		DbPassword: dbPassword,
		DbUsername: dbUsername,
		DbPort:     dbPort,
		DbHost:     dbHost,
		Port:       port,
	}, nil
}

func getEnv(key string, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}

func getEnvInt(key string, fallback int) int {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	intVal, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return fallback
	}
	return int(intVal)
}
