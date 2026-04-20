package service

import (
	"context"
	"sortedstartup/chatservice/dao"
	pb "sortedstartup/chatservice/proto"
	"strings"
	"testing"
)

func setupTestChatService(t *testing.T) *ChatService {
	config := &dao.Config{}
	config.Database.Type = dao.DatabaseTypeSQLite
	config.Database.SQLite.URL = ":memory:"

	chatService := &ChatService{}
	dbConn := chatService.Init(config)

	daoInstance, err := dao.NewSQLiteInMemoryDAO(dbConn)
	if err != nil {
		t.Fatalf("failed to create SQLiteDAO: %v", err)
	}
	chatService.dao = daoInstance

	// Register cleanup to run after this test completes
	t.Cleanup(func() {
		// Cleanup logic here, e.g., closing db connection
		dbConn.Close()
	})

	return chatService
}

func TestRenameChat(t *testing.T) {
	chatService := setupTestChatService(t)
	daoInstance := chatService.dao.(*dao.SQLiteDAO)

	// Insert initial chat data using DAO
	chat_id, err := chatService.CreateChat(context.Background(), "user456", "Old Chat Name", "")
	if err != nil {
		t.Fatalf("failed to insert initial chat: %v", err)
	}

	// Rename chat
	_, err = chatService.RenameItem(context.Background(), "user456", chat_id, "New Chat Name", pb.RenameItemRequest_CHAT)
	if err != nil {
		t.Errorf("RenameItem error = %v", err)
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

// CGO_CFLAGS="-I$(pwd)/sqlite3" go test -v -tags "sqlite_fts5" -run ^TestAddModel$ sortedstartup/chatservice/service
func TestAddModel(t *testing.T) {
	chatService := setupTestChatService(t)

	tests := []struct {
		name             string
		modelId          string
		providerName     string
		modelName        string
		inputTokenCost   int32
		outputTokenCost  int32
		cachedTokenCost  int32
		isEmbeddingModel bool
		url              string
		wantErr          bool
		expectedLabel    string
	}{
		{
			name:             "Successful addition with all fields",
			modelId:          "test-model-1",
			providerName:     "test-provider",
			modelName:        "Test Model 1",
			inputTokenCost:   10,
			outputTokenCost:  20,
			cachedTokenCost:  5,
			isEmbeddingModel: false,
			url:              "https://api.test.com",
			wantErr:          false,
			expectedLabel:    "Test Model 1",
		},
		{
			name:             "Fallback to modelId if modelName is empty",
			modelId:          "test-model-2",
			providerName:     "test-provider",
			modelName:        "",
			inputTokenCost:   15,
			outputTokenCost:  25,
			cachedTokenCost:  10,
			isEmbeddingModel: true,
			url:              "https://api.test2.com",
			wantErr:          false,
			expectedLabel:    "test-model-2",
		},
		{
			name:             "Fails when modelId is empty",
			modelId:          "",
			providerName:     "test-provider",
			modelName:        "Invalid Model",
			inputTokenCost:   0,
			outputTokenCost:  0,
			cachedTokenCost:  0,
			isEmbeddingModel: false,
			url:              "",
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := chatService.AddModel(
				context.Background(),
				tt.modelId,
				tt.providerName,
				tt.modelName,
				tt.inputTokenCost,
				tt.outputTokenCost,
				tt.cachedTokenCost,
				tt.isEmbeddingModel,
				tt.url,
			)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if msg != "Model added successfully" {
				t.Errorf("expected success message, got %s", msg)
			}

			// Verify model was added
			models, err := chatService.ListModel(context.Background())
			if err != nil {
				t.Fatalf("ListModel failed: %v", err)
			}

			found := false
			for _, m := range models {
				if m.Id == tt.modelId {
					found = true
					if m.Label != tt.expectedLabel {
						t.Errorf("expected label %s, got %s", tt.expectedLabel, m.Label)
					}
					if m.Provider != tt.providerName {
						t.Errorf("expected provider %s, got %s", tt.providerName, m.Provider)
					}
					if m.CachedTokenCost != float32(tt.cachedTokenCost) {
						t.Errorf("expected cached token cost %v, got %v", tt.cachedTokenCost, m.CachedTokenCost)
					}
					if m.IsEmbeddingModel != tt.isEmbeddingModel {
						t.Errorf("expected isEmbeddingModel %v, got %v", tt.isEmbeddingModel, m.IsEmbeddingModel)
					}
					break
				}
			}

			if !found {
				t.Errorf("model %s not found in ListModel response", tt.modelId)
			}
		})
	}
}
