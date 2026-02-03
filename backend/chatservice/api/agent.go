package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"sortedstartup/chatservice/agents"
	db "sortedstartup/chatservice/dao"
	pb "sortedstartup/chatservice/proto"
	"sortedstartup/chatservice/settings"
	"sortedstartup/chatservice/sortedagents"
	"sortedstartup/common/auth"
)

const GenericConnectionError = "[ API/Connection Error ]"

type AgentServiceAPI struct {
	pb.UnimplementedAgentServiceServer
	dao             db.AgentDAO
	settingsManager *settings.SettingsManager
}

func NewAgentService(daoFactory db.DAOFactory, settingsManager *settings.SettingsManager) (*AgentServiceAPI, error) {
	dao, err := daoFactory.CreateAgentDAO()
	if err != nil {
		return nil, err
	}
	return &AgentServiceAPI{dao: dao, settingsManager: settingsManager}, nil
}

func (s *AgentServiceAPI) CreateAgent(ctx context.Context, req *pb.CreateAgentRequest) (*pb.CreateAgentResponse, error) {
	slog.Info("api:CreateAgent", "name", req.Name)

	agentID := uuid.New().String()

	// Serialize local tools to JSON string
	localToolsJSON, err := json.Marshal(req.LocalTools)
	if err != nil {
		slog.Error("api:CreateAgent", "error", "failed to marshal local tools", "details", err)
		return nil, status.Error(codes.Internal, "failed to serialize local tools")
	}

	// Serialize MCP servers to JSON string
	mcpServersJSON, err := serializeMCPServers(req.McpServers)
	if err != nil {
		slog.Error("api:CreateAgent", "error", "failed to marshal mcp servers", "details", err)
		return nil, status.Error(codes.Internal, "failed to serialize mcp servers")
	}

	agent := db.AgentRow{
		ID:           agentID,
		Name:         req.Name,
		Description:  req.Description,
		SystemPrompt: req.SystemPrompt,
		Provider:     req.Provider,
		Model:        req.Model,
		LocalTools:   string(localToolsJSON),
		MCPServers:   mcpServersJSON,
	}

	if err := s.dao.CreateAgent(agent); err != nil {
		slog.Error("api:CreateAgent", "error", err)
		return nil, status.Error(codes.Internal, "failed to create agent")
	}

	return &pb.CreateAgentResponse{
		Message: "Agent created successfully",
		AgentId: agentID,
	}, nil
}

func (s *AgentServiceAPI) UpdateAgent(ctx context.Context, req *pb.UpdateAgentRequest) (*pb.UpdateAgentResponse, error) {
	slog.Info("api:UpdateAgent", "agentID", req.AgentId)

	// Serialize MCP servers to JSON string
	mcpServersJSON, err := serializeMCPServers(req.McpServers)
	if err != nil {
		slog.Error("api:UpdateAgent", "error", "failed to marshal mcp servers", "details", err)
		return nil, status.Error(codes.Internal, "failed to serialize mcp servers")
	}

	agent := db.AgentRow{
		ID:           req.AgentId,
		Name:         req.Name,
		Description:  req.Description,
		SystemPrompt: req.SystemPrompt,
		Provider:     req.Provider,
		Model:        req.Model,
		MCPServers:   mcpServersJSON,
	}

	if err := s.dao.UpdateAgent(agent); err != nil {
		slog.Error("api:UpdateAgent", "error", err)
		return nil, status.Error(codes.Internal, "failed to update agent")
	}

	return &pb.UpdateAgentResponse{
		Message: "Agent updated successfully",
	}, nil
}

