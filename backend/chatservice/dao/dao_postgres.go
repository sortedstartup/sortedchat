package dao

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	proto "sortedstartup/chatservice/proto"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// PostgresDAO implements the DAO interface using PostgreSQL and sqlx
type PostgresDAO struct {
	db *sqlx.DB
}

// NewPostgresDAO creates a new PostgreSQL DAO instance
func NewPostgresDAO(config *PostgresConfig) (*PostgresDAO, error) {
	slog.Info("dao_postgres:NewPostgresDAO")
	dsn := config.GetPostgresDSN()

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		slog.Error("dao_postgres:NewPostgresDAO", "message", "failed to open PostgreSQL connection", "error", err)
		return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.Pool.MaxOpenConnections)
	db.SetMaxIdleConns(config.Pool.MaxIdleConnections)
	db.SetConnMaxLifetime(config.Pool.ConnectionMaxLifetime)

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		slog.Error("dao_postgres:NewPostgresDAO", "message", "failed to ping PostgreSQL database", "error", err)
		return nil, fmt.Errorf("failed to ping PostgreSQL database: %w", err)
	}

	slog.Info("PostgreSQL DAO created successfully",
		"host", config.Host,
		"port", config.Port,
		"database", config.Database,
		"max_open_conns", config.Pool.MaxOpenConnections)

	return &PostgresDAO{db: db}, nil
}

// NewPostgresDAOWithDB creates a new PostgreSQL DAO instance using a shared database connection
func NewPostgresDAOWithDB(db *sqlx.DB) (*PostgresDAO, error) {
	return &PostgresDAO{db: db}, nil
}

// Close closes the database connection
// Note: When using shared connection pool, this method should not be called
// as the connection is managed by the factory
func (p *PostgresDAO) Close() error {
	// Do nothing - connection is managed by the factory
	return nil
}

// CreateChat creates a new chat with the given ID and name
func (p *PostgresDAO) CreateChat(userID string, chatId string, name string, projectID string) error {
	slog.Info("dao_postgres:CreateChat", "userID", userID, "chatId", chatId, "name", name, "projectID", projectID)
	if projectID == "" || projectID == "null" {
		_, err := p.db.Exec("INSERT INTO chat_list (chat_id, name, user_id) VALUES ($1, $2, $3)", chatId, name, userID)
		if err != nil {
			slog.Error("dao_postgres:CreateChat", "message", "failed to create chat", "error", err, "userID", userID, "chatId", chatId, "projectID", projectID)
			return fmt.Errorf("failed to create chat")
		}
	} else {
		_, err := p.db.Exec("INSERT INTO chat_list (chat_id, name, project_id, user_id) VALUES ($1, $2, $3, $4)", chatId, name, projectID, userID)
		if err != nil {
			slog.Error("dao_postgres:CreateChat", "message", "failed to create chat", "error", err, "userID", userID, "chatId", chatId, "projectID", projectID)
			return fmt.Errorf("failed to create chat")
		}
	}
	return nil
}

func (p *PostgresDAO) GetChatName(userID string, chatId string) (string, error) {
	slog.Info("dao_postgres:GetChatName", "userID", userID, "chatId", chatId)
	var name string
	err := p.db.Get(&name, "SELECT name FROM chat_list WHERE chat_id = $1 AND user_id = $2", chatId, userID)
	if err != nil {
		slog.Error("dao_postgres:GetChatName", "message", "failed to get chat name", "error", err, "userID", userID, "chatId", chatId)
		return "", fmt.Errorf("failed to get chat name")
	}
	return name, nil
}

func (p *PostgresDAO) SaveChatName(userID string, chatId string, name string) error {
	slog.Info("dao_postgres:SaveChatName", "userID", userID, "chatId", chatId, "name", name)
	_, err := p.db.Exec("UPDATE chat_list SET name = $1 WHERE chat_id = $2 AND user_id = $3", name, chatId, userID)
	if err != nil {
		slog.Error("dao_postgres:SaveChatName", "message", "failed to save chat name", "error", err, "userID", userID, "chatId", chatId)
		return fmt.Errorf("failed to save chat name")
	}
	return nil
}

// AddChatMessage adds a message to a chat
func (p *PostgresDAO) AddChatMessage(userID string, chatId string, role string, content string, model string, inputTokens int, outputTokens int, cachedTokens int, references string, ragEnabled bool) (string, error) {
	slog.Info("dao_postgres:AddChatMessage", "userID", userID, "chatId", chatId, "role", role, "model", model)
	var messageId string

	// Handle empty or whitespace references by setting it to empty JSON object
	trimmedRef := strings.TrimSpace(references)
	if trimmedRef == "" {
		trimmedRef = "[]"
	}

	// Validate references JSON
	var temp interface{}
	if err := json.Unmarshal([]byte(trimmedRef), &temp); err != nil {
		slog.Error("dao_postgres:AddChatMessage", "message", "invalid JSON format for references field", "error", err, "userID", userID, "chatId", chatId, "role", role, "model", model)
		return "", fmt.Errorf("invalid JSON format for references field")
	}

	// Insert into DB
	err := p.db.Get(&messageId,
		`INSERT INTO chat_messages
        (chat_id, role, content, user_id, rag_enabled, model, input_token_count, output_token_count, cached_token_count, document_references)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::jsonb)
        RETURNING id`,
		chatId, role, content, userID, ragEnabled, model, inputTokens, outputTokens, cachedTokens, trimmedRef)
	if err != nil {
		slog.Error("dao_postgres:AddChatMessage", "message", "failed to add chat message", "error", err, "userID", userID, "chatId", chatId, "role", role, "model", model)
		return "", fmt.Errorf("failed to add chat message")
	}

	return messageId, nil
}

