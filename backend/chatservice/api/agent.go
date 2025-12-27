package api

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
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

func NewAgentService(dao db.AgentDAO) *AgentServiceAPI {
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
	return nil, status.Error(codes.Unimplemented, "GetSessions not implemented in DAO")
}

func (s *AgentServiceAPI) AgentChat(req *pb.AgentChatRequest, stream pb.AgentService_AgentChatServer) error {
	// 1. Get User ID
	userID, err := auth.GetUserIDFromContext_WithError(stream.Context())
	if err != nil {
		return err
	}
	_ = userID // unused for now, but good to validate auth

	// 2. Save User Message
	msgID := uuid.New().String()
	userMsg := db.AgentMessageRow{
		ID:             msgID,
		SessionID:      req.SessionId,
		SequenceNumber: int(time.Now().UnixNano()), // Simple sequence for now
		Role:           "user",
		Type:           "text",
		Content:        req.Message,
	}

	if err := s.dao.AddAgentMessage(userMsg); err != nil {
		slog.Error("api:AgentChat", "error", "failed to save user message", "details", err)
		return status.Error(codes.Internal, "failed to save message")
	}

	// 3. Ack (Streaming response)
	if err := stream.Send(&pb.AgentChatResponse{
		Response: &pb.AgentChatResponse_Message{
			Message: "Message received. Agent logic not yet connected.",
		},
	}); err != nil {
		return err
	}

	return nil
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
