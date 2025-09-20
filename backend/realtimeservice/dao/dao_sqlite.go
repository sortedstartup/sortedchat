package dao

import (

	// sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"

	"log/slog"

	"github.com/google/uuid"
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

	// Set busy timeout to 10 seconds
	_, err = db.Exec("PRAGMA busy_timeout = 10000;")
	if err != nil {
		return nil, err
	}

	// Enable WAL mode
	_, err = db.Exec("PRAGMA journal_mode = WAL;")
	if err != nil {
		return nil, err
	}

	return &SQLiteDAO{db: db}, nil
}

func (d *SQLiteDAO) Infer(dummy string) error {
	return nil
}

func (d *SQLiteDAO) CreateAudioChat(userID string, modelName string, startTime string, endTime string) (string, error) {
	id := uuid.New().String()
	_, err := d.db.Exec("INSERT INTO realtimeservice_audio_chat (id, model_name, user_id, start_time, end_time) VALUES (?, ?, ?, ?, ?)", id, modelName, userID, startTime, endTime)
	if err != nil {
		slog.Error("Failed to create audio chat sqlite", "error", err)
		return "", err
	}
	return id, nil
}

func (d *SQLiteDAO) UpdateAudioChat(userID string, id string, endTime string) error {
	_, err := d.db.Exec("UPDATE realtimeservice_audio_chat SET end_time = ? WHERE id = ? AND user_id = ?", endTime, id, userID)
	if err != nil {
		slog.Error("Failed to update audio chat sqlite", "error", err)
		return err
	}
	return nil
}