// GetChatMessages retrieves all messages for a given chat - need to add cost
func (p *PostgresDAO) GetChatMessages(userID string, chatId string) ([]ChatMessageRow, error) {
	slog.Info("dao_postgres:GetChatMessages", "userID", userID, "chatId", chatId)
	var messages []ChatMessageRow
	err := p.db.Select(&messages, "SELECT role, content, id, COALESCE(document_references::text, '') as document_references, rag_enabled, COALESCE(model, '') as model, COALESCE(input_token_count, 0) as input_token_count, COALESCE(output_token_count, 0) as output_token_count, COALESCE(cost, 0) as cost, COALESCE(cached_token_count, 0) as cached_token_count FROM chat_messages WHERE chat_id = $1 AND user_id = $2 ORDER BY id", chatId, userID)
	if err != nil {
		slog.Error("dao_postgres:GetChatMessages", "message", "failed to get chat messages", "error", err, "userID", userID, "chatId", chatId)
		return nil, fmt.Errorf("failed to get chat messages")
	}
	return messages, nil
}

// GetChatList retrieves all chats for a user
func (p *PostgresDAO) GetChatList(userID string, projectID string, softDeleted bool) ([]*proto.ChatInfo, error) {
	slog.Info("dao_postgres:GetChatList", "userID", userID, "projectID", projectID, "softDeleted", softDeleted)
	var chats []ChatInfoRow

	query := "SELECT chat_id, name FROM chat_list WHERE soft_deleted = ? AND parent_chat_id IS NULL"
	args := []interface{}{softDeleted}

	if projectID == "" || projectID == "null" {
		query += " AND project_id IS NULL AND user_id = ?"
		args = append(args, userID)
	} else {
		query += " AND project_id = ? AND user_id = ?"
		args = append(args, projectID, userID)
	}

	query += " ORDER BY id DESC"

	query = p.db.Rebind(query)
	err := p.db.Select(&chats, query, args...)
	if err != nil {
		slog.Error("dao_postgres:GetChatList", "message", "failed to get chat list", "error", err, "userID", userID, "projectID", projectID, "softDeleted", softDeleted)
		return nil, fmt.Errorf("failed to get chat list")
	}

	var result []*proto.ChatInfo
	for _, c := range chats {
		result = append(result, &proto.ChatInfo{
			ChatId: c.Id,
			Name:   c.Name,
		})
	}
	return result, nil
}

func (p *PostgresDAO) AddChatMessageWithTokens(
	userID string,
	chatId string,
	role string,
	content string,
	model string,
	inputTokens int,
	outputTokens int,
	cachedTokens int,
	references string,
	ragEnabled bool,
) (MessageSummary, error) {
	slog.Info("dao_postgres:AddChatMessageWithTokens", "userID", userID, "chatId", chatId, "role", role, "model", model)

	var messageId string
	var cost float64
	var referencesValue interface{}
	if references == "" {
		referencesValue = nil
	} else {
		referencesValue = references
	}

	sqlQuery := `
		WITH ins AS (
			INSERT INTO chat_messages (
				chat_id, role, content, model,
				input_token_count, output_token_count, cached_token_count,
				user_id, document_references, rag_enabled
			)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
			RETURNING id
		),
		prices AS (
			SELECT input_token_cost, output_token_cost, cached_token_cost
			FROM model_metadata WHERE id = $4
		),
		calc AS (
			SELECT (p.input_token_cost * $5
				+ p.output_token_cost * $6
				+ p.cached_token_cost * $7) / 1000000.0 AS cost
			FROM prices p
		),
		upd_msg AS (
			UPDATE chat_messages m
			SET cost = c.cost
			FROM calc c
			WHERE m.id = (SELECT id FROM ins)
		)
		, upd_chat AS (
			UPDATE chat_list cl
			SET cost               = COALESCE(cl.cost,0) + c.cost,
				input_token_count  = COALESCE(cl.input_token_count,0) + $5,
				output_token_count = COALESCE(cl.output_token_count,0) + $6,
				cached_token_count = COALESCE(cl.cached_token_count,0) + $7
			FROM calc c
			WHERE cl.chat_id = $11 AND cl.user_id = $8
		)
		SELECT ins.id, calc.cost
		FROM ins, calc;
		`

	err := p.db.QueryRow(sqlQuery,
		chatId, role, content, model,
		inputTokens, outputTokens, cachedTokens,
		userID, referencesValue, ragEnabled,
		chatId, // $11
	).Scan(&messageId, &cost)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Error("dao_postgres:AddChatMessageWithTokens", "message", "no message inserted or updated, no result returned", "error", err, "userID", userID, "chatId", chatId, "role", role, "model", model)
			return MessageSummary{}, fmt.Errorf("no message inserted or updated, no result returned")
		}
		slog.Error("dao_postgres:AddChatMessageWithTokens", "message", "failed to add chat message with tokens", "error", err, "userID", userID, "chatId", chatId, "role", role, "model", model)
		return MessageSummary{}, fmt.Errorf("failed to add chat message with tokens")
	}

	return MessageSummary{
		MessageId:        messageId,
		Model:            model,
		InputTokenCount:  inputTokens,
		OutputTokenCount: outputTokens,
		CachedTokenCount: cachedTokens,
		Cost:             cost,
	}, nil
}

// GetModels retrieves all available models
func (p *PostgresDAO) GetModels() ([]*proto.ModelListInfo, error) {
	slog.Info("dao_postgres:GetModels")
	var models []Models
	err := p.db.Select(&models, "SELECT id, name,provider,url,input_token_cost,output_token_cost,COALESCE(capabilities, '{}'::jsonb)::text AS capabilities FROM model_metadata")
	if err != nil {
		slog.Error("dao_postgres:GetModels", "message", "failed to get models", "error", err)
		return nil, fmt.Errorf("failed to get models")
	}

	var result []*proto.ModelListInfo
	for _, m := range models {
		// Parse capabilities JSON
		capabilities, err := parseCapabilities(m.Capabilities)
		if err != nil {
			slog.Error("dao_postgres:GetModels", "message", "failed to parse capabilities for model", "error", err, "modelID", m.ID)
			return nil, fmt.Errorf("failed to parse capabilities for model")
		}

		result = append(result, &proto.ModelListInfo{
			Id:              m.ID,
			Label:           m.Name,
			Provider:        m.Provider,
			Url:             m.URL,
			InputTokenCost:  m.InputTokenCost,
			OutputTokenCost: m.OutputTokenCost,
			Capabilities:    capabilities,
		})
	}
	return result, nil
}

