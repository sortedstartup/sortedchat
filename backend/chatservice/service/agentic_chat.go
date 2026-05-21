package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"sortedstartup/chatservice/dao"
	pb "sortedstartup/chatservice/proto"
	"sortedstartup/chatservice/rag"
	"sortedstartup/chatservice/sortedagents"
)

const MAX_TURNS = 4
const chatAgentInstructions = `You are SortedChat’s default assistant.

Answer from your own knowledge and reasoning by default. Use web_search only when fresh, external, verifiable, or source-backed information is needed, such as news, prices, laws, schedules, product details, live data, recent updates, or explicit search requests.
When you search, ground the answer in the results and include relevant source URLs.
`

func (s *ChatService) shouldUseAgenticChat(capabilities *pb.ModelCapabilities, hasImages bool, history []dao.ChatMessageRow, providerSettings *pb.ProviderSettings) bool {
	slog.Debug("service:Chat", "message", "checking if agentic chat should be used")
	if hasImages || providerSettings == nil || strings.TrimSpace(providerSettings.ApiUrl) == "" {
		return false
	}

	webSearchSettings, err := s.getWebSearchSettings()
	if err != nil {
		slog.Error("service:Chat", "message", "failed to load websearch settings", "error", err)
		return false
	}

	if strings.TrimSpace(webSearchSettings.BraveSearchAPIKey) == "" {
		return false
	}

	// Today our agent loop is text-only, so we gate on text chat capability
	// rather than provider name. Tool-calling capability can be added here once
	// it exists in model metadata.
	if capabilities == nil || capabilities.Text == nil || !capabilities.Text.Input || !capabilities.Text.Output {
		return false
	}

	for _, msg := range history {
		if _, ok := extractTextOnlyChatMessage(msg); !ok {
			return false
		}
	}

	return true
}

func extractTextOnlyChatMessage(msg dao.ChatMessageRow) (string, bool) {
	slog.Debug("service:Chat", "message", "extracting text from chat message", "messageId", msg)
	if msg.ContentImage != "" {
		return "", false
	}

	if strings.TrimSpace(msg.Content) == "" {
		return "", true
	}

	if !(strings.HasPrefix(msg.Content, "[") && strings.HasSuffix(msg.Content, "]")) {
		return msg.Content, true
	}

	var contents []*pb.MessageContent
	if err := json.Unmarshal([]byte(msg.Content), &contents); err != nil {
		return "", false
	}

	var textParts []string
	for _, content := range contents {
		if content.Type != "text" {
			return "", false
		}
		if strings.TrimSpace(content.Text) != "" {
			textParts = append(textParts, content.Text)
		}
	}

	return strings.TrimSpace(strings.Join(textParts, "\n")), true
}

func (s *ChatService) buildAgenticSession(chatID string, history []dao.ChatMessageRow) (sortedagents.Session, error) {
	slog.Debug("service:Chat", "message", "building agentic session from chat history", "chatId", chatID)
	messages := []sortedagents.Message{
		{Role: "system", Content: chatAgentInstructions},
	}

	for _, msg := range history {
		text, ok := extractTextOnlyChatMessage(msg)
		if !ok {
			return nil, fmt.Errorf("chat history contains non-text content")
		}

		messages = append(messages, sortedagents.Message{
			Role:    msg.Role,
			Content: text,
		})
	}

	return sortedagents.NewSessionFromMessages(chatID, messages), nil
}

