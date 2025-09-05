package service

import (
	"context"
	"sortedstartup/chatservice/dao"
	"strings"
	"testing"
)

func TestRenameChat(t *testing.T) {
	config := &dao.Config{}
	config.Database.Type = dao.DatabaseTypeSQLite
	config.Database.SQLite.URL = ":memory:"

	// Run migrations and seed using service.Init
	chatService := &ChatService{}
	dbConn := chatService.Init(config)

	daoInstance, err := dao.NewSQLiteInMemoryDAO(dbConn)
	if err != nil {
		t.Fatalf("failed to create SQLiteDAO: %v", err)
	}
	chatService.dao = daoInstance

	// Insert initial chat data using DAO
	chat_id, err := chatService.CreateChat(context.Background(), "user456", "Old Chat Name", "")
	if err != nil {
		t.Fatalf("failed to insert initial chat: %v", err)
	}

	// Rename chat
	err = chatService.RenameChat(context.Background(), "user456", chat_id, "New Chat Name")
	if err != nil {
		t.Errorf("RenameChat error = %v", err)
	}

	// Verify change persisted in DB
	newName, err := daoInstance.GetChatName("user456", chat_id)
	if err != nil {
		t.Fatalf("failed to query chat: %v", err)
	}
	if strings.TrimSpace(newName) != "New Chat Name" {
		t.Errorf("expected chat name to be 'New Chat Name', got '%s'", newName)
	}
}
