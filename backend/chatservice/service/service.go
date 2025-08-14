package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"sortedstartup/chatservice/dao"
	"sortedstartup/chatservice/events"
	pb "sortedstartup/chatservice/proto"
	"sortedstartup/chatservice/queue"
	"sortedstartup/chatservice/rag"
	settings "sortedstartup/chatservice/settings"
	"sortedstartup/chatservice/store"

	"github.com/google/uuid"
)

type ChatService struct {
	dao                dao.DAO
	settingsDAO        dao.SettingsDAO
	store              *store.DiskObjectStore
	queue              queue.Queue
	pipeline           rag.RAGIndexingPipeline
	embeddingsProvider rag.Embedder
	settingsManager    *settings.SettingsManager
}

type GenerateEmbeddingMessage struct {
	DocsID string `json:"docs_id"`
}

func NewChatService(queue queue.Queue, settingsManager *settings.SettingsManager, daoFactory dao.DAOFactory) (*ChatService, error) {
	daoInstance, err := daoFactory.CreateDAO()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize DAO: %v", err)
	}

	settingsDAOInstance, err := daoFactory.CreateSettingsDAO()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize settings DAO: %v", err)
	}

	storeInstance, err := store.NewDiskObjectStore("filestore")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize object store: %v", err)
	}

	embeddingsProvider := &rag.OLLamaEmbedder{
		SettingsManager: settingsManager,
		Model:           "nomic-embed-text",
	}

	pipeline := rag.NewPipeline(
		&rag.TextExtractor{},
		&rag.EqualSizeChunker{ChunkSize: 512},
		embeddingsProvider,
	)

	return &ChatService{
		dao:                daoInstance,
		settingsDAO:        settingsDAOInstance,
		store:              storeInstance,
		queue:              queue,
		pipeline:           pipeline,
		embeddingsProvider: embeddingsProvider,
		settingsManager:    settingsManager,
	}, nil
}

