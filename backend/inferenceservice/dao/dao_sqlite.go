package dao

import (

	// sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteDAO implements the DAO interface using SQLite and sqlx
type SQLiteDAO struct {
	db *sqlx.DB
}

// NewSQLiteDAO creates a new SQLite DAO instance
func NewSQLiteDAO(sqliteUrl string) (*SQLiteDAO, error) {

	db, err := sqlx.Open("sqlite3", sqliteUrl)
	if err != nil {
		return nil, err
	}

	return &SQLiteDAO{db: db}, nil
}

func (d *SQLiteDAO) Infer(dummy string) error {
	return nil
}

func (d *SQLiteDAO) DownloadModel(userID string, modelName string, url string) error {
	return nil
}

func (d *SQLiteDAO) GetModelByName(modelName string) (*ModelMetadata, error) {
	query := `SELECT id, name, url, provider, input_token_cost, output_token_cost, progress, is_downloaded, is_downloadable, status FROM inference_model_metadata WHERE name = ?`

	var model ModelMetadata
	err := d.db.Get(&model, query, modelName)
	if err != nil {
		return nil, err
	}

	return &model, nil
}

func (d *SQLiteDAO) GetAllModels() ([]*ModelMetadata, error) {
	query := `SELECT id, name, url, provider, input_token_cost, output_token_cost, progress, is_downloaded, is_downloadable, status FROM inference_model_metadata ORDER BY name`

	var models []*ModelMetadata
	err := d.db.Select(&models, query)
	if err != nil {
		return nil, err
	}

	return models, nil
}

func (d *SQLiteDAO) UpdateModelProgress(id string, progress *DownloadProgress) error {
	isDownloaded := progress.Status == StatusCompleted
	query := `UPDATE inference_model_metadata SET progress = ?, is_downloaded = ?, status = ? WHERE id = ?`

	_, err := d.db.Exec(query, progress.ToJSON(), isDownloaded, progress.Status, id)
	return err
}
