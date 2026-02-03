package dao

import (
	"fmt"
	"log"
	"log/slog"
	"time"

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

	// Set busy timeout to 30 seconds
	_, err = db.Exec("PRAGMA busy_timeout = 30000;")
	if err != nil {
		slog.Error("dao_sqlite_agent:NewSQLiteAgentsDAO", "message", "failed to set busy timeout", "error", err)
	}

	// Enable WAL mode
	_, err = db.Exec("PRAGMA journal_mode = WAL;")
	if err != nil {
		slog.Error("dao_sqlite_agent:NewSQLiteAgentsDAO", "message", "failed to set WAL mode", "error", err)
	}

	return &SQLiteAgentsDAO{db: db}
}

func NewSQLiteAgentsDAOWithDB(db *sqlx.DB) *SQLiteAgentsDAO {
	return &SQLiteAgentsDAO{db: db}
}

// Agent CRUD
func (s *SQLiteAgentsDAO) CreateAgent(agent AgentRow) error {
	_, err := s.db.NamedExec(`
		INSERT INTO agents (id, name, description, system_prompt, provider, model, local_tools, mcp_servers, created_at, updated_at)
		VALUES (:id, :name, :description, :system_prompt, :provider, :model, :local_tools, :mcp_servers, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, agent)
	if err != nil {
		slog.Error("dao_sqlite:CreateAgent", "message", "failed to create agent", "error", err, "agentID", agent.ID)
		return fmt.Errorf("failed to create agent")
	}
	return nil
}

func (s *SQLiteAgentsDAO) UpdateAgent(agent AgentRow) error {
	_, err := s.db.NamedExec(`
		UPDATE agents 
		SET name = :name, 
			description = :description, 
			system_prompt = :system_prompt, 
			provider = :provider, 
			model = :model, 
			mcp_servers = :mcp_servers,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = :id
	`, agent)
	if err != nil {
		slog.Error("dao_sqlite:UpdateAgent", "message", "failed to update agent", "error", err, "agentID", agent.ID)
		return fmt.Errorf("failed to update agent")
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
	// Retry logic for transient database errors (e.g. locking)
	maxRetries := 3
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		_, err := s.db.NamedExec(`
			INSERT INTO agent_messages (
				id, session_id, sequence_number, role, type, 
				content, tool_name, tool_call_id, tool_args, thought_signature, 
				success, error_message, run_time_ms, created_at
			) VALUES (
				:id, :session_id, :sequence_number, :role, :type,
				:content, :tool_name, :tool_call_id, :tool_args, :thought_signature,
				:success, :error_message, :run_time_ms, CURRENT_TIMESTAMP
			)
		`, message)
		if err == nil {
			return nil
		}
		
		lastErr = err
		slog.Warn("dao_sqlite:AddAgentMessage", "message", "failed to add agent message, retrying...", "attempt", i+1, "error", err)
		
		if i < maxRetries-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	slog.Error("dao_sqlite:AddAgentMessage", "message", "failed to add agent message after retries", "error", lastErr, "messageID", message.ID)
	return fmt.Errorf("failed to add agent message: %w", lastErr)
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

// Agent File Operations
func (s *SQLiteAgentsDAO) SaveAgentFile(agentID, docsID, fileName, filePath string, fileSize int64, userID string) error {
	_, err := s.db.Exec(`
		INSERT INTO agent_docs (id, agent_id, docs_id, file_name, file_path, file_size, uploaded_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, docsID, agentID, docsID, fileName, filePath, fileSize, userID)
	if err != nil {
		slog.Error("dao_sqlite:SaveAgentFile", "message", "failed to save agent file", "error", err, "agentID", agentID, "docsID", docsID)
		return fmt.Errorf("failed to save agent file")
	}
	return nil
}

func (s *SQLiteAgentsDAO) GetAgentFiles(agentID string) ([]AgentDocumentRow, error) {
	var files []AgentDocumentRow
	err := s.db.Select(&files, `
		SELECT * FROM agent_docs 
		WHERE agent_id = ? 
		ORDER BY created_at DESC
	`, agentID)
	if err != nil {
		slog.Error("dao_sqlite:GetAgentFiles", "message", "failed to get agent files", "error", err, "agentID", agentID)
		return nil, fmt.Errorf("failed to get agent files")
	}
	return files, nil
}

func (s *SQLiteAgentsDAO) GetAgentFileByPath(agentID, filePath string) (*AgentDocumentRow, error) {
	var file AgentDocumentRow
	err := s.db.Get(&file, `
		SELECT * FROM agent_docs 
		WHERE agent_id = ? AND file_path = ?
	`, agentID, filePath)
	if err != nil {
		slog.Error("dao_sqlite:GetAgentFileByPath", "message", "failed to get agent file by path", "error", err, "agentID", agentID, "filePath", filePath)
		return nil, fmt.Errorf("failed to get agent file by path")
	}
	return &file, nil
}

func (s *SQLiteAgentsDAO) DeleteAgentFile(agentID, docsID string) error {
	_, err := s.db.Exec(`
		DELETE FROM agent_docs 
		WHERE agent_id = ? AND docs_id = ?
	`, agentID, docsID)
	if err != nil {
		slog.Error("dao_sqlite:DeleteAgentFile", "message", "failed to delete agent file", "error", err, "agentID", agentID, "docsID", docsID)
		return fmt.Errorf("failed to delete agent file")
	}
	return nil
}