func (s *AgentServiceAPI) GetAgents(ctx context.Context, req *pb.GetAgentsRequest) (*pb.GetAgentsResponse, error) {
	agents, err := s.dao.GetAgents()
	if err != nil {
		slog.Error("api:GetAgents", "error", err)
		return nil, status.Error(codes.Internal, "failed to get agents")
	}

	var pbAgents []*pb.Agent
	for _, a := range agents {
		// Parse local tools from JSON string
		var localTools []string
		if a.LocalTools != "" {
			if err := json.Unmarshal([]byte(a.LocalTools), &localTools); err != nil {
				slog.Warn("api:GetAgents", "warning", "failed to parse local tools", "agentID", a.ID)
				localTools = []string{}
			}
		}

		// Parse MCP servers from JSON string
		mcpServers, err := deserializeMCPServers(a.MCPServers)
		if err != nil {
			slog.Warn("api:GetAgents", "warning", "failed to parse mcp servers", "agentID", a.ID, "error", err)
			mcpServers = []*pb.MCPServer{}
		}

		pbAgents = append(pbAgents, &pb.Agent{
			Id:           a.ID,
			Name:         a.Name,
			Description:  a.Description,
			SystemPrompt: a.SystemPrompt,
			Provider:     a.Provider,
			Model:        a.Model,
			LocalTools:   localTools,
			McpServers:   mcpServers,
		})
	}

	return &pb.GetAgentsResponse{Agents: pbAgents}, nil
}

func (s *AgentServiceAPI) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.CreateSessionResponse, error) {
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		return nil, err
	}

	sessionID := uuid.New().String()
	session := db.AgentSessionRow{
		ID:      sessionID,
		AgentID: req.AgentId,
		UserID:  userID,
		Status:  "active",
		// Title optional
	}

	if err := s.dao.CreateSession(session); err != nil {
		slog.Error("api:CreateSession", "error", err)
		return nil, status.Error(codes.Internal, "failed to create session")
	}

	return &pb.CreateSessionResponse{
		Message:   "Session created successfully",
		SessionId: sessionID,
	}, nil
}

func (s *AgentServiceAPI) GetSession(ctx context.Context, req *pb.GetSessionRequest) (*pb.GetSessionResponse, error) {
	session, err := s.dao.GetSession(req.SessionId)
	if err != nil {
		slog.Error("api:GetSession", "error", err)
		return nil, status.Error(codes.Internal, "failed to get session")
	}

	return &pb.GetSessionResponse{
		Session: &pb.Session{
			Id:      session.ID,
			AgentId: session.AgentID,
		},
	}, nil
}

func (s *AgentServiceAPI) GetSessions(ctx context.Context, req *pb.GetSessionsRequest) (*pb.GetSessionsResponse, error) {
	sessions, err := s.dao.GetAgentSessions(req.AgentId)
	if err != nil {
		slog.Error("api:GetSessions", "error", err)
		return nil, status.Error(codes.Internal, "failed to get sessions")
	}

	var pbSessions []*pb.Session
	for _, sess := range sessions {
		pbSessions = append(pbSessions, &pb.Session{
			Id:      sess.ID,
			AgentId: sess.AgentID,
			// UserId and Status are not in proto definition
		})
	}

	return &pb.GetSessionsResponse{Sessions: pbSessions}, nil
}

