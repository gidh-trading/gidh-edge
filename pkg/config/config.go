package config

import (
	"gidh-edge/pkg/env"
	"os"
)

type Config struct {
	API APIConfig
	DB  DBConfig
	App AppConfig
}

type APIConfig struct {
	Port      string
	EngineURL string
}
type DBConfig struct{ ConnString string }
type AppConfig struct{ LogLevel string }

func Load() *Config {
	_ = env.Load(".env") // Load .env if it exists

	return &Config{
		API: APIConfig{
			Port:      getEnv("API_PORT", "8080"),
			EngineURL: getEnv("ENGINE_URL", "http://localhost:8081"), // Added Engine URL
		},
		DB: DBConfig{
			ConnString: getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/gidh?sslmode=disable"),
		},
		App: AppConfig{
			LogLevel: getEnv("LOG_LEVEL", "info"),
		},
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
