package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"sortedstartup/chatservice/dao"
	db "sortedstartup/chatservice/dao"
	pb "sortedstartup/chatservice/proto"
	"sortedstartup/chatservice/queue"
	"sortedstartup/chatservice/rag"
	"sortedstartup/chatservice/store"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SettingService struct {
	pb.UnimplementedSettingServiceServer
	dao *dao.SQLiteSettingsDAO
}

func NewSettingService() *SettingService {
	dao := dao.NewSQLiteSettingsDAO(SQLITE_DB_URL)
	return &SettingService{dao: dao}
}

func (s *SettingService) Init() {
	// since right now the Setting is in chatservice so chatservice handles migrations
}

func (s *SettingService) GetSetting(ctx context.Context, req *pb.GetSettingRequest) (*pb.GetSettingResponse, error) {
	settings, err := s.dao.GetSettings()
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}

	return &pb.GetSettingResponse{
		Settings: settings,
	}, nil
}

func (s *SettingService) SetSetting(ctx context.Context, req *pb.SetSettingRequest) (*pb.SetSettingResponse, error) {

	err := s.dao.SetSettings(req.Settings)
	if err != nil {
		return nil, fmt.Errorf("failed to set settings: %w", err)
	}

	return &pb.SetSettingResponse{
		Message: "Setting Saved",
	}, nil
}

type ChatService struct {
	pb.UnimplementedSortedChatServer
	dao                *dao.SQLiteDAO
	store              *store.DiskObjectStore
	queue              queue.Queue
	pipeline           rag.RAGIndexingPipeline
	embeddingsProvider rag.Embedder
}

var SQLITE_DB_URL = "db.sqlite"

func NewChatService(mux *http.ServeMux) *ChatService {

	daoInstance, err := dao.NewSQLiteDAO(SQLITE_DB_URL)
	if err != nil {
		log.Fatalf("Failed to initialize DAO: %v", err)
	}

	storeInstance, err := store.NewDiskObjectStore("filestore")
	if err != nil {
		log.Fatalf("Failed to initialize object store: %v", err)
	}

	ollama_url := os.Getenv("OLLAMA_URL")
	embeddingsProvider := &rag.OLLamaEmbedder{
		//TODO: read from config
		URL:   ollama_url + "/v1/embeddings",
		Model: "nomic-embed-text",
	}

	pipeline := rag.NewPipeline(
		&rag.TextExtractor{},
		&rag.EqualSizeChunker{ChunkSize: 512},
		embeddingsProvider,
	)

	s := &ChatService{
		dao:                daoInstance,
		store:              storeInstance,
		queue:              queue.NewInMemoryQueue(),
		pipeline:           pipeline,
		embeddingsProvider: embeddingsProvider,
	}

	s.registerRoutes(mux)

	s.EmbeddingSubscriber()

	return s
}

