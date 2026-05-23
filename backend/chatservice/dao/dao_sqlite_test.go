package dao

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"strings"
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

func TestAddChatMessageWithTokens(t *testing.T) {
	// Define model metadata
	modelMetadata := []struct {
		model            string
		modelDisplayName string
		modelProvider    string
		inputTokenCost   float64
		outputTokenCost  float64
		cachedTokenCost  float64
	}{
		{
			model:            "sanskar-4.1",
			modelDisplayName: "SANSKAR-4",
			modelProvider:    "SANSKAR",
			inputTokenCost:   0.025,
			outputTokenCost:  0.050,
			cachedTokenCost:  0.001,
		},
		{
			model:            "gpt-4-turbo",
			modelDisplayName: "GPT-4 Turbo",
			modelProvider:    "OPENAI",
			inputTokenCost:   0.010,
			outputTokenCost:  0.030,
			cachedTokenCost:  0.005,
		},
		{
			model:            "claude-3-sonnet",
			modelDisplayName: "Claude-3 Sonnet",
			modelProvider:    "ANTHROPIC",
			inputTokenCost:   0.003,
			outputTokenCost:  0.015,
			cachedTokenCost:  0.0003,
		},
		{
			model:            "test-model-small",
			modelDisplayName: "Test Model Small",
			modelProvider:    "TEST",
			inputTokenCost:   0.0001,
			outputTokenCost:  0.0002,
			cachedTokenCost:  0.0001,
		},
	}

	// Define token test cases
	tokenTestCases := []struct {
		name          string
		role          string
		message       string
		inputTokens   int
		outputTokens  int
		cachedTokens  int
		references    string
		ragEnabled    bool
		expectError   bool
		errorContains string
	}{
		{
			name:         "Valid assistant message",
			role:         "assistant",
			message:      "Hello, how can I help you?",
			inputTokens:  4,
			outputTokens: 10,
			cachedTokens: 0,
			references:   "",
			ragEnabled:   false,
			expectError:  false,
		},
		{
			name:         "Valid user message with cached tokens",
			role:         "user",
			message:      "What is the weather today?",
			inputTokens:  5,
			outputTokens: 0,
			cachedTokens: 5,
			references:   "",
			ragEnabled:   false,
			expectError:  false,
		},
		{
			name:         "System message with references and RAG enabled",
			role:         "system",
			message:      "You are a helpful assistant.",
			inputTokens:  6,
			outputTokens: 0,
			cachedTokens: 0,
			references:   "doc1.pdf,doc2.txt",
			ragEnabled:   true,
			expectError:  false,
		},
		{
			name:         "Small token counts",
			role:         "assistant",
			message:      "Short response with few tokens.",
			inputTokens:  3,
			outputTokens: 7,
			cachedTokens: 2,
			references:   "",
			ragEnabled:   false,
			expectError:  false,
		},
		{
			name:         "Zero tokens",
			role:         "user",
			message:      "",
			inputTokens:  0,
			outputTokens: 0,
			cachedTokens: 0,
			references:   "",
			ragEnabled:   false,
			expectError:  false,
		},
		{
			name:         "High token counts",
			role:         "assistant",
			message:      "This is a very long response that would consume many tokens in a real scenario.",
			inputTokens:  100,
			outputTokens: 250,
			cachedTokens: 50,
			references:   "",
			ragEnabled:   false,
			expectError:  false,
		},
		{
			name:         "Only input tokens",
			role:         "user",
			message:      "User query with only input tokens",
			inputTokens:  15,
			outputTokens: 0,
			cachedTokens: 0,
			references:   "",
			ragEnabled:   false,
			expectError:  false,
		},
		{
			name:         "Only cached tokens",
			role:         "user",
			message:      "Message using only cached tokens",
			inputTokens:  0,
			outputTokens: 0,
			cachedTokens: 20,
			references:   "",
			ragEnabled:   false,
			expectError:  false,
		},
	}

	daoInstance := SetupSQLiteInMemoryTestDB(t)

	for _, model := range modelMetadata {
		t.Run(fmt.Sprintf("Model_%s", model.model), func(t *testing.T) {
			// Add the model once per model test group
			err := daoInstance.UpsertModel(model.model, model.modelDisplayName, "", model.modelProvider,
				model.inputTokenCost, model.outputTokenCost, model.cachedTokenCost, false)
			if err != nil {
				t.Fatalf("failed to add model: %v", err)
			}

			for _, tc := range tokenTestCases {
				t.Run(tc.name, func(t *testing.T) {
					// Generate unique IDs for this test case to avoid conflicts
					userID := fmt.Sprintf("user_%s_%s", model.model, strings.ReplaceAll(tc.name, " ", "_"))
					chatID := fmt.Sprintf("chat_%s_%s", model.model, strings.ReplaceAll(tc.name, " ", "_"))

					// Pre-insert a chat into chat_list to satisfy foreign key constraint
					err := daoInstance.CreateChat(userID, chatID, "Test Chat", "")
					if err != nil {
						t.Fatalf("failed to create chat: %v", err)
					}

					// Execute the function under test
					summary, err := daoInstance.AddChatMessageWithTokens(
						userID,
						chatID,
						tc.role,
						tc.message,
						"",
						model.model,
						tc.inputTokens,
						tc.outputTokens,
						tc.cachedTokens,
						0,
						0,
						0,
						tc.references,
						tc.ragEnabled,
						nil,
					)

					// Check error expectations
					if tc.expectError {
						if err == nil {
							t.Fatalf("expected error but got none")
						}
						if tc.errorContains != "" && !strings.Contains(err.Error(), tc.errorContains) {
							t.Fatalf("expected error containing '%s', got '%s'", tc.errorContains, err.Error())
						}
						return // Skip further assertions for error cases
					}

					if err != nil {
						t.Fatalf("AddChatMessageWithTokens failed: %v", err)
					}

					// Verify the returned summary
					if summary.Model != model.model {
						t.Errorf("expected model '%s', got '%s'", model.model, summary.Model)
					}

					if summary.InputTokenCount != tc.inputTokens {
						t.Errorf("expected input token count %d, got %d", tc.inputTokens, summary.InputTokenCount)
					}

					if summary.OutputTokenCount != tc.outputTokens {
						t.Errorf("expected output token count %d, got %d", tc.outputTokens, summary.OutputTokenCount)
					}

					if summary.CachedTokenCount != tc.cachedTokens {
						t.Errorf("expected cached token count %d, got %d", tc.cachedTokens, summary.CachedTokenCount)
					}

					// Calculate expected cost
					expectedCost := (float64(tc.inputTokens)*model.inputTokenCost +
						float64(tc.outputTokens)*model.outputTokenCost +
						float64(tc.cachedTokens)*model.cachedTokenCost) / 1000000.0

					// Direct equality check for cost
					if summary.Cost != expectedCost {
						t.Errorf("expected cost %f, got %f", expectedCost, summary.Cost)
					}

					// Verify the message was inserted into chat_messages table
					var count int
					err = daoInstance.db.QueryRow(`
						SELECT COUNT(*) FROM chat_messages 
						WHERE chat_id = ? AND user_id = ? AND role = ?`,
						chatID, userID, tc.role).Scan(&count)
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
						chatID, userID).Scan(&totalCost, &totalInput, &totalOutput, &totalCached)
					if err != nil {
						t.Fatalf("failed to query chat_list: %v", err)
					}

					if totalCost != expectedCost {
						t.Errorf("expected chat_list cost %f, got %f", expectedCost, totalCost)
					}
					if totalInput != tc.inputTokens {
						t.Errorf("expected chat_list input tokens %d, got %d", tc.inputTokens, totalInput)
					}
					if totalOutput != tc.outputTokens {
						t.Errorf("expected chat_list output tokens %d, got %d", tc.outputTokens, totalOutput)
					}
					if totalCached != tc.cachedTokens {
						t.Errorf("expected chat_list cached tokens %d, got %d", tc.cachedTokens, totalCached)
					}
				})
			}
		})
	}
}