// SearchChatMessages performs full text search across chat messages
func (p *PostgresDAO) SearchChatMessages(userID string, query string) ([]proto.SearchResult, error) {
	slog.Info("dao_postgres:SearchChatMessages", "userID", userID, "query", query)
	// Input validation and sanitization
	if userID == "" || query == "" {
		slog.Error("dao_postgres:SearchChatMessages", "message", "userID and query are required", "userID", userID, "query", query)
		return nil, errors.New("userID and query are required")
	}

	// Sanitize query to prevent injection
	sanitizedQuery := sanitizeFTSQuery(query)
	if sanitizedQuery == "" {
		slog.Error("dao_postgres:SearchChatMessages", "message", "query contains no searchable terms", "userID", userID, "query", query)
		return nil, fmt.Errorf("query contains no searchable terms")
	}

	sqlQuery := `
		SELECT
			cm.chat_id,
			cl.name AS chat_name,
			string_agg(
				CASE
					WHEN length(cm.content) > 150
					THEN left(cm.content, 150) || '...'
					ELSE cm.content
				END,
				E'\n-----\n'
				ORDER BY cm.created_at
			) AS matched_text
		FROM chat_messages cm
		JOIN chat_list cl ON cm.chat_id = cl.chat_id
		WHERE cm.user_id = $2
		AND cl.user_id = $2
		AND cm.content_tsvector @@ to_tsquery('english', $1)
		GROUP BY cm.chat_id, cl.name
		ORDER BY max(ts_rank_cd(cm.content_tsvector, to_tsquery('english', $1))) DESC, cm.chat_id
		LIMIT 20`

	rows, err := p.db.Query(sqlQuery, sanitizedQuery, userID)
	if err != nil {
		slog.Error("dao_postgres:SearchChatMessages", "message", "failed to execute FTS query", "error", err, "userID", userID, "query", query)
		return nil, fmt.Errorf("failed to execute FTS query")
	}
	defer rows.Close()

	var results []proto.SearchResult
	for rows.Next() {

		var chatId, chatName, matchedText string

		err := rows.Scan(&chatId, &chatName, &matchedText)
		if err != nil {
			slog.Error("dao_postgres:SearchChatMessages", "message", "failed to scan search result", "error", err, "userID", userID, "query", query)
			return nil, fmt.Errorf("failed to scan search result")
		}

		results = append(results, proto.SearchResult{
			ChatId:      chatId,
			ChatName:    chatName,
			MatchedText: matchedText,
		})
	}

	return results, nil
}

