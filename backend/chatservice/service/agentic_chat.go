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
	"sortedstartup/chatservice/types"
)

type ChatMessageMetadata struct {
	WebSearches []ChatWebSearch `json:"websearches,omitempty"`
	Sources     []ChatSource    `json:"sources,omitempty"`
}

type ChatWebSearch struct {
	Query string `json:"query"`
}

type ChatSource struct {
	URL string `json:"url"`
}

func (c *ChatMessageMetadata) ToProto() *pb.AssistantMessageMetadata {
	if c == nil {
		return nil
	}
	proto := &pb.AssistantMessageMetadata{}
	for _, search := range c.WebSearches {
		proto.Websearches = append(proto.Websearches, &pb.AssistantMessageMetadata_WebSearch{
			Query: search.Query,
		})
	}
	for _, source := range c.Sources {
		proto.Sources = append(proto.Sources, &pb.AssistantMessageMetadata_Source{
			Url: source.URL,
		})
	}
	return proto
}

func (c *ChatMessageMetadata) ToJSON() (string, error) {
	if c == nil {
		return "", nil
	}
	bytes, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func NewChatMessageMetadata(webSearchQueries []string, sourceURLs []string) *ChatMessageMetadata {
	metadata := &ChatMessageMetadata{}
	seenQueries := make(map[string]struct{})
	seenSources := make(map[string]struct{})

	for _, query := range webSearchQueries {
		trimmed := strings.TrimSpace(query)
		if trimmed == "" {
			continue
		}
		if _, exists := seenQueries[trimmed]; exists {
			continue
		}
		seenQueries[trimmed] = struct{}{}
		metadata.WebSearches = append(metadata.WebSearches, ChatWebSearch{
			Query: trimmed,
		})
	}

	for _, sourceURL := range sourceURLs {
		trimmed := strings.TrimSpace(sourceURL)
		if trimmed == "" {
			continue
		}
		if _, exists := seenSources[trimmed]; exists {
			continue
		}
		seenSources[trimmed] = struct{}{}
		metadata.Sources = append(metadata.Sources, ChatSource{
			URL: trimmed,
		})
	}

	if len(metadata.WebSearches) == 0 && len(metadata.Sources) == 0 {
		return nil
	}

	return metadata
}

// this function convert normal text to sortedagents message struct type
// we have to check that if message is in string format or in content part format and then convert it accordingly to sortedagents message struct type
func buildSortedAgentsMessage(role string, content any) sortedagents.Message {
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
	case []*pb.MessageContent:
		parts := make(sortedagents.ContentParts, 0, len(value))
		for _, part := range value {
			if part == nil {
				continue
			}

			converted := sortedagents.ContentPart{
				Type: part.Type,
				Text: part.Text,
			}
			if part.ImageUrl != nil {
				converted.ImageURL = &sortedagents.ImageURLPart{URL: part.ImageUrl.Url}
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
		{
			Role:    "system",
			Content: sortedagents.TextContent(prompt),
		},
	}

	for _, msg := range history {
		slog.Debug("service:Chat", "message", "building sortedagents content from chat message", "messageId", msg.Id)

		var contents []*pb.MessageContent
		if strings.HasPrefix(msg.Content, "[") && strings.HasSuffix(msg.Content, "]") {
			var textContents []*pb.MessageContent
			if err := json.Unmarshal([]byte(msg.Content), &textContents); err != nil {
				contents = append(contents, &pb.MessageContent{
					Type: "text",
					Text: msg.Content,
				})
			} else {
				contents = append(contents, textContents...)
			}
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
	userContent any,
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
	var webSearchQueries []string
	var sourceURLs []string
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
		metadata := NewChatMessageMetadata(webSearchQueries, sourceURLs)
		metadataJSON, err := metadata.ToJSON()
		if err != nil {
			slog.Error("service:Chat", "message", "failed to serialize assistant metadata", "error", err, "chatId", chatID, "userID", userID, "projectID", projectID)
			metadataJSON = ""
		}

		if _, err := s.dao.AddChatMessageWithTokenCount(userID, chatID, "assistant", assistantText, "", model, nonCachedInputTokens, outputTokens, cachedTokens, searchCost, referencesJSON, metadataJSON, ragEnabled); err != nil {
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
				if query, ok := e.Args["query"].(string); ok {
					webSearchQueries = append(webSearchQueries, query)
				}
			case "browser_scrape":
				if url, ok := e.Args["url"].(string); ok {
					sourceURLs = append(sourceURLs, url)
				}
			}

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
				if resultMap, ok := e.Result.(map[string]any); ok {
					if results, ok := resultMap["results"].([]map[string]string); ok {
						for _, result := range results {
							if url := strings.TrimSpace(result["url"]); url != "" {
								sourceURLs = append(sourceURLs, url)
							}
						}
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
		metadata := NewChatMessageMetadata(webSearchQueries, sourceURLs)
		metadataJSON, err := metadata.ToJSON()
		if err != nil {
			slog.Error("service:Chat", "message", "failed to serialize assistant metadata", "error", err, "chatId", chatID, "userID", userID, "projectID", projectID)
			metadataJSON = ""
		}
		daoSummary, err := s.dao.AddChatMessageWithTokenCount(userID, chatID, "assistant", assistantText, "", model, nonCachedInputTokens, outputTokens, cachedTokens, searchCost, referencesJSON, metadataJSON, ragEnabled)
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
				Metadata:     metadata.ToProto(),
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
