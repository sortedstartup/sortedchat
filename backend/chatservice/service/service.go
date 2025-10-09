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
	"time"

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
	slog.Info("service:Chat", "userID", userID, "chatId", req.GetChatId(), "model", req.GetModel())

	projectID := req.GetProjectContext().GetProjectId()
	ragEnabled := req.GetProjectContext().GetRagEnabled()

	apiKey := s.settingsManager.GetSettings().OpenAIAPIKey
	if apiKey == "" {
		slog.Error("service:Chat", "error", "OpenAI API key not set")
		return fmt.Errorf("OpenAI API key not set")
	}

	chatId := req.ChatId
	if chatId == "" {
		slog.Error("service:Chat", "error", "Chat ID is required to maintain context")
		return fmt.Errorf("Chat ID is required to maintain context")
	}

	isDeleted, err := s.dao.IsChatDeleted(chatId, userID)
	if err != nil {
		slog.Error("service:Chat", "error", "failed to get chat status", "error", err, "chatId", chatId, "userID", userID)
		return fmt.Errorf("failed to get chat status")
	}

	if isDeleted {
		slog.Error("service:Chat", "error", "Chat is deleted, please create a new chat", "chatId", chatId, "userID", userID)
		return fmt.Errorf("Chat is deleted, please create a new chat")
	}

	model := req.Model
	if model == "" {
		slog.Error("service:Chat", "error", "model is required", "chatId", chatId, "userID", userID)
		return fmt.Errorf("model is required")
	}

	// Get chat history using DAO
	history, err := s.dao.GetChatMessages(userID, chatId)
	if err != nil {
		slog.Error("service:Chat", "message", "failed to fetch message history", "error", err, "chatId", chatId, "userID", userID)
		return fmt.Errorf("failed to fetch chat message")
	}

	userMessage := req.Text
	var ragChunks []rag.Result

	// First, check if this chat is in context of a project and retrieve similar chunks
	if projectID != "" && projectID != "null" && ragEnabled { // if this chat is in context of a project
		chunks, err := s.retrieveSimilarChunks(ctx, userID, projectID, req.Text)
		if err != nil {
			slog.Error("service:Chat", "message", "failed to retrieve similar chunks", "error", err, "chatId", chatId, "userID", userID, "projectID", projectID)
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
					slog.Error("service:Chat", "message", "failed to get document metadata", "error", err, "docs_id", docsID)
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
					slog.Error("service:Chat", "message", "failed to send document reference summary", "error", err, "chatId", chatId, "userID", userID, "projectID", projectID)
					return fmt.Errorf("failed to send document reference summary")
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
			slog.Error("failed to marshal RAG document references for user message", "error", err, "chatId", chatId, "userID", userID, "projectID", projectID)
		} else {
			referencesJSON = string(referencesBytes)
		}
	}
	requestMessageId, err = s.dao.AddChatMessage(userID, chatId, "user", req.Text, model, 0, 0, 0, referencesJSON, ragEnabled)
	if err != nil {
		slog.Error("service:Chat", "error", "failed to insert user message", "error", err, "chatId", chatId, "userID", userID, "projectID", projectID)
		return fmt.Errorf("failed to insert user message")
	}
	if err := stream(&pb.ChatResponse{
		Response: &pb.ChatResponse_RequestMessageId{
			RequestMessageId: requestMessageId,
		},
	}); err != nil {
		slog.Error("service:Chat", "error", "failed to send message summary", "error", err, "chatId", chatId, "userID", userID, "projectID", projectID)
		return fmt.Errorf("error while processing request, please try again")
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
		slog.Error("service:Chat", "message", "failed to marshal request", "error", err, "chatId", chatId, "userID", userID, "projectID", projectID)
		return fmt.Errorf("error while processing request, please try again")
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.settingsManager.GetSettings().OpenAIAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		slog.Error("service:Chat", "message", "failed to create request", "error", err, "chatId", chatId, "userID", userID, "projectID", projectID)
		return fmt.Errorf("failed to create request, please try again")
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
			if err := stream(&pb.ChatResponse{
				Response: &pb.ChatResponse_Progress{
					Progress: &pb.ChatProgress{State: pb.ChatProgress_REQUEST_SENT_TO_LLM, Message: ""}},
			}); err != nil {
				slog.Error("service:Chat", "message", "failed to send progress (sent)", "error", err, "chatId", chatId, "userID", userID, "projectID", projectID)
			}
		},
		GotFirstResponseByte: func() {
			if err := stream(&pb.ChatResponse{
				Response: &pb.ChatResponse_Progress{
					Progress: &pb.ChatProgress{State: pb.ChatProgress_FIRST_RESPONSE_RECEIVED, Message: ""}},
			}); err != nil {
				slog.Error("service:Chat", "message", "failed to send progress (first byte)", "error", err, "chatId", chatId, "userID", userID, "projectID", projectID)
			}

		},
	}

	httpReq = httpReq.WithContext(httptrace.WithClientTrace(httpReq.Context(), trace))
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		slog.Error("service:Chat", "message", "OpenAI request failed", "error", err, "chatId", chatId, "userID", userID, "projectID", projectID)
		return fmt.Errorf("OpenAI request failed, please try again")
	}
	defer resp.Body.Close()

	// UI can show request sent, useful because sometimes there is a delay from the API server

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		slog.Error("service:Chat", "message", "OpenAI API error", "status", resp.StatusCode, "body", string(bodyBytes), "chatId", chatId, "userID", userID, "projectID", projectID, "bodyBytes", string(bodyBytes))
		return fmt.Errorf("OpenAI API error, please try again")
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
					slog.Error("service:Chat", "message", "failed to marshal RAG document references for partial response", "error", err, "chatId", chatId, "userID", userID, "projectID", projectID)
				} else {
					partialReferencesJSON = string(partialRefsBytes)
				}
			}
			nonCachedInputTokens := inputTokens - cachedTokens
			if nonCachedInputTokens < 0 {
				slog.Warn("service:Chat", "message", "cachedTokens > inputTokens, setting non-cached input tokens to 0", "inputTokens", inputTokens, "cachedTokens", cachedTokens, "chatId", chatId, "userID", userID, "projectID", projectID)
				nonCachedInputTokens = 0
			}
			_, err := s.dao.AddChatMessageWithTokens(userID, chatId, "assistant", assistantText, model, nonCachedInputTokens, outputTokens, cachedTokens, partialReferencesJSON, ragEnabled)
			if err != nil {
				slog.Error("service:Chat", "message", "failed to save partial assistant message", "error", err, "chatId", chatId, "userID", userID, "projectID", projectID)
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
			slog.Error("service:Chat", "message", "failed to parse chunk", "error", err, "chatId", chatId, "userID", userID, "projectID", projectID)
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
					slog.Error("service:Chat", "message", "failed to send stream response", "error", err, "chatId", chatId, "userID", userID, "projectID", projectID)
					savePartialResponse()
					return fmt.Errorf("failed to send stream response, please try again")
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		if ctx.Err() != nil {
			slog.Info("service:Chat", "message", "streaming cancelled by client, saving partial response", "error", err, "chatId", chatId, "userID", userID, "projectID", projectID)
		} else {
			slog.Error("service:Chat", "message", "scanner error occurred, saving partial response", "error", err, "chatId", chatId, "userID", userID, "projectID", projectID)
		}
		savePartialResponse() //save partial response
		return fmt.Errorf("error while processing request, please try again")
	}

	err = stream(&pb.ChatResponse{
		Response: &pb.ChatResponse_Progress{Progress: &pb.ChatProgress{State: pb.ChatProgress_TOKENS_STOPPED, Message: "Response finished"}},
	})
	if err != nil {
		slog.Error("service:Chat", "message", "failed to send completion progress", "error", err, "chatId", chatId, "userID", userID, "projectID", projectID)
		return fmt.Errorf("error while processing request, please try again")
	}

	// Normal completion - save full response
	assistantText := fullResponse.String()
	if assistantText != "" {
		var finalReferencesJSON string
		if len(ragChunks) > 0 {
			finalRagDocs := s.createRAGDocumentJSONFromChunks(ragChunks)
			finalRefsBytes, err := json.Marshal(finalRagDocs)
			if err != nil {
				slog.Error("service:Chat", "message", "failed to marshal RAG document references", "error", err, "chatId", chatId, "userID", userID, "projectID", projectID)
			} else {
				finalReferencesJSON = string(finalRefsBytes)
			}
		}
		nonCachedInputTokens := inputTokens - cachedTokens
		if nonCachedInputTokens < 0 {
			slog.Warn("service:Chat", "message", "cachedTokens > inputTokens, setting non-cached input tokens to 0", "inputTokens", inputTokens, "cachedTokens", cachedTokens, "chatId", chatId, "userID", userID, "projectID", projectID)
			nonCachedInputTokens = 0
		}
		// TODO : scope for optimization, can be 1 sql call internally
		daoSummary, err := s.dao.AddChatMessageWithTokens(userID, chatId, "assistant", assistantText, model, nonCachedInputTokens, outputTokens, cachedTokens, finalReferencesJSON, ragEnabled)
		if err != nil {
			slog.Error("service:Chat", "message", "failed to insert assistant message", "error", err, "chatId", chatId, "userID", userID, "projectID", projectID)
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
				slog.Error("service:Chat", "message", "failed to send message summary", "error", err, "chatId", chatId, "userID", userID, "projectID", projectID)
				return fmt.Errorf("failed to send message summary, please try again")
			}
		}
	}

	chatInfo, err := s.dao.GetChatMetadata(userID, chatId)
	if err != nil {
		slog.Error("service:Chat", "message", "failed to get chat metadata", "error", err, "chatId", chatId, "userID", userID, "projectID", projectID)
		return fmt.Errorf("error while processing request, please try again")
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
		slog.Error("service:Chat", "message", "failed to send chat metadata", "error", err, "chatId", chatId, "userID", userID, "projectID", projectID)
		return fmt.Errorf("error while processing request, please try again")
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
		slog.Error("service:GenerateChatName", "message", "chat ID is required", "chatId", chatId, "userID", userID)
		return "", fmt.Errorf("chat ID is required")
	}

	if message == "" {
		slog.Error("service:GenerateChatName", "message", "message is required", "chatId", chatId, "userID", userID)
		return "", fmt.Errorf("message is required")
	}

	if model == "" {
		slog.Error("service:GenerateChatName", "message", "model is required", "chatId", chatId, "userID", userID)
		return "", fmt.Errorf("model is required")
	}

	apiKey := s.settingsManager.GetSettings().OpenAIAPIKey
	if apiKey == "" {
		slog.Error("service:GenerateChatName", "message", "OpenAI API key not set", "chatId", chatId, "userID", userID)
		return "", fmt.Errorf("OpenAI API key not set")
	}

	name, err := s.dao.GetChatName(userID, chatId)
	if err != nil {
		slog.Error("service:GenerateChatName", "message", "failed to get chat name", "error", err, "chatId", chatId, "userID", userID)
		return "", fmt.Errorf("error while processing request, please try again")
	}

	if name != "" {
		slog.Error("service:GenerateChatName", "message", "Chat name already exists", "chatId", chatId, "userID", userID)
		return "", fmt.Errorf("Chat name already exists")
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
		slog.Error("service:GenerateChatName", "message", "failed to marshal request", "error", err, "chatId", chatId, "userID", userID)
		return "", fmt.Errorf("error while processing request, please try again")
	}

	httpReq, err := http.NewRequest("POST", s.settingsManager.GetSettings().OpenAIAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		slog.Error("service:GenerateChatName", "message", "failed to create request", "error", err, "chatId", chatId, "userID", userID)
		return "", fmt.Errorf("failed to create request, please try again")
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		slog.Error("service:GenerateChatName", "message", "OpenAI request failed", "error", err, "chatId", chatId, "userID", userID)
		return "", fmt.Errorf("OpenAI request failed, please try again")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		slog.Error("service:GenerateChatName", "message", "OpenAI API error", "status", resp.StatusCode, "body", string(body), "chatId", chatId, "userID", userID)
		return "", fmt.Errorf("OpenAI API error, please try again")
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("service:GenerateChatName", "message", "failed to read response body", "error", err, "chatId", chatId, "userID", userID)
		return "", fmt.Errorf("error while processing request, please try again")
	}

	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(respBody, &openAIResp); err != nil {
		slog.Error("service:GenerateChatName", "message", "failed to parse OpenAI response", "error", err, "chatId", chatId, "userID", userID)
		return "", fmt.Errorf("error while processing request, please try again")
	}

	if len(openAIResp.Choices) == 0 {
		slog.Error("service:GenerateChatName", "message", "no choices returned from OpenAI", "chatId", chatId, "userID", userID)
		return "", fmt.Errorf("no choices returned from OpenAI, please try again")
	}

	chatName := openAIResp.Choices[0].Message.Content

	if err := s.dao.SaveChatName(userID, chatId, chatName); err != nil {
		slog.Error("service:GenerateChatName", "message", "failed to save chat name", "error", err, "chatId", chatId, "userID", userID)
		return "", fmt.Errorf("error while processing request, please try again")
	}

	return chatName, nil
}