func (s *AgentServiceAPI) AgentChat(req *pb.AgentChatRequest, stream pb.AgentService_AgentChatServer) error {
	ctx := stream.Context()

	// 1. Get User ID
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		return err
	}
	_ = userID // unused for now, but good to validate auth

	// 2. Get session and agent info
	session, err := s.dao.GetSession(req.SessionId)
	if err != nil {
		slog.Error("api:AgentChat", "error", "failed to get session", "details", err)
		return status.Error(codes.Internal, "failed to get session")
	}

	agentRow, err := s.dao.GetAgent(session.AgentID)
	if err != nil {
		slog.Error("api:AgentChat", "error", "failed to get agent", "details", err)
		return status.Error(codes.Internal, "failed to get agent")
	}

	// 3. Get message history for context & sequence number
	messages, err := s.dao.GetAgentMessages(req.SessionId)
	if err != nil {
		slog.Error("api:AgentChat", "error", "failed to get message history", "details", err)
		return status.Error(codes.Internal, "failed to get message history")
	}

	// Repair DB state: If the last message was from a user (indicating a previous failure),
	// insert a blank assistant message to close the turn and maintain User -> Assistant alternation.
	if len(messages) > 0 && messages[len(messages)-1].Role == "user" {
		lastMsg := messages[len(messages)-1]
		repairSeq := lastMsg.SequenceNumber + 1

		slog.Info("api:AgentChat", "info", "Repairing dangling user message", "lastMsgID", lastMsg.ID)

		// Insert placeholder assistant message
		placeholderMsg := db.AgentMessageRow{
			ID:             uuid.New().String(),
			SessionID:      req.SessionId,
			SequenceNumber: repairSeq,
			Role:           "assistant",
			Type:           "text",
			Content:        " ", // Blank content as requested
			Success:        false,
			ErrorMessage:   strPtr("System: Conversation turn repaired"),
		}

		if err := s.dao.AddAgentMessage(placeholderMsg); err != nil {
			slog.Error("api:AgentChat", "error", "failed to save repair message", "details", err)
			// Continue anyway, as the next user message might still be saved, but we tried.
		}

		// Append to local messages slice so context sent to LLM is correct
		messages = append(messages, placeholderMsg)
	}

	// Determine sequence number
	nextSeq := 1
	if len(messages) > 0 {
		nextSeq = messages[len(messages)-1].SequenceNumber + 1
	}

	// 4. Save User Message
	msgID := uuid.New().String()
	userMsg := db.AgentMessageRow{
		ID:             msgID,
		SessionID:      req.SessionId,
		SequenceNumber: nextSeq,
		Role:           "user",
		Type:           "text",
		Content:        req.Message,
		Success:        true,
	}

	if err := s.dao.AddAgentMessage(userMsg); err != nil {
		slog.Error("api:AgentChat", "error", "failed to save user message", "details", err)
		return status.Error(codes.Internal, "failed to save message")
	}
	nextSeq++

	// 5. Create and run agent with callbacks
	if err := s.runAgentWithCallbacks(ctx, userID, req.SessionId, *agentRow, messages, req.Message, stream, &nextSeq); err != nil {
		slog.Error("api:AgentChat", "error", "agent execution failed", "details", err)
		return status.Error(codes.Internal, fmt.Sprintf("agent execution failed: %v", err))
	}

	return nil
}

