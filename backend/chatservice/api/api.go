package api

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	"sortedstartup/chatservice/ai"
	db "sortedstartup/chatservice/dao"
	pb "sortedstartup/chatservice/proto"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedSortedChatServer
	dao          db.DAO
	modelManager *ai.ModelManager
}

func (s *Server) Chat(req *pb.ChatRequest, stream grpc.ServerStreamingServer[pb.ChatResponse]) error {
	chatId := req.ChatId
	if chatId == "" {
		return fmt.Errorf("Chat ID is required to maintain context")
	}

	model := req.Model
	if model == "" {
		return fmt.Errorf("Model is required")
	}

	// Parse provider and model - support both "provider/model" and legacy "model" formats
	var providerName, modelName string

	if strings.Contains(model, "/") {
		// New format: "provider/model" (e.g., "openai/gpt-4o")
		parts := strings.SplitN(model, "/", 2)
		if len(parts) != 2 {
			return fmt.Errorf("Invalid model format. Use 'provider/model' (e.g., 'openai/gpt-4o') or just 'model' for auto-detection")
		}
		providerName = parts[0]
		modelName = parts[1]
	} else {
		// Legacy format: just "model" (e.g., "gpt-4o") - auto-detect provider
		modelName = model
		providerName = s.detectProviderForModel(modelName)
		if providerName == "" {
			return fmt.Errorf("Could not auto-detect provider for model '%s'. Use format 'provider/model' (e.g., 'openai/gpt-4o')", modelName)
		}
	}

	// Get the provider
	provider, exists := s.modelManager.GetProvider(providerName)
	if !exists {
		return fmt.Errorf("Provider '%s' not found or not configured", providerName)
	}

	// Validate that the provider supports this model
	supportedModels := provider.SupportedModels()
	modelSupported := false
	for _, supportedModel := range supportedModels {
		if supportedModel == modelName {
			modelSupported = true
			break
		}
	}
	if !modelSupported {
		return fmt.Errorf("Model '%s' is not supported by provider '%s'. Supported models: %v",
			modelName, providerName, supportedModels)
	}

	// Get chat history using DAO
	history, err := s.dao.GetChatMessages(chatId)
	if err != nil {
		slog.Error("failed to fetch message history", "error", err)
		return fmt.Errorf("failed to fetch message history: %v", err)
	}

	// Convert DAO messages to AI messages
	var messages []ai.ChatMessage

	// Add system message if no history
	if len(history) == 0 {
		messages = append(messages, ai.NewTextMessage("system", "You are a helpful assistant"))
	} else {
		// Convert history to AI format
		for _, h := range history {
			messages = append(messages, ai.NewTextMessage(h.Role, h.Content))
		}
	}

	// Parse user input for multimodal content
	userMessage := s.parseUserInput(req.Text)
	messages = append(messages, userMessage)

	// Add user message to database
	err = s.dao.AddChatMessage(chatId, "user", req.Text)
	if err != nil {
		return fmt.Errorf("failed to insert user message: %v", err)
	}

	// Create AI request
	aiRequest := ai.ChatRequest{
		Model:    modelName,
		Messages: messages,
		Stream:   true,
	}

	// Get streaming response from AI provider
	responseStream, err := provider.Chat(context.Background(), aiRequest)
	if err != nil {
		return fmt.Errorf("AI request failed: %v", err)
	}

	var assistantText strings.Builder
	var inputTokens, outputTokens int

	// Stream responses back to client
	for response := range responseStream {
		switch response.Type {
		case "text_delta":
			// Stream incremental text
			stream.Send(&pb.ChatResponse{Text: response.Delta})
			assistantText.WriteString(response.Delta)

		case "completion":
			// Final completion
			inputTokens = response.InputTokens
			outputTokens = response.OutputTokens

		case "error":
			return fmt.Errorf("AI provider error: %s", response.Error)
		}
	}

	// Save assistant response to database with original model format for consistency
	err = s.dao.AddChatMessageWithTokens(chatId, "assistant", assistantText.String(), model, inputTokens, outputTokens)
	if err != nil {
		log.Printf("Failed to insert assistant message: %v", err)
	}

	return nil
}

// detectProviderForModel attempts to auto-detect the provider based on the model name
func (s *Server) detectProviderForModel(modelName string) string {
	// Check each registered provider to see if it supports this model
	allProviders := s.modelManager.GetAllProviders()
	for providerName, provider := range allProviders {
		supportedModels := provider.SupportedModels()
		for _, supportedModel := range supportedModels {
			if supportedModel == modelName {
				return providerName
			}
		}
	}

	// Fallback: try to guess based on common model naming patterns
	switch {
	case strings.HasPrefix(modelName, "gpt-") || strings.HasPrefix(modelName, "o1-") || modelName == "chatgpt-4o-latest":
		if _, exists := s.modelManager.GetProvider("openai"); exists {
			return "openai"
		}
	case strings.HasPrefix(modelName, "claude-"):
		if _, exists := s.modelManager.GetProvider("claude"); exists {
			return "claude"
		}
	case strings.HasPrefix(modelName, "gemini-"):
		if _, exists := s.modelManager.GetProvider("gemini"); exists {
			return "gemini"
		}
	}

	return "" // Could not detect
}