// Project CRUD
func (p *PostgresDAO) CreateProject(userID string, id string, name string, description string, additionalData string) (string, error) {
	slog.Info("dao_postgres:CreateProject", "userID", userID, "id", id)
	_, err := p.db.Exec(`
		INSERT INTO project (id, name, description, additional_data, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, id, name, description, additionalData, userID)
	if err != nil {
		slog.Error("dao_postgres:CreateProject", "message", "failed to create project", "error", err, "userID", userID, "id", id)
		return "", fmt.Errorf("failed to create project")
	}
	return id, nil
}

// GetProjects retrieves all projects for a user
func (p *PostgresDAO) GetProjects(userID string) ([]ProjectRow, error) {
	slog.Info("dao_postgres:GetProjects", "userID", userID)
	var projects []ProjectRow
	err := p.db.Select(&projects, `SELECT id, name, description, additional_data, created_at, updated_at FROM project WHERE user_id = $1`, userID)
	if err != nil {
		slog.Error("dao_postgres:GetProjects", "message", "failed to get projects", "error", err, "userID", userID)
		return nil, fmt.Errorf("failed to get projects")
	}
	return projects, nil
}

func (p *PostgresDAO) FileSave(userID string, project_id string, docs_id string, file_name string, file_size int64) error {
	slog.Info("dao_postgres:FileSave", "userID", userID, "project_id", project_id, "docs_id", docs_id, "file_name", file_name, "file_size", file_size)
	size_kb := file_size / 1024
	_, err := p.db.Exec("INSERT INTO project_docs (project_id, docs_id, file_name, file_size, embedding_status, user_id) VALUES ($1, $2, $3, $4, $5, $6)",
		project_id, docs_id, file_name, size_kb, int32(proto.Embedding_Status_STATUS_QUEUED), userID)
	if err != nil {
		slog.Error("dao_postgres:FileSave", "message", "failed to save file", "error", err, "userID", userID, "project_id", project_id, "docs_id", docs_id, "file_name", file_name, "file_size", file_size)
		return fmt.Errorf("failed to save file")
	}
	return nil
}

func (p *PostgresDAO) UpdateEmbeddingStatus(docs_id string, status int32) error {
	slog.Info("dao_postgres:UpdateEmbeddingStatus", "docs_id", docs_id, "status", status)
	_, err := p.db.Exec("UPDATE project_docs SET embedding_status = $1 WHERE docs_id = $2", status, docs_id)
	if err != nil {
		slog.Error("dao_postgres:UpdateEmbeddingStatus", "message", "failed to update embedding status", "error", err, "docs_id", docs_id, "status", status)
		return fmt.Errorf("failed to update embedding status")
	}
	return nil
}

func (p *PostgresDAO) FetchErrorDocs(userID string, project_id string) ([]string, error) {
	slog.Info("dao_postgres:FetchErrorDocs", "userID", userID, "project_id", project_id)
	var docs_list []string
	err := p.db.Select(&docs_list, "SELECT docs_id FROM project_docs WHERE project_id = $1 AND embedding_status = $2 AND user_id = $3",
		project_id, int32(proto.Embedding_Status_STATUS_ERROR), userID)
	if err != nil {
		slog.Error("dao_postgres:FetchErrorDocs", "message", "failed to fetch error docs", "error", err, "userID", userID, "project_id", project_id)
		return nil, fmt.Errorf("failed to fetch error docs")
	}
	return docs_list, nil
}

func (p *PostgresDAO) TotalUsedSize(userID string, projectID string) (int64, error) {
	slog.Info("dao_postgres:TotalUsedSize", "userID", userID, "projectID", projectID)
	var total int64
	err := p.db.Get(&total, `
		SELECT COALESCE(SUM(file_size), 0)
		FROM project_docs
		WHERE project_id = $1 AND user_id = $2
	`, projectID, userID)
	if err != nil {
		slog.Error("dao_postgres:TotalUsedSize", "message", "failed to get total used size", "error", err, "userID", userID, "projectID", projectID)
		return 0, fmt.Errorf("failed to get total used size")
	}
	return total, err
}

func (p *PostgresDAO) FilesList(userID string, project_id string) ([]DocumentListRow, error) {
	slog.Info("dao_postgres:FilesList", "userID", userID, "project_id", project_id)
	var files []DocumentListRow
	err := p.db.Select(&files, `
		SELECT id, project_id, docs_id, file_name, created_at, updated_at, embedding_status
		FROM project_docs
		WHERE project_id = $1 AND user_id = $2
	`, project_id, userID)
	if err != nil {
		slog.Error("dao_postgres:FilesList", "message", "failed to get files list", "error", err, "userID", userID, "project_id", project_id)
		return nil, fmt.Errorf("failed to get files list")
	}
	return files, nil
}

func (p *PostgresDAO) GetFileMetadata(docsId string) (*DocumentListRow, error) {
	slog.Info("dao_postgres:GetFileMetadata", "docsId", docsId)
	var doc DocumentListRow
	err := p.db.Get(&doc, `SELECT * FROM project_docs WHERE docs_id = $1`, docsId)
	if err != nil {
		slog.Error("dao_postgres:GetFileMetadata", "message", "failed to get file metadata", "error", err, "docsId", docsId)
		return nil, fmt.Errorf("failed to get file metadata")
	}
	return &doc, nil
}

// SaveRAGChunk saves a chunk to rag_chunks table
func (p *PostgresDAO) SaveRAGChunk(userID string, chunkID, projectID, docsID string, startByte, endByte int) error {
	slog.Info("dao_postgres:SaveRAGChunk", "userID", userID, "chunkID", chunkID, "projectID", projectID, "docsID", docsID, "startByte", startByte, "endByte", endByte)
	_, err := p.db.Exec(`
		INSERT INTO rag_chunks (id, project_id, docs_id, start_byte, end_byte, user_id)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, chunkID, projectID, docsID, startByte, endByte, userID)
	if err != nil {
		slog.Error("dao_postgres:SaveRAGChunk", "message", "failed to save rag chunk", "error", err, "userID", userID, "chunkID", chunkID, "projectID", projectID, "docsID", docsID, "startByte", startByte, "endByte", endByte)
		return fmt.Errorf("failed to save rag chunk")
	}
	return nil
}

// SaveRAGChunkEmbedding stores vector embedding for a RAG chunk
func (p *PostgresDAO) SaveRAGChunkEmbedding(chunkID string, embedding []float64) error {
	slog.Info("dao_postgres:SaveRAGChunkEmbedding", "chunkID", chunkID)
	// Input validation
	if chunkID == "" {
		slog.Error("dao_postgres:SaveRAGChunkEmbedding", "message", "chunkID cannot be empty", "chunkID", chunkID)
		return fmt.Errorf("chunkID cannot be empty")
	}
	if len(embedding) == 0 {
		slog.Error("dao_postgres:SaveRAGChunkEmbedding", "message", "embedding cannot be empty", "chunkID", chunkID)
		return fmt.Errorf("embedding cannot be empty")
	}
	if len(embedding) != 768 { // Validate expected dimension (768 as per CHOICE 2)
		slog.Error("dao_postgres:SaveRAGChunkEmbedding", "message", "embedding dimension mismatch", "chunkID", chunkID, "embedding", embedding)
		return fmt.Errorf("embedding dimension mismatch: expected 768, got %d", len(embedding))
	}

	// Convert to pgvector format
	embeddingStr := vectorToString(embedding)

	// Use prepared statement to prevent SQL injection
	query := `
		UPDATE rag_chunks 
		SET embedding = $1, 
		    embedding_created_at = CURRENT_TIMESTAMP 
		WHERE id = $2`

	_, err := p.db.Exec(query, embeddingStr, chunkID)
	if err != nil {
		slog.Error("dao_postgres:SaveRAGChunkEmbedding", "message", "failed to save embedding for chunk", "error", err, "chunkID", chunkID)
		return fmt.Errorf("failed to save embedding for chunk")
	}

	return nil
}