// runAgentWithCallbacks creates an agent and executes it using sortedagents
func (s *AgentServiceAPI) runAgentWithCallbacks(
	ctx context.Context,
	userID string,
	sessionID string,
	agentRow db.AgentRow,
	messageHistory []db.AgentMessageRow,
	userMessage string,
	stream pb.AgentService_AgentChatServer,
	nextSeq *int,
) error {
	modelName := agentRow.Model
	slog.Info("Creating agent with sortedagents", "model", modelName, "agent", agentRow.Name)

	// Get provider settings
	providerSettings, err := s.settingsManager.GetProviderSetting(agentRow.Provider)
	if err != nil {
		slog.Error("Failed to get provider settings", "error", err, "provider", agentRow.Provider)
		return status.Error(codes.Internal, "failed to get provider settings")
	}
	if providerSettings == nil {
		slog.Error("Provider settings not found", "provider", agentRow.Provider)
		return status.Error(codes.NotFound, fmt.Sprintf("provider settings for %s not found", agentRow.Provider))
	}

	apiKey := providerSettings.ApiKey
	apiURL := providerSettings.ApiUrl

	// Create filesystem tools with sandboxed path: sortedchat-data/agents/agentid/sessionid
	workspacePath := filepath.Join("sortedchat-data", "agents", agentRow.ID, sessionID)
	fsTools, err := agents.NewFileSystemTools(agentRow.ID, workspacePath)
	if err != nil {
		return fmt.Errorf("failed to create filesystem tools: %w", err)
	}

	// Load agent files into workspace before starting the session
	agentFiles, err := s.dao.GetAgentFiles(agentRow.ID)
	if err != nil {
		slog.Warn("Failed to get agent files", "error", err, "agentID", agentRow.ID)
	} else if len(agentFiles) > 0 {
		// Convert to AgentFileInfo for loading
		var fileInfos []agents.AgentFileInfo
		for _, f := range agentFiles {
			fileInfos = append(fileInfos, agents.AgentFileInfo{
				DocsID:   f.DocsID,
				FilePath: f.FilePath,
			})
		}

		// Load files into workspace
		if err := agents.LoadAgentFilesIntoWorkspace(fileInfos, workspacePath, "filestore"); err != nil {
			slog.Warn("Failed to load agent files into workspace", "error", err, "agentID", agentRow.ID)
		} else {
			slog.Info("Loaded agent files into workspace", "count", len(fileInfos), "agentID", agentRow.ID, "workspace", workspacePath)
		}
	}

	// Build tools using sortedagents ToolBuilder
	toolBuilder := sortedagents.NewToolBuilder()

	// Add filesystem tools
	toolBuilder.AddTypedFunc(sortedagents.NewTool("read_file", "Reads the contents of a file from the agent's workspace. Optionally show line numbers. Returns the file content as a string.", fsTools.ReadFile))
	toolBuilder.AddTypedFunc(sortedagents.NewTool("write_file", "Writes content to a file in the agent's workspace. Creates parent directories if needed. Returns success message.", fsTools.WriteFile))
	toolBuilder.AddTypedFunc(sortedagents.NewTool("list_dir", "Lists files and directories in the specified path within the agent's workspace. Returns object with 'files' array containing file information (name, type, size).", fsTools.ListDir))
	toolBuilder.AddTypedFunc(sortedagents.NewTool("create_dir", "Creates a directory in the agent's workspace. Creates parent directories if needed. Returns success message.", fsTools.CreateDir))
	toolBuilder.AddTypedFunc(sortedagents.NewTool("file_exists", "Checks if a file or directory exists in the agent's workspace. Returns true if exists, false otherwise.", fsTools.FileExists))
	toolBuilder.AddTypedFunc(sortedagents.NewTool("move_file", "Moves or renames a file within the agent's workspace. Can move files between directories. Returns success message.", fsTools.MoveFile))
	toolBuilder.AddTypedFunc(sortedagents.NewTool("append_to_file", "Appends content to the end of a file in the agent's workspace. Creates the file if it doesn't exist. Returns success message.", fsTools.AppendToFile))
	toolBuilder.AddTypedFunc(sortedagents.NewTool("read_lines", "Reads specific lines from a file (1-indexed, inclusive). Useful for reading large files partially. Returns lines with line numbers.", fsTools.ReadLines))
	toolBuilder.AddTypedFunc(sortedagents.NewTool("delete_lines", "Deletes specific lines from a file (1-indexed, inclusive). Rewrites the file without the deleted lines. Returns success message.", fsTools.DeleteLines))
	toolBuilder.AddTypedFunc(sortedagents.NewTool("replace_lines", "Replaces specific lines in a file with new content (1-indexed, inclusive). Useful for precise file editing. Returns success message.", fsTools.ReplaceLines))
	toolBuilder.AddTypedFunc(sortedagents.NewTool("search_regex", "Searches for a regex pattern in a file. Returns array of matches with line numbers, line content, and matched text.", fsTools.SearchRegex))
	toolBuilder.AddTypedFunc(sortedagents.NewTool("regex_replace_all", "Replaces all occurrences of a regex pattern in a file with replacement text. Supports capture groups. Returns success message with count.", fsTools.RegexReplaceAll))

	// Add timestamp tool
	toolBuilder.AddFunc("get_timestamp", "Returns the current timestamp in RFC3339 format (ISO 8601). Useful for getting the current date and time.", getTimestamp)

	tools := toolBuilder.Build()
	slog.Debug("Added tools to agent", "count", len(tools), "workspace", workspacePath)

	// Create sortedagents session and load message history
	session := sortedagents.NewSessionWithID(sessionID)

	// Always add system prompt as the first message
	session.AddMessage(sortedagents.Message{
		Role:    "system",
		Content: agentRow.SystemPrompt,
	})

	// Load message history if exists
	if len(messageHistory) > 0 {
		for _, msg := range messageHistory {
			session.AddMessage(convertDBMessageToSortedAgentsMessage(msg))
		}
		slog.Debug("Loaded message history into sortedagents session", "count", len(messageHistory))
	}

	// Create agent
	agent := sortedagents.NewAgent(
		agentRow.Name,
		agentRow.SystemPrompt,
		modelName,
		tools,
	)

	// Create runner with configured LLM
	llm := sortedagents.NewOpenAILLMWithConfig(apiKey, apiURL, modelName)
	runner := sortedagents.NewRunnerWithLLM(llm)

	// Execute agent with streaming
	maxTurns := 15
	eventChan := runner.RunStream(ctx, agent, userMessage, maxTurns, session)

	// Track tool call IDs and accumulated text
	idCounter := newToolCallIDCounter()
	var accumulatedText string
	var toolCallStartTimes = make(map[string]time.Time)

	// Process events and stream to client
	for event := range eventChan {
		switch e := event.(type) {
		case *sortedagents.TextChunkEvent:
			// Stream text chunk to client
			accumulatedText += e.Chunk
			if err := stream.Send(&pb.AgentChatResponse{
				Response: &pb.AgentChatResponse_Content{
					Content: &pb.ContentEvent{
						Type:  "text",
						Text:  e.Chunk,
						Model: modelName,
					},
				},
			}); err != nil {
				slog.Error("Failed to send text chunk", "error", err)
				if accumulatedText != "" {
					s.saveAgentMessage(sessionID, *nextSeq, "assistant", "text", accumulatedText, nil, nil, nil, nil, false, strPtr(err.Error()), 0)
					*nextSeq++
				}
				return err
			}

		case *sortedagents.ToolCallStartEvent:
			// Flush any accumulated text before tool call
			if accumulatedText != "" {
				s.saveAgentMessage(sessionID, *nextSeq, "assistant", "text", accumulatedText, nil, nil, nil, nil, true, nil, 0)
				*nextSeq++
				accumulatedText = ""
			}

			// Generate unique tool call ID
			toolCallID := idCounter.getOrCreateID(e.ToolName)
			toolCallStartTimes[toolCallID] = time.Now()

			// Marshal arguments
			argsJSON, _ := json.Marshal(e.Args)
			argsStr := string(argsJSON)

			// Save tool call to DB
			var thoughtSigPtr *string
			if e.ThoughtSignature != "" {
				thoughtSigPtr = strPtr(e.ThoughtSignature)
			}
			s.saveAgentMessage(sessionID, *nextSeq, "assistant", "tool_call", "", strPtr(e.ToolName), strPtr(toolCallID), &argsStr, thoughtSigPtr, true, nil, 0)
			*nextSeq++

			// Stream tool call to client
			if err := stream.Send(&pb.AgentChatResponse{
				Response: &pb.AgentChatResponse_ToolCall{
					ToolCall: &pb.ToolCall{
						Id:            toolCallID,
						Name:          e.ToolName,
						ArgumentsJson: string(argsJSON),
					},
				},
			}); err != nil {
				slog.Error("Failed to send tool call", "error", err)
				return err
			}

		case *sortedagents.ToolCallEndEvent:
			// Get tool call ID
			toolCallID := idCounter.getOrCreateID(e.ToolName)

			// Calculate duration
			var durationMs int64
			if startTime, exists := toolCallStartTimes[toolCallID]; exists {
				durationMs = time.Since(startTime).Milliseconds()
				delete(toolCallStartTimes, toolCallID)
			}

			// Marshal result
			resultJSON, _ := json.Marshal(e.Result)
			resultStr := string(resultJSON)

			// Save tool result to DB
			success := e.Error == nil
			var content string
			if success {
				content = resultStr
			} else {
				content = fmt.Sprintf("Error: %v", e.Error)
			}
			var errMsg *string
			if !success {
				msg := e.Error.Error()
				errMsg = &msg
			}
			s.saveAgentMessage(sessionID, *nextSeq, "tool", "tool_result", content, strPtr(e.ToolName), strPtr(toolCallID), nil, nil, success, errMsg, durationMs)
			*nextSeq++

			// Stream tool result to client
			errorMessage := ""
			if e.Error != nil {
				errorMessage = e.Error.Error()
			}

			if err := stream.Send(&pb.AgentChatResponse{
				Response: &pb.AgentChatResponse_ToolResult{
					ToolResult: &pb.ToolResult{
						Id:           toolCallID,
						Name:         e.ToolName,
						ResultJson:   resultStr,
						Success:      success,
						ErrorMessage: errorMessage,
						DurationMs:   durationMs,
					},
				},
			}); err != nil {
				slog.Error("Failed to send tool result", "error", err)
				return err
			}

			// Clear the tool call ID after completion
			idCounter.clearID(e.ToolName)

		case *sortedagents.CompleteEvent:
			// Save any final accumulated text
			if accumulatedText != "" {
				s.saveAgentMessage(sessionID, *nextSeq, "assistant", "text", accumulatedText, nil, nil, nil, nil, true, nil, 0)
				*nextSeq++
			}
			slog.Info("Agent execution completed", "agent", agentRow.Name)
			return nil

		case *sortedagents.ErrorEvent:
			// Flush any accumulated text before error
			if accumulatedText != "" {
				s.saveAgentMessage(sessionID, *nextSeq, "assistant", "text", accumulatedText, nil, nil, nil, nil, true, nil, 0)
				*nextSeq++
			} else {
				// If no text was accumulated, save a generic error message to the DB
				// so the user sees something went wrong and the conversation role is preserved.
				// The real error is logged to the server logs.
				s.saveAgentMessage(sessionID, *nextSeq, "assistant", "text", GenericConnectionError, nil, nil, nil, nil, false, strPtr(e.Error.Error()), 0)
				*nextSeq++
			}

			// Determine error type
			errText := e.Error.Error()
			errType := pb.AgentChatError_UNKNOWN
			errCode := int32(0)

			if strings.Contains(strings.ToLower(errText), "loading model") {
				errType = pb.AgentChatError_MODEL_LOADING
				// Extract code if possible (e.g., from "status: 503")
				if strings.Contains(errText, "503") {
					errCode = 503
				} else if strings.Contains(errText, "500") {
					errCode = 500
				}
			} else if strings.Contains(strings.ToLower(errText), "api request failed") {
				errType = pb.AgentChatError_PROVIDER_ERROR
			}

			// Stream structured error to client
			if err := stream.Send(&pb.AgentChatResponse{
				Response: &pb.AgentChatResponse_Error{
					Error: &pb.AgentChatError{
						Type:    errType,
						Message: errText,
						Code:    errCode,
					},
				},
			}); err != nil {
				slog.Error("Failed to send error message", "error", err)
			}
			return e.Error
		}
	}

	// Save any remaining accumulated text
	if accumulatedText != "" {
		s.saveAgentMessage(sessionID, *nextSeq, "assistant", "text", accumulatedText, nil, nil, nil, nil, true, nil, 0)
		*nextSeq++
	}

	return nil
}

