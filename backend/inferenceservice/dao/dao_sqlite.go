package dao

import (

	// sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"

	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteDAO implements the DAO interface using SQLite and sqlx
type SQLiteDAO struct {
	db *sqlx.DB
}

// NewSQLiteDAO creates a new SQLite DAO instance
func NewSQLiteDAO(sqliteUrl string) (*SQLiteDAO, error) {
	slog.Debug("inferenceservice:dao_sqlite:NewSQLiteDAO")
	db, err := sqlx.Open("sqlite3", sqliteUrl)
	if err != nil {
		slog.Error("inferenceservice:dao_sqlite:NewSQLiteDAO", "message", "failed to open SQLite database", "error", err)
		return nil, err
	}

	// Set busy timeout to 10 seconds
	_, err = db.Exec("PRAGMA busy_timeout = 10000;")
	if err != nil {
		slog.Error("inferenceservice:dao_sqlite:NewSQLiteDAO", "message", "failed to set busy timeout", "error", err)
		return nil, err
	}

	// Enable WAL mode
	_, err = db.Exec("PRAGMA journal_mode = WAL;")
	if err != nil {
		slog.Error("inferenceservice:dao_sqlite:NewSQLiteDAO", "message", "failed to set WAL mode", "error", err)
		return nil, err
	}

	return &SQLiteDAO{db: db}, nil
}

func (d *SQLiteDAO) Infer(dummy string) error {
	return nil
}

func (d *SQLiteDAO) GetModelByName(modelName string) (*ModelMetadata, error) {
	query := `SELECT id, name, url, provider, input_token_cost, output_token_cost, progress, is_downloaded, is_downloadable, status, filestore_id FROM shared_models_metadata WHERE name = ?`

	var model ModelMetadata
	err := d.db.Get(&model, query, modelName)
	if err != nil {
		slog.Error("inferenceservice:dao_sqlite:GetModelByName", "message", "failed to get model by name", "error", err)
		return nil, fmt.Errorf("failed to get model by name")
	}

	return &model, nil
}

func (d *SQLiteDAO) GetAllModels() ([]*ModelMetadata, error) {
	query := `SELECT id, name, url, provider, input_token_cost, output_token_cost, progress, is_downloaded, is_downloadable, status, filestore_id, capabilities, cached_token_cost, is_enabled FROM shared_models_metadata ORDER BY name`

	var models []*ModelMetadata
	err := d.db.Select(&models, query)
	if err != nil {
		slog.Error("inferenceservice:dao_sqlite:GetAllModels", "message", "failed to get all models", "error", err)
		return nil, fmt.Errorf("failed to get all models")
	}

	return models, nil
}

func (d *SQLiteDAO) UpdateModelProgress(id string, progress *DownloadProgress) error {
	slog.Info("inferenceservice:dao_sqlite:UpdateModelProgress", "id", id, "progress", progress.Status)
	isDownloaded := progress.Status == StatusCompleted
	query := `UPDATE shared_models_metadata SET progress = ?, is_downloaded = ?, status = ? WHERE id = ?`

	progressJSON, err := progress.ToJSON()
	if err != nil {
		slog.Error("inferenceservice:dao_sqlite:UpdateModelProgress", "message", "failed to convert progress to JSON", "error", err)
		return fmt.Errorf("failed to convert progress to JSON")
	}

	_, err = d.db.Exec(query, progressJSON, isDownloaded, progress.Status, id)
	if err != nil {
		slog.Error("inferenceservice:dao_sqlite:UpdateModelProgress", "message", "failed to update model progress", "error", err)
		return fmt.Errorf("failed to update model progress")
	}
	return nil
}

func (d *SQLiteDAO) UpdateModelFileStoreID(id string, filestoreID string) error {
	query := `UPDATE shared_models_metadata SET filestore_id = ? WHERE id = ?`
	_, err := d.db.Exec(query, filestoreID, id)
	if err != nil {
		slog.Error("inferenceservice:dao_sqlite:UpdateModelFileStoreID", "message", "failed to update model file store ID", "error", err)
		return fmt.Errorf("failed to update model file store ID")
	}
	return nil
}

func (d *SQLiteDAO) ResetModelToInitialState(id string) error {
	query := `UPDATE shared_models_metadata SET progress = '', status = 0, is_downloaded = false, filestore_id = NULL WHERE id = ?`
	_, err := d.db.Exec(query, id)
	if err != nil {
		slog.Error("inferenceservice:dao_sqlite:ResetModelToInitialState", "message", "failed to reset model to initial state", "error", err)
		return fmt.Errorf("failed to reset model to initial state")
	}
	return nil
}