// GetTopSimilarRAGChunks retrieves most similar chunks using cosine similarity
func (p *PostgresDAO) GetTopSimilarRAGChunks(userID string, queryEmbedding string, projectID string) ([]RAGChunkRow, error) {
	slog.Info("dao_postgres:GetTopSimilarRAGChunks", "userID", userID, "queryEmbedding", queryEmbedding, "projectID", projectID)
	// Input validation
	if userID == "" || queryEmbedding == "" || projectID == "" {
		slog.Error("dao_postgres:GetTopSimilarRAGChunks", "message", "userID, queryEmbedding, and projectID are required", "userID", userID, "queryEmbedding", queryEmbedding, "projectID", projectID)
		return nil, fmt.Errorf("userID, queryEmbedding, and projectID are required")
	}

	// Validate embedding format
	if !isValidEmbeddingFormat(queryEmbedding) {
		slog.Error("dao_postgres:GetTopSimilarRAGChunks", "message", "invalid embedding format", "userID", userID, "queryEmbedding", queryEmbedding, "projectID", projectID)
		return nil, fmt.Errorf("invalid embedding format")
	}

	query := `
		SELECT id, project_id, docs_id, start_byte, end_byte,1-(embedding <=> $1) AS similarity
    FROM rag_chunks 
    WHERE user_id = $2 
      AND project_id = $3
      AND embedding IS NOT NULL
    ORDER BY embedding <=> $1  -- Cosine distance (smaller = more similar)
    LIMIT 2`

	var chunks []RAGChunkRow
	rows, err := p.db.Query(query, queryEmbedding, userID, projectID)
	if err != nil {
		slog.Error("dao_postgres:GetTopSimilarRAGChunks", "message", "failed to query similar chunks", "error", err, "userID", userID, "queryEmbedding", queryEmbedding, "projectID", projectID)
		return nil, fmt.Errorf("failed to query similar chunks")
	}
	defer rows.Close()

	for rows.Next() {
		var chunk RAGChunkRow

		err := rows.Scan(&chunk.ID, &chunk.ProjectID, &chunk.DocsID,
			&chunk.StartByte, &chunk.EndByte, &chunk.Similarity)
		if err != nil {
			slog.Error("dao_postgres:GetTopSimilarRAGChunks", "message", "failed to scan chunk row", "error", err, "userID", userID, "queryEmbedding", queryEmbedding, "projectID", projectID)
			return nil, fmt.Errorf("failed to scan chunk row")
		}
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

func (p *PostgresDAO) IsMainBranch(userID string, source_chat_id string) (bool, error) {
	slog.Info("dao_postgres:IsMainBranch", "userID", userID, "source_chat_id", source_chat_id)
	var isMainBranch bool
	err := p.db.Get(&isMainBranch, `SELECT is_main_branch FROM chat_list WHERE chat_id = $1 AND user_id = $2`, source_chat_id, userID)
	if err != nil {
		slog.Error("dao_postgres:IsMainBranch", "message", "failed to get is main branch", "error", err, "userID", userID, "source_chat_id", source_chat_id)
		return false, fmt.Errorf("failed to get is main branch")
	}
	return isMainBranch, err
}

func (p *PostgresDAO) BranchChat(userID string, source_chat_id string, parent_message_id string, new_chat_id string, branch_name string) error {
	slog.Info("dao_postgres:BranchChat", "userID", userID, "source_chat_id", source_chat_id, "parent_message_id", parent_message_id, "new_chat_id", new_chat_id, "branch_name", branch_name)
	tx, err := p.db.Beginx()
	if err != nil {
		slog.Error("dao_postgres:BranchChat", "message", "failed to begin transaction", "error", err, "userID", userID, "source_chat_id", source_chat_id, "parent_message_id", parent_message_id, "new_chat_id", new_chat_id, "branch_name", branch_name)
		return fmt.Errorf("failed to begin transaction")
	}
	defer tx.Rollback()

	// Use CTE to find project_id from source chat and insert the new branch chat
	_, err = tx.Exec(`WITH source_chat AS (
						SELECT project_id 
						FROM chat_list 
						WHERE chat_id = $1 AND user_id = $2
					)
					INSERT INTO chat_list (chat_id, name, project_id, parent_chat_id, parent_message_id, is_main_branch, user_id)
					SELECT $3, $4, COALESCE(source_chat.project_id, NULL), $1, $5, FALSE, $2
					FROM source_chat`, source_chat_id, userID, new_chat_id, branch_name, parent_message_id)
	if err != nil {
		slog.Error("dao_postgres:BranchChat", "message", "failed to insert branch chat", "error", err, "userID", userID, "source_chat_id", source_chat_id, "parent_message_id", parent_message_id, "new_chat_id", new_chat_id, "branch_name", branch_name)
		return fmt.Errorf("failed to insert branch chat")
	}

	// Copy messages up to branch point
	_, err = tx.Exec(`INSERT INTO chat_messages (chat_id, role, content, model, error, input_token_count, output_token_count, created_at, user_id)
					  SELECT $1, role, content, model, error, input_token_count, output_token_count, created_at, $2
					  FROM chat_messages 
					  WHERE chat_id = $3 AND id <= $4 AND user_id = $5
					  ORDER BY id`, new_chat_id, userID, source_chat_id, parent_message_id, userID)
	if err != nil {
		slog.Error("dao_postgres:BranchChat", "message", "failed to insert messages", "error", err, "userID", userID, "source_chat_id", source_chat_id, "parent_message_id", parent_message_id, "new_chat_id", new_chat_id, "branch_name", branch_name)
		return fmt.Errorf("failed to insert messages")
	}

	return tx.Commit()
}

func (p *PostgresDAO) GetChatBranches(userID string, chatId string, isMain bool) ([]ChatInfoRow, error) {
	slog.Info("dao_postgres:GetChatBranches", "userID", userID, "chatId", chatId, "isMain", isMain)
	var chats []ChatInfoRow
	var err error

	if isMain {
		err = p.db.Select(&chats, `SELECT chat_id, name FROM chat_list WHERE parent_chat_id = $1`, chatId)
	} else {
		err = p.db.Select(&chats, `
			SELECT c1.chat_id, c1.name 
			FROM chat_list c1
			JOIN chat_list c2 ON c1.chat_id = c2.parent_chat_id
			WHERE c2.chat_id = $1
		`, chatId)
	}

	if err != nil {
		slog.Error("dao_postgres:GetChatBranches", "message", "failed to get chat branches", "error", err, "userID", userID, "chatId", chatId, "isMain", isMain)
		return nil, fmt.Errorf("failed to get chat branches")
	}

	return chats, nil
}

// Helper function to convert float64 slice to pgvector string format
func vectorToString(embedding []float64) string {
	slog.Info("dao_postgres:vectorToString", "embedding", embedding)
	strValues := make([]string, len(embedding))
	for i, v := range embedding {
		strValues[i] = strconv.FormatFloat(v, 'f', -1, 64)
	}
	return "[" + strings.Join(strValues, ",") + "]"
}

// Helper function to validate embedding format
func isValidEmbeddingFormat(embedding string) bool {
	slog.Info("dao_postgres:isValidEmbeddingFormat", "embedding", embedding)
	// Basic validation: should start with [ and end with ]
	if !strings.HasPrefix(embedding, "[") || !strings.HasSuffix(embedding, "]") {
		return false
	}

	// Trim brackets and check content
	content := embedding[1 : len(embedding)-1]
	if content == "" {
		return true // An empty vector `[]` is valid.
	}

	parts := strings.Split(content, ",")
	for _, part := range parts {
		trimmedPart := strings.TrimSpace(part)
		if trimmedPart == "" {
			return false // Disallow empty elements like in "[1,,2]"
		}
		if _, err := strconv.ParseFloat(trimmedPart, 64); err != nil {
			return false // Each part must be a valid float
		}
	}

	return true
}

// Pre-compiled regexes for better performance
var (
	dangerousCharsRegex = regexp.MustCompile(`[;&|<>(){}[\]\\'"*?]`)
	validWordRegex      = regexp.MustCompile(`^[a-zA-Z0-9]{2,}$`)
)

// sanitizeFTSQuery cleans and prepares query for PostgreSQL FTS
func sanitizeFTSQuery(query string) string {
	slog.Info("dao_postgres:sanitizeFTSQuery", "query", query)
	// Remove potentially dangerous characters
	cleaned := dangerousCharsRegex.ReplaceAllString(query, " ")

	// Limit query length
	if len(cleaned) > 500 {
		cleaned = cleaned[:500]
	}

	// Split and rejoin to create safe tsquery
	words := strings.Fields(cleaned)
	validWords := make([]string, 0, len(words))

	for _, word := range words {
		// Only include words with letters/numbers, min 2 chars
		if validWordRegex.MatchString(word) {
			validWords = append(validWords, word)
		}
	}

	if len(validWords) == 0 {
		return ""
	}

	return strings.Join(validWords, " & ")
}

func (p *PostgresDAO) DeleteDocument(userID string, projectID string, docID string) error {
	slog.Info("dao_postgres:DeleteDocument", "userID", userID, "projectID", projectID, "docID", docID)
	// Start a transaction to ensure all operations succeed or fail together
	tx, err := p.db.Beginx()
	if err != nil {
		slog.Error("dao_postgres:DeleteDocument", "message", "failed to begin transaction", "error", err, "userID", userID, "projectID", projectID, "docID", docID)
		return fmt.Errorf("failed to begin transaction")
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				slog.Error("dao_postgres:DeleteDocument", "message", "transaction rollback failed", "original_error", err, "rollback_error", rbErr)
			}
		}
	}()

	// Delete from rag_chunks first (document chunks)
	_, err = tx.Exec("DELETE FROM rag_chunks WHERE project_id = $1 AND docs_id = $2 AND user_id = $3", projectID, docID, userID)
	if err != nil {
		slog.Error("dao_postgres:DeleteDocument", "message", "failed to delete from rag_chunks", "error", err, "userID", userID, "projectID", projectID, "docID", docID)
		return fmt.Errorf("failed to delete from rag_chunks")
	}

	// Delete from project_docs (document metadata)
	_, err = tx.Exec("DELETE FROM project_docs WHERE project_id = $1 AND docs_id = $2 AND user_id = $3", projectID, docID, userID)
	if err != nil {
		slog.Error("dao_postgres:DeleteDocument", "message", "failed to delete from project_docs", "error", err, "userID", userID, "projectID", projectID, "docID", docID)
		return fmt.Errorf("failed to delete from project_docs")
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		slog.Error("dao_postgres:DeleteDocument", "message", "failed to commit transaction", "error", err, "userID", userID, "projectID", projectID, "docID", docID)
		return fmt.Errorf("failed to commit transaction")
	}

	return nil
}