// Helper to save agent message to DB
func (s *AgentServiceAPI) saveAgentMessage(
	sessionID string,
	seq int,
	role string,
	msgType string,
	content string,
	toolName *string,
	toolCallID *string,
	toolArgs *string,
	thoughtSignature *string,
	success bool,
	errorMessage *string,
	runTimeMs int64,
) {
	msgID := uuid.New().String()
	msg := db.AgentMessageRow{
		ID:               msgID,
		SessionID:        sessionID,
		SequenceNumber:   seq,
		Role:             role,
		Type:             msgType,
		Content:          content,
		ToolName:         toolName,
		ToolCallID:       toolCallID,
		ToolArgs:         toolArgs,
		ThoughtSignature: thoughtSignature,
		Success:          success,
		ErrorMessage:     errorMessage,
		RunTimeMs:        runTimeMs,
	}

	if err := s.dao.AddAgentMessage(msg); err != nil {
		slog.Error("Failed to save agent message", "error", err)
	}
}

// Helper for string pointer
func strPtr(s string) *string {
	return &s
}

func (s *AgentServiceAPI) GetAgentMessages(ctx context.Context, req *pb.GetAgentMessagesRequest) (*pb.GetAgentMessagesResponse, error) {
	msgs, err := s.dao.GetAgentMessages(req.SessionId)
	if err != nil {
		slog.Error("api:GetAgentMessages", "error", err)
		return nil, status.Error(codes.Internal, "failed to get messages")
	}

	var pbMsgs []*pb.AgentMessage
	for _, m := range msgs {
		pbMsgs = append(pbMsgs, &pb.AgentMessage{
			Id:               m.ID,
			SessionId:        m.SessionID,
			SequenceNumber:   int32(m.SequenceNumber),
			Role:             m.Role,
			Type:             m.Type,
			Content:          m.Content,
			ToolName:         getStringValue(m.ToolName),
			ToolCallId:       getStringValue(m.ToolCallID),
			ToolArgs:         getStringValue(m.ToolArgs),
			Success:          m.Success,
			ErrorMessage:     getStringValue(m.ErrorMessage),
			RunTimeMs:        m.RunTimeMs,
			ThoughtSignature: getStringValue(m.ThoughtSignature),
		})
	}

	return &pb.GetAgentMessagesResponse{Messages: pbMsgs}, nil
}

