package dao

import (
	"fmt"
	"log/slog"
)

// DAOFactory interface for creating DAO instances
type DAOFactory interface {
	CreateDAO() (DAO, error)
	CreateSettingsDAO() (SettingsDAO, error)
	Close() error
}

// SQLiteDAOFactory implements DAOFactory for SQLite
type SQLiteDAOFactory struct {
	config *Config
}

// PostgresDAOFactory implements DAOFactory for PostgreSQL
type PostgresDAOFactory struct {
	config *Config
}

// NewDAOFactory creates the appropriate DAO factory based on configuration
func NewDAOFactory(config *Config) (DAOFactory, error) {
	if config == nil {
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
		return &PostgresDAOFactory{config: config}, nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Database.Type)
	}
}

// SQLiteDAOFactory implementation

func (f *SQLiteDAOFactory) CreateDAO() (DAO, error) {
	return NewSQLiteDAO(f.config.Database.SQLite.URL)
}

func (f *SQLiteDAOFactory) CreateSettingsDAO() (SettingsDAO, error) {
	return NewSQLiteSettingsDAO(f.config.Database.SQLite.URL), nil
}

func (f *SQLiteDAOFactory) Close() error {
	// SQLite connections are closed by individual DAOs
	return nil
}

// PostgresDAOFactory implementation

func (f *PostgresDAOFactory) CreateDAO() (DAO, error) {
	return NewPostgresDAO(&f.config.Database.Postgres)
}

func (f *PostgresDAOFactory) CreateSettingsDAO() (SettingsDAO, error) {
	return NewPostgresSettingsDAO(&f.config.Database.Postgres)
}

func (f *PostgresDAOFactory) Close() error {
	// PostgreSQL connections are closed by individual DAOs
	return nil
}