func (p *PostgresDAO) SoftDeleteChat(userID string, chatId string) error {
	slog.Info("dao_postgres:SoftDeleteChat", "userID", userID, "chatId", chatId)
	_, err := p.db.Exec(`
        WITH RECURSIVE chat_hierarchy AS (
            SELECT chat_id
            FROM chat_list
            WHERE chat_id = $1 AND user_id = $2
            UNION ALL
            SELECT c.chat_id
            FROM chat_list c
            JOIN chat_hierarchy h ON c.parent_chat_id = h.chat_id
			WHERE c.user_id = $2
        )
        UPDATE chat_list
        SET soft_deleted = TRUE
        WHERE chat_id IN (SELECT chat_id FROM chat_hierarchy)
        AND user_id = $2;
    `, chatId, userID)
	if err != nil {
		slog.Error("dao_postgres:SoftDeleteChat", "message", "failed to soft delete chat", "error", err, "userID", userID, "chatId", chatId)
		return fmt.Errorf("failed to soft delete chat")
	}
	return err
}

func (p *PostgresDAO) DeleteChat(userID string, chatId string) (err error) {
	slog.Info("dao_postgres:DeleteChat", "userID", userID, "chatId", chatId)
	tx, err := p.db.Beginx()
	if err != nil {
		slog.Error("dao_postgres:DeleteChat", "message", "failed to begin transaction", "error", err, "userID", userID, "chatId", chatId)
		return fmt.Errorf("failed to begin transaction")
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				slog.Error("dao_postgres:DeleteChat", "message", "transaction rollback failed", "original_error", err, "rollback_error", rbErr)
			}
			slog.Error("dao_postgres:DeleteChat", "message", "transaction rollback failed", "original_error", err)
		}
	}()

	// Delete messages under the hierarchy first
	_, err = tx.Exec(`
        WITH RECURSIVE chat_hierarchy AS (
            SELECT chat_id
            FROM chat_list
            WHERE chat_id = $1 AND user_id = $2
            UNION ALL
            SELECT c.chat_id
            FROM chat_list c
            JOIN chat_hierarchy h ON c.parent_chat_id = h.chat_id
            WHERE c.user_id = $2
        )
        DELETE FROM chat_messages
        WHERE user_id = $2
          AND chat_id IN (SELECT chat_id FROM chat_hierarchy);
    `, chatId, userID)
	if err != nil {
		slog.Error("dao_postgres:DeleteChat", "message", "failed to delete chat messages", "error", err, "userID", userID, "chatId", chatId)
		return fmt.Errorf("failed to delete chat messages")
	}

	// Delete chats from the hierarchy
	_, err = tx.Exec(`
        WITH RECURSIVE chat_hierarchy AS (
            SELECT chat_id
            FROM chat_list
            WHERE chat_id = $1 AND user_id = $2
            UNION ALL
            SELECT c.chat_id
            FROM chat_list c
            JOIN chat_hierarchy h ON c.parent_chat_id = h.chat_id
            WHERE c.user_id = $2
        )
        DELETE FROM chat_list
        WHERE user_id = $2
          AND chat_id IN (SELECT chat_id FROM chat_hierarchy);
    `, chatId, userID)
	if err != nil {
		slog.Error("dao_postgres:DeleteChat", "message", "failed to delete chat list", "error", err, "userID", userID, "chatId", chatId)
		return fmt.Errorf("failed to delete chat list")
	}

	err = tx.Commit()
	if err != nil {
		slog.Error("dao_postgres:DeleteChat", "message", "failed to commit transaction", "error", err, "userID", userID, "chatId", chatId)
		return fmt.Errorf("failed to commit transaction")
	}
	return nil
}

