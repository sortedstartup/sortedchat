package dao

import (
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// PostgresDAO implements the DAO interface using PostgreSQL and sqlx
type PostgresDAO struct {
	db *sqlx.DB
}

// NewPostgresDAO creates a new PostgreSQL DAO instance
func NewPostgresDAO(config *PostgresConfig) (*PostgresDAO, error) {
	dsn := config.GetPostgresDSN()

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.Pool.MaxOpenConnections)
	db.SetMaxIdleConns(config.Pool.MaxIdleConnections)
	db.SetConnMaxLifetime(config.Pool.ConnectionMaxLifetime)

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping PostgreSQL database: %w", err)
	}

	slog.Info("PostgreSQL DAO created successfully",
		"host", config.Host,
		"port", config.Port,
		"database", config.Database,
		"max_open_conns", config.Pool.MaxOpenConnections)

	return &PostgresDAO{db: db}, nil
}

// NewPostgresDAOWithDB creates a new PostgreSQL DAO instance using a shared database connection
func NewPostgresDAOWithDB(db *sqlx.DB) (*PostgresDAO, error) {
	return &PostgresDAO{db: db}, nil
}

func (d *PostgresDAO) Infer(dummy string) error {
	return nil
}

func (d *PostgresDAO) CreateAudioChat(userID string, modelName string, startTime string, endTime string) (string, error) {
	id := uuid.New().String()
	_, err := d.db.Exec("INSERT INTO realtimeservice_audio_chat (id, model_name, user_id, start_time, end_time) VALUES ($1, $2, $3, $4, $5)", id, modelName, userID, startTime, endTime)

	if err != nil {
		slog.Error("Failed to create audio chat postgres", "error", err)
		return "", err
	}
	return id, nil
}

func (d *PostgresDAO) UpdateAudioChat(userID string, id string, endTime string) error {
	_, err := d.db.Exec("UPDATE realtimeservice_audio_chat SET end_time = $1 WHERE id = $2 AND user_id = $3", endTime, id, userID)
	if err != nil {
		slog.Error("Failed to update audio chat postgres", "error", err)
		return err
	}
	return err
}
