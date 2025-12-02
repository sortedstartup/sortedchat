package dao

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
)

// DatabaseType represents the type of database being used
type DatabaseType string

const (
	DatabaseTypeSQLite   DatabaseType = "sqlite"
	DatabaseTypePostgres DatabaseType = "postgres"
)

// DatabaseConfig holds the configuration for database connections
type DatabaseConfig struct {
	Type     DatabaseType   `koanf:"type" json:"type"`
	SQLite   SQLiteConfig   `koanf:"sqlite" json:"sqlite"`
	Postgres PostgresConfig `koanf:"postgres" json:"postgres"`
}

// SQLiteConfig holds SQLite-specific configuration
type SQLiteConfig struct {
	URL string `koanf:"url" json:"url"`
}

// PostgresConfig holds PostgreSQL-specific configuration
type PostgresConfig struct {
	Host     string     `koanf:"host" json:"host"`
	Port     int        `koanf:"port" json:"port"`
	Database string     `koanf:"database" json:"database"`
	Username string     `koanf:"username" json:"username"`
	Password string     `koanf:"password" json:"password"`
	SSLMode  string     `koanf:"ssl_mode" json:"ssl_mode"`
	Pool     PoolConfig `koanf:"pool" json:"pool"`
}

// PoolConfig holds database connection pool configuration
type PoolConfig struct {
	MaxOpenConnections    int           `koanf:"max_open_connections" json:"max_open_connections"`
	MaxIdleConnections    int           `koanf:"max_idle_connections" json:"max_idle_connections"`
	ConnectionMaxLifetime time.Duration `koanf:"connection_max_lifetime" json:"connection_max_lifetime"`
}

// Config represents the complete application configuration
type Config struct {
	Database DatabaseConfig `koanf:"database" json:"database"`
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() (*Config, error) {
	slog.Debug("authservice:dao:LoadConfig")
	k := koanf.New(".")

	// Default configuration
	defaultConfig := &Config{
		Database: DatabaseConfig{
			Type: DatabaseTypeSQLite,
			SQLite: SQLiteConfig{
				URL: "db.sqlite",
			},
			Postgres: PostgresConfig{
				Host:     "localhost",
				Port:     5432,
				Database: "sortedchat",
				Username: "postgres",
				Password: "",
				SSLMode:  "disable",
				Pool: PoolConfig{
					MaxOpenConnections:    25,
					MaxIdleConnections:    5,
					ConnectionMaxLifetime: 5 * time.Minute,
				},
			},
		},
	}

	// Load defaults first
	if err := k.Load(structs.Provider(defaultConfig, "koanf"), nil); err != nil {
		return nil, fmt.Errorf("failed to load default config: %w", err)
	}

	// Override with environment variables if present
	loadFromEnv(k)

	// TODO: Load from config.yaml file if it exists (future enhancement)

	var config Config
	if err := k.Unmarshal("", &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	slog.Info("Configuration loaded",
		"database_type", config.Database.Type,
		"postgres_host", config.Database.Postgres.Host,
		"postgres_port", config.Database.Postgres.Port,
	)
slog.Info("Configuration loaded from environment variables for authservice", "database_type", config.Database.Type, "postgres_host", config.Database.Postgres.Host, "postgres_port", config.Database.Postgres.Port)
	return &config, nil
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv(k *koanf.Koanf) {
	// Database type
	if dbType := os.Getenv("DB_TYPE"); dbType != "" {
		k.Set("database.type", dbType)
	}

	// PostgreSQL configuration
	if host := os.Getenv("POSTGRES_HOST"); host != "" {
		k.Set("database.postgres.host", host)
	}
	if port := os.Getenv("POSTGRES_PORT"); port != "" {
		k.Set("database.postgres.port", port)
	}
	if database := os.Getenv("POSTGRES_DATABASE"); database != "" {
		k.Set("database.postgres.database", database)
	}
	if username := os.Getenv("POSTGRES_USERNAME"); username != "" {
		k.Set("database.postgres.username", username)
	}
	if password := os.Getenv("POSTGRES_PASSWORD"); password != "" {
		k.Set("database.postgres.password", password)
	}
	if sslMode := os.Getenv("POSTGRES_SSL_MODE"); sslMode != "" {
		k.Set("database.postgres.ssl_mode", sslMode)
	}
	if maxOpen := os.Getenv("POSTGRES_MAX_OPEN_CONNECTIONS"); maxOpen != "" {
		k.Set("database.postgres.pool.max_open_connections", maxOpen)
	}
	if maxIdle := os.Getenv("POSTGRES_MAX_IDLE_CONNECTIONS"); maxIdle != "" {
		k.Set("database.postgres.pool.max_idle_connections", maxIdle)
	}
	if maxLifetime := os.Getenv("POSTGRES_CONNECTION_MAX_LIFETIME"); maxLifetime != "" {
		if duration, err := time.ParseDuration(maxLifetime); err == nil {
			k.Set("database.postgres.pool.connection_max_lifetime", duration)
		} else {
			slog.Warn("Invalid format for POSTGRES_CONNECTION_MAX_LIFETIME, using default", "value", maxLifetime, "error", err)
		}
	}

	// SQLite configuration
	if sqliteURL := os.Getenv("SQLITE_URL"); sqliteURL != "" {
		k.Set("database.sqlite.url", sqliteURL)
	}
}

// validateConfig validates the loaded configuration
func validateConfig(config *Config) error {
	switch config.Database.Type {
	case DatabaseTypeSQLite:
		if config.Database.SQLite.URL == "" {
			return fmt.Errorf("sqlite URL is required when database type is sqlite")
		}
	case DatabaseTypePostgres:
		pg := config.Database.Postgres
		if pg.Host == "" {
			return fmt.Errorf("postgres host is required when database type is postgres")
		}
		if pg.Database == "" {
			return fmt.Errorf("postgres database name is required when database type is postgres")
		}
		if pg.Username == "" {
			return fmt.Errorf("postgres username is required when database type is postgres")
		}
		if pg.Port <= 0 || pg.Port > 65535 {
			return fmt.Errorf("postgres port must be between 1 and 65535")
		}
		if !isValidSSLMode(pg.SSLMode) {
			return fmt.Errorf("invalid postgres ssl_mode: %s", pg.SSLMode)
		}
	default:
		return fmt.Errorf("unsupported database type: %s", config.Database.Type)
	}

	return nil
}

// isValidSSLMode checks if the SSL mode is valid for PostgreSQL
func isValidSSLMode(mode string) bool {
	validModes := []string{"disable", "require", "verify-ca", "verify-full"}
	for _, valid := range validModes {
		if strings.EqualFold(mode, valid) {
			return true
		}
	}
	return false
}

// GetPostgresDSN builds a PostgreSQL connection string from the configuration
func (c *PostgresConfig) GetPostgresDSN() string {
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		c.Host, c.Port, c.Database, c.Username, c.Password, c.SSLMode)
}
