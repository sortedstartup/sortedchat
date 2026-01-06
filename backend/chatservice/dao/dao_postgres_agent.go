package dao

import (
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
)

type PostgresAgentsDAO struct {
	db *sqlx.DB
}

func NewPostgresAgentsDAO(db *sqlx.DB) *PostgresAgentsDAO {
	return &PostgresAgentsDAO{db: db}
}

// Agent CRUD
func (p *PostgresAgentsDAO) CreateAgent(agent AgentRow) error {
	_, err := p.db.NamedExec(`
		INSERT INTO agents (id, name, description, system_prompt, provider, model, local_tools, created_at, updated_at)
		VALUES (:id, :name, :description, :system_prompt, :provider, :model, :local_tools, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, agent)
	if err != nil {
		slog.Error("dao_postgres:CreateAgent", "message", "failed to create agent", "error", err, "agentID", agent.ID)
		return fmt.Errorf("failed to create agent")
	}
	return nil
}

func (p *PostgresAgentsDAO) GetAgents() ([]AgentRow, error) {
	var agents []AgentRow
	err := p.db.Select(&agents, `SELECT * FROM agents ORDER BY created_at DESC`)
	if err != nil {
		slog.Error("dao_postgres:GetAgents", "message", "failed to get agents", "error", err)
		return nil, fmt.Errorf("failed to get agents")
	}
	return agents, nil
}

func (p *PostgresAgentsDAO) GetAgent(agentID string) (*AgentRow, error) {
	var agent AgentRow
	err := p.db.Get(&agent, `SELECT * FROM agents WHERE id = $1`, agentID)
	if err != nil {
		slog.Error("dao_postgres:GetAgent", "message", "failed to get agent", "error", err, "agentID", agentID)
		return nil, fmt.Errorf("failed to get agent")
	}
	return &agent, nil
}

// Session CRUD
func (p *PostgresAgentsDAO) CreateSession(session AgentSessionRow) error {
	_, err := p.db.NamedExec(`
		INSERT INTO agent_sessions (
			id, agent_id, user_id, status, title, 
			parent_session_id, created_at, updated_at
		) VALUES (
			:id, :agent_id, :user_id, :status, :title,
			:parent_session_id, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		)
	`, session)
	if err != nil {
		slog.Error("dao_postgres:CreateSession", "message", "failed to create session", "error", err, "sessionID", session.ID)
		return fmt.Errorf("failed to create session")
	}
	return nil
}

func (p *PostgresAgentsDAO) GetSession(sessionID string) (*AgentSessionRow, error) {
	var session AgentSessionRow
	err := p.db.Get(&session, `SELECT * FROM agent_sessions WHERE id = $1`, sessionID)
	if err != nil {
		slog.Error("dao_postgres:GetSession", "message", "failed to get session", "error", err, "sessionID", sessionID)
		return nil, fmt.Errorf("failed to get session")
	}
	return &session, nil
}

func (p *PostgresAgentsDAO) GetAgentSessions(agentID string) ([]AgentSessionRow, error) {
	var sessions []AgentSessionRow
	err := p.db.Select(&sessions, `
		SELECT * FROM agent_sessions 
		WHERE agent_id = $1 
		ORDER BY created_at DESC
	`, agentID)
	if err != nil {
		slog.Error("dao_postgres:GetAgentSessions", "message", "failed to get agent sessions", "error", err, "agentID", agentID)
		return nil, fmt.Errorf("failed to get agent sessions")
	}
	return sessions, nil
}

// Message CRUD
func (p *PostgresAgentsDAO) AddAgentMessage(message AgentMessageRow) error {
	_, err := p.db.NamedExec(`
		INSERT INTO agent_messages (
			id, session_id, sequence_number, role, type, 
			content, tool_name, tool_call_id, tool_args, created_at
		) VALUES (
			:id, :session_id, :sequence_number, :role, :type,
			:content, :tool_name, :tool_call_id, :tool_args, CURRENT_TIMESTAMP
		)
	`, message)
	if err != nil {
		slog.Error("dao_postgres:AddAgentMessage", "message", "failed to add agent message", "error", err, "messageID", message.ID)
		return fmt.Errorf("failed to add agent message")
	}
	return nil
}

func (p *PostgresAgentsDAO) GetAgentMessages(sessionID string) ([]AgentMessageRow, error) {
	var messages []AgentMessageRow
	err := p.db.Select(&messages, `
		SELECT * FROM agent_messages 
		WHERE session_id = $1 
		ORDER BY sequence_number ASC
	`, sessionID)
	if err != nil {
		slog.Error("dao_postgres:GetAgentMessages", "message", "failed to get agent messages", "error", err, "sessionID", sessionID)
		return nil, fmt.Errorf("failed to get agent messages")
	}
	return messages, nil
}

// Agent File Operations
func (p *PostgresAgentsDAO) SaveAgentFile(agentID, docsID, fileName, filePath string, fileSize int64, userID string) error {
	_, err := p.db.Exec(`
		INSERT INTO agent_docs (id, agent_id, docs_id, file_name, file_path, file_size, uploaded_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, docsID, agentID, docsID, fileName, filePath, fileSize, userID)
	if err != nil {
		slog.Error("dao_postgres:SaveAgentFile", "message", "failed to save agent file", "error", err, "agentID", agentID, "docsID", docsID)
		return fmt.Errorf("failed to save agent file")
	}
	return nil
}

func (p *PostgresAgentsDAO) GetAgentFiles(agentID string) ([]AgentDocumentRow, error) {
	var files []AgentDocumentRow
	err := p.db.Select(&files, `
		SELECT * FROM agent_docs 
		WHERE agent_id = $1 
		ORDER BY created_at DESC
	`, agentID)
	if err != nil {
		slog.Error("dao_postgres:GetAgentFiles", "message", "failed to get agent files", "error", err, "agentID", agentID)
		return nil, fmt.Errorf("failed to get agent files")
	}
	return files, nil
}

func (p *PostgresAgentsDAO) GetAgentFileByPath(agentID, filePath string) (*AgentDocumentRow, error) {
	var file AgentDocumentRow
	err := p.db.Get(&file, `
		SELECT * FROM agent_docs 
		WHERE agent_id = $1 AND file_path = $2
	`, agentID, filePath)
	if err != nil {
		slog.Error("dao_postgres:GetAgentFileByPath", "message", "failed to get agent file by path", "error", err, "agentID", agentID, "filePath", filePath)
		return nil, fmt.Errorf("failed to get agent file by path")
	}
	return &file, nil
}

func (p *PostgresAgentsDAO) DeleteAgentFile(agentID, docsID string) error {
	_, err := p.db.Exec(`
		DELETE FROM agent_docs 
		WHERE agent_id = $1 AND docs_id = $2
	`, agentID, docsID)
	if err != nil {
		slog.Error("dao_postgres:DeleteAgentFile", "message", "failed to delete agent file", "error", err, "agentID", agentID, "docsID", docsID)
		return fmt.Errorf("failed to delete agent file")
	}
	return nil
}


