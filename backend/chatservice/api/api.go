package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"io"
	"strings"

	"sortedstartup/chatservice/dao"
	db "sortedstartup/chatservice/dao"
	pb "sortedstartup/chatservice/proto"
	"sortedstartup/chatservice/queue"
	"sortedstartup/chatservice/store"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedSortedChatServer
	dao   *dao.SQLiteDAO
	store *store.DiskObjectStore
	queue queue.Queue
}

func NewServer(mux *http.ServeMux) *Server {
	daoInstance, err := dao.NewSQLiteDAO("chatservice.db")
	if err != nil {
		log.Fatalf("Failed to initialize DAO: %v", err)
	}

	storeInstance, err := store.NewDiskObjectStore("filestore")
	if err != nil {
		log.Fatalf("Failed to initialize object store: %v", err)
	}

	s := &Server{
		dao:   daoInstance,
		store: storeInstance,
		queue: queue.NewInMemoryQueue(),
	}

	s.registerRoutes(mux)

	s.EmbeddingSubscriber()

	return s
}


func (s *Server) Chat(req *pb.ChatRequest, stream grpc.ServerStreamingServer[pb.ChatResponse]) error {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("OpenAI API key not set")
	}

	chatId := req.ChatId
	if chatId == "" {
		return fmt.Errorf("Chat ID is required to maintain context")
	}

	model := req.Model
	if model == "" {
		return fmt.Errorf("Model is required")
	}

	// Get chat history using DAO
	history, err := s.dao.GetChatMessages(chatId)
	if err != nil {
		slog.Error("failed to fetch message history", "error", err)
		return fmt.Errorf("failed to fetch message history: %v", err)
	}

	var messages []map[string]interface{}
	
	if len(history) == 0 {
		messages = append(messages, map[string]interface{}{
			"role":    "system",
			"content": "You are a helpful assistant",
		})
	}

	for _, msg := range history {
		messages = append(messages, map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}
	

	err = s.dao.AddChatMessage(chatId, "user", req.Text)
	if err != nil {
		return fmt.Errorf("failed to insert user message: %v", err)
	}

	messages = append(messages, map[string]interface{}{
		"role":    "user",
		"content": req.Text,
	})

	requestBody := map[string]interface{}{
		"model":    model,
		"messages": messages,
		"stream":   true,
		"stream_options": map[string]interface{}{
			"include_usage": true,
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	httpReq, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("OpenAI request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("OpenAI API error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var fullResponse strings.Builder
	var inputTokens, outputTokens int

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		if line == "" {
			continue
		}
		
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		
		if data == "[DONE]" {
			break
		}

		var chunk map[string]interface{}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			log.Printf("Failed to parse chunk: %v", err)
			continue
		}

		if usage, ok := chunk["usage"].(map[string]interface{}); ok {
			if promptTokens, ok := usage["prompt_tokens"].(float64); ok {
				inputTokens = int(promptTokens)
			}
			if completionTokens, ok := usage["completion_tokens"].(float64); ok {
				outputTokens = int(completionTokens)
			}
			log.Printf("Token usage found: input=%d, output=%d", inputTokens, outputTokens)
		}

		choices, ok := chunk["choices"].([]interface{})
		if !ok || len(choices) == 0 {
			continue
		}

		choice, ok := choices[0].(map[string]interface{})
		if !ok {
			continue
		}

		if delta, ok := choice["delta"].(map[string]interface{}); ok {
			if content, ok := delta["content"].(string); ok && content != "" {
				fullResponse.WriteString(content)
				
				if err := stream.Send(&pb.ChatResponse{Text: content}); err != nil {
					return fmt.Errorf("failed to send stream response: %v", err)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stream: %v", err)
	}

	assistantText := fullResponse.String()
	if assistantText != "" {
		err := s.dao.AddChatMessageWithTokens(chatId, "assistant", assistantText, model, inputTokens, outputTokens)
		if err != nil {
			log.Printf("Failed to insert assistant message: %v", err)
		}
	}

	return nil
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
	models, err := s.dao.GetModels()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch models: %v", err)
	}

	pbModels := make([]*pb.ModelListInfo, 0, len(models))
	for _, m := range models {
		pbModels = append(pbModels, &pb.ModelListInfo{
			Id:    m.Id,
			Label: m.Label,
		})
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

func (s *Server) CreateProject(ctx context.Context, req *pb.CreateProjectRequest) (*pb.CreateProjectResponse, error) {
	id := uuid.New().String()

	name := req.Name
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	description := req.Description
	additionalData := req.AdditionalData

	projectID, err := s.dao.CreateProject(id, name, description, additionalData)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return &pb.CreateProjectResponse{
		Message:   "Project created successfully",
		ProjectId: projectID,
	}, nil
}

func (s *Server) GetProjects(ctx context.Context, req *pb.GetProjectsRequest) (*pb.GetProjectsResponse, error) {
	projects, err := s.dao.GetProjects()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch project list: %w", err)
	}

	var pbProjects []*pb.Project
	for _, p := range projects {
		pbProjects = append(pbProjects, &pb.Project{
			Id:             p.ID,
			Name:           p.Name,
			Description:    p.Description,
			AdditionalData: p.AdditionalData,
			CreatedAt:      p.CreatedAt,
			UpdatedAt:      p.UpdatedAt,
		})
	}

	return &pb.GetProjectsResponse{Projects: pbProjects}, nil
}

func (s *Server) ListDocuments(ctx context.Context, req *pb.ListDocumentsRequest) (*pb.ListDocumentsResponse, error) {
	docs, err := s.dao.FilesList(req.GetProjectId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch documents: %v", err)
	}

	var result []*pb.Document
	for _, doc := range docs {
		result = append(result, &pb.Document{
			Id:        doc.ID,
			ProjectId: doc.ProjectID,
			DocsId:    doc.DocsID,
			FileName:  doc.FileName,
			CreatedAt: doc.CreatedAt,
			UpdatedAt: doc.UpdatedAt,
		})
	}

	return &pb.ListDocumentsResponse{
		Documents: result,
	}, nil

}

func (s *Server) Init() {
	// Initialize DAO
	sqliteDAO, err := db.NewSQLiteDAO("chatservice.db")
	if err != nil {
		log.Fatalf("Failed to initialize DAO: %v", err)
	}
	s.dao = sqliteDAO

	//db.InitDB()
	// TODO: handle migration for postgres also
	db.MigrateSQLite("chatservice.db")
	db.SeedSqlite("chatservice.db")
}

func (s *Server) EmbeddingSubscriber() {
	go func() {

		sub, err := s.queue.Subscribe(context.Background(), "generate.embedding")
		if err != nil {
			fmt.Printf("Failed %v\n", err)
			return
		}
		for msg := range sub {
			fmt.Println(msg)
			var payload GenerateEmbeddingMessage
			if err := json.Unmarshal(msg.Data, &payload); err == nil {
				fmt.Println(payload)
				fmt.Printf("docs_id: %v\n", payload.DocsID)
			}
		}
	}()
}
