package dao

import (
	"database/sql"
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
	dsn := config.GetPostgresDSN()

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.Pool.MaxOpenConnections)
	db.SetMaxIdleConns(config.Pool.MaxIdleConnections)
	db.SetConnMaxLifetime(config.Pool.ConnectionMaxLifetime)

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
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
	if projectID == "" || projectID == "null" {
		_, err := p.db.Exec("INSERT INTO chat_list (chat_id, name, user_id) VALUES ($1, $2, $3)", chatId, name, userID)
		return err
	} else {
		_, err := p.db.Exec("INSERT INTO chat_list (chat_id, name, project_id, user_id) VALUES ($1, $2, $3, $4)", chatId, name, projectID, userID)
		return err
	}
}

func (p *PostgresDAO) GetChatName(userID string, chatId string) (string, error) {
	var name string
	err := p.db.Get(&name, "SELECT name FROM chat_list WHERE chat_id = $1 AND user_id = $2", chatId, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get chat name: %w", err)
	}
	return name, nil
}

func (p *PostgresDAO) SaveChatName(userID string, chatId string, name string) error {
	_, err := p.db.Exec("UPDATE chat_list SET name = $1 WHERE chat_id = $2 AND user_id = $3", name, chatId, userID)
	if err != nil {
		return fmt.Errorf("failed to save chat name: %w", err)
	}
	return nil
}

// AddChatMessage adds a message to a chat
func (p *PostgresDAO) AddChatMessage(userID string, chatId string, role string, content string) error {
	_, err := p.db.Exec("INSERT INTO chat_messages (chat_id, role, content, user_id) VALUES ($1, $2, $3, $4)", chatId, role, content, userID)
	return err
}

// GetChatMessages retrieves all messages for a given chat
func (p *PostgresDAO) GetChatMessages(userID string, chatId string) ([]ChatMessageRow, error) {
	var messages []ChatMessageRow
	err := p.db.Select(&messages, "SELECT role, content, id FROM chat_messages WHERE chat_id = $1 AND user_id = $2 ORDER BY id", chatId, userID)
	return messages, err
}

