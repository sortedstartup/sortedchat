package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
	"google.golang.org/genai"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"sortedstartup/chatservice/agents"
	db "sortedstartup/chatservice/dao"
	pb "sortedstartup/chatservice/proto"
	"sortedstartup/common/auth"
)

type AgentServiceAPI struct {
	pb.UnimplementedAgentServiceServer
	dao db.AgentDAO
}

func NewAgentService(daoFactory db.DAOFactory) (*AgentServiceAPI, error) {
	dao, err := daoFactory.CreateAgentDAO()
	if err != nil {
		return nil, err
	}
	return &AgentServiceAPI{dao: dao}, nil
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

	agent := db.AgentRow{
		ID:           agentID,
		Name:         req.Name,
		Description:  req.Description,
		SystemPrompt: req.SystemPrompt,
		Provider:     req.Provider,
		Model:        req.Model,
		LocalTools:   string(localToolsJSON),
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

func (s *AgentServiceAPI) GetAgents(ctx context.Context, req *pb.GetAgentsRequest) (*pb.GetAgentsResponse, error) {
	agents, err := s.dao.GetAgents()
	if err != nil {
		slog.Error("api:GetAgents", "error", err)
		return nil, status.Error(codes.Internal, "failed to get agents")
	}

	var pbAgents []*pb.Agent
	for _, a := range agents {
		pbAgents = append(pbAgents, &pb.Agent{
			Id:           a.ID,
			Name:         a.Name,
			Description:  a.Description,
			SystemPrompt: a.SystemPrompt,
			Provider:     a.Provider,
			Model:        a.Model,
			// LocalTools: parse or leave empty? Proto expects repeated string.
			// Ignoring complexity of parsing back for now as per "use dao" instruction focus.
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

// runAgentWithCallbacks creates an agent with callbacks and executes it
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
	// Create Gemini model
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("GOOGLE_API_KEY environment variable not set")
	}

	// Hardcode model for now (temporary)
	modelName := "gemini-3-flash-preview"
	slog.Info("Creating agent with model (hardcoded)", "model", modelName, "agent", agentRow.Name)

	geminiModel, err := gemini.NewModel(ctx, modelName, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}

	// Create streaming callbacks with shared state
	toolCallIDCounter := 0
	toolState := make(map[string]toolCallState)

	beforeModelCallback := s.createBeforeModelCallback(stream, modelName)
	afterModelCallback := s.createAfterModelCallback(stream, modelName)
	beforeToolCallback := s.createBeforeToolCallback(stream, modelName, toolState, &toolCallIDCounter)
	afterToolCallback := s.createAfterToolCallback(stream, modelName, toolState)

	// Parse local tools (if any)
	var tools []tool.Tool

	// Create filesystem tools with sandboxed path: ./agentid/sessionid
	workspacePath := filepath.Join(".", agentRow.ID, sessionID)
	fsTools, err := agents.NewFileSystemTools(agentRow.ID, workspacePath)
	if err != nil {
		return fmt.Errorf("failed to create filesystem tools: %w", err)
	}

	// Get filesystem tools for agent
	fileSystemTools, err := fsTools.GetTools()
	if err != nil {
		return fmt.Errorf("failed to get filesystem tools: %w", err)
	}
	tools = append(tools, fileSystemTools...)
	slog.Debug("Added filesystem tools to agent", "count", len(fileSystemTools), "workspace", workspacePath)

	// Add timestamp tool
	timestampTool, err := functiontool.New(functiontool.Config{
		Name:        "get_timestamp",
		Description: "Returns the current timestamp in RFC3339 format (ISO 8601). Useful for getting the current date and time.",
	}, getTimestamp)
	if err != nil {
		return fmt.Errorf("failed to create timestamp tool: %w", err)
	}
	tools = append(tools, timestampTool)

	// Create agent with callbacks
	llmAgent, err := llmagent.New(llmagent.Config{
		Name:                 agentRow.Name,
		Model:                geminiModel,
		Description:          agentRow.Description,
		Instruction:          agentRow.SystemPrompt,
		Tools:                tools,
		BeforeModelCallbacks: []llmagent.BeforeModelCallback{beforeModelCallback},
		AfterModelCallbacks:  []llmagent.AfterModelCallback{afterModelCallback},
		BeforeToolCallbacks:  []llmagent.BeforeToolCallback{beforeToolCallback},
		AfterToolCallbacks:   []llmagent.AfterToolCallback{afterToolCallback},
	})
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	// Create session service (in-memory for now)
	sessionService := session.InMemoryService()

	// Create the session in ADK session service with our database session ID
	createResp, err := sessionService.Create(ctx, &session.CreateRequest{
		AppName:   "sortedchat",
		UserID:    userID,
		SessionID: sessionID, // Use our database session ID
		State:     make(map[string]any),
	})
	if err != nil {
		// Session might already exist, which is fine - just log it
		slog.Debug("Session creation in ADK", "sessionID", sessionID, "note", "may already exist")
	}

	// Populate session history from messageHistory
	// This ensures the LLM has context of previous messages in this session
	if createResp != nil && len(messageHistory) > 0 {
		adkSession := createResp.Session
		for _, msg := range messageHistory {
			event := session.NewEvent(fmt.Sprintf("history_%d", msg.SequenceNumber))
			// Set the Author field to match our agent name - this is required by ADK runner
			event.Author = agentRow.Name

			// Convert our message types to genai.Content
			switch msg.Type {
			case "text":
				event.Content = &genai.Content{
					Role: msg.Role,
					Parts: []*genai.Part{
						{Text: msg.Content},
					},
				}
			case "tool_call":
				// Parse tool args
				var args map[string]any
				if msg.ToolArgs != nil {
					if err := json.Unmarshal([]byte(*msg.ToolArgs), &args); err != nil {
						slog.Warn("Failed to unmarshal tool_call arguments, skipping message",
							"error", err,
							"sequence", msg.SequenceNumber,
							"tool_name", getStringValue(msg.ToolName),
							"tool_call_id", msg.ToolCallID)
						continue
					}
				}
				event.Content = &genai.Content{
					Role: "model",
					Parts: []*genai.Part{
						{FunctionCall: &genai.FunctionCall{
							Name: getStringValue(msg.ToolName),
							Args: args,
						}},
					},
				}
			case "tool_result":
				// Parse tool result
				var result map[string]any
				if err := json.Unmarshal([]byte(msg.Content), &result); err != nil {
					slog.Warn("Failed to unmarshal tool_result content, skipping message",
						"error", err,
						"sequence", msg.SequenceNumber,
						"tool_name", getStringValue(msg.ToolName),
						"tool_call_id", msg.ToolCallID)
					continue
				}
				event.Content = &genai.Content{
					Role: "function",
					Parts: []*genai.Part{
						{FunctionResponse: &genai.FunctionResponse{
							Name:     getStringValue(msg.ToolName),
							Response: result,
						}},
					},
				}
			}

			if event.Content != nil {
				if err := sessionService.AppendEvent(ctx, adkSession, event); err != nil {
					slog.Warn("Failed to append history event", "error", err, "seq", msg.SequenceNumber)
				}
			}
		}
		slog.Debug("Loaded message history into ADK session", "count", len(messageHistory))
	}

	// Create runner
	r, err := runner.New(runner.Config{
		AppName:        "sortedchat",
		Agent:          llmAgent,
		SessionService: sessionService,
	})
	if err != nil {
		return fmt.Errorf("failed to create runner: %w", err)
	}

	// Create user message content
	userContent := &genai.Content{
		Role: "user",
		Parts: []*genai.Part{
			{Text: userMessage},
		},
	}

	// Execute agent and stream events
	var accumulatedText string

	// Use the actual database session ID
	// StreamingModeSSE enables token-by-token streaming from the LLM
	for event, err := range r.Run(ctx, userID, sessionID, userContent, agent.RunConfig{
		StreamingMode: agent.StreamingModeSSE,
	}) {
		if err != nil {
			slog.Error("Agent execution error", "error", err)
			return fmt.Errorf("agent execution error: %w", err)
		}

		// Stream to client (text is streamed here for true token-by-token streaming)
		if err := s.streamEventToClient(event, stream, modelName); err != nil {
			return err
		}

		// Accumulate and Save to DB
		if event.Content != nil {
			for _, part := range event.Content.Parts {
				// Text
				if part.Text != "" {
					accumulatedText += part.Text
				}

				// Tool Call
				if part.FunctionCall != nil {
					// Flush pending text
					if accumulatedText != "" {
						s.saveAgentMessage(sessionID, *nextSeq, "assistant", "text", accumulatedText, nil, nil, nil)
						*nextSeq++
						accumulatedText = ""
					}
					// Save Tool Call
					argsBytes, _ := json.Marshal(part.FunctionCall.Args)
					argsStr := string(argsBytes)
					// Look up tool call ID from state (generated in BeforeToolCallback)
					var toolCallIDPtr *string
					if state, ok := toolState[part.FunctionCall.Name]; ok {
						toolCallIDPtr = &state.id
					}
					s.saveAgentMessage(sessionID, *nextSeq, "assistant", "tool_call", "", strPtr(part.FunctionCall.Name), toolCallIDPtr, &argsStr)
					*nextSeq++
				}

				// Tool Result
				if part.FunctionResponse != nil {
					// Flush pending text (safety)
					if accumulatedText != "" {
						s.saveAgentMessage(sessionID, *nextSeq, "assistant", "text", accumulatedText, nil, nil, nil)
						*nextSeq++
						accumulatedText = ""
					}
					// Save Tool Result
					resBytes, _ := json.Marshal(part.FunctionResponse.Response)
					resStr := string(resBytes)
					// Look up tool call ID from state (generated in BeforeToolCallback)
					var toolCallIDPtr *string
					if state, ok := toolState[part.FunctionResponse.Name]; ok {
						toolCallIDPtr = &state.id
						// Clean up after saving - no longer needed
						delete(toolState, part.FunctionResponse.Name)
					}
					s.saveAgentMessage(sessionID, *nextSeq, "tool", "tool_result", resStr, strPtr(part.FunctionResponse.Name), toolCallIDPtr, nil)
					*nextSeq++
				}
			}
		}
	}

	// Final flush of accumulated text
	if accumulatedText != "" {
		s.saveAgentMessage(sessionID, *nextSeq, "assistant", "text", accumulatedText, nil, nil, nil)
		*nextSeq++
	}

	return nil
}

// streamEventToClient streams agent events to the client
// Text is streamed here (from event loop) for true token-by-token streaming
// Tool calls and results are handled by callbacks (beforeToolCallback, afterToolCallback)
func (s *AgentServiceAPI) streamEventToClient(event *session.Event, stream pb.AgentService_AgentChatServer, modelName string) error {
	// Handle errors
	if event.ErrorMessage != "" {
		if err := stream.Send(&pb.AgentChatResponse{
			Response: &pb.AgentChatResponse_Error{
				Error: event.ErrorMessage,
			},
		}); err != nil {
			return err
		}
		return nil
	}

	// Only stream text from PARTIAL events (streaming chunks)
	// Skip non-partial events to avoid duplicating the complete response
	// With SSE streaming: Partial=true means streaming chunk, Partial=false means final complete response
	if !event.Partial {
		return nil
	}

	// Stream text content (this gives us true streaming, token by token)
	if event.Content != nil {
		for _, part := range event.Content.Parts {
			if part.Text != "" {
				if err := stream.Send(&pb.AgentChatResponse{
					Response: &pb.AgentChatResponse_Content{
						Content: &pb.ContentEvent{
							Type:  "text",
							Text:  part.Text,
							Model: modelName,
						},
					},
				}); err != nil {
					slog.Error("Failed to send text message", "error", err)
					return err
				}
			}
			// Note: FunctionCall and FunctionResponse are handled by callbacks
			// to provide better UX (tool_call before execution, tool_result after)
		}
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
) {
	msgID := uuid.New().String()
	msg := db.AgentMessageRow{
		ID:             msgID,
		SessionID:      sessionID,
		SequenceNumber: seq,
		Role:           role,
		Type:           msgType,
		Content:        content,
		ToolName:       toolName,
		ToolCallID:     toolCallID,
		ToolArgs:       toolArgs,
	}
	if err := s.dao.AddAgentMessage(msg); err != nil {
		slog.Error("Failed to save agent message", "error", err)
	}
}

// Helper for string pointer
func strPtr(s string) *string {
	return &s
}

// toolCallState tracks state for a tool call (ID and start time)
type toolCallState struct {
	id        string
	startTime time.Time
}

// createBeforeModelCallback creates a callback that fires before LLM calls
func (s *AgentServiceAPI) createBeforeModelCallback(stream pb.AgentService_AgentChatServer, modelName string) llmagent.BeforeModelCallback {
	return func(ctx agent.CallbackContext, req *model.LLMRequest) (*model.LLMResponse, error) {
		// Just log for debugging, don't stream to UI (it's noise for the user)
		slog.Debug("[Callback] BeforeModel triggered", "agent", ctx.AgentName(), "contents_count", len(req.Contents))
		return nil, nil // Proceed with original request
	}
}

// createAfterModelCallback creates a callback that fires after LLM responds
// Note: Text streaming is handled by streamEventToClient (event loop), not here
func (s *AgentServiceAPI) createAfterModelCallback(stream pb.AgentService_AgentChatServer, modelName string) llmagent.AfterModelCallback {
	return func(ctx agent.CallbackContext, resp *model.LLMResponse, respErr error) (*model.LLMResponse, error) {
		if respErr != nil {
			slog.Error("[Callback] LLM error", "error", respErr)
			return nil, respErr
		}

		// Just log for debugging - text streaming is handled by event loop
		slog.Debug("[Callback] AfterModel triggered", "agent", ctx.AgentName())

		return nil, nil // Proceed with original response
	}
}

// createBeforeToolCallback creates a callback that fires before tool execution
func (s *AgentServiceAPI) createBeforeToolCallback(stream pb.AgentService_AgentChatServer, modelName string, toolState map[string]toolCallState, toolCallIDCounter *int) llmagent.BeforeToolCallback {
	return func(ctx tool.Context, t tool.Tool, args map[string]any) (map[string]any, error) {
		// Generate unique tool call ID and track start time
		*toolCallIDCounter++
		toolCallID := fmt.Sprintf("call_%d", *toolCallIDCounter)
		toolState[t.Name()] = toolCallState{
			id:        toolCallID,
			startTime: time.Now(),
		}

		// Stream tool call notification (this is the ONLY place we send ToolCall)
		argsJSON, _ := json.Marshal(args)

		if err := stream.Send(&pb.AgentChatResponse{
			Response: &pb.AgentChatResponse_ToolCall{
				ToolCall: &pb.ToolCall{
					Id:            toolCallID,
					Name:          t.Name(),
					ArgumentsJson: string(argsJSON),
				},
			},
		}); err != nil {
			slog.Error("Failed to send tool call message", "error", err)
		}

		return nil, nil // Proceed with original args
	}
}

// createAfterToolCallback creates a callback that fires after tool execution
func (s *AgentServiceAPI) createAfterToolCallback(stream pb.AgentService_AgentChatServer, modelName string, toolState map[string]toolCallState) llmagent.AfterToolCallback {
	return func(ctx tool.Context, t tool.Tool, args, result map[string]any, err error) (map[string]any, error) {
		// Get the tool call ID and calculate duration from stored state
		var toolCallID string
		var durationMs int64
		if state, ok := toolState[t.Name()]; ok {
			toolCallID = state.id
			durationMs = time.Since(state.startTime).Milliseconds()
			// Don't delete yet - the main loop needs it to save FunctionResponse with the ID
		} else {
			toolCallID = "unknown"
		}

		// Stream tool result
		resultJSON, _ := json.Marshal(result)

		success := err == nil
		errorMessage := ""
		if err != nil {
			errorMessage = err.Error()
			slog.Error("[Callback] Tool execution error", "tool", t.Name(), "error", err)
		}

		if streamErr := stream.Send(&pb.AgentChatResponse{
			Response: &pb.AgentChatResponse_ToolResult{
				ToolResult: &pb.ToolResult{
					Id:           toolCallID,
					Name:         t.Name(),
					ResultJson:   string(resultJSON),
					Success:      success,
					ErrorMessage: errorMessage,
					DurationMs:   durationMs,
				},
			},
		}); streamErr != nil {
			slog.Error("Failed to send tool result message", "error", streamErr)
		}

		return nil, nil // Proceed with original result
	}
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
			Id:             m.ID,
			SessionId:      m.SessionID,
			SequenceNumber: int32(m.SequenceNumber),
			Role:           m.Role,
			Type:           m.Type,
			Content:        m.Content,
			ToolName:       getStringValue(m.ToolName),
			ToolCallId:     getStringValue(m.ToolCallID),
			ToolArgs:       getStringValue(m.ToolArgs),
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

// Debug tools for testing

// AddArgs defines arguments for the add tool
type AddArgs struct {
	A float64 `json:"a" jsonschema:"First number to add"`
	B float64 `json:"b" jsonschema:"Second number to add"`
}

// add is a debug tool that adds two numbers
func add(ctx tool.Context, args *AddArgs) (float64, error) {
	result := args.A + args.B
	return result, nil
}

// SubArgs defines arguments for the subtract tool
type SubArgs struct {
	A float64 `json:"a" jsonschema:"First number (minuend)"`
	B float64 `json:"b" jsonschema:"Second number to subtract (subtrahend)"`
}

// subtract is a debug tool that subtracts two numbers
func subtract(ctx tool.Context, args *SubArgs) (float64, error) {
	result := args.A - args.B
	return result, nil
}

// GetTimestampArgs defines arguments for the timestamp tool (empty - no args needed)
type GetTimestampArgs struct{}

// TimestampResponse contains the current timestamp information
type TimestampResponse struct {
	Timestamp string `json:"timestamp" jsonschema:"Current timestamp in RFC3339 format (ISO 8601)"`
	Unix      int64  `json:"unix" jsonschema:"Unix timestamp in seconds since epoch"`
	UnixMilli int64  `json:"unix_milli" jsonschema:"Unix timestamp in milliseconds since epoch"`
}

// getTimestamp returns the current timestamp
func getTimestamp(ctx tool.Context, args *GetTimestampArgs) (TimestampResponse, error) {
	now := time.Now()
	return TimestampResponse{
		Timestamp: now.Format(time.RFC3339),
		Unix:      now.Unix(),
		UnixMilli: now.UnixMilli(),
	}, nil
}