func (p *PostgresDAO) RestoreChat(userID string, chatId string) error {
	slog.Info("dao_postgres:RestoreChat", "userID", userID, "chatId", chatId)
	_, err := p.db.Exec(`
        WITH RECURSIVE chat_hierarchy AS (
            SELECT chat_id
            FROM chat_list
            WHERE chat_id = $1 AND user_id = $2
            UNION ALL
            SELECT c.chat_id
            FROM chat_list c
            JOIN chat_hierarchy h ON c.parent_chat_id = h.chat_id
			WHERE c.user_id = $2
        )
        UPDATE chat_list
        SET soft_deleted = FALSE
        WHERE chat_id IN (SELECT chat_id FROM chat_hierarchy)
        AND user_id = $2;
    `, chatId, userID)
	if err != nil {
		slog.Error("dao_postgres:RestoreChat", "message", "failed to restore chat", "error", err, "userID", userID, "chatId", chatId)
		return fmt.Errorf("failed to restore chat")
	}
	return err
}

func (p *PostgresDAO) IsChatDeleted(chatId string, userID string) (bool, error) {
	slog.Info("dao_postgres:IsChatDeleted", "userID", userID, "chatId", chatId)
	var isDeleted bool
	err := p.db.Get(&isDeleted, "SELECT soft_deleted FROM chat_list WHERE chat_id = $1 AND user_id = $2", chatId, userID)
	if err != nil {
		slog.Error("dao_postgres:IsChatDeleted", "message", "failed to get chat deleted", "error", err, "userID", userID, "chatId", chatId)
		return false, fmt.Errorf("failed to get chat deleted")
	}
	return isDeleted, nil
}

func (p *PostgresDAO) GetChatMetadata(userID string, chatId string) (ChatInfoRow, error) {
	slog.Info("dao_postgres:GetChatMetadata", "userID", userID, "chatId", chatId)
	var chat ChatInfoRow
	err := p.db.Get(&chat, "SELECT chat_id, name, COALESCE(cost,0) AS cost, COALESCE(input_token_count,0) AS input_token_count, COALESCE(output_token_count,0) AS output_token_count, COALESCE(cached_token_count,0) AS cached_token_count FROM chat_list WHERE chat_id = $1 AND user_id = $2", chatId, userID)
	if err != nil {
		slog.Error("dao_postgres:GetChatMetadata", "message", "failed to get chat metadata", "error", err, "userID", userID, "chatId", chatId)
		return ChatInfoRow{}, fmt.Errorf("failed to get chat metadata")
	}
	return chat, nil
}

func (p *PostgresDAO) RenameChat(userID string, chatId string, name string) error {
	slog.Info("dao_postgres:RenameChat", "userID", userID, "chatId", chatId, "name", name)
	result, err := p.db.Exec("UPDATE chat_list SET name = $1 WHERE chat_id = $2 AND user_id = $3", name, chatId, userID)
	if err != nil {
		slog.Error("dao_postgres:RenameChat", "message", "failed to rename chat", "error", err, "userID", userID, "chatId", chatId, "name", name)
		return fmt.Errorf("failed to rename chat")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("dao_postgres:RenameChat", "message", "failed to get rows affected", "error", err, "userID", userID, "chatId", chatId, "name", name)
		return fmt.Errorf("failed to get rows affected")
	}
	if rowsAffected == 0 {
		slog.Error("dao_postgres:RenameChat", "message", "chat not found or permission denied", "userID", userID, "chatId", chatId, "name", name)
		return fmt.Errorf("chat not found or permission denied")
	}
	return nil
}