func (s *AgentServiceAPI) GetAgentMessage(ctx context.Context, req *pb.GetAgentMessageRequest) (*pb.GetAgentMessageResponse, error) {
	return nil, status.Error(codes.Unimplemented, "GetAgentMessage not implemented in DAO")
}

func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// getTimestamp returns the current timestamp (sortedagents format)
func getTimestamp(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	now := time.Now()
	return map[string]interface{}{
		"timestamp":  now.Format(time.RFC3339),
		"unix":       now.Unix(),
		"unix_milli": now.UnixMilli(),
	}, nil
}

// convertDBMessageToSortedAgentsMessage converts a DB message to sortedagents.Message format
func convertDBMessageToSortedAgentsMessage(msg db.AgentMessageRow) sortedagents.Message {
	switch msg.Type {
	case "text":
		return sortedagents.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	case "tool_call":
		sig := getStringValue(msg.ThoughtSignature)
		var extra *sortedagents.ExtraContent
		if sig != "" {
			extra = &sortedagents.ExtraContent{
				Google: &sortedagents.GoogleExtra{
					ThoughtSignature: sig,
				},
			}
		}
		return sortedagents.Message{
			Role: "assistant",
			ToolCalls: []sortedagents.ToolCall{{
				ID:   getStringValue(msg.ToolCallID),
				Type: "function",
				Function: sortedagents.Function{
					Name:             getStringValue(msg.ToolName),
					Arguments:        getStringValue(msg.ToolArgs),
					ThoughtSignature: sig,
				},
				ThoughtSignature: sig,
				ExtraContent:     extra,
			}},
		}
	case "tool_result":
		return sortedagents.Message{
			Role:       "tool",
			Content:    msg.Content,
			ToolCallID: getStringValue(msg.ToolCallID),
		}
	default:
		return sortedagents.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}
}

