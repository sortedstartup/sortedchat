package service

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptrace"
	"os"
	"strings"

	"sortedstartup/chatservice/dao"
	"sortedstartup/chatservice/events"
	pb "sortedstartup/chatservice/proto"
	"sortedstartup/chatservice/queue"
	"sortedstartup/chatservice/rag"
	settings "sortedstartup/chatservice/settings"
	"sortedstartup/chatservice/store"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"

	"github.com/google/uuid"
)

// RAGDocumentChunk represents a chunk within a document
type RAGDocumentChunk struct {
	StartByte  int     `json:"startByte"`
	EndByte    int     `json:"endByte"`
	ChunkText  string  `json:"chunk_text"`
	Similarity float64 `json:"similarity"`
}

// RAGDocumentJSON represents a document with its chunks for JSON storage
type RAGDocumentJSON struct {
	DocID  string             `json:"doc_id"`
	Chunks []RAGDocumentChunk `json:"chunks"`
}

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

const MAX_CHAT_NAME_LENGTH = 50
const MIN_CHAT_NAME_LENGTH = 1

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

	projectID := req.GetProjectContext().GetProjectId()
	ragEnabled := req.GetProjectContext().GetRagEnabled()

	apiKey := s.settingsManager.GetSettings().OpenAIAPIKey
	if apiKey == "" {
		return fmt.Errorf("OpenAI API key not set")
	}

	chatId := req.ChatId
	if chatId == "" {
		return fmt.Errorf("Chat ID is required to maintain context")
	}

	isDeleted, err := s.dao.IsChatDeleted(chatId, userID)
	if err != nil {
		return fmt.Errorf("error occured while checking chat id ")
	}

	if isDeleted {
		return fmt.Errorf("Chat is deleted, please create a new chat")
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

	userMessage := req.Text
	var ragChunks []rag.Result

	// First, check if this chat is in context of a project and retrieve similar chunks
	if projectID != "" && projectID != "null" && ragEnabled { // if this chat is in context of a project
		chunks, err := s.retrieveSimilarChunks(ctx, userID, projectID, req.Text)
		if err != nil {
			slog.Error("failed to retrieve similar chunks", "error", err)
		} else if len(chunks.Results) > 0 {
			userMessage = chunks.Prompt
			ragChunks = chunks.Results

			// Group chunks by document ID to create summary
			docChunksMap := make(map[string][]rag.Result)
			for _, result := range chunks.Results {
				docChunksMap[result.Chunk.DocsID] = append(docChunksMap[result.Chunk.DocsID], result)
			}

			// Create document reference summary list for streaming
			var summaries []*pb.RAGDocumentReferenceSummaryList_Summary
			for docsID, docChunks := range docChunksMap {
				// Get document metadata
				docMeta, err := s.dao.GetFileMetadata(docsID)
				if err != nil {
					slog.Error("failed to get document metadata", "error", err, "docs_id", docsID)
					continue
				}

				summary := &pb.RAGDocumentReferenceSummaryList_Summary{
					DocId:      docsID,
					FileName:   docMeta.FileName,
					ChunkCount: int32(len(docChunks)),
				}
				summaries = append(summaries, summary)
			}

			// Send document reference summary
			if len(summaries) > 0 {
				response := &pb.ChatResponse{
					Response: &pb.ChatResponse_DocumentReference{
						DocumentReference: &pb.RAGDocumentReferenceSummaryList{
							Summary: summaries,
						},
					},
				}

				if err := stream(response); err != nil {
					return fmt.Errorf("failed to send document reference summary: %v", err)
				}
			}
		}
	}

	// Save user message with RAG document references if available
	var requestMessageId string
	var referencesJSON string
	if len(ragChunks) > 0 {
		ragDocuments := s.createRAGDocumentJSONFromChunks(ragChunks)
		referencesBytes, err := json.Marshal(ragDocuments)
		if err != nil {
			slog.Error("failed to marshal RAG document references for user message", "error", err)
		} else {
			referencesJSON = string(referencesBytes)
		}
	}
	requestMessageId, err = s.dao.AddChatMessage(userID, chatId, "user", req.Text, model, 0, 0, 0, referencesJSON, ragEnabled)
	if err != nil {
		return fmt.Errorf("failed to insert user message: %v", err)
	}
	if err := stream(&pb.ChatResponse{
		Response: &pb.ChatResponse_RequestMessageId{
			RequestMessageId: requestMessageId,
		},
	}); err != nil {
		return fmt.Errorf("failed to send message summary: %v", err)
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

	stream(&pb.ChatResponse{
		Response: &pb.ChatResponse_Progress{Progress: &pb.ChatProgress{State: pb.ChatProgress_SENDING_REQUEST_TO_LLM, Message: "Request sent"}},
	})

	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.settingsManager.GetSettings().OpenAIAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	// UI can show request sent, useful because sometimes there is a delay from the API server
	stream(&pb.ChatResponse{
		Response: &pb.ChatResponse_Progress{Progress: &pb.ChatProgress{State: pb.ChatProgress_SENDING_REQUEST_TO_LLM, Message: "Request sent"}},
	})

	// This is awesome!, in go I was easily able to find out exactly when the request was sent
	trace := &httptrace.ClientTrace{
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			stream(&pb.ChatResponse{
				Response: &pb.ChatResponse_Progress{Progress: &pb.ChatProgress{State: pb.ChatProgress_REQUEST_SENT_TO_LLM, Message: ""}},
			})
		},
		GotFirstResponseByte: func() {
			stream(&pb.ChatResponse{
				Response: &pb.ChatResponse_Progress{Progress: &pb.ChatProgress{State: pb.ChatProgress_FIRST_RESPONSE_RECEIVED, Message: ""}},
			})

		},
	}

	httpReq = httpReq.WithContext(httptrace.WithClientTrace(httpReq.Context(), trace))
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("OpenAI request failed: %v", err)
	}
	defer resp.Body.Close()

	// UI can show request sent, useful because sometimes there is a delay from the API server

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("OpenAI API error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var fullResponse strings.Builder
	var inputTokens, outputTokens, cachedTokens int

	scanner := bufio.NewScanner(resp.Body)
	firstToken := true

	// Function to save partial response
	savePartialResponse := func() {
		assistantText := fullResponse.String()
		if assistantText != "" {
			var partialReferencesJSON string
			if len(ragChunks) > 0 {
				partialRagDocs := s.createRAGDocumentJSONFromChunks(ragChunks)
				partialRefsBytes, err := json.Marshal(partialRagDocs)
				if err != nil {
					slog.Error("failed to marshal RAG document references for partial response", "error", err)
				} else {
					partialReferencesJSON = string(partialRefsBytes)
				}
			}
			_, err := s.dao.AddChatMessageWithTokens(userID, chatId, "assistant", assistantText, model, inputTokens, outputTokens, cachedTokens, partialReferencesJSON, ragEnabled)
			if err != nil {
				slog.Error("failed to save partial assistant message", "error", err)
			} else {
				slog.Info("saved partial response to database", "length", len(assistantText))
			}
		}
	}

	for scanner.Scan() {
		if firstToken {
			stream(&pb.ChatResponse{
				Response: &pb.ChatResponse_Progress{Progress: &pb.ChatProgress{State: pb.ChatProgress_FIRST_TOKEN_RECEIVED, Message: "First token received"}},
			})
			firstToken = false
		}
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
			if promptTokensDetails, ok := usage["prompt_tokens_details"].(map[string]interface{}); ok {
				if cachedTokensVal, ok := promptTokensDetails["cached_tokens"].(float64); ok {
					cachedTokens = int(cachedTokensVal)
				}
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
				if err := stream(&pb.ChatResponse{Response: &pb.ChatResponse_Text{Text: content}}); err != nil {
					// If streaming fails, save partial response before returning error
					slog.Error("failed to send stream response, saving partial response", "error", err)
					savePartialResponse()
					return fmt.Errorf("failed to send stream response: %v", err)
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		if ctx.Err() != nil {
			slog.Info("streaming cancelled by client, saving partial response")
		} else {
			slog.Error("scanner error occurred, saving partial response", "error", err)
		}
		savePartialResponse() //save partial response
		return fmt.Errorf("error reading stream: %v", err)
	}

	err = stream(&pb.ChatResponse{
		Response: &pb.ChatResponse_Progress{Progress: &pb.ChatProgress{State: pb.ChatProgress_TOKENS_STOPPED, Message: "Response finished"}},
	})
	if err != nil {
		return fmt.Errorf("failed to send completion progress: %v", err)
	}

	// Normal completion - save full response
	assistantText := fullResponse.String()
	if assistantText != "" {
		var finalReferencesJSON string
		if len(ragChunks) > 0 {
			finalRagDocs := s.createRAGDocumentJSONFromChunks(ragChunks)
			finalRefsBytes, err := json.Marshal(finalRagDocs)
			if err != nil {
				slog.Error("failed to marshal RAG document references", "error", err)
			} else {
				finalReferencesJSON = string(finalRefsBytes)
			}
		}
		// TODO : scope for optimization, can be 1 sql call internally
		daoSummary, err := s.dao.AddChatMessageWithTokens(userID, chatId, "assistant", assistantText, model, inputTokens, outputTokens, cachedTokens, finalReferencesJSON, ragEnabled)
		if err != nil {
			log.Printf("Failed to insert assistant message: %v", err)
		} else {
			pbSummary := &pb.ResponseSummary{
				MessageId:    daoSummary.MessageId,
				Model:        model,
				InputTokens:  int32(daoSummary.InputTokenCount),
				OutputTokens: int32(daoSummary.OutputTokenCount),
				CachedTokens: int32(daoSummary.CachedTokenCount),
				Cost:         float32(daoSummary.Cost),
			}
			if err := stream(&pb.ChatResponse{Response: &pb.ChatResponse_Summary{Summary: pbSummary}}); err != nil {
				return fmt.Errorf("failed to send message summary: %v", err)
			}
		}
	}

	chatInfo, err := s.dao.GetChatMetadata(userID, chatId)
	if err != nil {
		return fmt.Errorf("failed to get chat metadata: %v", err)
	}

	chatInfoPb := &pb.ChatInfo{
		ChatId:           chatId,
		Cost:             float32(chatInfo.Cost),
		InputTokenCount:  int32(chatInfo.InputTokenCount),
		OutputTokenCount: int32(chatInfo.OutputTokenCount),
		CachedTokenCount: int32(chatInfo.CachedTokenCount),
	}

	if err := stream(&pb.ChatResponse{
		Response: &pb.ChatResponse_ChatMetadata{
			ChatMetadata: chatInfoPb,
		},
	}); err != nil {
		return fmt.Errorf("failed to send chat metadata: %v", err)
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

func (s *ChatService) GetHistory(ctx context.Context, userID string, chatId string) ([]*pb.ChatMessage, *pb.ChatInfo, error) {
	if chatId == "" {
		return nil, nil, fmt.Errorf("chat ID is required")
	}

	messages, err := s.dao.GetChatMessages(userID, chatId)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch history: %v", err)
	}

	var pbMessages []*pb.ChatMessage
	for _, m := range messages {
		pbMessage := &pb.ChatMessage{
			Role:         m.Role,
			Content:      m.Content,
			MessageId:    m.Id,
			RagEnabled:   m.RagEnabled,
			Model:        m.Model,
			InputTokens:  int32(m.InputTokenCount),
			OutputTokens: int32(m.OutputTokenCount),
			CachedTokens: int32(m.CachedTokenCount),
			Cost:         float32(m.Cost),
		}

		if m.DocumentReferences != "" {
			var ragDocuments []RAGDocumentJSON
			if err := json.Unmarshal([]byte(m.DocumentReferences), &ragDocuments); err == nil {
				// Convert RAGDocumentJSON to pb.RAGDocumentReference
				var references []*pb.RAGDocumentReference
				for _, doc := range ragDocuments {
					// Get document metadata
					docMeta, err := s.dao.GetFileMetadata(doc.DocID)
					if err != nil {
						slog.Error("failed to get document metadata in history", "error", err, "docs_id", doc.DocID)
						continue
					}

					// Convert chunks to proto format
					var protoChunks []*pb.RAGDocumentReference_Chunk
					for _, chunk := range doc.Chunks {
						protoChunks = append(protoChunks, &pb.RAGDocumentReference_Chunk{
							ChunkText:   chunk.ChunkText,
							StartByte:   int32(chunk.StartByte),
							EndByte:     int32(chunk.EndByte),
							Simillarity: float32(chunk.Similarity),
						})
					}

					reference := &pb.RAGDocumentReference{
						DocId:    doc.DocID,
						FileName: docMeta.FileName,
						Chunks:   protoChunks,
					}
					references = append(references, reference)
				}
				pbMessage.References = references
			}
		}

		pbMessages = append(pbMessages, pbMessage)
	}

	chatInfo, err := s.dao.GetChatMetadata(userID, chatId)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get chat metadata: %v", err)
	}
	pbChatInfo := &pb.ChatInfo{
		ChatId:           chatId,
		Cost:             float32(chatInfo.Cost),
		InputTokenCount:  int32(chatInfo.InputTokenCount),
		OutputTokenCount: int32(chatInfo.OutputTokenCount),
		CachedTokenCount: int32(chatInfo.CachedTokenCount),
	}

	return pbMessages, pbChatInfo, nil
}

func (s *ChatService) GetChatList(ctx context.Context, userID string, projectID string, soft_deleted bool) ([]*pb.ChatInfo, error) {
	chats, err := s.dao.GetChatList(userID, projectID, soft_deleted)
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

	return models, nil
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
				Similarity: v.Similarity * 100,
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
					err := s.dao.SaveRAGChunk(docMeta.User, chunk.ID, chunk.ProjectID, chunk.DocsID, chunk.StartByte, chunk.EndByte)
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

// createRAGDocumentJSONFromChunks converts RAG chunks to the requested JSON structure
func (s *ChatService) createRAGDocumentJSONFromChunks(ragChunks []rag.Result) []RAGDocumentJSON {
	// Group chunks by document ID
	docChunksMap := make(map[string][]RAGDocumentChunk)

	for _, result := range ragChunks {
		chunk := RAGDocumentChunk{
			StartByte:  result.Chunk.StartByte,
			EndByte:    result.Chunk.EndByte,
			ChunkText:  strings.ToValidUTF8(result.Chunk.Text, ""),
			Similarity: result.Similarity,
		}
		docChunksMap[result.Chunk.DocsID] = append(docChunksMap[result.Chunk.DocsID], chunk)
	}

	// Convert to the final structure
	var ragDocuments []RAGDocumentJSON
	for docID, chunks := range docChunksMap {
		ragDocuments = append(ragDocuments, RAGDocumentJSON{
			DocID:  docID,
			Chunks: chunks,
		})
	}

	return ragDocuments
}

// GetRAGDocumentReference retrieves RAG document references for a specific message
func (s *ChatService) GetRAGDocumentReference(ctx context.Context, userID string, req *pb.RAGDocumentReferenceRequest) (*pb.RAGDocumentReferenceResponse, error) {

	// Validate request
	if req.MessageId == "" {
		return nil, fmt.Errorf("message_id is required")
	}

	// Get the chat message by ID
	message, err := s.dao.GetChatMessageByID(userID, req.MessageId)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %v", err)
	}

	// Parse the document references JSON
	if message.DocumentReferences == "" {
		return &pb.RAGDocumentReferenceResponse{
			Reference: nil,
		}, nil
	}

	var ragDocuments []RAGDocumentJSON
	if err := json.Unmarshal([]byte(message.DocumentReferences), &ragDocuments); err != nil {
		return nil, fmt.Errorf("failed to parse document references: %v", err)
	}

	// If docId is specified, filter for that specific document
	if req.DocId != "" {
		for _, doc := range ragDocuments {
			if doc.DocID == req.DocId {
				// Get document metadata
				docMeta, err := s.dao.GetFileMetadata(doc.DocID)
				if err != nil {
					return nil, fmt.Errorf("failed to get document metadata: %v", err)
				}

				// Convert chunks to proto format
				var protoChunks []*pb.RAGDocumentReference_Chunk
				for _, chunk := range doc.Chunks {
					protoChunks = append(protoChunks, &pb.RAGDocumentReference_Chunk{
						ChunkText:   chunk.ChunkText,
						StartByte:   int32(chunk.StartByte),
						EndByte:     int32(chunk.EndByte),
						Simillarity: float32(chunk.Similarity),
					})
				}

				return &pb.RAGDocumentReferenceResponse{
					Reference: &pb.RAGDocumentReference{
						DocId:    doc.DocID,
						FileName: docMeta.FileName,
						Chunks:   protoChunks,
					},
				}, nil
			}
		}
		// Document not found - return empty result
		return &pb.RAGDocumentReferenceResponse{
			Reference: nil,
		}, nil
	}

	// If no specific docId, return the first document (or you could return all)
	if len(ragDocuments) > 0 {
		doc := ragDocuments[0]

		// Get document metadata
		docMeta, err := s.dao.GetFileMetadata(doc.DocID)
		if err != nil {
			return nil, fmt.Errorf("failed to get document metadata: %v", err)
		}

		// Convert chunks to proto format
		var protoChunks []*pb.RAGDocumentReference_Chunk
		for _, chunk := range doc.Chunks {
			protoChunks = append(protoChunks, &pb.RAGDocumentReference_Chunk{
				ChunkText:   chunk.ChunkText,
				StartByte:   int32(chunk.StartByte),
				EndByte:     int32(chunk.EndByte),
				Simillarity: float32(chunk.Similarity),
			})
		}

		return &pb.RAGDocumentReferenceResponse{
			Reference: &pb.RAGDocumentReference{
				DocId:    doc.DocID,
				FileName: docMeta.FileName,
				Chunks:   protoChunks,
			},
		}, nil
	}

	return &pb.RAGDocumentReferenceResponse{
		Reference: nil,
	}, nil
}

func (s *ChatService) DeleteDocument(ctx context.Context, userID string, projectID string, docID string) error {
	if projectID == "" || docID == "" {
		return fmt.Errorf("project_id and doc_id are required")
	}

	err := s.dao.DeleteDocument(userID, projectID, docID)
	if err != nil {
		return fmt.Errorf("failed to delete document: %v", err)
	}

	//TODO: What if this operation fails?
	err = s.store.DeleteObject(ctx, docID)
	if err != nil {
		return fmt.Errorf("failed to delete object: %v", err)
	}

	return nil
}

func (s *ChatService) DeleteChat(ctx context.Context, userID string, chatId string, operation pb.DeleteChatRequest_Operation) error {
	if chatId == "" {
		return fmt.Errorf("chat ID is required")
	}

	switch operation {
	case pb.DeleteChatRequest_DELETE:
		if err := s.dao.DeleteChat(userID, chatId); err != nil {
			return fmt.Errorf("failed to delete chat: %v", err)
		}
	case pb.DeleteChatRequest_SOFT_DELETE:
		if err := s.dao.SoftDeleteChat(userID, chatId); err != nil {
			return fmt.Errorf("failed to delete chat: %v", err)
		}
	default:
		return fmt.Errorf("unsupported delete operation: %v", operation)
	}

	return nil
}

func (s *ChatService) RestoreChat(ctx context.Context, userID string, chatId string) error {
	if chatId == "" {
		return fmt.Errorf("chat ID is required")
	}

	err := s.dao.RestoreChat(userID, chatId)
	if err != nil {
		return fmt.Errorf("failed to restore chat: %v", err)
	}

	return nil
}

func (s *ChatService) RenameChat(ctx context.Context, userID string, chatId string, name string) error {
	if chatId == "" {
		return fmt.Errorf("chat ID is required")
	}

	trimmedName := strings.TrimSpace(name)

	if len(trimmedName) < MIN_CHAT_NAME_LENGTH {
		return fmt.Errorf("name must be at least %d characters", MIN_CHAT_NAME_LENGTH)
	}

	if len(trimmedName) > MAX_CHAT_NAME_LENGTH {
		return fmt.Errorf("name must be less than %d characters", MAX_CHAT_NAME_LENGTH)
	}

	err := s.dao.RenameChat(userID, chatId, trimmedName)
	if err != nil {
		return fmt.Errorf("failed to rename chat: %v", err)
	}
	return nil
}

func (s *ChatService) Init(config *dao.Config) *sql.DB {

	//for sqlite we pass db connnection to migrate and seed functions
	//for postgres we pass dsn(URL) to migrate and seed functions
	switch config.Database.Type {
	case dao.DatabaseTypeSQLite:
		//Create DB and run migrations
		sqlite_vec.Auto()
		sqlDB, err := sql.Open("sqlite3", config.Database.SQLite.URL)
		if err != nil {
			log.Fatalf("failed to open database: %v", err)
		}
		// defer sqlDB.Close() //lets not close it here
		slog.Info("ChatService: Running SQLite migrations")
		if err := dao.MigrateDB_UsingConnectionDefaults(sqlDB); err != nil {
			log.Fatalf("ChatService: Failed to migrate SQLite database: %v", err)
		}
		if err := dao.SeedDB_UsingConnectionDefaults(sqlDB); err != nil {
			log.Fatalf("ChatService: Failed to seed SQLite database: %v", err)
		}
		return sqlDB
	case dao.DatabaseTypePostgres:
		slog.Info("ChatService: Running PostgreSQL migrations")
		dsn := config.Database.Postgres.GetPostgresDSN()
		if err := dao.MigratePostgres(dsn); err != nil {
			log.Fatalf("ChatService: Failed to migrate PostgreSQL database: %v", err)
		}
		if err := dao.SeedPostgres(dsn); err != nil {
			log.Fatalf("ChatService: Failed to seed PostgreSQL database: %v", err)
		}
	default:
		log.Fatalf("ChatService: Unsupported database type: %s", config.Database.Type)
	}
	return nil
}