func (s *ChatService) Chat(req *pb.ChatRequest, stream grpc.ServerStreamingServer[pb.ChatResponse]) error {
	projectID := req.GetProjectId()

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

	userMessage := req.Text

	if projectID != "" && projectID != "null" { // if this chat is in context of a project
		chunks, err := s.retrieveSimilarChunks(context.Background(), projectID, req.Text)
		if err != nil {
			// TODO: user message also
			slog.Error("failed to retrieve similar chunks", "error", err)
		} else if len(chunks.Results) > 0 {
			fmt.Println("Retrieved chunks:", chunks.Prompt)
			userMessage = chunks.Prompt
		}
	}

	history = append(history, dao.ChatMessageRow{Role: "user", Content: userMessage})

	requestBody := map[string]interface{}{
		"model":    model,
		"messages": history,
		"stream":   true,
		"stream_options": map[string]interface{}{
			"include_usage": true,
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	httpReq, err := http.NewRequest("POST", os.Getenv("OPENAI_API_URL"), bytes.NewBuffer(jsonData))
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

func (s *ChatService) GetHistory(ctx context.Context, req *pb.GetHistoryRequest) (*pb.GetHistoryResponse, error) {
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

func (s *ChatService) GetChatList(ctx context.Context, req *pb.GetChatListRequest) (*pb.GetChatListResponse, error) {
	projectID := req.GetProjectId()

	chats, err := s.dao.GetChatList(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch chat list: %v", err)
	}
	return &pb.GetChatListResponse{Chats: chats}, nil

}

func (s *ChatService) CreateChat(ctx context.Context, req *pb.CreateChatRequest) (*pb.CreateChatResponse, error) {
	name := req.Name
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	chatId := uuid.New().String()
	projectID := req.GetProjectId()

	err := s.dao.CreateChat(chatId, name, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert chat record: %w", err)
	}

	return &pb.CreateChatResponse{
		Message: "Chat created successfully",
		ChatId:  chatId,
	}, nil
}

func (s *ChatService) ListModel(ctx context.Context, req *pb.ListModelsRequest) (*pb.ListModelsResponse, error) {
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

func (s *ChatService) SearchChat(ctx context.Context, req *pb.ChatSearchRequest) (*pb.ChatSearchResponse, error) {
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

func (s *ChatService) CreateProject(ctx context.Context, req *pb.CreateProjectRequest) (*pb.CreateProjectResponse, error) {
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

func (s *ChatService) GetProjects(ctx context.Context, req *pb.GetProjectsRequest) (*pb.GetProjectsResponse, error) {
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

func (s *ChatService) ListDocuments(ctx context.Context, req *pb.ListDocumentsRequest) (*pb.ListDocumentsResponse, error) {
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

func (s *ChatService) retrieveSimilarChunks(ctx context.Context, projectID string, query string) (*rag.Response, error) {
	if projectID == "" || query == "" {
		return nil, fmt.Errorf("project_id and query are required")
	}

	// TODO: tech debt, need to refactor this
	embedding, err := s.embeddingsProvider.Embed(ctx, []rag.Chunk{
		{
			ID:        "0",
			ProjectID: projectID,
			DocsID:    "0",
			StartByte: 0,
			EndByte:   len(query),
			Text:      query,
		},
	})

	if err != nil {
		return nil, err
	}
	if len(embedding) == 0 {
		return nil, fmt.Errorf("embedding could not be created")
	}

	params := rag.SearchParams{TopK: 2, ProjectID: projectID}
	retriever := func(ctx context.Context, embedding []float64, params rag.SearchParams) ([]rag.Result, error) {
		embBytes, err := json.Marshal(embedding)
		if err != nil {
			return nil, err
		}
		vecRows, err := s.dao.GetTopSimilarRAGChunks(string(embBytes), projectID)
		if err != nil {
			return nil, err
		}
		var results []rag.Result
		for _, v := range vecRows {
			_, reader, err := s.store.GetObject(ctx, v.DocsID)
			if err != nil {
				return nil, fmt.Errorf("failed to get object for docsID %s: %w", v.DocsID, err)
			}
			data, err := io.ReadAll(reader)
			if err != nil {
				return nil, fmt.Errorf("failed to read object for docsID %s: %w", v.DocsID, err)
			}
			if v.StartByte < 0 || v.EndByte > len(data) || v.StartByte > v.EndByte {
				return nil, fmt.Errorf("invalid chunk byte range for docsID %s: %d-%d (file size %d)", v.DocsID, v.StartByte, v.EndByte, len(data))
			}
			chunkText := string(data[v.StartByte:v.EndByte])
			results = append(results, rag.Result{
				Chunk: rag.Chunk{
					ID:        v.ID,
					ProjectID: v.ProjectID,
					DocsID:    v.DocsID,
					StartByte: v.StartByte,
					EndByte:   v.EndByte,
					Text:      chunkText,
				},
				Similarity: 0,
			})
		}
		return results, nil
	}
	response, err := rag.BasicRetrievePipeline(ctx, retriever, rag.BasicPromptBuilder, embedding[0].Vector, query, params)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *ChatService) Init() {
	//db.InitDB()
	// TODO: handle migration for postgres also
	db.MigrateSQLite(SQLITE_DB_URL)
	db.SeedSqlite(SQLITE_DB_URL)
}

func (s *ChatService) EmbeddingSubscriber() {
	go func() {
		sub, err := s.queue.Subscribe(context.Background(), "generate.embedding")
		if err != nil {
			fmt.Printf("Failed %v\n", err)
			return
		}
		for msg := range sub {
			var payload GenerateEmbeddingMessage
			if err := json.Unmarshal(msg.Data, &payload); err == nil {

				// Fetch project_id for docs_id
				docMeta, err := s.dao.GetFileMetadata(payload.DocsID)
				if err != nil {
					fmt.Printf("Failed to fetch file metadata: %v\n", err)
					continue
				}

				filePath := "filestore/objects/" + payload.DocsID
				f, err := os.Open(filePath)
				if err != nil {
					fmt.Printf("Failed :%v\n", err)
					continue
				}
				defer f.Close()

				metadata := map[string]string{
					"project_id": docMeta.ProjectID,
					"docs_id":    payload.DocsID,
					"source":     docMeta.FileName,
				}

				result, err := s.pipeline.RunWithChunks(context.Background(), f, "text/plain", metadata)
				if err != nil {
					fmt.Printf("Pipeline error: %v\n", err)
					continue
				}

				// Save each chunk to rag_chunks
				for _, chunk := range result.Chunks {
					err := s.dao.SaveRAGChunk(chunk.ID, chunk.ProjectID, chunk.DocsID, chunk.StartByte, chunk.EndByte)
					if err != nil {
						fmt.Printf("Failed to save chunk: %v", err)
					}

					for _, emb := range result.Embeddings {
						if emb.ChunkID == chunk.ID {
							err := s.dao.SaveRAGChunkEmbedding(chunk.ID, emb.Vector)
							if err != nil {
								fmt.Printf("Failed to save embedding: %v\n", err)
							}
							break
						}
					}
				}
			}
		}
	}()
}