func (s *ChatService) GetHistory(ctx context.Context, userID string, chatId string) ([]*pb.ChatMessage, *pb.ChatInfo, error) {
	if chatId == "" {
		slog.Error("service:GetHistory", "message", "chat ID is required", "chatId", chatId, "userID", userID)
		return nil, nil, fmt.Errorf("chat ID is required")
	}

	messages, err := s.dao.GetChatMessages(userID, chatId)
	if err != nil {
		slog.Error("service:GetHistory", "message", "failed to fetch history", "error", err, "chatId", chatId, "userID", userID)
		return nil, nil, fmt.Errorf("failed to fetch history, please try again")
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
		slog.Error("service:GetHistory", "message", "failed to get chat metadata", "error", err, "chatId", chatId, "userID", userID)
		return nil, nil, fmt.Errorf("error while processing request, please try again")
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
	slog.Info("service:GetChatList", "projectID", projectID, "userID", userID)
	chats, err := s.dao.GetChatList(userID, projectID, soft_deleted)
	if err != nil {
		slog.Error("service:GetChatList", "message", "failed to get chat metadata", "error", err, "userID", userID, "projectID", projectID)
		return nil, fmt.Errorf("error while processing request, please try again")
	}
	return chats, nil
}

func (s *ChatService) CreateChat(ctx context.Context, userID string, name string, projectID string) (string, error) {
	slog.Info("service:CreateChat", "projectID", projectID, "userID", userID)
	chatId := uuid.New().String()

	err := s.dao.CreateChat(userID, chatId, name, projectID)
	if err != nil {
		slog.Error("service:CreateChat", "message", "failed to insert chat record", "error", err, "projectID", projectID, "userID", userID)
		return "", fmt.Errorf("error while processing request, please try again")
	}

	return chatId, nil
}

func (s *ChatService) ListModel(ctx context.Context) ([]*pb.ModelListInfo, error) {
	slog.Info("service:ListModel")
	models, err := s.dao.GetModels()
	if err != nil {
		slog.Error("service:ListModel", "message", "failed to fetch models", "error", err)
		return nil, fmt.Errorf("error while processing request, please try again")
	}

	return models, nil
}

func (s *ChatService) SearchChat(ctx context.Context, userID string, query string) ([]*pb.SearchResult, error) {
	if query == "" {
		slog.Error("service:SearchChat", "error", "query is required", "userID", userID)
		return nil, fmt.Errorf("query is required")
	}

	results, err := s.dao.SearchChatMessages(userID, query)
	if err != nil {
		slog.Error("service:SearchChat", "message", "search failed", "error", err, "userID", userID, "query", query)
		return nil, fmt.Errorf("error while processing request, please try again")
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
	slog.Info("service:CreateProject", "userID", userID, "name", name)
	id := uuid.New().String()

	if name == "" {
		slog.Error("service:CreateProject", "message", "name is required", "userID", userID, "name", name)
		return "", fmt.Errorf("name is required")
	}

	projectID, err := s.dao.CreateProject(userID, id, name, description, additionalData)
	if err != nil {
		slog.Error("service:CreateProject", "message", "failed to create project", "error", err, "userID", userID, "name", name)
		return "", fmt.Errorf("error while processing request, please try again")
	}

	return projectID, nil
}

func (s *ChatService) GetProjects(ctx context.Context, userID string) ([]dao.ProjectRow, error) {
	slog.Info("service:GetProjects", "userID", userID)
	projects, err := s.dao.GetProjects(userID)
	if err != nil {
		slog.Error("service:GetProjects", "error", "failed to fetch project list", "error", err, "userID", userID)
		return nil, fmt.Errorf("error while processing request, please try again")
	}

	return projects, nil
}

func (s *ChatService) ListDocuments(ctx context.Context, userID string, projectID string) ([]dao.DocumentListRow, error) {
	slog.Info("service:ListDocuments", "userID", userID, "projectID", projectID)
	docs, err := s.dao.FilesList(userID, projectID)
	if err != nil {
		slog.Error("service:ListDocuments", "message", "failed to fetch documents", "error", err, "userID", userID, "projectID", projectID)
		return nil, fmt.Errorf("error while processing request, please try again")
	}

	return docs, nil
}

func (s *ChatService) UploadFile(ctx context.Context, userID string, projectID string, file multipart.File, header *multipart.FileHeader, maxFileSize int64, maxProjectSize int64) (string, error) {
	slog.Info("service:UploadFile", "userID", userID, "projectID", projectID)
	if projectID == "" {
		slog.Error("service:UploadFile", "message", "project_id is required", "userID", userID, "projectID", projectID)
		return "", fmt.Errorf("project_id is required")
	}

	fileSize := header.Size
	if fileSize > maxFileSize {
		slog.Error("service:UploadFile", "error", "file exceeds limit", "userID", userID, "projectID", projectID, "fileSize", fileSize, "maxFileSize", maxFileSize)
		return "", fmt.Errorf("file exceeds %d MB limit", maxFileSize/(1024*1024))
	}

	// Check total project size
	totalUsed, err := s.dao.TotalUsedSize(userID, projectID)
	if err != nil {
		slog.Error("service:UploadFile", "message", "failed to fetch usage", "error", err, "userID", userID, "projectID", projectID)
		return "", fmt.Errorf("error while processing request, please try again")
	}

	if totalUsed+fileSize > maxProjectSize {
		slog.Error("service:UploadFile", "message", "project storage exceeds limit", "userID", userID, "projectID", projectID, "totalUsed", totalUsed, "fileSize", fileSize, "maxProjectSize", maxProjectSize)
		return "", fmt.Errorf("project storage exceeds %d MB", maxProjectSize/(1024*1024))
	}

	// Generate object ID and store file
	objectID := uuid.New().String()

	if err := s.store.StoreObject(ctx, objectID, file); err != nil {
		slog.Error("service:UploadFile", "message", "failed to store file", "error", err, "userID", userID, "projectID", projectID, "objectID", objectID)
		return "", fmt.Errorf("failed to store file,please try again")
	}

	// Save file metadata to database
	if err := s.dao.FileSave(userID, projectID, objectID, header.Filename, fileSize); err != nil {
		slog.Error("service:UploadFile", "message", "failed to save metadata", "error", err, "userID", userID, "projectID", projectID, "objectID", objectID)
		return "", fmt.Errorf("error while processing request, please try again")
	}

	// Publish embedding generation event
	msg := GenerateEmbeddingMessage{DocsID: objectID}
	msgBytes, _ := json.Marshal(msg)
	err = s.queue.Publish(ctx, events.GENERATE_EMBEDDINGS, msgBytes)
	if err != nil {
		// Log error but don't fail the upload
		slog.Error("service:UploadFile", "message", "failed to publish embedding generation event", "error", err, "userID", userID, "projectID", projectID, "objectID", objectID)
	}

	return objectID, nil
}

func (s *ChatService) retrieveSimilarChunks(ctx context.Context, userID string, projectID string, query string) (*rag.Response, error) {
	if projectID == "" || query == "" {
		slog.Error("service:retrieveSimilarChunks", "message", "project_id and query are required", "userID", userID, "projectID", projectID, "query", query)
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
		slog.Error("service:retrieveSimilarChunks", "message", "failed to embed query", "error", err, "userID", userID, "projectID", projectID, "query", query)
		return nil, fmt.Errorf("error while processing request, please try again")
	}
	if len(embedding) == 0 {
		slog.Error("service:retrieveSimilarChunks", "message", "embedding could not be created", "userID", userID, "projectID", projectID, "query", query)
		return nil, fmt.Errorf("embedding could not be created, please try again")
	}

	params := rag.SearchParams{TopK: 2, ProjectID: projectID}
	retriever := func(ctx context.Context, embedding []float64, params rag.SearchParams) ([]rag.Result, error) {
		embBytes, err := json.Marshal(embedding)
		if err != nil {
			slog.Error("service:retrieveSimilarChunks", "message", "failed to marshal embedding", "error", err, "userID", userID, "projectID", projectID, "query", query)
			return nil, fmt.Errorf("error while processing request, please try again")
		}
		vecRows, err := s.dao.GetTopSimilarRAGChunks(userID, string(embBytes), projectID)
		if err != nil {
			slog.Error("service:retrieveSimilarChunks", "message", "failed to get top similar chunks", "error", err, "userID", userID, "projectID", projectID, "query", query)
			return nil, fmt.Errorf("error while processing request, please try again")
		}
		var results []rag.Result
		for _, v := range vecRows {
			_, reader, err := s.store.GetObject(ctx, v.DocsID)
			if err != nil {
				slog.Error("service:retrieveSimilarChunks", "message", "failed to get object for docsID", "error", err, "userID", userID, "projectID", projectID, "query", query, "docsID", v.DocsID)
				return nil, fmt.Errorf("failed to get document, please try again")
			}
			data, err := io.ReadAll(reader)
			if rc, ok := reader.(io.ReadCloser); ok {
				_ = rc.Close()
			}
			if err != nil {
				slog.Error("service:retrieveSimilarChunks", "message", "failed to read object for docsID", "error", err, "userID", userID, "projectID", projectID, "query", query, "docsID", v.DocsID)
				return nil, fmt.Errorf("failed to read document, please try again")
			}
			if v.StartByte < 0 || v.EndByte > len(data) || v.StartByte > v.EndByte {
				slog.Error("service:retrieveSimilarChunks", "message", "invalid chunk byte range for docsID", "userID", userID, "projectID", projectID, "query", query, "docsID", v.DocsID, "startByte", v.StartByte, "endByte", v.EndByte, "fileSize", len(data))
				return nil, fmt.Errorf("invalid chunk byte range for document, please try again")
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
		slog.Error("service:retrieveSimilarChunks", "message", "failed to retrieve pipeline", "error", err, "userID", userID, "projectID", projectID, "query", query)
		return nil, fmt.Errorf("error while processing request, please try again")
	}

	return response, nil
}

func (s *ChatService) SubmitGenerateEmbeddingsJob(ctx context.Context, userID string, projectID string) error {
	if projectID == "" {
		slog.Error("service:SubmitGenerateEmbeddingsJob", "message", "project_id is required", "userID", userID, "projectID", projectID)
		return fmt.Errorf("project_id is required")
	}

	docs, error := s.dao.FetchErrorDocs(userID, projectID)
	if error != nil {
		slog.Error("service:SubmitGenerateEmbeddingsJob", "message", "failed to fetch error docs", "error", error, "userID", userID, "projectID", projectID)
		return fmt.Errorf("failed to check embedding status, please try again")
	}

	for _, docsID := range docs {
		msg := GenerateEmbeddingMessage{DocsID: docsID}
		msgBytes, _ := json.Marshal(msg)
		err := s.queue.Publish(ctx, "generate.embedding", msgBytes)
		if err != nil {
			slog.Error("service:SubmitGenerateEmbeddingsJob", "message", "failed to publish job", "error", err, "userID", userID, "projectID", projectID, "docsID", docsID)
			return fmt.Errorf("failed to publish job, please try again")
		}

		if updateErr := s.dao.UpdateEmbeddingStatus(docsID, int32(pb.Embedding_Status_STATUS_QUEUED)); updateErr != nil {
			slog.Error("service:SubmitGenerateEmbeddingsJob", "message", "failed to update embedding status", "error", updateErr, "userID", userID, "projectID", projectID, "docsID", docsID)
			return fmt.Errorf("failed to update embedding status, please try again")
		}
	}

	return nil
}

func (s *ChatService) BranchAChat(ctx context.Context, userID string, sourceChatId string, branchFromMessageId string, branchName string) (string, error) {
	if sourceChatId == "" {
		slog.Error("service:BranchAChat", "message", "parent id is required", "userID", userID, "sourceChatId", sourceChatId)
		return "", fmt.Errorf("parent id is required")
	}

	if branchFromMessageId == "" {
		slog.Error("service:BranchAChat", "message", "message id is required", "userID", userID, "sourceChatId", sourceChatId, "branchFromMessageId", branchFromMessageId)
		return "", fmt.Errorf("message id is required")
	}

	isMain, err := s.dao.IsMainBranch(userID, sourceChatId)
	if err != nil || !isMain {
		slog.Error("service:BranchAChat", "message", "can only branch from main branch chats", "userID", userID, "sourceChatId", sourceChatId)
		return "", fmt.Errorf("its not a main branch, try branching from a main branch")
	}

	newChatId := uuid.New().String()

	err = s.dao.BranchChat(userID, sourceChatId, branchFromMessageId, newChatId, branchName)
	if err != nil {
		slog.Error("service:BranchAChat", "message", "failed to create branch", "error", err, "userID", userID, "sourceChatId", sourceChatId, "branchFromMessageId", branchFromMessageId, "branchName", branchName)
		return "", fmt.Errorf("failed to create branch, please try again")
	}

	return newChatId, nil
}

func (s *ChatService) ListChatBranch(ctx context.Context, userID string, chatId string) ([]dao.ChatInfoRow, error) {
	slog.Info("service:ListChatBranch", "userID", userID, "chatId", chatId)
	if chatId == "" {
		slog.Error("service:ListChatBranch", "message", "chat id is required", "userID", userID, "chatId", chatId)
		return nil, fmt.Errorf("Chat Id is required")
	}

	isMain, err := s.dao.IsMainBranch(userID, chatId)
	if err != nil {
		slog.Error("service:ListChatBranch", "message", "failed to get main branch status", "error", err, "userID", userID, "chatId", chatId)
		return nil, fmt.Errorf("failed to get main branch status, please try again")
	}

	innerChats, err := s.dao.GetChatBranches(userID, chatId, isMain)
	if err != nil {
		slog.Error("service:ListChatBranch", "message", "failed to get inner chat list", "error", err, "userID", userID, "chatId", chatId)
		return nil, fmt.Errorf("failed to get inner chat list, please try again")
	}

	return innerChats, nil
}

func (s *ChatService) EmbeddingSubscriber() {
	slog.Info("service:EmbeddingSubscriber")
	go func() {
		sub, err := s.queue.Subscribe(context.Background(), events.GENERATE_EMBEDDINGS)
		if err != nil {
			slog.Error("service:EmbeddingSubscriber", "message", "failed to subscribe to embedding generation event", "error", err)
			return
		}

		for msg := range sub {
			var payload GenerateEmbeddingMessage
			if err := json.Unmarshal(msg.Data, &payload); err == nil {

				if updateErr := s.dao.UpdateEmbeddingStatus(payload.DocsID, int32(pb.Embedding_Status_STATUS_IN_PROGRESS)); updateErr != nil {
					slog.Error("service:EmbeddingSubscriber", "message", "failed to update embedding status to in-progress", "error", updateErr, "docsID", payload.DocsID)
					continue
				}

				// Fetch project_id for docs_id
				docMeta, err := s.dao.GetFileMetadata(payload.DocsID)
				if err != nil {
					slog.Error("service:EmbeddingSubscriber", "message", "failed to fetch file metadata from database", "error", err, "docsID", payload.DocsID)
					continue
				}

				filePath := "filestore/objects/" + payload.DocsID
				f, err := os.Open(filePath)
				if err != nil {
					slog.Error("service:EmbeddingSubscriber", "message", "failed to open file", "error", err, "docsID", payload.DocsID, "filePath", filePath)
					continue
				}

				metadata := map[string]string{
					"project_id": docMeta.ProjectID,
					"docs_id":    payload.DocsID,
					"source":     docMeta.FileName,
				}

				result, err := s.pipeline.RunWithChunks(context.Background(), f, "text/plain", metadata) //WILL LOGGIFY THIS LATER
				if err != nil {
					if cerr := f.Close(); cerr != nil {
						slog.Error("service:EmbeddingSubscriber", "message", "failed to close file after pipeline error", "error", cerr, "docsID", payload.DocsID)
					}
					slog.Error("service:EmbeddingSubscriber", "message", "failed to run pipeline", "error", err, "metadata", metadata)
					if updateErr := s.dao.UpdateEmbeddingStatus(payload.DocsID, int32(pb.Embedding_Status_STATUS_ERROR)); updateErr != nil {
						slog.Error("service:EmbeddingSubscriber", "message", "failed to update embedding status to error", "error", updateErr, "docsID", payload.DocsID)
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
						slog.Error("service:EmbeddingSubscriber", "message", "failed to save chunk", "error", err, "chunkID", chunk.ID, "projectID", chunk.ProjectID, "docsID", chunk.DocsID)
					}

					if emb, ok := embeddingMap[chunk.ID]; ok {
						if err := s.dao.SaveRAGChunkEmbedding(chunk.ID, emb.Vector); err != nil {
							slog.Error("service:EmbeddingSubscriber", "message", "failed to save embedding", "error", err, "chunkID", chunk.ID)
						}
					}
				}
				if updateErr := s.dao.UpdateEmbeddingStatus(payload.DocsID, int32(pb.Embedding_Status_STATUS_SUCCESS)); updateErr != nil {
					slog.Error("service:EmbeddingSubscriber", "message", "failed to update embedding status to success", "error", updateErr, "docsID", payload.DocsID)
				}
				if cerr := f.Close(); cerr != nil {
					slog.Error("service:EmbeddingSubscriber", "message", "failed to close file", "error", cerr, "docsID", payload.DocsID)
				}
			} else {
				slog.Error("service:EmbeddingSubscriber", "message", "failed to unmarshal message", "error", err, "msgID", msg.ID, "msgSubject", msg.Subject)
				continue
			}
		}
	}()
}

// createRAGDocumentJSONFromChunks converts RAG chunks to the requested JSON structure
func (s *ChatService) createRAGDocumentJSONFromChunks(ragChunks []rag.Result) []RAGDocumentJSON {
	slog.Info("service:createRAGDocumentJSONFromChunks", "ChunksCount", len(ragChunks))
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
		slog.Info("service:createRAGDocumentJSONFromChunks", "docID", docID, "chunkCount", len(chunks))

		ragDocuments = append(ragDocuments, RAGDocumentJSON{
			DocID:  docID,
			Chunks: chunks,
		})
	}

	return ragDocuments
}

// GetRAGDocumentReference retrieves RAG document references for a specific message
func (s *ChatService) GetRAGDocumentReference(ctx context.Context, userID string, req *pb.RAGDocumentReferenceRequest) (*pb.RAGDocumentReferenceResponse, error) {
	slog.Info("service:GetRAGDocumentReference", "userID", userID, "req", req)
	// Validate request
	if req.MessageId == "" {
		slog.Error("service:GetRAGDocumentReference", "message", "message_id is required", "userID", userID, "req", req)
		return nil, fmt.Errorf("message_id is required")
	}

	// Get the chat message by ID
	message, err := s.dao.GetChatMessageByID(userID, req.MessageId)
	if err != nil {
		slog.Error("service:GetRAGDocumentReference", "message", "failed to get message", "error", err, "userID", userID, "req", req)
		return nil, fmt.Errorf("failed to get message, please try again")
	}

	// Parse the document references JSON
	if message.DocumentReferences == "" {
		return &pb.RAGDocumentReferenceResponse{
			Reference: nil,
		}, nil
	}

	var ragDocuments []RAGDocumentJSON
	if err := json.Unmarshal([]byte(message.DocumentReferences), &ragDocuments); err != nil {
		slog.Error("service:GetRAGDocumentReference", "message", "failed to parse document references", "error", err, "userID", userID, "req", req)
		return nil, fmt.Errorf("failed to parse document references, please try again")
	}

	// If docId is specified, filter for that specific document
	if req.DocId != "" {
		for _, doc := range ragDocuments {
			if doc.DocID == req.DocId {
				// Get document metadata
				docMeta, err := s.dao.GetFileMetadata(doc.DocID)
				if err != nil {
					slog.Error("service:GetRAGDocumentReference", "message", "failed to get document metadata", "error", err, "userID", userID, "req", req, "docID", doc.DocID)
					return nil, fmt.Errorf("failed to get document metadata, please try again")
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
			slog.Error("service:GetRAGDocumentReference", "message", "failed to get document metadata", "error", err, "userID", userID, "req", req, "docID", doc.DocID)
			return nil, fmt.Errorf("failed to get document metadata, please try again")
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
	slog.Info("service:DeleteDocument", "userID", userID, "projectID", projectID, "docID", docID)
	if projectID == "" || docID == "" {
		slog.Error("service:DeleteDocument", "message", "project_id and doc_id are required", "userID", userID, "projectID", projectID, "docID", docID)
		return fmt.Errorf("project_id and doc_id are required")
	}

	err := s.dao.DeleteDocument(userID, projectID, docID)
	if err != nil {
		slog.Error("service:DeleteDocument", "message", "failed to delete document", "error", err, "userID", userID, "projectID", projectID, "docID", docID)
		return fmt.Errorf("failed to delete document, please try again")
	}

	//TODO: What if this operation fails?
	err = s.store.DeleteObject(ctx, docID)
	if err != nil {
		slog.Error("service:DeleteDocument", "message", "failed to delete object", "error", err, "userID", userID, "projectID", projectID, "docID", docID)
		return fmt.Errorf("failed to delete object, please try again")
	}

	return nil
}

func (s *ChatService) DeleteChat(ctx context.Context, userID string, chatId string, operation pb.DeleteChatRequest_Operation) error {
	slog.Info("service:DeleteChat", "userID", userID, "chatId", chatId, "operation", operation)
	if chatId == "" {
		slog.Error("service:DeleteChat", "message", "chat ID is required", "userID", userID, "chatId", chatId, "operation", operation)
		return fmt.Errorf("chat ID is required")
	}

	switch operation {
	case pb.DeleteChatRequest_DELETE:
		if err := s.dao.DeleteChat(userID, chatId); err != nil {
			slog.Error("service:DeleteChat", "message", "failed to delete chat", "error", err, "userID", userID, "chatId", chatId, "operation", operation)
			return fmt.Errorf("failed to delete chat, please try again")
		}
	case pb.DeleteChatRequest_SOFT_DELETE:
		if err := s.dao.SoftDeleteChat(userID, chatId); err != nil {
			slog.Error("service:DeleteChat", "message", "failed to soft delete chat", "error", err, "userID", userID, "chatId", chatId, "operation", operation)
			return fmt.Errorf("failed to soft delete chat, please try again")
		}
	default:
		slog.Error("service:DeleteChat", "message", "unsupported delete operation", "userID", userID, "chatId", chatId, "operation", operation)
		return fmt.Errorf("unsupported delete operation: %v", operation)
	}

	return nil
}

func (s *ChatService) RestoreChat(ctx context.Context, userID string, chatId string) error {
	if chatId == "" {
		slog.Error("service:RestoreChat", "message", "chat ID is required", "userID", userID, "chatId", chatId)
		return fmt.Errorf("chat ID is required")
	}

	err := s.dao.RestoreChat(userID, chatId)
	if err != nil {
		slog.Error("service:RestoreChat", "message", "failed to restore chat", "error", err, "userID", userID, "chatId", chatId)
		return fmt.Errorf("failed to restore chat, please try again")
	}

	return nil
}

func (s *ChatService) RenameChat(ctx context.Context, userID string, chatId string, name string) error {
	if chatId == "" {
		slog.Error("service:RenameChat", "message", "chat ID is required", "userID", userID, "chatId", chatId, "name", name)
		return fmt.Errorf("chat ID is required")
	}

	trimmedName := strings.TrimSpace(name)

	if len(trimmedName) < MIN_CHAT_NAME_LENGTH {
		slog.Error("service:RenameChat", "message", fmt.Sprintf("name must be at least %d characters", MIN_CHAT_NAME_LENGTH), "userID", userID, "chatId", chatId, "name", name)
		return fmt.Errorf("name must be at least %d characters", MIN_CHAT_NAME_LENGTH)
	}

	if len(trimmedName) > MAX_CHAT_NAME_LENGTH {
		slog.Error("service:RenameChat", "message", fmt.Sprintf("name must be less than %d characters", MAX_CHAT_NAME_LENGTH), "userID", userID, "chatId", chatId, "name", name)
		return fmt.Errorf("name must be less than %d characters", MAX_CHAT_NAME_LENGTH)
	}

	err := s.dao.RenameChat(userID, chatId, trimmedName)
	if err != nil {
		slog.Error("service:RenameChat", "message", "failed to rename chat", "error", err, "userID", userID, "chatId", chatId, "name", name)
		return fmt.Errorf("failed to rename chat, please try again")
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
