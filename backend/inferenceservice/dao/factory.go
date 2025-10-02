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
}

// PostgresDAOFactory implements DAOFactory for PostgreSQL
type PostgresDAOFactory struct {
	config *Config
	db     *sqlx.DB // Shared connection pool
}

// NewDAOFactory creates the appropriate DAO factory based on configuration
func NewDAOFactory(config *Config) (DAOFactory, error) {
	slog.Info("inferenceservice:dao:NewDAOFactory")
	if config == nil {
		slog.Error("inferenceservice:dao:NewDAOFactory", "error", "config cannot be nil")
		return nil, fmt.Errorf("config cannot be nil")
	}

	switch config.Database.Type {
	case DatabaseTypeSQLite:
		slog.Info("Creating SQLite DAO factory", "url", config.Database.SQLite.URL)
		return &SQLiteDAOFactory{config: config}, nil
	case DatabaseTypePostgres:
		slog.Info("Creating PostgreSQL DAO factory",
			"host", config.Database.Postgres.Host,
			"port", config.Database.Postgres.Port,
			"database", config.Database.Postgres.Database)

		// Create shared connection pool
		dsn := config.Database.Postgres.GetPostgresDSN()
		db, err := sqlx.Open("postgres", dsn)
		if err != nil {
			slog.Error("inferenceservice:dao:NewDAOFactory", "error", "failed to open PostgreSQL connection", "error", err)
			return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
		}

		// Configure connection pool
		db.SetMaxOpenConns(config.Database.Postgres.Pool.MaxOpenConnections)
		db.SetMaxIdleConns(config.Database.Postgres.Pool.MaxIdleConnections)
		db.SetConnMaxLifetime(config.Database.Postgres.Pool.ConnectionMaxLifetime)

		// Test the connection
		if err := db.Ping(); err != nil {
			db.Close()
			slog.Error("inferenceservice:dao:NewDAOFactory", "error", "failed to ping PostgreSQL database", "error", err)
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
		slog.Error("inferenceservice:dao:NewDAOFactory", "error", "unsupported database type", "database_type", config.Database.Type)
		return nil, fmt.Errorf("unsupported database type: %s", config.Database.Type)
	}
}

// SQLiteDAOFactory implementation

func (f *SQLiteDAOFactory) CreateDAO() (DAO, error) {
	slog.Info("inferenceservice:dao:SQLiteDAOFactory:CreateDAO")
	return NewSQLiteDAO(f.config.Database.SQLite.URL)
}

func (f *SQLiteDAOFactory) Close() error {
	// SQLite connections are closed by individual DAOs
	return nil
}

// PostgresDAOFactory implementation
func (f *PostgresDAOFactory) CreateDAO() (DAO, error) {
	slog.Info("inferenceservice:dao:PostgresDAOFactory:CreateDAO")
	return NewPostgresDAOWithDB(f.db)
}

func (f *PostgresDAOFactory) Close() error {
	if f.db != nil {
		return f.db.Close()
	}
	return nil
}
