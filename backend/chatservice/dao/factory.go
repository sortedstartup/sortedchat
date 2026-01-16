package dao

import (
	"fmt"
	"log/slog"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"github.com/jmoiron/sqlx"
)

// DAOFactory interface for creating DAO instances
type DAOFactory interface {
	CreateDAO() (DAO, error)
	CreateSettingsDAO() (SettingsDAO, error)
	CreateAgentDAO() (AgentDAO, error)
	Close() error
}

// SQLiteDAOFactory implements DAOFactory for SQLite
type SQLiteDAOFactory struct {
	config *Config
	db     *sqlx.DB // Shared connection pool
}

// PostgresDAOFactory implements DAOFactory for PostgreSQL
type PostgresDAOFactory struct {
	config *Config
	db     *sqlx.DB // Shared connection pool
}

// NewDAOFactory creates the appropriate DAO factory based on configuration
func NewDAOFactory(config *Config) (DAOFactory, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	switch config.Database.Type {
	case DatabaseTypeSQLite:
		slog.Debug("Creating SQLite DAO factory", "url", config.Database.SQLite.URL)

		sqlite_vec.Auto()

		// Create shared connection pool for SQLite
		db, err := sqlx.Open("sqlite3", config.Database.SQLite.URL)
		if err != nil {
			slog.Error("Failed to open SQLite connection", "error", err)
			return nil, fmt.Errorf("failed to open SQLite connection: %w", err)
		}

		// Enable WAL mode and busy timeout for concurrency
		if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
			slog.Error("Failed to enable WAL mode", "error", err)
			return nil, err
		}
		if _, err := db.Exec("PRAGMA busy_timeout=30000;"); err != nil {
			slog.Error("Failed to set busy_timeout", "error", err)
			return nil, err
		}

		return &SQLiteDAOFactory{
			config: config,
			db:     db,
		}, nil
	case DatabaseTypePostgres:
		slog.Debug("Creating PostgreSQL DAO factory",
			"host", config.Database.Postgres.Host,
			"port", config.Database.Postgres.Port,
			"database", config.Database.Postgres.Database)

		// Create shared connection pool
		dsn := config.Database.Postgres.GetPostgresDSN()
		db, err := sqlx.Open("postgres", dsn)
		if err != nil {
			slog.Error("Failed to open PostgreSQL connection", "error", err)
			return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
		}

		// Configure connection pool
		db.SetMaxOpenConns(config.Database.Postgres.Pool.MaxOpenConnections)
		db.SetMaxIdleConns(config.Database.Postgres.Pool.MaxIdleConnections)
		db.SetConnMaxLifetime(config.Database.Postgres.Pool.ConnectionMaxLifetime)

		// Test the connection
		if err := db.Ping(); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to ping PostgreSQL database: %w", err)
		}

		slog.Info("PostgreSQL connection pool created successfully",
			"host", config.Database.Postgres.Host,
			"port", config.Database.Postgres.Port,
			"database", config.Database.Postgres.Database,
			"max_open_conns", config.Database.Postgres.Pool.MaxOpenConnections)

		return &PostgresDAOFactory{
			config: config,
			db:     db,
		}, nil
	default:
		slog.Error("Unsupported database type", "database_type", config.Database.Type)
		return nil, fmt.Errorf("unsupported database type: %s", config.Database.Type)
	}
}

// SQLiteDAOFactory implementation

func (f *SQLiteDAOFactory) CreateDAO() (DAO, error) {
	return NewSQLiteDAOWithDB(f.db), nil
}

func (f *SQLiteDAOFactory) CreateSettingsDAO() (SettingsDAO, error) {
	return NewSQLiteSettingsDAOWithDB(f.db), nil
}

func (f *SQLiteDAOFactory) CreateAgentDAO() (AgentDAO, error) {
	return NewSQLiteAgentsDAOWithDB(f.db), nil
}

func (f *SQLiteDAOFactory) Close() error {
	if f.db != nil {
		return f.db.Close()
	}
	return nil
}

// PostgresDAOFactory implementation

func (f *PostgresDAOFactory) CreateDAO() (DAO, error) {
	return NewPostgresDAOWithDB(f.db)
}

func (f *PostgresDAOFactory) CreateSettingsDAO() (SettingsDAO, error) {
	return NewPostgresSettingsDAOWithDB(f.db)
}

func (f *PostgresDAOFactory) CreateAgentDAO() (AgentDAO, error) {
	return NewPostgresAgentsDAO(f.db), nil
}

func (f *PostgresDAOFactory) Close() error {
	if f.db != nil {
		return f.db.Close()
	}
	return nil
}
