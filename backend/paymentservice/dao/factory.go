package dao

import (
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
)

// DAOFactory interface for creating DAO instances
type DAOFactory interface {
	CreateDAO() (DAO, error)
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
		slog.Info("Creating SQLite DAO factory", "url", config.Database.SQLite.URL)
		db, err := sqlx.Open("sqlite3", config.Database.SQLite.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to open SQLite connection: %w", err)
		}
		if err := db.Ping(); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to ping SQLite database: %w", err)
		}
		return &SQLiteDAOFactory{config: config, db: db}, nil
	case DatabaseTypePostgres:
		slog.Info("Creating PostgreSQL DAO factory",
			"host", config.Database.Postgres.Host,
			"port", config.Database.Postgres.Port,
			"database", config.Database.Postgres.Database)

		// Create shared connection pool
		dsn := config.Database.Postgres.GetPostgresDSN()
		db, err := sqlx.Open("postgres", dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
		}

		// Configure connection pool
		if v := config.Database.Postgres.Pool.MaxOpenConnections; v > 0 {
			db.SetMaxOpenConns(v)
		} else if v < 0 {
			return nil, fmt.Errorf("invalid postgres pool: MaxOpenConnections cannot be negative: %d", v)
		}
		if v := config.Database.Postgres.Pool.MaxIdleConnections; v >= 0 {
			db.SetMaxIdleConns(v)
		} else {
			return nil, fmt.Errorf("invalid postgres pool: MaxIdleConnections cannot be negative: %d", v)
		}
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
		return nil, fmt.Errorf("unsupported database type: %s", config.Database.Type)
	}
}

// SQLiteDAOFactory implementation

func (f *SQLiteDAOFactory) CreateDAO() (DAO, error) {
	return NewSQLiteDAO(f.db)
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

func (f *PostgresDAOFactory) Close() error {
	if f.db != nil {
		return f.db.Close()
	}
	return nil
}