// GetChatList retrieves all chats for a user
func (p *PostgresDAO) GetChatList(userID string, projectID string) ([]*proto.ChatInfo, error) {
	var chats []ChatInfoRow
	var err error

	if projectID == "" || projectID == "null" {
		err = p.db.Select(&chats, "SELECT chat_id, name FROM chat_list WHERE project_id IS NULL AND user_id = $1", userID)
	} else {
		err = p.db.Select(&chats, "SELECT chat_id, name FROM chat_list WHERE project_id = $1 AND user_id = $2", projectID, userID)
	}

	if err != nil {
		return nil, err
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

func (p *PostgresDAO) AddChatMessageWithTokens(userID string, chatId string, role string, content string, model string, inputTokens int, outputTokens int) (int64, error) {
	// PostgreSQL doesn't have LastInsertId(), so we use RETURNING
	var messageId int64
	err := p.db.Get(&messageId, `
		INSERT INTO chat_messages (chat_id, role, content, model, input_token_count, output_token_count, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`,
		chatId, role, content, model, inputTokens, outputTokens, userID)
	if err != nil {
		return 0, err
	}

	return messageId, nil
}

// GetModels retrieves all available models
func (p *PostgresDAO) GetModels() ([]proto.ModelListInfo, error) {
	var models []struct {
		ID   string `db:"id"`
		Name string `db:"name"`
	}
	err := p.db.Select(&models, "SELECT id, name FROM model_metadata")
	if err != nil {
		return nil, err
	}

	var result []proto.ModelListInfo
	for _, m := range models {
		result = append(result, proto.ModelListInfo{
			Id:    m.ID,
			Label: m.Name,
		})
	}
	return result, nil
}

// SearchChatMessages performs full text search across chat messages
func (p *PostgresDAO) SearchChatMessages(userID string, query string) ([]proto.SearchResult, error) {
	// Input validation and sanitization
	if userID == "" || query == "" {
		return nil, errors.New("userID and query are required")
	}

	// Sanitize query to prevent injection
	sanitizedQuery := sanitizeFTSQuery(query)
	if sanitizedQuery == "" {
		return nil, errors.New("query contains no searchable terms")
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
		return nil, fmt.Errorf("failed to execute FTS query: %w", err)
	}
	fmt.Println("rows", rows)
	defer rows.Close()

	var results []proto.SearchResult
	for rows.Next() {

		var chatId, chatName, matchedText string

		err := rows.Scan(&chatId, &chatName, &matchedText)
		if err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
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
	_, err := p.db.Exec(`
		INSERT INTO project (id, name, description, additional_data, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, id, name, description, additionalData, userID)
	if err != nil {
		return "", err
	}
	return id, nil
}

// GetProjects retrieves all projects for a user
func (p *PostgresDAO) GetProjects(userID string) ([]ProjectRow, error) {
	var projects []ProjectRow
	err := p.db.Select(&projects, `SELECT id, name, description, additional_data, created_at, updated_at FROM project WHERE user_id = $1`, userID)
	return projects, err
}

func (p *PostgresDAO) FileSave(userID string, project_id string, docs_id string, file_name string, file_size int64) error {
	size_kb := file_size / 1024
	_, err := p.db.Exec("INSERT INTO project_docs (project_id, docs_id, file_name, file_size, embedding_status, user_id) VALUES ($1, $2, $3, $4, $5, $6)",
		project_id, docs_id, file_name, size_kb, int32(proto.Embedding_Status_STATUS_QUEUED), userID)
	return err
}

func (p *PostgresDAO) UpdateEmbeddingStatus(docs_id string, status int32) error {
	_, err := p.db.Exec("UPDATE project_docs SET embedding_status = $1 WHERE docs_id = $2", status, docs_id)
	return err
}

func (p *PostgresDAO) FetchErrorDocs(userID string, project_id string) ([]string, error) {
	var docs_list []string
	err := p.db.Select(&docs_list, "SELECT docs_id FROM project_docs WHERE project_id = $1 AND embedding_status = $2 AND user_id = $3",
		project_id, int32(proto.Embedding_Status_STATUS_ERROR), userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch error docs: %w", err)
	}
	return docs_list, nil
}

func (p *PostgresDAO) TotalUsedSize(userID string, projectID string) (int64, error) {
	var total int64
	err := p.db.Get(&total, `
		SELECT COALESCE(SUM(file_size), 0)
		FROM project_docs
		WHERE project_id = $1 AND user_id = $2
	`, projectID, userID)
	return total, err
}

func (p *PostgresDAO) FilesList(userID string, project_id string) ([]DocumentListRow, error) {
	var files []DocumentListRow
	err := p.db.Select(&files, `
		SELECT id, project_id, docs_id, file_name, created_at, updated_at, embedding_status
		FROM project_docs
		WHERE project_id = $1 AND user_id = $2
	`, project_id, userID)
	return files, err
}

func (p *PostgresDAO) GetFileMetadata(docsId string) (*DocumentListRow, error) {
	var doc DocumentListRow
	err := p.db.Get(&doc, `SELECT * FROM project_docs WHERE docs_id = $1`, docsId)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// SaveRAGChunk saves a chunk to rag_chunks table
func (p *PostgresDAO) SaveRAGChunk(userID string, chunkID, projectID, docsID string, startByte, endByte int) error {
	_, err := p.db.Exec(`
		INSERT INTO rag_chunks (id, project_id, docs_id, start_byte, end_byte, user_id)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, chunkID, projectID, docsID, startByte, endByte, userID)
	return err
}

// SaveRAGChunkEmbedding stores vector embedding for a RAG chunk
func (p *PostgresDAO) SaveRAGChunkEmbedding(chunkID string, embedding []float64) error {
	// Input validation
	if chunkID == "" {
		return errors.New("chunkID cannot be empty")
	}
	if len(embedding) == 0 {
		return errors.New("embedding cannot be empty")
	}
	if len(embedding) != 768 { // Validate expected dimension (768 as per CHOICE 2)
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
		return fmt.Errorf("failed to save embedding for chunk %s: %w", chunkID, err)
	}

	return nil
}

// GetTopSimilarRAGChunks retrieves most similar chunks using cosine similarity
func (p *PostgresDAO) GetTopSimilarRAGChunks(userID string, queryEmbedding string, projectID string) ([]RAGChunkRow, error) {
	// Input validation
	if userID == "" || queryEmbedding == "" || projectID == "" {
		return nil, errors.New("userID, queryEmbedding, and projectID are required")
	}

	// Validate embedding format
	if !isValidEmbeddingFormat(queryEmbedding) {
		return nil, errors.New("invalid embedding format")
	}

	query := `
		SELECT id, project_id, docs_id, start_byte, end_byte
    FROM rag_chunks 
    WHERE user_id = $2 
      AND project_id = $3
      AND embedding IS NOT NULL
    ORDER BY embedding <=> $1  -- Cosine distance (smaller = more similar)
    LIMIT 10`

	var chunks []RAGChunkRow
	rows, err := p.db.Query(query, queryEmbedding, userID, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query similar chunks: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var chunk RAGChunkRow

		err := rows.Scan(&chunk.ID, &chunk.ProjectID, &chunk.DocsID,
			&chunk.StartByte, &chunk.EndByte)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chunk row: %w", err)
		}
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

func (p *PostgresDAO) IsMainBranch(userID string, source_chat_id string) (bool, error) {
	var isMainBranch bool
	err := p.db.Get(&isMainBranch, `SELECT is_main_branch FROM chat_list WHERE chat_id = $1 AND user_id = $2`, source_chat_id, userID)
	return isMainBranch, err
}

func (p *PostgresDAO) BranchChat(userID string, source_chat_id string, parent_message_id string, new_chat_id string, branch_name string) error {
	tx, err := p.db.Beginx()
	if err != nil {
		return err
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
		return err
	}

	// Copy messages up to branch point
	_, err = tx.Exec(`INSERT INTO chat_messages (chat_id, role, content, model, error, input_token_count, output_token_count, created_at, user_id)
					  SELECT $1, role, content, model, error, input_token_count, output_token_count, created_at, $2
					  FROM chat_messages 
					  WHERE chat_id = $3 AND id <= $4 AND user_id = $5
					  ORDER BY id`, new_chat_id, userID, source_chat_id, parent_message_id, userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (p *PostgresDAO) GetChatBranches(userID string, chatId string, isMain bool) ([]ChatInfoRow, error) {
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
		return nil, err
	}

	return chats, nil
}

// Helper function to convert float64 slice to pgvector string format
func vectorToString(embedding []float64) string {
	strValues := make([]string, len(embedding))
	for i, v := range embedding {
		strValues[i] = strconv.FormatFloat(v, 'f', -1, 64)
	}
	return "[" + strings.Join(strValues, ",") + "]"
}

// Helper function to validate embedding format
func isValidEmbeddingFormat(embedding string) bool {
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

// PostgresSettingsDAO implements the SettingsDAO interface using PostgreSQL
type PostgresSettingsDAO struct {
	db *sqlx.DB
}

func NewPostgresSettingsDAO(config *PostgresConfig) (*PostgresSettingsDAO, error) {
	dsn := config.GetPostgresDSN()

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL connection for settings: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.Pool.MaxOpenConnections)
	db.SetMaxIdleConns(config.Pool.MaxIdleConnections)
	db.SetConnMaxLifetime(config.Pool.ConnectionMaxLifetime)

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping PostgreSQL database for settings: %w", err)
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

	var dbSetting dbSettings
	err := p.db.Get(&dbSetting, "SELECT name, settings FROM settings WHERE name = $1", settingName)
	if err != nil {
		// Preserve sql.ErrNoRows so callers can distinguish between no rows and actual database errors
		if err == sql.ErrNoRows {
			return "", err
		}
		return "", fmt.Errorf("failed to get setting '%s' from database: %w", settingName, err)
	}

	return dbSetting.Settings, nil
}

func (p *PostgresSettingsDAO) SetSettingValue(settingName string, settingValue string) error {
	query := `
        INSERT INTO settings (name, settings) VALUES ($1, $2)
        ON CONFLICT(name) DO UPDATE SET settings = EXCLUDED.settings, updated_at = CURRENT_TIMESTAMP
    `

	_, err := p.db.Exec(query, settingName, settingValue)
	if err != nil {
		return fmt.Errorf("failed to upsert settings: %w", err)
	}
	return nil
}
