package db

import (
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sqlx.DB

func InitDB() {
	var err error
	DB, err = sqlx.Open("sqlite3", "./chat_history.db")
	if err != nil {
		log.Fatal(err)
	}

	schema := `
    CREATE TABLE IF NOT EXISTS chat_messages (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        chat_id TEXT NOT NULL,
        role TEXT NOT NULL,
        content TEXT NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );

	CREATE TABLE IF NOT EXISTS chat_list (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chat_id TEXT NOT NULL,
		name TEXT NOT NULL
	);
    `
	_, err = DB.Exec(schema)
	if err != nil {
		log.Fatalf("failed to create schema: %v", err)
	}
}
