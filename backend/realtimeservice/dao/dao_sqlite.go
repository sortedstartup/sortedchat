package dao

import (

	// sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"

	"fmt"
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
	slog.Info("RealtimeService:dao_sqlite:NewSQLiteDAO")
	db, err := sqlx.Open("sqlite3", sqliteUrl)
	if err != nil {
		slog.Error("RealtimeService:dao_sqlite:NewSQLiteDAO", "message", "failed to open SQLite database", "error", err)
		return nil, fmt.Errorf("failed to open SQLite database")
	}

	// Set busy timeout to 30 seconds
	_, err = db.Exec("PRAGMA busy_timeout = 30000;")
	if err != nil {
		slog.Error("RealtimeService:dao_sqlite:NewSQLiteDAO", "message", "failed to set busy timeout", "error", err)
		return nil, err
	}

	// Enable WAL mode
	_, err = db.Exec("PRAGMA journal_mode = WAL;")
	if err != nil {
		slog.Error("RealtimeService:dao_sqlite:NewSQLiteDAO", "message", "failed to set WAL mode", "error", err)
		return nil, err
	}

	return &SQLiteDAO{db: db}, nil
}

// NewSQLiteDAOWithDB creates a new SQLite DAO instance using a shared database connection
func NewSQLiteDAOWithDB(db *sqlx.DB) (*SQLiteDAO, error) {
	slog.Info("RealtimeService:dao_sqlite:NewSQLiteDAOWithDB")
	return &SQLiteDAO{db: db}, nil
}

func (d *SQLiteDAO) Infer(dummy string) error {
	return nil
}

func (d *SQLiteDAO) CreateAudioChat(userID string, modelName string, startTime string, endTime string) (string, error) {
	slog.Info("RealtimeService:dao_sqlite:CreateAudioChat")
	id := uuid.New().String()
	_, err := d.db.Exec("INSERT INTO realtimeservice_audio_chat (id, model_name, user_id, start_time, end_time) VALUES (?, ?, ?, ?, ?)", id, modelName, userID, startTime, endTime)
	if err != nil {
		slog.Error("RealtimeService:dao_sqlite:CreateAudioChat", "message", "failed to create audio chat", "error", err)
		return "", fmt.Errorf("failed to create audio chat")
	}
	return id, nil
}

func (d *SQLiteDAO) UpdateAudioChat(userID string, id string, endTime string) error {
	slog.Info("RealtimeService:dao_sqlite:UpdateAudioChat")
	_, err := d.db.Exec("UPDATE realtimeservice_audio_chat SET end_time = ? WHERE id = ? AND user_id = ?", endTime, id, userID)
	if err != nil {
		slog.Error("RealtimeService:dao_sqlite:UpdateAudioChat", "message", "failed to update audio chat", "error", err)
		return fmt.Errorf("failed to update audio chat")
	}
	return nil
}
