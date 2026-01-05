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
	slog.Info("inferenceservice:dao_postgres:NewPostgresDAO")
	dsn := config.GetPostgresDSN()

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		slog.Error("inferenceservice:dao_postgres:NewPostgresDAO", "message", "failed to open PostgreSQL connection", "error", err)
		return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.Pool.MaxOpenConnections)
	db.SetMaxIdleConns(config.Pool.MaxIdleConnections)
	db.SetConnMaxLifetime(config.Pool.ConnectionMaxLifetime)

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		slog.Error("inferenceservice:dao_postgres:NewPostgresDAO", "message", "failed to ping PostgreSQL database", "error", err)
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
	slog.Info("inferenceservice:dao_postgres:NewPostgresDAOWithDB")
	return &PostgresDAO{db: db}, nil
}

func (d *PostgresDAO) Infer(dummy string) error {
	return nil
}

func (d *PostgresDAO) GetModelByName(modelName string) (*ModelMetadata, error) {
	slog.Info("inferenceservice:dao_postgres:GetModelByName")
	query := `SELECT id, name, url, provider, input_token_cost, output_token_cost, progress, is_downloaded, is_downloadable, status, filestore_id FROM shared_models_metadata WHERE name = $1`

	var model ModelMetadata
	err := d.db.Get(&model, query, modelName)
	if err != nil {
		slog.Error("inferenceservice:dao_postgres:GetModelByName", "message", "failed to get model by name", "error", err)
		return nil, fmt.Errorf("failed to get model by name")
	}

	return &model, nil
}

func (d *PostgresDAO) GetAllModels() ([]*ModelMetadata, error) {
	slog.Info("inferenceservice:dao_postgres:GetAllModels")
	query := `SELECT id, name, url, provider, input_token_cost, output_token_cost, progress, is_downloaded, is_downloadable, status, filestore_id, cached_token_cost, is_enabled, is_embedding_model FROM shared_models_metadata ORDER BY name`

	var models []*ModelMetadata
	err := d.db.Select(&models, query)
	if err != nil {
		slog.Error("inferenceservice:dao_postgres:GetAllModels", "message", "failed to get all models", "error", err)
		return nil, fmt.Errorf("failed to get all models")
	}

	return models, nil
}

func (d *PostgresDAO) UpdateModelProgress(id string, progress *DownloadProgress) error {
	slog.Info("inferenceservice:dao_postgres:UpdateModelProgress")
	isDownloaded := progress.Status == StatusCompleted
	query := `UPDATE shared_models_metadata SET progress = $1, is_downloaded = $2, status = $3 WHERE id = $4`

	progressJSON, err := progress.ToJSON()
	if err != nil {
		slog.Error("inferenceservice:dao_postgres:UpdateModelProgress", "message", "failed to convert progress to JSON", "error", err)
		return fmt.Errorf("failed to convert progress to JSON")
	}

	_, err = d.db.Exec(query, progressJSON, isDownloaded, progress.Status, id)
	if err != nil {
		slog.Error("inferenceservice:dao_postgres:UpdateModelProgress", "message", "failed to update model progress", "error", err)
		return fmt.Errorf("failed to update model progress")
	}
	return nil
}

func (d *PostgresDAO) UpdateModelFileStoreID(id string, filestoreID string) error {
	slog.Info("inferenceservice:dao_postgres:UpdateModelFileStoreID")
	query := `UPDATE shared_models_metadata SET filestore_id = $1 WHERE id = $2`
	_, err := d.db.Exec(query, filestoreID, id)
	if err != nil {
		slog.Error("inferenceservice:dao_postgres:UpdateModelFileStoreID", "message", "failed to update model file store ID", "error", err)
		return fmt.Errorf("failed to update model file store ID")
	}
	return nil
}

func (d *PostgresDAO) ResetModelToInitialState(id string) error {
	slog.Info("inferenceservice:dao_postgres:ResetModelToInitialState")
	query := `UPDATE shared_models_metadata SET progress = '', status = 0, is_downloaded = false, filestore_id = NULL WHERE id = $1`
	_, err := d.db.Exec(query, id)
	if err != nil {
		slog.Error("inferenceservice:dao_postgres:ResetModelToInitialState", "message", "failed to reset model to initial state", "error", err)
		return fmt.Errorf("failed to reset model to initial state")
	}
	return nil
}
