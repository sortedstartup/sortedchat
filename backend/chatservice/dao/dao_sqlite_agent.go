package dao

import (
	"fmt"
	"log"
	"log/slog"

	"github.com/jmoiron/sqlx"
)

type SQLiteAgentsDAO struct {
	db *sqlx.DB
}

func NewSQLiteAgentsDAO(sqliteUrl string) *SQLiteAgentsDAO {
	db, err := sqlx.Open("sqlite3", sqliteUrl)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	return &SQLiteAgentsDAO{db: db}
}

// Agent CRUD
func (s *SQLiteAgentsDAO) CreateAgent(agent AgentRow) error {
	_, err := s.db.NamedExec(`
		INSERT INTO agents (id, name, description, system_prompt, provider, model, local_tools, created_at, updated_at)
		VALUES (:id, :name, :description, :system_prompt, :provider, :model, :local_tools, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, agent)
	if err != nil {
		slog.Error("dao_sqlite:CreateAgent", "message", "failed to create agent", "error", err, "agentID", agent.ID)
		return fmt.Errorf("failed to create agent")
	}
	return nil
}

func (s *SQLiteAgentsDAO) GetAgents() ([]AgentRow, error) {
	var agents []AgentRow
	err := s.db.Select(&agents, `SELECT * FROM agents ORDER BY created_at DESC`)
	if err != nil {
		slog.Error("dao_sqlite:GetAgents", "message", "failed to get agents", "error", err)
		return nil, fmt.Errorf("failed to get agents")
	}
	return agents, nil
}

func (s *SQLiteAgentsDAO) GetAgent(agentID string) (*AgentRow, error) {
	var agent AgentRow
	err := s.db.Get(&agent, `SELECT * FROM agents WHERE id = ?`, agentID)
	if err != nil {
		slog.Error("dao_sqlite:GetAgent", "message", "failed to get agent", "error", err, "agentID", agentID)
		return nil, fmt.Errorf("failed to get agent")
	}
	return &agent, nil
}

// Session CRUD
func (s *SQLiteAgentsDAO) CreateSession(session AgentSessionRow) error {
	_, err := s.db.NamedExec(`
		INSERT INTO agent_sessions (
			id, agent_id, user_id, status, title, 
			parent_session_id, created_at, updated_at
		) VALUES (
			:id, :agent_id, :user_id, :status, :title,
			:parent_session_id, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		)
	`, session)
	if err != nil {
		slog.Error("dao_sqlite:CreateSession", "message", "failed to create session", "error", err, "sessionID", session.ID)
		return fmt.Errorf("failed to create session")
	}
	return nil
}

func (s *SQLiteAgentsDAO) GetSession(sessionID string) (*AgentSessionRow, error) {
	var session AgentSessionRow
	err := s.db.Get(&session, `SELECT * FROM agent_sessions WHERE id = ?`, sessionID)
	if err != nil {
		slog.Error("dao_sqlite:GetSession", "message", "failed to get session", "error", err, "sessionID", sessionID)
		return nil, fmt.Errorf("failed to get session")
	}
	return &session, nil
}

func (s *SQLiteAgentsDAO) GetAgentSessions(agentID string) ([]AgentSessionRow, error) {
	var sessions []AgentSessionRow
	err := s.db.Select(&sessions, `
		SELECT * FROM agent_sessions 
		WHERE agent_id = ? 
		ORDER BY created_at DESC
	`, agentID)
	if err != nil {
		slog.Error("dao_sqlite:GetAgentSessions", "message", "failed to get agent sessions", "error", err, "agentID", agentID)
		return nil, fmt.Errorf("failed to get agent sessions")
	}
	return sessions, nil
}

// Message CRUD
func (s *SQLiteAgentsDAO) AddAgentMessage(message AgentMessageRow) error {
	_, err := s.db.NamedExec(`
		INSERT INTO agent_messages (
			id, session_id, sequence_number, role, type, 
			content, tool_name, tool_call_id, tool_args, created_at
		) VALUES (
			:id, :session_id, :sequence_number, :role, :type,
			:content, :tool_name, :tool_call_id, :tool_args, CURRENT_TIMESTAMP
		)
	`, message)
	if err != nil {
		slog.Error("dao_sqlite:AddAgentMessage", "message", "failed to add agent message", "error", err, "messageID", message.ID)
		return fmt.Errorf("failed to add agent message")
	}
	return nil
}

func (s *SQLiteAgentsDAO) GetAgentMessages(sessionID string) ([]AgentMessageRow, error) {
	var messages []AgentMessageRow
	err := s.db.Select(&messages, `
		SELECT * FROM agent_messages 
		WHERE session_id = ? 
		ORDER BY sequence_number ASC
	`, sessionID)
	if err != nil {
		slog.Error("dao_sqlite:GetAgentMessages", "message", "failed to get agent messages", "error", err, "sessionID", sessionID)
		return nil, fmt.Errorf("failed to get agent messages")
	}
	return messages, nil
}
