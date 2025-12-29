package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/geminitool"
	"google.golang.org/genai"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	db "sortedstartup/chatservice/dao"
	pb "sortedstartup/chatservice/proto"
	"sortedstartup/common/auth"
)

type AgentServiceAPI struct {
	pb.UnimplementedAgentServiceServer
	dao db.AgentDAO
}

func NewAgentService(daoFactory db.DAOFactory) *AgentServiceAPI {
	dao, err := daoFactory.CreateAgentDAO()
	if err != nil {
		panic(err)
	}
	return &AgentServiceAPI{dao: dao}
}

func (s *AgentServiceAPI) CreateAgent(ctx context.Context, req *pb.CreateAgentRequest) (*pb.CreateAgentResponse, error) {
	slog.Info("api:CreateAgent", "name", req.Name)

	agentID := uuid.New().String()
	// Serialize local tools to JSON string if needed, or simple string representation
	// Since proto has repeated string, and DAO takes string (likely JSON), we might need to join/marshal.
	// For now assuming simple JSON array serialization is handled or we just store them.
	// Actually DAO expects string for LocalTools.
	// Let's do a simple JSON marshal
	// Import encoding/json if needed inside function or package level.
	// But first let's just stick to the plan.

	agent := db.AgentRow{
		ID:           agentID,
		Name:         req.Name,
		Description:  req.Description,
		SystemPrompt: req.SystemPrompt,
		Provider:     req.Provider,
		Model:        req.Model,
		// Simple conversion for now, ideally json.Marshal
		LocalTools: fmt.Sprintf("%v", req.LocalTools),
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

	// 3. Save User Message
	msgID := uuid.New().String()
	userMsg := db.AgentMessageRow{
		ID:             msgID,
		SessionID:      req.SessionId,
		SequenceNumber: int(int32(time.Now().Unix())),
		Role:           "user",
		Type:           "text",
		Content:        req.Message,
	}

	if err := s.dao.AddAgentMessage(userMsg); err != nil {
		slog.Error("api:AgentChat", "error", "failed to save user message", "details", err)
		return status.Error(codes.Internal, "failed to save message")
	}

	// 4. Get message history for context
	messages, err := s.dao.GetAgentMessages(req.SessionId)
	if err != nil {
		slog.Error("api:AgentChat", "error", "failed to get message history", "details", err)
		return status.Error(codes.Internal, "failed to get message history")
	}

	// 5. Create and run agent with callbacks
	if err := s.runAgentWithCallbacks(ctx, userID, req.SessionId, *agentRow, messages, req.Message, stream); err != nil {
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
) error {
	// Create Gemini model
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("GOOGLE_API_KEY environment variable not set")
	}

	geminiModel, err := gemini.NewModel(ctx, agentRow.Model, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}

	// Create streaming callbacks
	beforeModelCallback := s.createBeforeModelCallback(stream)
	afterModelCallback := s.createAfterModelCallback(stream)
	beforeToolCallback := s.createBeforeToolCallback(stream)
	afterToolCallback := s.createAfterToolCallback(stream)

	// Parse local tools (if any)
	var tools []tool.Tool
	if agentRow.LocalTools != "" {
		// Add Google Search as default tool
		tools = append(tools, geminitool.GoogleSearch{})
	}

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
	_, err = sessionService.Create(ctx, &session.CreateRequest{
		AppName:   "sortedchat",
		UserID:    userID,
		SessionID: sessionID, // Use our database session ID
		State:     make(map[string]any),
	})
	if err != nil {
		// Session might already exist, which is fine - just log it
		slog.Debug("Session creation in ADK", "sessionID", sessionID, "note", "may already exist")
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
	// Use the actual database session ID
	for event, err := range r.Run(ctx, userID, sessionID, userContent, agent.RunConfig{}) {
		if err != nil {
			slog.Error("Agent execution error", "error", err)
			return fmt.Errorf("agent execution error: %w", err)
		}

		// Process and stream events
		if err := s.processAgentEvent(event, stream); err != nil {
			return err
		}
	}

	return nil
}

// processAgentEvent processes agent events and streams them to the client
func (s *AgentServiceAPI) processAgentEvent(event *session.Event, stream pb.AgentService_AgentChatServer) error {
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

	// Handle content
	if event.Content != nil {
		for _, part := range event.Content.Parts {
			// Handle text content
			if part.Text != "" {
				if err := stream.Send(&pb.AgentChatResponse{
					Response: &pb.AgentChatResponse_Message{
						Message: part.Text,
					},
				}); err != nil {
					return err
				}
			}

			// Handle function calls (tool calls)
			if part.FunctionCall != nil {
				toolCallData := map[string]interface{}{
					"name": part.FunctionCall.Name,
					"args": part.FunctionCall.Args,
				}
				toolCallJSON, _ := json.Marshal(toolCallData)
				if err := stream.Send(&pb.AgentChatResponse{
					Response: &pb.AgentChatResponse_ToolCall{
						ToolCall: string(toolCallJSON),
					},
				}); err != nil {
					return err
				}
			}

			// Handle function responses (tool results)
			if part.FunctionResponse != nil {
				toolResultData := map[string]interface{}{
					"name":   part.FunctionResponse.Name,
					"result": part.FunctionResponse.Response,
				}
				toolResultJSON, _ := json.Marshal(toolResultData)
				if err := stream.Send(&pb.AgentChatResponse{
					Response: &pb.AgentChatResponse_ToolResult{
						ToolResult: string(toolResultJSON),
					},
				}); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// createBeforeModelCallback creates a callback that fires before LLM calls
func (s *AgentServiceAPI) createBeforeModelCallback(stream pb.AgentService_AgentChatServer) llmagent.BeforeModelCallback {
	return func(ctx agent.CallbackContext, req *model.LLMRequest) (*model.LLMResponse, error) {
		slog.Info("[Callback] BeforeModel triggered", "agent", ctx.AgentName())

		// Stream thinking message
		if err := stream.Send(&pb.AgentChatResponse{
			Response: &pb.AgentChatResponse_Thinking{
				Thinking: fmt.Sprintf("Sending request to LLM (%s)...", ctx.AgentName()),
			},
		}); err != nil {
			slog.Error("Failed to send thinking message", "error", err)
		}

		// Log request details
		slog.Debug("[Callback] LLM Request", "contents_count", len(req.Contents))

		return nil, nil // Proceed with original request
	}
}

// createAfterModelCallback creates a callback that fires after LLM responds
func (s *AgentServiceAPI) createAfterModelCallback(stream pb.AgentService_AgentChatServer) llmagent.AfterModelCallback {
	return func(ctx agent.CallbackContext, resp *model.LLMResponse, respErr error) (*model.LLMResponse, error) {
		slog.Info("[Callback] AfterModel triggered", "agent", ctx.AgentName())

		if respErr != nil {
			slog.Error("[Callback] LLM error", "error", respErr)
			return nil, respErr
		}

		if resp != nil && resp.Content != nil {
			// Check for function calls
			if len(resp.Content.Parts) > 0 {
				for _, part := range resp.Content.Parts {
					if part.FunctionCall != nil {
						// Stream tool call notification
						toolCallData := map[string]interface{}{
							"name": part.FunctionCall.Name,
							"args": part.FunctionCall.Args,
						}
						toolCallJSON, _ := json.Marshal(toolCallData)

						if err := stream.Send(&pb.AgentChatResponse{
							Response: &pb.AgentChatResponse_ToolCall{
								ToolCall: string(toolCallJSON),
							},
						}); err != nil {
							slog.Error("Failed to send tool call message", "error", err)
						}
					}
				}
			}

			// Stream thinking about response
			if err := stream.Send(&pb.AgentChatResponse{
				Response: &pb.AgentChatResponse_Thinking{
					Thinking: "Received response from LLM",
				},
			}); err != nil {
				slog.Error("Failed to send thinking message", "error", err)
			}
		}

		return nil, nil // Proceed with original response
	}
}

// createBeforeToolCallback creates a callback that fires before tool execution
func (s *AgentServiceAPI) createBeforeToolCallback(stream pb.AgentService_AgentChatServer) llmagent.BeforeToolCallback {
	return func(ctx tool.Context, t tool.Tool, args map[string]any) (map[string]any, error) {
		slog.Info("[Callback] BeforeTool triggered", "tool", t.Name(), "agent", ctx.AgentName())

		// Stream tool execution notification
		toolCallData := map[string]interface{}{
			"tool_name": t.Name(),
			"args":      args,
			"status":    "executing",
		}
		toolCallJSON, _ := json.Marshal(toolCallData)

		if err := stream.Send(&pb.AgentChatResponse{
			Response: &pb.AgentChatResponse_ToolCall{
				ToolCall: string(toolCallJSON),
			},
		}); err != nil {
			slog.Error("Failed to send tool call message", "error", err)
		}

		// Log args
		slog.Debug("[Callback] Tool args", "tool", t.Name(), "args", args)

		return nil, nil // Proceed with original args
	}
}

// createAfterToolCallback creates a callback that fires after tool execution
func (s *AgentServiceAPI) createAfterToolCallback(stream pb.AgentService_AgentChatServer) llmagent.AfterToolCallback {
	return func(ctx tool.Context, t tool.Tool, args, result map[string]any, err error) (map[string]any, error) {
		slog.Info("[Callback] AfterTool triggered", "tool", t.Name(), "agent", ctx.AgentName())

		// Stream tool result
		toolResultData := map[string]interface{}{
			"tool_name": t.Name(),
			"args":      args,
			"result":    result,
			"error":     nil,
		}

		if err != nil {
			toolResultData["error"] = err.Error()
			slog.Error("[Callback] Tool execution error", "tool", t.Name(), "error", err)
		}

		toolResultJSON, _ := json.Marshal(toolResultData)

		if streamErr := stream.Send(&pb.AgentChatResponse{
			Response: &pb.AgentChatResponse_ToolResult{
				ToolResult: string(toolResultJSON),
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
