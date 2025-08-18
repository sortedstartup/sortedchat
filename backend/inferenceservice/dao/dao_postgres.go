package dao

import (
	"fmt"
	"log/slog"

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

func (d *PostgresDAO) DownloadModel(userID string, modelName string, url string) error {
	return nil
}

func (d *PostgresDAO) GetModelByName(modelName string) (*ModelMetadata, error) {
	query := `SELECT id, name, url, provider, input_token_cost, output_token_cost, progress, is_downloaded, is_downloadable, status FROM inference_model_metadata WHERE name = $1`

	var model ModelMetadata
	err := d.db.Get(&model, query, modelName)
	if err != nil {
		return nil, err
	}

	return &model, nil
}

func (d *PostgresDAO) GetAllModels() ([]*ModelMetadata, error) {
	query := `SELECT id, name, url, provider, input_token_cost, output_token_cost, progress, is_downloaded, is_downloadable, status FROM inference_model_metadata ORDER BY name`

	var models []*ModelMetadata
	err := d.db.Select(&models, query)
	if err != nil {
		return nil, err
	}

	return models, nil
}

func (d *PostgresDAO) UpdateModelProgress(id string, progress *DownloadProgress) error {
	isDownloaded := progress.Status == StatusCompleted
	query := `UPDATE inference_model_metadata SET progress = $1, is_downloaded = $2, status = $3 WHERE id = $4`

	_, err := d.db.Exec(query, progress.ToJSON(), isDownloaded, progress.Status, id)
	return err
}