func (s *ChatService) runAgenticChat(
	ctx context.Context,
	chatID string,
	userID string,
	projectID string,
	model string,
	prompt string,
	history []dao.ChatMessageRow,
	ragChunks []rag.Result,
	ragEnabled bool,
	providerSettings *pb.ProviderSettings,
	stream func(*pb.ChatResponse) error,
) error {
	slog.Debug("service:Chat", "message", "running agentic chat", "chatId", chatID, "userID", userID, "projectID", projectID)
	session, err := s.buildAgenticSession(chatID, history)
	if err != nil {
		slog.Error("service:Chat", "message", "failed to build agentic session", "error", err, "chatId", chatID, "userID", userID, "projectID", projectID)
		return fmt.Errorf("failed to prepare chat context")
	}

	webSearchSettings, err := s.getWebSearchSettings()
	if err != nil {
		slog.Error("service:Chat", "message", "failed to load websearch settings", "error", err, "chatId", chatID, "userID", userID, "projectID", projectID)
		return fmt.Errorf("failed to load web search settings")
	}

	agent := sortedagents.NewAgent(
		"chat-agent",
		chatAgentInstructions,
		model,
		[]sortedagents.Tool{NewBraveSearchToolWithAPIKey(webSearchSettings.BraveSearchAPIKey)},
	)
	runner := sortedagents.NewRunnerWithLLM(sortedagents.NewOpenAILLMWithConfig(
		providerSettings.ApiKey,
		providerSettings.ApiUrl,
		model,
	))

	if err := stream(&pb.ChatResponse{
		Response: &pb.ChatResponse_Progress{
			Progress: &pb.ChatProgress{
				State:   pb.ChatProgress_SENDING_REQUEST_TO_LLM,
				Message: "Sending request to LLM",
			},
		},
	}); err != nil {
		return fmt.Errorf("error while processing request, please try again")
	}

	events := runner.RunStream(ctx, agent, prompt, MAX_TURNS, session)

	if err := stream(&pb.ChatResponse{
		Response: &pb.ChatResponse_Progress{
			Progress: &pb.ChatProgress{
				State:   pb.ChatProgress_REQUEST_SENT_TO_LLM,
				Message: "Request sent",
			},
		},
	}); err != nil {
		return fmt.Errorf("error while processing request, please try again")
	}

	var fullResponse strings.Builder
	var inputTokens, outputTokens, cachedTokens int
	firstToken := true
	completed := false

	savePartialResponse := func() {
		assistantText := fullResponse.String()
		if assistantText == "" {
			return
		}

		var referencesJSON string
		if len(ragChunks) > 0 {
			ragDocs := s.createRAGDocumentJSONFromChunks(ragChunks)
			referencesBytes, err := json.Marshal(ragDocs)
			if err != nil {
				slog.Error("service:Chat", "message", "failed to marshal RAG document references for partial agentic response", "error", err, "chatId", chatID, "userID", userID, "projectID", projectID)
			} else {
				referencesJSON = string(referencesBytes)
			}
		}

		nonCachedInputTokens := inputTokens - cachedTokens
		if nonCachedInputTokens < 0 {
			nonCachedInputTokens = 0
		}

		if _, err := s.dao.AddChatMessageWithTokens(userID, chatID, "assistant", assistantText, "", model, nonCachedInputTokens, outputTokens, cachedTokens, referencesJSON, ragEnabled); err != nil {
			slog.Error("service:Chat", "message", "failed to save partial agentic assistant message", "error", err, "chatId", chatID, "userID", userID, "projectID", projectID)
		}
	}

	for event := range events {
		switch e := event.(type) {
		case *sortedagents.TextChunkEvent:
			if firstToken {
				if err := stream(&pb.ChatResponse{
					Response: &pb.ChatResponse_Progress{
						Progress: &pb.ChatProgress{
							State:   pb.ChatProgress_FIRST_TOKEN_RECEIVED,
							Message: "First token received",
						},
					},
				}); err != nil {
					savePartialResponse()
					return fmt.Errorf("error while processing request, please try again")
				}
				firstToken = false
			}

			fullResponse.WriteString(e.Chunk)
			if err := stream(&pb.ChatResponse{
				Response: &pb.ChatResponse_Text{Text: e.Chunk},
			}); err != nil {
				savePartialResponse()
				return fmt.Errorf("error while processing request, please try again")
			}

		case *sortedagents.ToolCallStartEvent:
			if e.ToolName != "web_search" {
				continue
			}
			if err := stream(&pb.ChatResponse{
				Response: &pb.ChatResponse_Progress{
					Progress: &pb.ChatProgress{
						State:   pb.ChatProgress_REQUEST_SENT_TO_LLM,
						Message: "Searching the web",
					},
				},
			}); err != nil {
				savePartialResponse()
				return fmt.Errorf("error while processing request, please try again")
			}

		case *sortedagents.UsageEvent:
			inputTokens = e.InputTokens
			outputTokens = e.OutputTokens
			cachedTokens = e.CachedTokens

		case *sortedagents.CompleteEvent:
			completed = true

		case *sortedagents.ErrorEvent:
			if ctx.Err() != nil {
				slog.Info("service:Chat", "message", "agentic streaming cancelled by client, saving partial response", "error", e.Error, "chatId", chatID, "userID", userID, "projectID", projectID)
			} else {
				slog.Error("service:Chat", "message", "agentic chat failed", "error", e.Error, "chatId", chatID, "userID", userID, "projectID", projectID)
			}
			savePartialResponse()
			return fmt.Errorf("error while processing request, please try again")
		}
	}

	if !completed && ctx.Err() != nil {
		savePartialResponse()
		return fmt.Errorf("error while processing request, please try again")
	}

	if err := stream(&pb.ChatResponse{
		Response: &pb.ChatResponse_Progress{
			Progress: &pb.ChatProgress{
				State:   pb.ChatProgress_TOKENS_STOPPED,
				Message: "Response finished",
			},
		},
	}); err != nil {
		return fmt.Errorf("error while processing request, please try again")
	}

	assistantText := fullResponse.String()
	if assistantText != "" {
		var referencesJSON string
		if len(ragChunks) > 0 {
			ragDocs := s.createRAGDocumentJSONFromChunks(ragChunks)
			referencesBytes, err := json.Marshal(ragDocs)
			if err != nil {
				slog.Error("service:Chat", "message", "failed to marshal RAG document references for final agentic response", "error", err, "chatId", chatID, "userID", userID, "projectID", projectID)
			} else {
				referencesJSON = string(referencesBytes)
			}
		}

		nonCachedInputTokens := inputTokens - cachedTokens
		if nonCachedInputTokens < 0 {
			nonCachedInputTokens = 0
		}

		daoSummary, err := s.dao.AddChatMessageWithTokens(userID, chatID, "assistant", assistantText, "", model, nonCachedInputTokens, outputTokens, cachedTokens, referencesJSON, ragEnabled)
		if err != nil {
			slog.Error("service:Chat", "message", "failed to insert agentic assistant message", "error", err, "chatId", chatID, "userID", userID, "projectID", projectID)
		} else {
			pbSummary := &pb.ResponseSummary{
				MessageId:    daoSummary.MessageId,
				Model:        model,
				InputTokens:  int32(daoSummary.InputTokenCount),
				OutputTokens: int32(daoSummary.OutputTokenCount),
				CachedTokens: int32(daoSummary.CachedTokenCount),
				Cost:         float32(daoSummary.Cost),
			}
			if err := stream(&pb.ChatResponse{
				Response: &pb.ChatResponse_Summary{Summary: pbSummary},
			}); err != nil {
				return fmt.Errorf("failed to send message summary, please try again")
			}
		}
	}

	chatInfo, err := s.dao.GetChatMetadata(userID, chatID)
	if err != nil {
		slog.Error("service:Chat", "message", "failed to get chat metadata after agentic response", "error", err, "chatId", chatID, "userID", userID, "projectID", projectID)
		return fmt.Errorf("error while processing request, please try again")
	}

	return stream(&pb.ChatResponse{
		Response: &pb.ChatResponse_ChatMetadata{
			ChatMetadata: &pb.ChatInfo{
				ChatId:           chatID,
				Cost:             float32(chatInfo.Cost),
				InputTokenCount:  int32(chatInfo.InputTokenCount),
				OutputTokenCount: int32(chatInfo.OutputTokenCount),
				CachedTokenCount: int32(chatInfo.CachedTokenCount),
			},
		},
	})
}
