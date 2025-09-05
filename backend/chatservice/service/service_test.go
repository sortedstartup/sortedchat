package service

import (
	"context"
	"fmt"
	"sortedstartup/chatservice/dao"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite db: %v", err)
	}
	schema := `
	CREATE TABLE chat_list (
		chat_id TEXT PRIMARY KEY,
		name TEXT,
		user_id TEXT,
		project_id TEXT,
		parent_chat_id TEXT,
		parent_message_id TEXT,
		is_main_branch BOOLEAN DEFAULT 1,
		soft_deleted BOOLEAN DEFAULT 0
	);`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}
	fmt.Println("Database schema created successfully")
	return db
}

func TestRenameChat(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.Exec("INSERT INTO chat_list (chat_id, name, user_id) VALUES (?, ?, ?)", "chat123", "Old Chat Name", "user456")
	if err != nil {
		t.Fatalf("failed to insert initial chat: %v", err)
	}

	daoInstance, err := dao.NewSQLiteDAO(":memory:")
	if err != nil {
		t.Fatalf("failed to create SQLiteDAO: %v", err)
	}

	// daoInstance.db = db // overwrite the db kaise ?

	chatService := &ChatService{dao: daoInstance}

	err = chatService.RenameChat(context.Background(), "user456", "chat123", "New Chat Name")
	if err != nil {
		fmt.Println("saskdsfs", err)
		t.Errorf("RenameChat error = %v", err)
	}

	var newName string
	err = db.Get(&newName, "SELECT name FROM chat_list WHERE chat_id=?", "chat123")
	if err != nil {
		t.Fatalf("failed to query chat: %v", err)
	}
	if strings.TrimSpace(newName) != "New Chat Name" {
		t.Errorf("expected chat name to be 'New Chat Name', got '%s'", newName)
	}
}