// parseUserInput parses user input to detect images and create multimodal content
func (s *Server) parseUserInput(text string) ai.ChatMessage {
	// Use the utility function to parse multimodal input
	contents, err := ai.ParseMultimodalInput(text)
	if err != nil {
		// If parsing fails, fallback to simple text
		contents = []ai.MessageContent{ai.NewTextContent(text)}
	}

	return ai.NewMultimodalMessage("user", contents)
}

func (s *Server) GetHistory(ctx context.Context, req *pb.GetHistoryRequest) (*pb.GetHistoryResponse, error) {
	chatId := req.ChatId
	if chatId == "" {
		return nil, fmt.Errorf("chat ID is required")
	}

	messages, err := s.dao.GetChatMessages(chatId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch history: %v", err)
	}

	var pbMessages []*pb.ChatMessage
	for _, m := range messages {
		pbMessages = append(pbMessages, &pb.ChatMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	return &pb.GetHistoryResponse{
		History: pbMessages,
	}, nil
}

type ChatRow struct {
	ChatID string `db:"chat_id"`
	Name   string `db:"name"`
}

func (s *Server) GetChatList(ctx context.Context, req *pb.GetChatListRequest) (*pb.GetChatListResponse, error) {
	chats, err := s.dao.GetChatList()
	if err != nil {
		slog.Error("failed to fetch chat list", "error", err)
		return nil, fmt.Errorf("failed to fetch chat list: %v", err)
	}

	return &pb.GetChatListResponse{Chats: chats}, nil
}

func (s *Server) CreateChat(ctx context.Context, req *pb.CreateChatRequest) (*pb.CreateChatResponse, error) {
	name := req.Name
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	chatId := uuid.New().String()

	err := s.dao.CreateChat(chatId, name)
	if err != nil {
		return nil, fmt.Errorf("failed to insert chat record: %w", err)
	}

	return &pb.CreateChatResponse{
		Message: "Chat created successfully",
		ChatId:  chatId, // return chatId so the frontend can use it for messages
	}, nil
}

func (s *Server) ListModel(ctx context.Context, req *pb.ListModelsRequest) (*pb.ListModelsResponse, error) {
	// Get models from both DAO (for static model info) and ModelManager (for available providers)
	daoModels, err := s.dao.GetModels()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch models: %v", err)
	}

	// Get models from all registered providers
	supportedModels := s.modelManager.GetSupportedModels()

	pbModels := make([]*pb.ModelListInfo, 0)

	// Add DAO models
	for _, m := range daoModels {
		pbModels = append(pbModels, &pb.ModelListInfo{
			Id:    m.Id,
			Label: m.Label,
		})
	}

	// Add provider models with provider prefix
	for providerName, models := range supportedModels {
		for _, model := range models {
			modelId := fmt.Sprintf("%s/%s", providerName, model)
			modelLabel := fmt.Sprintf("%s (%s)", model, providerName)
			pbModels = append(pbModels, &pb.ModelListInfo{
				Id:    modelId,
				Label: modelLabel,
			})
		}
	}

	return &pb.ListModelsResponse{Models: pbModels}, nil
}

type ChatSearchRow struct {
	ChatID      string `db:"chat_id"`
	ChatName    string `db:"chat_name"`
	MatchedText string `db:"aggregated_snippets"`
}

func (s *Server) SearchChat(ctx context.Context, req *pb.ChatSearchRequest) (*pb.ChatSearchResponse, error) {
	query := req.Query
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	results, err := s.dao.SearchChatMessages(query)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	var pbResults []*pb.SearchResult
	for _, result := range results {
		pbResults = append(pbResults, &pb.SearchResult{
			ChatId:      result.ChatId,
			ChatName:    result.ChatName,
			MatchedText: result.MatchedText,
		})
	}

	return &pb.ChatSearchResponse{
		Query:   query,
		Results: pbResults,
	}, nil
}

func (s *Server) Init() {
	// Initialize DAO
	sqliteDAO, err := db.NewSQLiteDAO("chatservice.db")
	if err != nil {
		log.Fatalf("Failed to initialize DAO: %v", err)
	}
	s.dao = sqliteDAO

	// Initialize ModelManager and register providers
	s.modelManager = ai.NewModelManager()

	// Register OpenAI provider
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		openaiProvider := ai.NewOpenAIProvider(apiKey)
		s.modelManager.RegisterProvider(openaiProvider)
		log.Printf("Registered OpenAI provider")
	}

	// Register Claude provider
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		claudeProvider := ai.NewClaudeProvider(apiKey)
		s.modelManager.RegisterProvider(claudeProvider)
		log.Printf("Registered Claude provider")
	}

	// Register Gemini provider
	if apiKey := os.Getenv("GEMINI_API_KEY"); apiKey != "" {
		geminiProvider := ai.NewGeminiProvider(apiKey)
		s.modelManager.RegisterProvider(geminiProvider)
		log.Printf("Registered Gemini provider")
	}

	// Database initialization
	db.MigrateSQLite("chatservice.db")
	db.SeedSqlite("chatservice.db")
}