// toolCallIDCounter generates unique tool call IDs for streaming
type toolCallIDCounter struct {
	counter     int
	callToIDMap map[string]string // toolName -> toolCallID
}

func newToolCallIDCounter() *toolCallIDCounter {
	return &toolCallIDCounter{
		counter:     0,
		callToIDMap: make(map[string]string),
	}
}

func (tc *toolCallIDCounter) getOrCreateID(toolName string) string {
	if id, exists := tc.callToIDMap[toolName]; exists {
		return id
	}
	tc.counter++
	id := fmt.Sprintf("call_%d", tc.counter)
	tc.callToIDMap[toolName] = id
	return id
}

func (tc *toolCallIDCounter) clearID(toolName string) {
	delete(tc.callToIDMap, toolName)
}

// serializeMCPServers converts proto MCP servers to JSON string for storage
func serializeMCPServers(servers []*pb.MCPServer) (string, error) {
	if len(servers) == 0 {
		return "[]", nil
	}

	// Convert proto to a more storage-friendly format
	type MCPServerStorage struct {
		ServerName           string            `json:"server_name"`
		IsEnabled            bool              `json:"is_enabled"`
		TransportType        string            `json:"transport_type"` // "stdio" or "http"
		Command              string            `json:"command,omitempty"`
		Arguments            []string          `json:"arguments,omitempty"`
		EnvironmentVariables map[string]string `json:"environment_variables,omitempty"`
		URL                  string            `json:"url,omitempty"`
		Headers              map[string]string `json:"headers,omitempty"`
		TimeoutSeconds       int32             `json:"timeout_seconds,omitempty"`
	}

	var storageServers []MCPServerStorage
	for _, s := range servers {
		storage := MCPServerStorage{
			ServerName: s.ServerName,
			IsEnabled:  s.IsEnabled,
		}

		switch t := s.Transport.(type) {
		case *pb.MCPServer_Stdio:
			storage.TransportType = "stdio"
			storage.Command = t.Stdio.Command
			storage.Arguments = t.Stdio.Arguments
			storage.EnvironmentVariables = t.Stdio.EnvironmentVariables
		case *pb.MCPServer_Http:
			storage.TransportType = "http"
			storage.URL = t.Http.Url
			storage.Headers = t.Http.Headers
			storage.TimeoutSeconds = t.Http.TimeoutSeconds
		}

		storageServers = append(storageServers, storage)
	}

	data, err := json.Marshal(storageServers)
	if err != nil {
		return "[]", fmt.Errorf("failed to marshal MCP servers: %w", err)
	}

	return string(data), nil
}

