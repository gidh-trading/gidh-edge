package config

import (
	"gidh-edge/pkg/env"
	"os"
	"strings"
)

type Config struct {
	API  APIConfig
	DB   DBConfig
	App  AppConfig
	Kite KiteConfig
}
type KiteConfig struct {
	APIKey      string
	AccessToken string
}

type APIConfig struct {
	Port      string
	EngineURL string
}

type DBConfig struct {
	ConnString string
}

type AppConfig struct {
	LogLevel          string
	Mode              string
	BacktestBackupDir string
	BacktestDataDir   string
}

func Load() *Config {
	_ = env.Load(".env") // Load .env if it exists

	// Determine the mode, defaulting to "live" if not set
	mode := strings.ToLower(getEnv("MODE", "backtest"))

	var port, engineURL, dbURL string

	if mode == "backtest" {
		port = getEnv("BACKTEST_API_PORT", "8081")
		engineURL = getEnv("BACKTEST_ENGINE_URL", "http://localhost:9091")
		dbURL = getEnv("BACKTEST_DATABASE_URL", "postgres://postgres:password@localhost:5432/gidh_backtest?sslmode=disable")
	} else {
		// Default to live configuration
		// We use standard variables as fallbacks here for backwards compatibility
		port = getEnv("LIVE_API_PORT", "8080")
		engineURL = getEnv("LIVE_ENGINE_URL", "http://localhost:9090")
		dbURL = getEnv("LIVE_DATABASE_URL", "")
	}

	return &Config{
		API: APIConfig{
			Port:      port,
			EngineURL: engineURL,
		},
		DB: DBConfig{
			ConnString: dbURL,
		},
		App: AppConfig{
			LogLevel:          getEnv("LOG_LEVEL", "info"),
			Mode:              mode,
			BacktestBackupDir: getEnv("BACKTEST_BACKUP_DIR", ""),
			BacktestDataDir:   getEnv("BACKTEST_DATA_DIR", ""),
		},
		Kite: KiteConfig{
			APIKey:      getEnv("KITE_API_KEY", ""),
			AccessToken: getEnv("KITE_ACCESS_TOKEN", ""),
		},
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
