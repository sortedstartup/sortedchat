package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"sortedstartup/chatservice/dao"
	pb "sortedstartup/chatservice/proto"
	"sortedstartup/chatservice/rag"
	settings "sortedstartup/chatservice/settings"
	"sortedstartup/chatservice/sortedagents"
)

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

func (s *ChatService) buildAgenticSession(chatID string, prompt string, history []dao.ChatMessageRow) (sortedagents.Session, error) {
	slog.Debug("service:Chat", "message", "building agentic session from chat history", "chatId", chatID)
	messages := []sortedagents.Message{
		{Role: "system", Content: prompt},
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
	agentPrompt, err := s.getChatDefaultPrompt()
	if err != nil {
		slog.Error("service:Chat", "message", "failed to load chat default prompt", "error", err, "chatId", chatID, "userID", userID, "projectID", projectID)
		return fmt.Errorf("failed to load chat prompt")
	}

	session, err := s.buildAgenticSession(chatID, agentPrompt, history)
	if err != nil {
		slog.Error("service:Chat", "message", "failed to build agentic session", "error", err, "chatId", chatID, "userID", userID, "projectID", projectID)
		return fmt.Errorf("failed to prepare chat context")
	}

	webSearchSettings, err := s.getWebSearchSettings()
	if err != nil {
		slog.Error("service:Chat", "message", "failed to load websearch settings", "error", err, "chatId", chatID, "userID", userID, "projectID", projectID)
		return fmt.Errorf("failed to load web search settings")
	}

	scrapeSettings, err := s.getCloudflareScrapeSettings()
	if err != nil {
		slog.Error("service:Chat", "message", "failed to load cloudflare scrape settings", "error", err, "chatId", chatID, "userID", userID, "projectID", projectID)
		return fmt.Errorf("failed to load scrape settings")
	}

	maxTurns, err := s.getAgenticMaxTurns()
	if err != nil {
		slog.Warn("service:Chat", "message", "failed to load agentic max turns, using default", "error", err, "chatId", chatID, "userID", userID, "projectID", projectID)
		return fmt.Errorf("failed to load agentic max turns setting")
	}

	tools := []sortedagents.Tool{
		NewBraveSearchToolWithConfig(webSearchSettings.APIURL, webSearchSettings.APIKey),
	}
	if strings.TrimSpace(scrapeSettings.APIURL) != "" && strings.TrimSpace(scrapeSettings.APIKey) != "" {
		slog.Info("Cloudflare scrape tool configured, adding to agent tools", "chatId", chatID, "userID", userID, "projectID", projectID)
		tools = append(tools, NewBrowserScrapeToolWithConfig(scrapeSettings.APIURL, scrapeSettings.APIKey))
	} else {
		slog.Info("Cloudflare scrape tool not configured, skipping", "chatId", chatID, "userID", userID, "projectID", projectID)
	}

	agent := sortedagents.NewAgent(
		"chat-agent",
		agentPrompt,
		model,
		tools,
	)
	apiKey := providerSettings.ApiKey
	apiURL := providerSettings.ApiUrl
	llm := sortedagents.NewOpenAILLMWithConfig(apiKey, apiURL, model)
	runner := sortedagents.NewRunnerWithLLM(llm)

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

	// maxTurnsValue, err := s.getSettingValue(AGENTIC_MAX_TURNS_KEY, defaultAgenticMaxTurns)
	// if err != nil {
	// 	slog.Warn("service:Chat", "message", "failed to load agentic max turns, using default", "error", err, "chatId", chatID, "userID", userID, "projectID", projectID)
	// 	maxTurnsValue = defaultAgenticMaxTurns
	// }
	// maxTurns, err := strconv.Atoi(maxTurnsValue)
	// if err != nil || maxTurns <= 0 {
	// 	maxTurns, _ = strconv.Atoi(defaultAgenticMaxTurns)
	// }

	events := runner.RunStream(ctx, agent, prompt, maxTurns, session)

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
	var successfulWebSearchCalls int
	searchCostPerRequest, err := parseBraveSearchCost(webSearchSettings.Cost)
	if err != nil {
		slog.Warn("service:Chat", "message", "invalid brave search request cost, using default", "cost", webSearchSettings.Cost, "error", err, "chatId", chatID, "userID", userID, "projectID", projectID)
		searchCostPerRequest, _ = parseBraveSearchCost(settings.DEFAULT_BRAVE_SEARCH_COST)
	}
	firstToken := true
	completed := false

	// - fullResponse accumulates streamed text chunks into one final assistant message.
	// - We stream chunks to the client immediately, but still need the full text for DB persistence.
	savePartialResponse := func() {
		assistantText := fullResponse.String()
		if assistantText == "" {
			return
		}

		// - referencesJSON stores which RAG docs/chunks were used for this assistant message.
		// - It is saved with the chat message so reference UI can be reconstructed later.
		var referencesJSON string
		if len(ragChunks) > 0 {
			// - createRAGDocumentJSONFromChunks groups retrieved chunks by document.
			// - We group retrieved chunks by document before saving references on a message.
			// - Example: chunk1/docA, chunk2/docA, chunk3/docB becomes docA with 2 chunks and docB with 1 chunk.
			// Group chunks by document ID
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

		searchCost := float64(successfulWebSearchCalls) * searchCostPerRequest

		if _, err := s.dao.AddChatMessageWithTokenCount(userID, chatID, "assistant", assistantText, "", model, nonCachedInputTokens, outputTokens, cachedTokens, searchCost, referencesJSON, ragEnabled); err != nil {
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
			switch e.ToolName {
			case "web_search":
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
			case "browser_scrape":
				if err := stream(&pb.ChatResponse{
					Response: &pb.ChatResponse_Progress{
						Progress: &pb.ChatProgress{
							State:   pb.ChatProgress_REQUEST_SENT_TO_LLM,
							Message: "Scraping web page",
						},
					},
				}); err != nil {
					savePartialResponse()
					return fmt.Errorf("error while processing request, please try again")
				}
			}

		case *sortedagents.UsageEvent:
			inputTokens = e.InputTokens
			outputTokens = e.OutputTokens
			cachedTokens = e.CachedTokens

		case *sortedagents.ToolCallEndEvent:
			if e.ToolName == "web_search" && e.Error == nil {
				successfulWebSearchCalls++
			}

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
		// - Final save mirrors partial save, but for a completed assistant response.
		// - We attach the same RAG references so the final message preserves its supporting context.
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

		searchCost := float64(successfulWebSearchCalls) * searchCostPerRequest
		daoSummary, err := s.dao.AddChatMessageWithTokenCount(userID, chatID, "assistant", assistantText, "", model, nonCachedInputTokens, outputTokens, cachedTokens, searchCost, referencesJSON, ragEnabled)
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

func parseBraveSearchCost(value string) (float64, error) {
	cost, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0, err
	}
	if cost < 0 {
		return 0, fmt.Errorf("cost must be non-negative")
	}
	return cost, nil
}
