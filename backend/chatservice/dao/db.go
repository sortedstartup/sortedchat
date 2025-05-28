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
		model TEXT ,
		error BOOLEAN,
		input_token_count INTEGER,
		output_token_count INTEGER,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );

	CREATE TABLE IF NOT EXISTS chat_list (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chat_id TEXT NOT NULL,
		name TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS model_metadata (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		url TEXT NOT NULL,
		provider TEXT,
		input_token_cost REAL,
		output_token_cost REAL
	);
	CREATE VIRTUAL TABLE IF NOT EXISTS chat_messages_fts USING fts5(
		message_content,
		content='chat_messages',
		content_rowid='id',
		tokenize='porter unicode61'
	);

	CREATE TRIGGER IF NOT EXISTS chat_messages_ai_fts 
	AFTER INSERT ON chat_messages 
	BEGIN
	INSERT INTO chat_messages_fts (rowid, message_content) VALUES (new.id, new.content);
	END;

	CREATE TRIGGER IF NOT EXISTS chat_messages_ad_fts 
	AFTER DELETE ON chat_messages
	BEGIN
	INSERT INTO chat_messages_fts (chat_messages_fts, rowid, message_content) VALUES ('delete', old.id, old.content);
	END;

	CREATE TRIGGER IF NOT EXISTS chat_messages_au_fts 
	AFTER UPDATE OF content ON chat_messages 
	BEGIN
	INSERT INTO chat_messages_fts (chat_messages_fts, rowid, message_content) VALUES ('delete', old.id, old.content);
	INSERT INTO chat_messages_fts (rowid, message_content) VALUES (new.id, new.content);
	END;
    `

	_, err = DB.Exec(schema)
	if err != nil {
		log.Fatalf("failed to create schema: %v", err)
	}

	var count int
	err = DB.Get(&count, "SELECT COUNT(*) FROM model_metadata")
	if err != nil {
		log.Fatalf("failed to count model_metadata: %v", err)
	}

	if count == 0 {
		models := []struct {
			ID              string
			Name            string
			URL             string
			Provider        string
			InputTokenCost  float64
			OutputTokenCost float64
		}{
			{"chatgpt-4o-latest", "ChatGPT-4o (Latest)", "https://api.openai.com/v1/responses", "openai", 0.01, 0.01},
			{"gpt-4-turbo", "GPT-4 Turbo", "https://api.openai.com/v1/responses", "openai", 0.01, 0.01},
			{"gpt-4.1", "GPT-4.1", "https://api.openai.com/v1/responses", "openai", 0.01, 0.01},
			{"gpt-4o", "GPT-4o", "https://api.openai.com/v1/responses", "openai", 0.01, 0.01},
			{"gpt-4o-mini", "GPT-4o Mini", "https://api.openai.com/v1/responses", "openai", 0.01, 0.01},
			{"o3-mini", "o3-mini", "https://api.openai.com/v1/responses", "openai", 0.01, 0.01},
			{"o4-mini", "o4-mini", "https://api.openai.com/v1/responses", "openai", 0.01, 0.01},
		}

		for _, m := range models {
			_, err := DB.Exec(`
				INSERT INTO model_metadata (id, name, url, provider, input_token_cost, output_token_cost)
				VALUES (?, ?, ?, ?, ?, ?)`,
				m.ID, m.Name, m.URL, m.Provider, m.InputTokenCost, m.OutputTokenCost,
			)
			if err != nil {
				log.Fatalf("failed to insert model metadata: %v", err)
			}
		}
	}

}