// deserializeMCPServers converts JSON string from storage to proto MCP servers
func deserializeMCPServers(jsonStr string) ([]*pb.MCPServer, error) {
	if jsonStr == "" || jsonStr == "[]" {
		return []*pb.MCPServer{}, nil
	}

	type MCPServerStorage struct {
		ServerName           string            `json:"server_name"`
		IsEnabled            bool              `json:"is_enabled"`
		TransportType        string            `json:"transport_type"`
		Command              string            `json:"command,omitempty"`
		Arguments            []string          `json:"arguments,omitempty"`
		EnvironmentVariables map[string]string `json:"environment_variables,omitempty"`
		URL                  string            `json:"url,omitempty"`
		Headers              map[string]string `json:"headers,omitempty"`
		TimeoutSeconds       int32             `json:"timeout_seconds,omitempty"`
	}

	var storageServers []MCPServerStorage
	if err := json.Unmarshal([]byte(jsonStr), &storageServers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal MCP servers: %w", err)
	}

	var servers []*pb.MCPServer
	for _, s := range storageServers {
		server := &pb.MCPServer{
			ServerName: s.ServerName,
			IsEnabled:  s.IsEnabled,
		}

		switch s.TransportType {
		case "stdio":
			server.Transport = &pb.MCPServer_Stdio{
				Stdio: &pb.MCPStdio{
					Command:              s.Command,
					Arguments:            s.Arguments,
					EnvironmentVariables: s.EnvironmentVariables,
				},
			}
		case "http":
			server.Transport = &pb.MCPServer_Http{
				Http: &pb.MCPHttp{
					Url:            s.URL,
					Headers:        s.Headers,
					TimeoutSeconds: s.TimeoutSeconds,
				},
			}
		}

		servers = append(servers, server)
	}

	return servers, nil
}
