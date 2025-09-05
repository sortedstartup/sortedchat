package dao

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"testing"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
)

func SetupSQLiteInMemoryTestDB(t *testing.T) *SQLiteDAO {
	// Initialize in-memory SQLite database
	sqlite_vec.Auto()
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	// defer sqlDB.Close() //lets not close it here

	slog.Info("ChatService: Running SQLite migrations")
	if err := MigrateDB_UsingConnectionDefaults(sqlDB); err != nil {
		log.Fatalf("ChatService: Failed to migrate SQLite database: %v", err)
	}
	if err := SeedDB_UsingConnectionDefaults(sqlDB); err != nil {
		log.Fatalf("ChatService: Failed to seed SQLite database: %v", err)
	}
	// Create in-memory SQLite instance
	daoInstance, err := NewSQLiteInMemoryDAO(sqlDB)
	if err != nil {
		t.Fatalf("failed to create SQLiteDAO: %v", err)
	}
	return daoInstance
}

// per 1M tokens
var model = "sanskar-4.1"
var newModelInputTokenCost = 0.025
var newModelOutputTokenCost = 0.050
var newModelCachedTokenCost = 0.001

func TestAddChatMessageWithTokens(t *testing.T) {

	daoInstance := SetupSQLiteInMemoryTestDB(t)

	// Pre-insert a chat into chat_list to satisfy foreign key constraint
	err := daoInstance.CreateChat("user123", "chat123", "Test Chat", "")
	if err != nil {
		t.Fatalf("failed to create chat: %v", err)
	}

	//add new model for testging
	err = daoInstance.AddModel(model, "SANSKAR-4", "", "SANSKAR", newModelInputTokenCost, newModelOutputTokenCost, newModelCachedTokenCost)
	if err != nil {
		t.Fatalf("failed to add model: %v", err)
	}

	// Test the function
	summary, err := daoInstance.AddChatMessageWithTokens(
		"user123",
		"chat123",
		"assistant",
		"Hello, how can I help you?",
		// "gpt-4.1",
		model,
		4,     // input tokens
		10,    // output tokens
		0,     // cached tokens
		"",    // references
		false, // rag enabled
	)

	// Assertions
	if err != nil {
		t.Fatalf("AddChatMessageWithTokens failed: %v", err)
	}

	// Verify the returned summary
	if summary.Model != model {
		t.Errorf("expected model 'gpt-4.1', got '%s'", summary.Model)
	}

	if summary.InputTokenCount != 4 {
		t.Errorf("expected input token count 4, got %d", summary.InputTokenCount)
	}

	if summary.OutputTokenCount != 10 {
		t.Errorf("expected output token count 10, got %d", summary.OutputTokenCount)
	}

	if summary.CachedTokenCount != 0 {
		t.Errorf("expected cached token count 0, got %d", summary.CachedTokenCount)
	}

	expectedCost := (4*newModelInputTokenCost + 10*newModelOutputTokenCost + 0*newModelCachedTokenCost) / 1000000.0
	fmt.Println("Expected Cost:", expectedCost)
	if summary.Cost != expectedCost {
		t.Errorf("expected cost %f, got %f", expectedCost, summary.Cost)
	}

	// Verify the message was inserted into chat_messages table
	var count int
	err = daoInstance.db.QueryRow(`
        SELECT COUNT(*) FROM chat_messages 
        WHERE chat_id = ? AND user_id = ? AND role = ?`,
		"chat123", "user123", "assistant").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query chat_messages: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 message in chat_messages, got %d", count)
	}

	// Verify chat_list was updated with totals
	var totalCost float64
	var totalInput, totalOutput, totalCached int
	err = daoInstance.db.QueryRow(`
        SELECT cost, input_token_count, output_token_count, cached_token_count
        FROM chat_list WHERE chat_id = ? AND user_id = ?`,
		"chat123", "user123").Scan(&totalCost, &totalInput, &totalOutput, &totalCached)
	if err != nil {
		t.Fatalf("failed to query chat_list: %v", err)
	}

	if totalCost != expectedCost {
		t.Errorf("expected chat_list cost %f, got %f", expectedCost, totalCost)
	}
	if totalInput != 4 {
		t.Errorf("expected chat_list input tokens 4, got %d", totalInput)
	}
	if totalOutput != 10 {
		t.Errorf("expected chat_list output tokens 10, got %d", totalOutput)
	}
	if totalCached != 0 {
		t.Errorf("expected chat_list cached tokens 0, got %d", totalCached)
	}
}