func (p *PostgresDAO) UpsertModel(modelID string, name string, url string, provider string, inputTokenCost float64, outputTokenCost float64, cachedTokenCost float64) error {
	slog.Info("dao_postgres:UpsertModel", "modelID", modelID, "name", name, "url", url, "provider", provider, "inputTokenCost", inputTokenCost, "outputTokenCost", outputTokenCost, "cachedTokenCost", cachedTokenCost)
	_, err := p.db.Exec(`
		INSERT INTO model_metadata (id, name, url, provider, input_token_cost, output_token_cost, cached_token_cost)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			url = EXCLUDED.url,
			provider = EXCLUDED.provider,
			input_token_cost = EXCLUDED.input_token_cost,
			output_token_cost = EXCLUDED.output_token_cost,
			cached_token_cost = EXCLUDED.cached_token_cost
	`, modelID, name, url, provider, inputTokenCost, outputTokenCost, cachedTokenCost)
	if err != nil {
		slog.Error("dao_postgres:UpsertModel", "message", "failed to upsert model", "error", err, "modelID", modelID, "name", name, "url", url, "provider", provider, "inputTokenCost", inputTokenCost, "outputTokenCost", outputTokenCost, "cachedTokenCost", cachedTokenCost)
		return fmt.Errorf("failed to upsert model")
	}
	return nil
}

// PostgresSettingsDAO implements the SettingsDAO interface using PostgreSQL
type PostgresSettingsDAO struct {
	db *sqlx.DB
}

func NewPostgresSettingsDAO(config *PostgresConfig) (*PostgresSettingsDAO, error) {
	slog.Info("dao_postgres:NewPostgresSettingsDAO", "config", config)
	dsn := config.GetPostgresDSN()

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		slog.Error("dao_postgres:NewPostgresSettingsDAO", "message", "failed to open PostgreSQL connection for settings", "error", err, "config", config)
		return nil, fmt.Errorf("failed to open PostgreSQL connection for settings ")
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.Pool.MaxOpenConnections)
	db.SetMaxIdleConns(config.Pool.MaxIdleConnections)
	db.SetConnMaxLifetime(config.Pool.ConnectionMaxLifetime)

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		slog.Error("dao_postgres:NewPostgresSettingsDAO", "message", "failed to ping PostgreSQL database for settings", "error", err, "config", config)
		return nil, fmt.Errorf("failed to ping PostgreSQL database for settings")
	}

	return &PostgresSettingsDAO{db: db}, nil
}

// NewPostgresSettingsDAOWithDB creates a new PostgreSQL Settings DAO instance using a shared database connection
func NewPostgresSettingsDAOWithDB(db *sqlx.DB) (*PostgresSettingsDAO, error) {
	return &PostgresSettingsDAO{db: db}, nil
}

func (p *PostgresSettingsDAO) Close() error {
	// Do nothing - connection is managed by the factory
	return nil
}

func (p *PostgresSettingsDAO) GetSettingValue(settingName string) (string, error) {
	slog.Info("dao_postgres:GetSettingValue", "settingName", settingName)
	var dbSetting dbSettings
	err := p.db.Get(&dbSetting, "SELECT name, settings FROM settings WHERE name = $1", settingName)
	if err != nil {
		// Preserve sql.ErrNoRows so callers can distinguish between no rows and actual database errors
		if err == sql.ErrNoRows {
			slog.Error("dao_postgres:GetSettingValue", "message", "setting not found", "error", err, "settingName", settingName)
			return "", err
		}
		slog.Error("dao_postgres:GetSettingValue", "message", "failed to get setting value", "error", err, "settingName", settingName)
		return "", fmt.Errorf("failed to get setting '%s' from database: %w", settingName, err)
	}

	return dbSetting.Settings, nil
}

func (p *PostgresSettingsDAO) SetSettingValue(settingName string, settingValue string) error {
	slog.Info("dao_postgres:SetSettingValue", "settingName", settingName, "settingValue", settingValue)
	query := `
        INSERT INTO settings (name, settings) VALUES ($1, $2)
        ON CONFLICT(name) DO UPDATE SET settings = EXCLUDED.settings, updated_at = CURRENT_TIMESTAMP
    `

	_, err := p.db.Exec(query, settingName, settingValue)
	if err != nil {
		slog.Error("dao_postgres:SetSettingValue", "message", "failed to upsert settings", "error", err, "settingName", settingName)
		return fmt.Errorf("failed to upsert settings")
	}
	return nil
}

// GetChatMessageByID retrieves a specific chat message by its ID
func (p *PostgresDAO) GetChatMessageByID(userID string, messageID string) (*ChatMessageRow, error) {
	slog.Info("dao_postgres:GetChatMessageByID", "userID", userID, "messageID", messageID)
	var message ChatMessageRow
	err := p.db.Get(&message, `
		SELECT role, content, id, COALESCE(document_references::text, '') as document_references 
		FROM chat_messages 
		WHERE id = $1 AND user_id = $2`, messageID, userID)
	if err != nil {
		slog.Error("dao_postgres:GetChatMessageByID", "message", "failed to get chat message", "error", err, "userID", userID, "messageID", messageID)
		return nil, fmt.Errorf("failed to get chat message")
	}
	return &message, nil
}

// UpdateChatMessageDocumentReferences updates the document references for a specific message
func (p *PostgresDAO) UpdateChatMessageDocumentReferences(userID string, messageID string, documentReferences string) error {
	slog.Info("dao_postgres:UpdateChatMessageDocumentReferences", "userID", userID, "messageID", messageID, "documentReferences", documentReferences)
	_, err := p.db.Exec(`
		UPDATE chat_messages 
		SET document_references = $1 
		WHERE id = $2 AND user_id = $3`, documentReferences, messageID, userID)
	if err != nil {
		slog.Error("dao_postgres:UpdateChatMessageDocumentReferences", "message", "failed to update chat message document references", "error", err, "userID", userID, "messageID", messageID, "documentReferences", documentReferences)
		return fmt.Errorf("failed to update chat message document references")
	}
	return nil
}