func (s *ChatService) Chat(ctx context.Context, userID string, req *pb.ChatRequest, stream func(*pb.ChatResponse) error) error {
	projectID := req.GetProjectId()

	apiKey := s.settingsManager.GetSettings().OpenAIAPIKey
	if apiKey == "" {
		return fmt.Errorf("OpenAI API key not set")
	}

	chatId := req.ChatId
	if chatId == "" {
		return fmt.Errorf("Chat ID is required to maintain context")
	}

	model := req.Model
	if model == "" {
		return fmt.Errorf("model is required")
	}

	// Get chat history using DAO
	history, err := s.dao.GetChatMessages(userID, chatId)
	if err != nil {
		slog.Error("failed to fetch message history", "error", err)
		return fmt.Errorf("failed to fetch message history: %v", err)
	}

	err = s.dao.AddChatMessage(userID, chatId, "user", req.Text)
	if err != nil {
		return fmt.Errorf("failed to insert user message: %v", err)
	}

	userMessage := req.Text

	if projectID != "" && projectID != "null" { // if this chat is in context of a project
		chunks, err := s.retrieveSimilarChunks(ctx, userID, projectID, req.Text)
		if err != nil {
			slog.Error("failed to retrieve similar chunks", "error", err)
		} else if len(chunks.Results) > 0 {
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

	httpReq, err := http.NewRequest("POST", s.settingsManager.GetSettings().OpenAIAPIURL, bytes.NewBuffer(jsonData))
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

				if err := stream(&pb.ChatResponse{Response: &pb.ChatResponse_Text{
					Text: content,
				}}); err != nil {
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
		messageId, err := s.dao.AddChatMessageWithTokens(userID, chatId, "assistant", assistantText, model, inputTokens, outputTokens)
		if err != nil {
			log.Printf("Failed to insert assistant message: %v", err)
		} else {
			summary := &pb.MessageSummary{
				MessageId: fmt.Sprintf("%d", messageId),
			}
			if err := stream(&pb.ChatResponse{
				Response: &pb.ChatResponse_Summary{
					Summary: summary,
				},
			}); err != nil {
				return fmt.Errorf("failed to send message summary: %v", err)
			}
		}
	}

	return nil
}

const (
	MAX_MESSAGE_LENGTH   = 500
	START_MESSAGE_LENGTH = 250
	END_MESSAGE_LENGTH   = 250
)

func (s *ChatService) GenerateChatName(ctx context.Context, userID string, chatId string, message string, model string) (string, error) {
	if chatId == "" {
		return "", fmt.Errorf("chat ID is required")
	}

	if message == "" {
		return "", fmt.Errorf("message is required")
	}

	if model == "" {
		return "", fmt.Errorf("model is required")
	}

	apiKey := s.settingsManager.GetSettings().OpenAIAPIKey
	if apiKey == "" {
		return "", fmt.Errorf("OpenAI API key not set")
	}

	name, err := s.dao.GetChatName(userID, chatId)
	if err != nil {
		return "", fmt.Errorf("failed to get chat name: %v", err)
	}

	if name != "" {
		return "", fmt.Errorf("Chat name already exists: %s", name)
	}

	words := strings.Fields(message)
	if len(words) > MAX_MESSAGE_LENGTH {
		start := strings.Join(words[:START_MESSAGE_LENGTH], " ")
		end := strings.Join(words[len(words)-END_MESSAGE_LENGTH:], " ")
		message = start + end
	}

	prompt := "Based on the given user message give me a most appropriate chat name of 1-5 word length: " + message

	requestBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"stream": false,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	httpReq, err := http.NewRequest("POST", s.settingsManager.GetSettings().OpenAIAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("OpenAI request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI API error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(respBody, &openAIResp); err != nil {
		return "", fmt.Errorf("failed to parse OpenAI response: %v", err)
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from OpenAI")
	}

	chatName := openAIResp.Choices[0].Message.Content

	if err := s.dao.SaveChatName(userID, chatId, chatName); err != nil {
		return "", fmt.Errorf("error while saving name: %v", err)
	}

	return chatName, nil
}

func (s *ChatService) GetHistory(ctx context.Context, userID string, chatId string) ([]*pb.ChatMessage, error) {
	if chatId == "" {
		return nil, fmt.Errorf("chat ID is required")
	}

	messages, err := s.dao.GetChatMessages(userID, chatId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch history: %v", err)
	}

	var pbMessages []*pb.ChatMessage
	for _, m := range messages {
		pbMessages = append(pbMessages, &pb.ChatMessage{
			Role:      m.Role,
			Content:   m.Content,
			MessageId: m.Id,
		})
	}

	return pbMessages, nil
}

func (s *ChatService) GetChatList(ctx context.Context, userID string, projectID string) ([]*pb.ChatInfo, error) {
	chats, err := s.dao.GetChatList(userID, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch chat list: %v", err)
	}
	return chats, nil
}

func (s *ChatService) CreateChat(ctx context.Context, userID string, name string, projectID string) (string, error) {
	chatId := uuid.New().String()

	err := s.dao.CreateChat(userID, chatId, name, projectID)
	if err != nil {
		return "", fmt.Errorf("failed to insert chat record: %w", err)
	}

	return chatId, nil
}

func (s *ChatService) ListModel(ctx context.Context) ([]*pb.ModelListInfo, error) {
	models, err := s.dao.GetModels()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch models: %v", err)
	}

	pbModels := make([]*pb.ModelListInfo, 0, len(models))
	for i := range models {
		pbModels = append(pbModels, &pb.ModelListInfo{
			Id:    models[i].Id,
			Label: models[i].Label,
		})
	}

	return pbModels, nil
}

func (s *ChatService) SearchChat(ctx context.Context, userID string, query string) ([]*pb.SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	results, err := s.dao.SearchChatMessages(userID, query)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	var pbResults []*pb.SearchResult
	for i := range results {
		pbResults = append(pbResults, &pb.SearchResult{
			ChatId:      results[i].ChatId,
			ChatName:    results[i].ChatName,
			MatchedText: results[i].MatchedText,
		})
	}

	return pbResults, nil
}

func (s *ChatService) CreateProject(ctx context.Context, userID string, name string, description string, additionalData string) (string, error) {
	id := uuid.New().String()

	if name == "" {
		return "", fmt.Errorf("name is required")
	}

	projectID, err := s.dao.CreateProject(userID, id, name, description, additionalData)
	if err != nil {
		return "", fmt.Errorf("failed to create project: %w", err)
	}

	return projectID, nil
}

func (s *ChatService) GetProjects(ctx context.Context, userID string) ([]dao.ProjectRow, error) {
	projects, err := s.dao.GetProjects(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch project list: %w", err)
	}

	return projects, nil
}

func (s *ChatService) ListDocuments(ctx context.Context, userID string, projectID string) ([]dao.DocumentListRow, error) {
	docs, err := s.dao.FilesList(userID, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch documents: %v", err)
	}

	return docs, nil
}

func (s *ChatService) UploadFile(ctx context.Context, userID string, projectID string, file multipart.File, header *multipart.FileHeader, maxFileSize int64, maxProjectSize int64) (string, error) {
	if projectID == "" {
		return "", fmt.Errorf("project_id is required")
	}

	fileSize := header.Size
	if fileSize > maxFileSize {
		return "", fmt.Errorf("file exceeds %d MB limit", maxFileSize/(1024*1024))
	}

	// Check total project size
	totalUsed, err := s.dao.TotalUsedSize(userID, projectID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch usage: %v", err)
	}

	if totalUsed+fileSize > maxProjectSize {
		return "", fmt.Errorf("project storage exceeds %d MB", maxProjectSize/(1024*1024))
	}

	// Generate object ID and store file
	objectID := uuid.New().String()

	if err := s.store.StoreObject(ctx, objectID, file); err != nil {
		return "", fmt.Errorf("failed to store file: %v", err)
	}

	// Save file metadata to database
	if err := s.dao.FileSave(userID, projectID, objectID, header.Filename, fileSize); err != nil {
		return "", fmt.Errorf("failed to save metadata: %v", err)
	}

	// Publish embedding generation event
	msg := GenerateEmbeddingMessage{DocsID: objectID}
	msgBytes, _ := json.Marshal(msg)
	err = s.queue.Publish(ctx, events.GENERATE_EMBEDDINGS, msgBytes)
	if err != nil {
		// Log error but don't fail the upload
		log.Printf("Failed to publish embedding generation event: %v", err)
	}

	return objectID, nil
}

func (s *ChatService) retrieveSimilarChunks(ctx context.Context, userID string, projectID string, query string) (*rag.Response, error) {
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
		vecRows, err := s.dao.GetTopSimilarRAGChunks(userID, string(embBytes), projectID)
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

func (s *ChatService) SubmitGenerateEmbeddingsJob(ctx context.Context, userID string, projectID string) error {
	if projectID == "" {
		return fmt.Errorf("project_id is required")
	}

	docs, error := s.dao.FetchErrorDocs(userID, projectID)
	if error != nil {
		return fmt.Errorf("failed to check embedding status: %v", error)
	}

	for _, docsID := range docs {
		msg := GenerateEmbeddingMessage{DocsID: docsID}
		msgBytes, _ := json.Marshal(msg)
		err := s.queue.Publish(ctx, "generate.embedding", msgBytes)
		if err != nil {
			return fmt.Errorf("failed to publish job: %v", err)
		}

		if updateErr := s.dao.UpdateEmbeddingStatus(docsID, int32(pb.Embedding_Status_STATUS_QUEUED)); updateErr != nil {
			fmt.Printf("Failed to update embedding status to error: %v\n", updateErr)
		}
	}

	return nil
}

func (s *ChatService) BranchAChat(ctx context.Context, userID string, sourceChatId string, branchFromMessageId string, branchName string) (string, error) {
	if sourceChatId == "" {
		return "", fmt.Errorf("parent id is required")
	}

	if branchFromMessageId == "" {
		return "", fmt.Errorf("message id is required")
	}

	isMain, err := s.dao.IsMainBranch(userID, sourceChatId)
	if err != nil || !isMain {
		return "", fmt.Errorf("can only branch from main branch chats")
	}

	newChatId := uuid.New().String()

	err = s.dao.BranchChat(userID, sourceChatId, branchFromMessageId, newChatId, branchName)
	if err != nil {
		return "", fmt.Errorf("failed to create branch: %v", err)
	}

	return newChatId, nil
}

func (s *ChatService) ListChatBranch(ctx context.Context, userID string, chatId string) ([]dao.ChatInfoRow, error) {
	if chatId == "" {
		return nil, fmt.Errorf("Chat Id is required")
	}

	isMain, err := s.dao.IsMainBranch(userID, chatId)
	if err != nil {
		return nil, fmt.Errorf("cannot identify chat id: %w", err)
	}

	innerChats, err := s.dao.GetChatBranches(userID, chatId, isMain)
	if err != nil {
		return nil, fmt.Errorf("failed to get inner chat list: %w", err)
	}

	return innerChats, nil
}

func (s *ChatService) EmbeddingSubscriber() {
	go func() {
		sub, err := s.queue.Subscribe(context.Background(), events.GENERATE_EMBEDDINGS)
		if err != nil {
			fmt.Printf("Failed %v\n", err)
			return
		}

		for msg := range sub {
			var payload GenerateEmbeddingMessage
			if err := json.Unmarshal(msg.Data, &payload); err == nil {

				if updateErr := s.dao.UpdateEmbeddingStatus(payload.DocsID, int32(pb.Embedding_Status_STATUS_IN_PROGRESS)); updateErr != nil {
					fmt.Printf("Failed to update embedding status to error: %v\n", updateErr)
				}

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

				metadata := map[string]string{
					"project_id": docMeta.ProjectID,
					"docs_id":    payload.DocsID,
					"source":     docMeta.FileName,
				}

				result, err := s.pipeline.RunWithChunks(context.Background(), f, "text/plain", metadata)
				if err != nil {
					fmt.Printf("Pipeline error: %v\n", err)
					if updateErr := s.dao.UpdateEmbeddingStatus(payload.DocsID, int32(pb.Embedding_Status_STATUS_ERROR)); updateErr != nil {
						fmt.Printf("Failed to update embedding status to error: %v\n", updateErr)
					}
					continue
				}

				embeddingMap := make(map[string]rag.Embedding, len(result.Embeddings))
				for i := range result.Embeddings {
					embeddingMap[result.Embeddings[i].ChunkID] = result.Embeddings[i]
				}
				for _, chunk := range result.Chunks {
					userID := "0" // TODO: Get actual user_id from document metadata when user system is fully implemented
					err := s.dao.SaveRAGChunk(userID, chunk.ID, chunk.ProjectID, chunk.DocsID, chunk.StartByte, chunk.EndByte)
					if err != nil {
						fmt.Printf("Failed to save chunk: %v", err)
					}

					if emb, ok := embeddingMap[chunk.ID]; ok {
						if err := s.dao.SaveRAGChunkEmbedding(chunk.ID, emb.Vector); err != nil {
							fmt.Printf("Failed to save embedding: %v\n", err)
						}
					}
				}
				if updateErr := s.dao.UpdateEmbeddingStatus(payload.DocsID, int32(pb.Embedding_Status_STATUS_SUCCESS)); updateErr != nil {
					fmt.Printf("Failed to update embedding status to success: %v\n", updateErr)
				}
			}
		}
	}()
}
