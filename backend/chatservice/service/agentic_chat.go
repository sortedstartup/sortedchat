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
	"sortedstartup/chatservice/sortedagents"
	"sortedstartup/chatservice/types"
)

func buildSortedAgentsMessage(role string, content interface{}) sortedagents.Message {
	var messageContent sortedagents.MessageContent

	switch value := content.(type) {
	case nil:
		messageContent = sortedagents.TextContent("")
	case string:
		messageContent = sortedagents.TextContent(value)
	case []types.ContentPart:
		parts := make(sortedagents.ContentParts, 0, len(value))
		for _, part := range value {
			converted := sortedagents.ContentPart{
				Type: part.Type,
				Text: part.Text,
			}
			if part.ImageURL != nil {
				converted.ImageURL = &sortedagents.ImageURLPart{URL: part.ImageURL.URL}
			}
			parts = append(parts, converted)
		}
		messageContent = parts
	}

	return sortedagents.Message{
		Role:    role,
		Content: messageContent,
	}
}

func (s *ChatService) buildAgenticSession(chatID string, prompt string, history []dao.ChatMessageRow) (sortedagents.Session, error) {
	slog.Debug("service:Chat", "message", "building agentic session from chat history", "chatId", chatID)
	messages := []sortedagents.Message{
		buildSortedAgentsMessage("system", prompt),
	}

	for _, msg := range history {
		slog.Debug("service:Chat", "message", "building sortedagents content from chat message", "messageId", msg.Id)

		var contents []*pb.MessageContent
		if strings.HasPrefix(msg.Content, "[") && strings.HasSuffix(msg.Content, "]") {
			var textContents []*pb.MessageContent
			if err := json.Unmarshal([]byte(msg.Content), &textContents); err != nil {
				return nil, fmt.Errorf("failed to parse text content: %w", err)
			}
			contents = append(contents, textContents...)
		} else if msg.Content != "" {
			contents = append(contents, &pb.MessageContent{
				Type: "text",
				Text: msg.Content,
			})
		}

		if msg.ContentImage != "" {
			var imageContents []*pb.MessageContent
			if err := json.Unmarshal([]byte(msg.ContentImage), &imageContents); err != nil {
				return nil, fmt.Errorf("failed to parse image content: %w", err)
			}
			contents = append(contents, imageContents...)
		}

		if len(contents) > 0 {
			messages = append(messages, buildSortedAgentsMessage(msg.Role, contents))
			continue
		}

		messages = append(messages, buildSortedAgentsMessage(msg.Role, msg.Content))
	}

	return sortedagents.NewSessionFromMessages(chatID, messages), nil
}

func (s *ChatService) runAgenticChat(
	ctx context.Context,
	chatID string,
	userID string,
	projectID string,
	model string,
	userContent interface{},
	history []dao.ChatMessageRow,
	ragChunks []rag.Result,
	ragEnabled bool,
	providerSettings *pb.ProviderSettings,
	stream func(*pb.ChatResponse) error,
) error {
	slog.Debug("service:Chat", "message", "running agentic chat", "chatId", chatID, "userID", userID, "projectID", projectID)
	agentPrompt, err := s.getSettingValue(CHAT_DEFAULT_PROMPT_KEY, defaultChatPrompt)
	if err != nil {
		slog.Error("service:Chat", "message", "failed to load chat default prompt", "error", err, "chatId", chatID, "userID", userID, "projectID", projectID)
		return fmt.Errorf("failed to load chat prompt")
	}

	session, err := s.buildAgenticSession(chatID, agentPrompt, history)
	if err != nil {
		slog.Error("service:Chat", "message", "failed to build agentic session", "error", err, "chatId", chatID, "userID", userID, "projectID", projectID)
		return fmt.Errorf("failed to prepare chat context")
	}

	webSearchSettings := webSearchSettings{
		APIURL: defaultBraveSearchAPIURL,
		Cost:   defaultBraveSearchCost,
	}
	if err := s.getJSONSetting(WEBSEARCH_SETTINGS_KEY, &webSearchSettings); err != nil {
		slog.Error("service:Chat", "message", "failed to load websearch settings", "error", err, "chatId", chatID, "userID", userID, "projectID", projectID)
		return fmt.Errorf("failed to load web search settings")
	}
	if strings.TrimSpace(webSearchSettings.APIURL) == "" {
		webSearchSettings.APIURL = defaultBraveSearchAPIURL
	}
	if strings.TrimSpace(webSearchSettings.Cost) == "" {
		webSearchSettings.Cost = defaultBraveSearchCost
	}

	scrapeSettings := cloudflareScrapeSettings{}
	if err := s.getJSONSetting(SCRAPE_CLOUDFLARE_SETTINGS_KEY, &scrapeSettings); err != nil {
		slog.Error("service:Chat", "message", "failed to load cloudflare scrape settings", "error", err, "chatId", chatID, "userID", userID, "projectID", projectID)
		return fmt.Errorf("failed to load scrape settings")
	}

	tools := []sortedagents.Tool{
		NewBraveSearchToolWithConfig(webSearchSettings.APIURL, webSearchSettings.APIKey),
		NewGetTimestampTool(),
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

	maxTurnsValue, err := s.getSettingValue(AGENTIC_MAX_TURNS_KEY, defaultAgenticMaxTurns)
	if err != nil {
		slog.Warn("service:Chat", "message", "failed to load agentic max turns, using default", "error", err, "chatId", chatID, "userID", userID, "projectID", projectID)
		maxTurnsValue = defaultAgenticMaxTurns
	}
	maxTurns, err := strconv.Atoi(maxTurnsValue)
	if err != nil || maxTurns <= 0 {
		maxTurns, _ = strconv.Atoi(defaultAgenticMaxTurns)
	}

	userMessage := buildSortedAgentsMessage("user", userContent)

	events := runner.RunStream(ctx, agent, userMessage, maxTurns, session)

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
	var scrapeAPIUsageTimeSeconds float64
	searchCostPerRequest, err := parseBraveSearchCost(webSearchSettings.Cost)
	if err != nil {
		slog.Warn("service:Chat", "message", "invalid brave search request cost, using default", "cost", webSearchSettings.Cost, "error", err, "chatId", chatID, "userID", userID, "projectID", projectID)
		searchCostPerRequest, _ = parseBraveSearchCost(defaultBraveSearchCost)
	}
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

		searchCost := float64(successfulWebSearchCalls) * searchCostPerRequest

		if _, err := s.dao.AddChatMessageWithTokens(userID, chatID, "assistant", assistantText, "", model, nonCachedInputTokens, outputTokens, cachedTokens, searchCost, successfulWebSearchCalls, scrapeAPIUsageTimeSeconds, referencesJSON, ragEnabled); err != nil {
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
			if e.ToolName == "browser_scrape" && e.Error == nil {
				if resultMap, ok := e.Result.(map[string]interface{}); ok {
					if usageSeconds, ok := resultMap[scrapeUsageSecondsResultKey].(float64); ok {
						scrapeAPIUsageTimeSeconds += usageSeconds
					}
				}
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
		daoSummary, err := s.dao.AddChatMessageWithTokens(userID, chatID, "assistant", assistantText, "", model, nonCachedInputTokens, outputTokens, cachedTokens, searchCost, successfulWebSearchCalls, scrapeAPIUsageTimeSeconds, referencesJSON, ragEnabled)
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
