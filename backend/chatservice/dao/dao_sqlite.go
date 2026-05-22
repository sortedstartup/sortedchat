package dao

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	proto "sortedstartup/chatservice/proto"
	"strings"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteDAO implements the DAO interface using SQLite and sqlx
type SQLiteDAO struct {
	db *sqlx.DB
}

// NewSQLiteDAO creates a new SQLite DAO instance
func NewSQLiteDAO(sqliteUrl string) (*SQLiteDAO, error) {
	sqlite_vec.Auto()

	db, err := sqlx.Open("sqlite3", sqliteUrl)
	if err != nil {
		return nil, err
	}

	// Set busy timeout to 30 seconds
	_, err = db.Exec("PRAGMA busy_timeout = 30000;")
	if err != nil {
		slog.Error("dao_sqlite:NewSQLiteDAO", "message", "failed to set busy timeout", "error", err)
		return nil, err
	}

	// Enable WAL mode
	_, err = db.Exec("PRAGMA journal_mode = WAL;")
	if err != nil {
		slog.Error("dao_sqlite:NewSQLiteDAO", "message", "failed to set WAL mode", "error", err)
		return nil, err
	}

	return &SQLiteDAO{db: db}, nil
}

// NewSQLiteDAOWithDB creates a new SQLite DAO instance with an existing connection
func NewSQLiteDAOWithDB(db *sqlx.DB) *SQLiteDAO {
	return &SQLiteDAO{db: db}
}

func NewSQLiteInMemoryDAO(dbConn *sql.DB) (*SQLiteDAO, error) {
	return &SQLiteDAO{db: sqlx.NewDb(dbConn, "sqlite3")}, nil
}

// CreateChat creates a new chat with the given ID and name
func (s *SQLiteDAO) CreateChat(userID string, chatId string, name string, projectID string) error {
	if projectID == "" || projectID == "null" {
		_, err := s.db.Exec("INSERT INTO chat_list (chat_id, name, user_id) VALUES (?, ?, ?)", chatId, name, userID)
		if err != nil {
			slog.Error("dao_sqlite:CreateChat", "message", "failed to create chat", "error", err, "projectID", projectID, "userID", userID)
			return fmt.Errorf("failed to create chat")
		}
		return nil
	} else {
		_, err := s.db.Exec("INSERT INTO chat_list (chat_id, name, project_id, user_id) VALUES (?, ?, ?, ?)", chatId, name, projectID, userID)
		if err != nil {
			slog.Error("dao_sqlite:CreateChat", "message", "failed to create chat", "error", err, "projectID", projectID, "userID", userID)
			return fmt.Errorf("failed to create chat")
		}
		return nil
	}
}

func (s *SQLiteDAO) GetChatName(userID string, chatId string) (string, error) {
	var name string
	err := s.db.Get(&name, "SELECT name FROM chat_list WHERE chat_id = ? AND user_id = ?", chatId, userID)
	if err != nil {
		slog.Error("dao_sqlite:GetChatName", "message", "failed to get chat name", "error", err, "chatId", chatId, "userID", userID)
		return "", fmt.Errorf("failed to get chat name")
	}
	return name, nil
}

func (s *SQLiteDAO) SaveChatName(userID string, chatId string, name string) error {
	_, err := s.db.Exec("UPDATE chat_list SET name = ? WHERE chat_id = ? AND user_id = ?", name, chatId, userID)
	if err != nil {
		slog.Error("dao_sqlite:SaveChatName", "message", "failed to save chat name", "error", err, "chatId", chatId, "userID", userID)
		return fmt.Errorf("failed to save chat name")
	}
	return nil
}

// AddChatMessage adds a message to a chat
func (s *SQLiteDAO) AddChatMessage(userID string, chatId string, role string, content string, contentImage string, model string, inputTokens int, outputTokens int, cachedTokens int, references string, ragEnabled bool) (string, error) {

	// Handle contentImage - use NULL if empty
	var contentImageValue interface{}
	if contentImage == "" {
		contentImageValue = nil
	} else {
		contentImageValue = contentImage
	}

	result, err := s.db.Exec("INSERT INTO chat_messages (chat_id, role, content, content_image, user_id, rag_enabled, model, input_token_count, output_token_count, cached_token_count, document_references) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", chatId, role, content, contentImageValue, userID, ragEnabled, model, inputTokens, outputTokens, cachedTokens, references)
	if err != nil {
		slog.Error("dao_sqlite:AddChatMessage", "message", "failed to add chat message", "error", err, "chatId", chatId, "userID", userID)
		return "", err
	}

	messageId, err := result.LastInsertId()
	if err != nil {
		slog.Error("dao_sqlite:AddChatMessage", "message", "failed to get last insert id", "error", err, "chatId", chatId, "userID", userID)
		return "", fmt.Errorf("failed to get last insert id") //change this error message
	}

	return fmt.Sprintf("%d", messageId), nil
}

func (s *SQLiteDAO) GetModelByID(modelID string) (*Models, error) {
	var model Models
	err := s.db.Get(&model,
		"SELECT id, name, provider, url, input_token_cost, output_token_cost, COALESCE(capabilities, '{}') AS capabilities FROM shared_models_metadata WHERE id = ?",
		modelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get model: %w", err)
	}
	return &model, nil
}

func (s *SQLiteDAO) GetChatMessages(userID string, chatId string) ([]ChatMessageRow, error) {
	var messages []ChatMessageRow
	err := s.db.Select(&messages, "SELECT role, content, COALESCE(content_image, '') as content_image, id, COALESCE(document_references, '') as document_references, (rag_enabled = 1) as rag_enabled,COALESCE(model, '') as model, COALESCE(input_token_count, 0) as input_token_count, COALESCE(output_token_count, 0) as output_token_count, COALESCE(cached_token_count, 0) as cached_token_count, COALESCE(cost, 0) as cost FROM chat_messages WHERE chat_id = ? AND user_id = ? ORDER BY id", chatId, userID)
	if err != nil {
		slog.Error("dao_sqlite:GetChatMessages", "message", "failed to get chat messages", "error", err, "chatId", chatId, "userID", userID)
		return nil, fmt.Errorf("failed to get chat messages")
	}
	return messages, nil
}

// GetChatList retrieves all chats for a user
func (s *SQLiteDAO) GetChatList(userID string, projectID string, softDeleted bool) ([]*proto.ChatInfo, error) {
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

	err := s.db.Select(&chats, query, args...)
	if err != nil {
		slog.Error("dao_sqlite:GetChatList", "message", "failed to get chat list", "error", err, "projectID", projectID, "userID", userID)
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

func (s *SQLiteDAO) AddChatMessageWithTokens(
	userID string,
	chatId string,
	role string,
	content string,
	contentImage string,
	model string,
	inputTokens int,
	outputTokens int,
	cachedTokens int,
	references string,
	ragEnabled bool,
) (MessageSummary, error) {
	// Handle contentImage - use NULL if empty
	var contentImageValue interface{}
	if contentImage == "" {
		contentImageValue = nil
	} else {
		contentImageValue = contentImage
	}

	// Insert the message first and capture its ID
	result, err := s.db.Exec(`
        INSERT INTO chat_messages (
            chat_id, role, content, content_image, model,
            input_token_count, output_token_count, cached_token_count,
            user_id, document_references, rag_enabled
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		chatId, role, content, contentImageValue, model,
		inputTokens, outputTokens, cachedTokens,
		userID, references, ragEnabled)
	if err != nil {
		slog.Error("dao_sqlite:AddChatMessageWithTokens", "message", "failed to add chat message with tokens", "error", err, "chatId", chatId, "userID", userID)
		return MessageSummary{}, fmt.Errorf("failed to add chat message with tokens")
	}

	messageId, err := result.LastInsertId()
	if err != nil {
		slog.Error("dao_sqlite:AddChatMessageWithTokens", "message", "failed to get last insert id", "error", err, "chatId", chatId, "userID", userID)
		return MessageSummary{}, fmt.Errorf("failed to get last insert id")
	}

	// Get the model metadata for cost calculation
	var inputCost, outputCost, cachedCost float64
	err = s.db.QueryRow(`
        SELECT input_token_cost, output_token_cost, cached_token_cost
        FROM shared_models_metadata
        WHERE id = ?`, model).Scan(&inputCost, &outputCost, &cachedCost)
	if err != nil {
		slog.Error("dao_sqlite:AddChatMessageWithTokens", "message", "failed to get model metadata", "error", err, "chatId", chatId, "userID", userID)
		return MessageSummary{}, fmt.Errorf("failed to get model metadata")
	}

	// Calculate the cost
	cost := (float64(inputTokens)*inputCost + float64(outputTokens)*outputCost + float64(cachedTokens)*cachedCost) / 1000000.0

	// Update the message with the calculated cost
	_, err = s.db.Exec(`
        UPDATE chat_messages 
        SET cost = ? 
        WHERE id = ?`, cost, messageId)
	if err != nil {
		slog.Error("dao_sqlite:AddChatMessageWithTokens", "message", "failed to update message cost", "error", err, "chatId", chatId, "userID", userID)
		return MessageSummary{}, fmt.Errorf("failed to update message cost")
	}

	// Update the chat_list with cumulative totals
	_, err = s.db.Exec(`
        UPDATE chat_list
        SET
            cost = COALESCE(cost, 0) + ?,
            input_token_count = COALESCE(input_token_count, 0) + ?,
            output_token_count = COALESCE(output_token_count, 0) + ?,
            cached_token_count = COALESCE(cached_token_count, 0) + ?
        WHERE chat_id = ? AND user_id = ?`,
		cost, inputTokens, outputTokens, cachedTokens, chatId, userID)
	if err != nil {
		slog.Error("dao_sqlite:AddChatMessageWithTokens", "message", "failed to update chat_list", "error", err, "chatId", chatId, "userID", userID)
		return MessageSummary{}, fmt.Errorf("failed to update chat list")
	}

	// Return the message summary
	summary := MessageSummary{
		MessageId:        fmt.Sprintf("%d", messageId),
		Model:            model,
		InputTokenCount:  inputTokens,
		OutputTokenCount: outputTokens,
		CachedTokenCount: cachedTokens,
		Cost:             cost,
	}

	return summary, nil
}

// SearchChatMessages searches chat messages using FTS
func (s *SQLiteDAO) SearchChatMessages(userID string, query string) ([]proto.SearchResult, error) {
	const searchSQL = `
        SELECT
            cm.chat_id as chat_id,
            cl.name AS chat_name,
            GROUP_CONCAT(
                CASE
                    WHEN LENGTH(cm.content) > 100 THEN SUBSTR(cm.content, 1, 100) || '...'
                    ELSE cm.content
                END,
                char(10) || '---' || char(10)
            ) AS aggregated_snippets
        FROM
            chat_messages_fts AS fts
        JOIN
            chat_messages AS cm ON fts.rowid = cm.id
        JOIN
            chat_list AS cl ON cm.chat_id = cl.chat_id
        WHERE
            fts.chat_messages_fts MATCH ? AND cm.user_id = ? AND cl.user_id = ?
        GROUP BY
            cm.chat_id, cl.name
        ORDER BY
            cm.chat_id;
    `

	var rows []struct {
		ChatID      string `db:"chat_id"`
		ChatName    string `db:"chat_name"`
		MatchedText string `db:"aggregated_snippets"`
	}

	err := s.db.Select(&rows, searchSQL, query, userID, userID)
	if err != nil {
		slog.Error("dao_sqlite:SearchChatMessages", "message", "failed to search chat messages", "error", err, "query", query, "userID", userID)
		return nil, fmt.Errorf("failed to search chat messages")
	}

	var results []proto.SearchResult
	for _, row := range rows {
		results = append(results, proto.SearchResult{
			ChatId:      row.ChatID,
			ChatName:    row.ChatName,
			MatchedText: row.MatchedText,
		})
	}

	return results, nil
}

// Project CRUD
func (s *SQLiteDAO) CreateProject(userID string, id string, name string, description string, additionalData string) (string, error) {
	_, err := s.db.Exec(`
		INSERT INTO project (id, name, description, additional_data, user_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, id, name, description, additionalData, userID)
	if err != nil {
		slog.Error("dao_sqlite:CreateProject", "message", "failed to create project", "error", err, "userID", userID, "name", name)
		return "", fmt.Errorf("failed to create project")
	}
	return id, nil
}

// GetProjectList retrieves all projects for a user
func (s *SQLiteDAO) GetProjects(userID string) ([]ProjectRow, error) {
	var projects []ProjectRow
	err := s.db.Select(&projects, `SELECT id, name, description, additional_data, created_at, updated_at FROM project WHERE user_id = ?`, userID)
	if err != nil {
		slog.Error("dao_sqlite:GetProjects", "message", "failed to get projects", "error", err, "userID", userID)
		return nil, fmt.Errorf("failed to get projects")
	}
	return projects, nil
}

func (s *SQLiteDAO) FileSave(userID string, project_id string, docs_id string, file_name string, file_size int64) error {
	size_kb := file_size / 1024
	_, err := s.db.Exec("INSERT INTO project_docs (project_id, docs_id, file_name,file_size,embedding_status, user_id) VALUES (?, ?, ?, ?, ?, ?)", project_id, docs_id, file_name, size_kb, int32(proto.Embedding_Status_STATUS_QUEUED), userID)
	if err != nil {
		slog.Error("dao_sqlite:FileSave", "message", "failed to save file", "error", err, "userID", userID, "project_id", project_id, "docs_id", docs_id, "file_size", file_size)
		return fmt.Errorf("failed to save file, please try again")
	}
	return nil
}

func (s *SQLiteDAO) UpdateEmbeddingStatus(docs_id string, status int32) error {
	_, err := s.db.Exec("UPDATE project_docs SET embedding_status = ? WHERE docs_id = ?", status, docs_id)
	if err != nil {
		slog.Error("dao_sqlite:UpdateEmbeddingStatus", "message", "failed to update embedding status", "error", err, "docs_id", docs_id, "status", status)
		return fmt.Errorf("failed to update embedding status, please try again")
	}
	return nil
}

func (s *SQLiteDAO) FetchErrorDocs(userID string, project_id string) ([]string, error) {
	var docs_list []string
	err := s.db.Select(&docs_list, "SELECT docs_id FROM project_docs WHERE project_id = ? AND embedding_status = ? AND user_id = ?", project_id, int32(proto.Embedding_Status_STATUS_ERROR), userID)
	if err != nil {
		slog.Error("dao_sqlite:FetchErrorDocs", "message", "failed to fetch error docs", "error", err, "userID", userID, "project_id", project_id)
		return nil, fmt.Errorf("failed to check embedding status, please try again")
	}
	return docs_list, nil
}

func (s *SQLiteDAO) TotalUsedSize(userID string, projectID string) (int64, error) {
	var total int64
	err := s.db.Get(&total, `
		SELECT COALESCE(SUM(file_size), 0)
		FROM project_docs
		WHERE project_id = ? AND user_id = ?
	`, projectID, userID)
	if err != nil {
		slog.Error("dao_sqlite:TotalUsedSize", "message", "failed to get total used size", "error", err, "userID", userID, "projectID", projectID)
		return 0, fmt.Errorf("failed to get total used size")
	}
	return total, nil
}

func (s *SQLiteDAO) FilesList(userID string, project_id string) ([]DocumentListRow, error) {
	var files []DocumentListRow
	err := s.db.Select(&files, `
		SELECT id, project_id, docs_id, file_name, created_at, updated_at,embedding_status
		FROM project_docs
		WHERE project_id = ? AND user_id = ?
	`, project_id, userID)
	if err != nil {
		slog.Error("dao_sqlite:FilesList", "message", "failed to get files list", "error", err, "userID", userID, "project_id", project_id)
		return nil, fmt.Errorf("failed to get files list")
	}
	return files, nil
}

func (s *SQLiteDAO) GetFileMetadata(docsId string) (*DocumentListRow, error) {
	var doc DocumentListRow
	err := s.db.Get(&doc, `SELECT * FROM project_docs WHERE docs_id = ?`, docsId)
	if err != nil {
		slog.Error("dao_sqlite:GetFileMetadata", "message", "failed to get file metadata", "error", err, "docsId", docsId)
		return nil, fmt.Errorf("failed to get document metadata")
	}
	return &doc, nil
}

// SaveRAGChunk saves a chunk to rag_chunks table
func (s *SQLiteDAO) SaveRAGChunk(userID string, chunkID, projectID, docsID string, startByte, endByte int) error {
	_, err := s.db.Exec(`
		INSERT INTO rag_chunks (id, project_id, docs_id, start_byte, end_byte, user_id)
		VALUES (?, ?, ?, ?, ?, ?)
	`, chunkID, projectID, docsID, startByte, endByte, userID)
	if err != nil {
		slog.Error("dao_sqlite:SaveRAGChunk", "message", "failed to save rag chunk", "error", err, "userID", userID, "chunkID", chunkID, "projectID", projectID, "docsID", docsID, "startByte", startByte, "endByte", endByte)
		return fmt.Errorf("failed to save rag chunk")
	}
	return err
}

func (s *SQLiteDAO) SaveRAGChunkEmbedding(chunkID string, vector []float64) error {
	arr, err := json.Marshal(vector)
	if err != nil {
		slog.Error("dao_sqlite:SaveRAGChunkEmbedding", "message", "failed to marshal vector", "error", err, "chunkID", chunkID)
		return fmt.Errorf("failed to marshal vector")
	}

	_, err = s.db.Exec("INSERT INTO rag_chunks_vec (id, embedding) VALUES (?, ?)", chunkID, string(arr))
	if err != nil {
		slog.Error("dao_sqlite:SaveRAGChunkEmbedding", "message", "failed to save rag chunk embedding", "error", err, "chunkID", chunkID)
		return fmt.Errorf("failed to save rag chunk embedding")
	}
	return nil
}

func (s *SQLiteDAO) GetTopSimilarRAGChunks(userID string, embedding string, projectID string) ([]RAGChunkRow, error) {
	var chunks []RAGChunkRow
	err := s.db.Select(&chunks, `
			SELECT 
			rc.id,
			rc.project_id,
			rc.docs_id,
			rc.start_byte,
			rc.end_byte,
			vec_distance_cosine(rcv.embedding, ?) AS similarity
			FROM rag_chunks rc
			JOIN rag_chunks_vec rcv ON rc.id = rcv.id
			WHERE rc.project_id = ? AND rc.user_id = ?
			ORDER BY similarity
			LIMIT 2
    `, embedding, projectID, userID)
	if err != nil {
		slog.Error("dao_sqlite:GetTopSimilarRAGChunks", "message", "failed to get top similar chunks", "error", err, "userID", userID, "embedding", embedding, "projectID", projectID)
		return nil, fmt.Errorf("failed to get top similar chunks")
	}
	return chunks, nil
}

func (s *SQLiteDAO) IsMainBranch(userID string, source_chat_id string) (bool, error) {
	var isMainBranch bool
	err := s.db.Get(&isMainBranch, `SELECT is_main_branch FROM chat_list WHERE chat_id = ? AND user_id = ?`, source_chat_id, userID)
	if err != nil {
		slog.Error("dao_sqlite:IsMainBranch", "message", "failed to get main branch status", "error", err, "userID", userID, "source_chat_id", source_chat_id)
		return false, fmt.Errorf("failed to get main branch status")
	}
	return isMainBranch, nil
}

func (s *SQLiteDAO) BranchChat(userID string, source_chat_id string, parent_message_id string, new_chat_id string, branch_name string) error {
	// Use CTE to find project_id from source chat and insert the new branch chat
	_, err := s.db.Exec(`WITH source_chat AS (
							SELECT project_id 
							FROM chat_list 
							WHERE chat_id = ? AND user_id = ?
						)
						INSERT INTO chat_list (chat_id, name, project_id, parent_chat_id, parent_message_id, is_main_branch, user_id)
						SELECT ?, ?, COALESCE(source_chat.project_id, NULL), ?, ?, FALSE, ?
						FROM source_chat`, source_chat_id, userID, new_chat_id, branch_name, source_chat_id, parent_message_id, userID)
	if err != nil {
		slog.Error("dao_sqlite:BranchChat", "message", "failed to branch chat", "error", err, "userID", userID, "source_chat_id", source_chat_id, "parent_message_id", parent_message_id, "new_chat_id", new_chat_id, "branch_name", branch_name)
		return fmt.Errorf("failed to branch chat, please try again")
	}

	//copy messages up to branch point
	_, err = s.db.Exec(`INSERT INTO chat_messages (chat_id, role, content, model, error, input_token_count, output_token_count, created_at, user_id)
						SELECT ?, role, content, model, error, input_token_count, output_token_count, created_at, ?
						FROM chat_messages 
						WHERE chat_id = ? AND id <= ? AND user_id = ?
						ORDER BY id;`, new_chat_id, userID, source_chat_id, parent_message_id, userID)
	if err != nil {
		slog.Error("dao_sqlite:BranchChat", "message", "failed to copy messages", "error", err, "userID", userID, "source_chat_id", source_chat_id, "parent_message_id", parent_message_id, "new_chat_id", new_chat_id, "branch_name", branch_name)
		return fmt.Errorf("failed to copy messages, please try again")
	}
	return nil
}

func (s *SQLiteDAO) GetChatBranches(userID string, chatId string, isMain bool) ([]ChatInfoRow, error) {
	var chats []ChatInfoRow
	var err error

	if isMain {
		err = s.db.Select(&chats, `SELECT chat_id, name FROM chat_list WHERE parent_chat_id = ?`, chatId)
		if err != nil {
			slog.Error("dao_sqlite:GetChatBranches", "message", "failed to get main branches", "error", err, "userID", userID, "chatId", chatId, "isMain", isMain)
			return nil, fmt.Errorf("failed to get main branches, please try again")
		}
	} else {
		err = s.db.Select(&chats, `
			SELECT c1.chat_id, c1.name 
			FROM chat_list c1
			JOIN chat_list c2 ON c1.chat_id = c2.parent_chat_id
			WHERE c2.chat_id = ?
		`, chatId)
		if err != nil {
			slog.Error("dao_sqlite:GetChatBranches", "message", "failed to get branches", "error", err, "userID", userID, "chatId", chatId, "isMain", isMain)
			return nil, fmt.Errorf("failed to get branches, please try again")
		}
	}

	if err != nil {
		slog.Error("dao_sqlite:GetChatBranches", "message", "failed to get branches", "error", err, "userID", userID, "chatId", chatId, "isMain", isMain)
		return nil, fmt.Errorf("failed to get branches, please try again")
	}
	return chats, nil
}

func (s *SQLiteDAO) DeleteDocument(userID string, projectID string, docID string) error {
	slog.Info("dao_sqlite:DeleteDocument", "userID", userID, "projectID", projectID, "docID", docID)
	// Start a transaction to ensure all operations succeed or fail together
	tx, err := s.db.Beginx()
	if err != nil {
		slog.Error("dao_sqlite:DeleteDocument", "message", "failed to begin transaction", "error", err, "userID", userID, "projectID", projectID, "docID", docID)
		return fmt.Errorf("failed to begin transaction, please try again")
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				slog.Error("dao_sqlite:DeleteDocument", "message", "transaction rollback failed", "error", rbErr, "userID", userID, "projectID", projectID, "docID", docID)
			}
			slog.Error("dao_sqlite:DeleteDocument", "message", "transaction rollback failed", "error", err, "userID", userID, "projectID", projectID, "docID", docID)
		}
	}()
	// Delete from rag_chunks_vec first (embeddings)
	_, err = tx.Exec("DELETE FROM rag_chunks_vec WHERE id IN (SELECT id FROM rag_chunks WHERE project_id = ? AND docs_id = ? AND user_id = ?)", projectID, docID, userID)
	if err != nil {
		slog.Error("dao_sqlite:DeleteDocument", "message", "failed to delete from rag_chunks_vec", "error", err, "userID", userID, "projectID", projectID, "docID", docID)
		return fmt.Errorf("failed to delete from rag_chunks_vec, please try again")
	}

	// Delete from rag_chunks (document chunks)
	_, err = tx.Exec("DELETE FROM rag_chunks WHERE project_id = ? AND docs_id = ? AND user_id = ?", projectID, docID, userID)
	if err != nil {
		slog.Error("dao_sqlite:DeleteDocument", "message", "failed to delete from rag_chunks", "error", err, "userID", userID, "projectID", projectID, "docID", docID)
		return fmt.Errorf("failed to delete from rag_chunks, please try again")
	}

	// Delete from project_docs (document metadata)
	_, err = tx.Exec("DELETE FROM project_docs WHERE project_id = ? AND docs_id = ? AND user_id = ?", projectID, docID, userID)
	if err != nil {
		slog.Error("dao_sqlite:DeleteDocument", "message", "failed to delete from project_docs", "error", err, "userID", userID, "projectID", projectID, "docID", docID)
		return fmt.Errorf("failed to delete from project_docs, please try again")
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		slog.Error("dao_sqlite:DeleteDocument", "message", "failed to commit transaction", "error", err, "userID", userID, "projectID", projectID, "docID", docID)
		return fmt.Errorf("failed to commit transaction, please try again")
	}
	return nil
}

func (s *SQLiteDAO) SoftDeleteChat(userID string, chatId string) error {
	slog.Info("dao_sqlite:SoftDeleteChat", "userID", userID, "chatId", chatId)
	_, err := s.db.Exec(`
        WITH RECURSIVE chat_hierarchy AS (
            SELECT chat_id FROM chat_list WHERE chat_id = ? AND user_id = ?
            UNION ALL
            SELECT c.chat_id FROM chat_list c
            JOIN chat_hierarchy h ON c.parent_chat_id = h.chat_id
			WHERE c.user_id = ?
        )
        UPDATE chat_list
        SET soft_deleted = TRUE
        WHERE chat_id IN (SELECT chat_id FROM chat_hierarchy)
        AND user_id = ?;
    `, chatId, userID, userID, userID)
	if err != nil {
		slog.Error("dao_sqlite:SoftDeleteChat", "message", "failed to soft delete chat", "error", err, "userID", userID, "chatId", chatId)
		return fmt.Errorf("failed to soft delete chat, please try again")
	}
	return nil
}
func (s *SQLiteDAO) DeleteChat(userID string, chatId string) error {
	slog.Info("dao_sqlite:DeleteChat", "userID", userID, "chatId", chatId)
	tx, err := s.db.Begin()
	if err != nil {
		slog.Error("dao_sqlite:DeleteChat", "message", "failed to begin transaction", "error", err, "userID", userID, "chatId", chatId)
		return fmt.Errorf("failed to begin transaction, please try again")
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				slog.Error("dao_sqlite:DeleteChat", "message", "transaction rollback failed", "error", rbErr, "userID", userID, "chatId", chatId)
			}
			slog.Error("dao_sqlite:DeleteChat", "message", "transaction rollback failed", "error", err, "userID", userID, "chatId", chatId)
		}
	}()

	// Create temporary table to hold chat IDs
	_, err = tx.Exec(`
        CREATE TEMP TABLE chat_ids_to_delete AS
            WITH RECURSIVE chat_hierarchy AS (
                SELECT chat_id FROM chat_list WHERE chat_id = ? AND user_id = ?
                UNION ALL
                SELECT c.chat_id FROM chat_list c
                JOIN chat_hierarchy h ON c.parent_chat_id = h.chat_id
                WHERE c.user_id = ?
            )
        SELECT chat_id FROM chat_hierarchy;
    `, chatId, userID, userID)
	if err != nil {
		slog.Error("dao_sqlite:DeleteChat", "message", "failed to create temporary table", "error", err, "userID", userID, "chatId", chatId)
		return fmt.Errorf("error while processing request, please try again")
	}

	// Delete messages using the temporary table
	_, err = tx.Exec(`
        DELETE FROM chat_messages
        WHERE user_id = ?
          AND chat_id IN (SELECT chat_id FROM chat_ids_to_delete);
    `, userID)
	if err != nil {
		slog.Error("dao_sqlite:DeleteChat", "message", "failed to delete messages", "error", err, "userID", userID, "chatId", chatId)
		return fmt.Errorf("failed to delete messages, please try again")
	}

	// Delete chats using the temporary table
	_, err = tx.Exec(`
        DELETE FROM chat_list 
        WHERE user_id = ? 
          AND chat_id IN (SELECT chat_id FROM chat_ids_to_delete);
    `, userID)
	if err != nil {
		slog.Error("dao_sqlite:DeleteChat", "message", "failed to delete chats", "error", err, "userID", userID, "chatId", chatId)
		return fmt.Errorf("failed to delete chat, please try again")
	}

	// Drop temp table if it exists
	_, err = tx.Exec(`DROP TABLE IF EXISTS chat_ids_to_delete;`)
	if err != nil {
		slog.Error("dao_sqlite:DeleteChat", "message", "failed to drop temp table", "error", err, "userID", userID, "chatId", chatId)
		return fmt.Errorf("failed to delete chat, please try again")
	}

	if commitErr := tx.Commit(); commitErr != nil {
		slog.Error("dao_sqlite:DeleteChat", "message", "failed to commit transaction", "error", err, "userID", userID, "chatId", chatId)
		return fmt.Errorf("failed to delete chat, please try again")
	}
	return nil
}

func (s *SQLiteDAO) RestoreChat(userID string, chatId string) error {
	slog.Debug("dao_sqlite:RestoreChat", "userID", userID, "chatId", chatId)
	_, err := s.db.Exec(`
        WITH RECURSIVE chat_hierarchy AS (
            SELECT chat_id FROM chat_list WHERE chat_id = ? AND user_id = ?
            UNION ALL
            SELECT c.chat_id FROM chat_list c
            JOIN chat_hierarchy h ON c.parent_chat_id = h.chat_id
			WHERE c.user_id = ?
        )
        UPDATE chat_list
        SET soft_deleted = FALSE
        WHERE chat_id IN (SELECT chat_id FROM chat_hierarchy)
        AND user_id = ?;
    `, chatId, userID, userID, userID)
	if err != nil {
		slog.Error("dao_sqlite:RestoreChat", "message", "failed to restore chat", "error", err, "userID", userID, "chatId", chatId)
		return fmt.Errorf("failed to restore chat, please try again")
	}
	return nil
}

func (s *SQLiteDAO) IsChatDeleted(chatId string, userID string) (bool, error) {
	var isDeleted bool
	err := s.db.Get(&isDeleted, "SELECT soft_deleted FROM chat_list WHERE chat_id = ? AND user_id = ?", chatId, userID)
	if err != nil {
		slog.Error("dao_sqlite:IsChatDeleted", "message", "failed to get chat deleted status", "error", err)
		return false, fmt.Errorf("failed to get chat status")
	}
	return isDeleted, err
}

func (s *SQLiteDAO) GetChatMetadata(userID string, chatId string) (ChatInfoRow, error) {
	var chat ChatInfoRow
	err := s.db.Get(&chat, "SELECT chat_id, name, COALESCE(cost, 0) AS cost, COALESCE(input_token_count, 0) AS input_token_count, COALESCE(output_token_count, 0) AS output_token_count, COALESCE(cached_token_count, 0) AS cached_token_count FROM chat_list WHERE chat_id = ? AND user_id = ?", chatId, userID)
	if err != nil {
		slog.Error("dao_sqlite:GetChatMetadata", "message", "failed to get chat metadata", "error", err, "chatId", chatId, "userID", userID)
		return ChatInfoRow{}, fmt.Errorf("failed to get chat metadata")
	}
	return chat, nil
}

func (s *SQLiteDAO) RenameChat(userID string, chatId string, name string) error {
	result, err := s.db.Exec("UPDATE chat_list SET name = ? WHERE chat_id = ? AND user_id = ?", name, chatId, userID)
	if err != nil {
		slog.Error("dao_sqlite:RenameChat", "message", "failed to rename chat", "error", err, "userID", userID, "chatId", chatId, "name", name)
		return fmt.Errorf("failed to rename chat, please try again")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("dao_sqlite:RenameChat", "message", "failed to get rows affected", "error", err, "userID", userID, "chatId", chatId, "name", name)
		return fmt.Errorf("failed to get rows affected, please try again")
	}
	if rowsAffected == 0 {
		slog.Error("dao_sqlite:RenameChat", "message", "chat not found or permission denied", "userID", userID, "chatId", chatId, "name", name)
		return fmt.Errorf("chat not found or permission denied")
	}
	return nil
}

func (s *SQLiteDAO) IsNameExists(userID string, chatId string, name string) (bool, error) {

	var exists bool
	//in query we are checking if the name exists and the chat id is not the same as the chat id passed in the function
	//in query 1 is like optimization to avoid scanning the whole table
	err := s.db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM chat_list WHERE name = ? AND user_id = ? AND chat_id != ?  AND soft_deleted = 0)", name, userID, chatId)
	if err != nil {
		slog.Error("dao_sqlite:IsNameExists", "message", "failed to check if name exists", "error", err, "userID", userID, "chatId", chatId, "name", name)
		return false, fmt.Errorf("failed to check if name exists")
	}
	return exists, nil
}

func (s *SQLiteDAO) UpsertModel(modelID string, name string, url string, provider string, inputTokenCost float64, outputTokenCost float64, cachedTokenCost float64, isEmbeddingModel bool) error {
	_, err := s.db.Exec(`
		INSERT INTO shared_models_metadata (id, name, url, provider, input_token_cost, output_token_cost, cached_token_cost, is_embedding_model, is_downloaded, is_downloadable, progress, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			url = excluded.url,
			provider = excluded.provider,
			input_token_cost = excluded.input_token_cost,
			output_token_cost = excluded.output_token_cost,
			cached_token_cost = excluded.cached_token_cost,
			is_embedding_model = excluded.is_embedding_model
	`, modelID, name, url, provider, inputTokenCost, outputTokenCost, cachedTokenCost, isEmbeddingModel, false, false, "", 0)
	if err != nil {
		slog.Error("dao_sqlite:UpsertModel", "message", "failed to upsert model", "error", err, "modelID", modelID)
		return fmt.Errorf("failed to upsert model")
	}
	return nil
}

func (s *SQLiteDAO) GetModelCatalogVersion() (*ModelCatalogVersion, error) {
	type catalogVersionRow struct {
		JSONSchemaVersion    sql.NullString `db:"json_schema_version"`
		ModelRevisionVersion sql.NullString `db:"model_revision_version"`
	}

	var row catalogVersionRow
	err := s.db.Get(&row, `
		SELECT json_schema_version, model_revision_version
		FROM shared_models_metadata
		WHERE COALESCE(model_revision_version, '') <> '' OR COALESCE(json_schema_version, '') <> ''
		ORDER BY model_revision_version DESC, id ASC
		LIMIT 1
	`)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		slog.Error("dao_sqlite:GetModelCatalogVersion", "message", "failed to fetch model catalog version", "error", err)
		return nil, fmt.Errorf("failed to fetch model catalog version")
	}

	if !row.JSONSchemaVersion.Valid && !row.ModelRevisionVersion.Valid {
		return nil, nil
	}

	return &ModelCatalogVersion{
		JSONSchemaVersion:    strings.TrimSpace(row.JSONSchemaVersion.String),
		ModelRevisionVersion: strings.TrimSpace(row.ModelRevisionVersion.String),
	}, nil
}

func (s *SQLiteDAO) UpsertHostedModel(model HostedModel, version ModelCatalogVersion) error {
	isEnabled := true
	if model.IsEnabled != nil {
		isEnabled = *model.IsEnabled
	}

	_, err := s.db.Exec(`
		INSERT INTO shared_models_metadata (
			id, name, url, provider, input_token_cost, output_token_cost, cached_token_cost,
			is_embedding_model, is_downloadable, capabilities, is_enabled, model_info,
			creator_name, modified_by, description, json_schema_version, model_revision_version
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			url = excluded.url,
			provider = excluded.provider,
			input_token_cost = excluded.input_token_cost,
			output_token_cost = excluded.output_token_cost,
			cached_token_cost = excluded.cached_token_cost,
			is_embedding_model = excluded.is_embedding_model,
			is_downloadable = excluded.is_downloadable,
			capabilities = excluded.capabilities,
			model_info = excluded.model_info,
			creator_name = excluded.creator_name,
			modified_by = excluded.modified_by,
			description = excluded.description,
			json_schema_version = excluded.json_schema_version,
			model_revision_version = excluded.model_revision_version
	`,
		model.ID,
		model.Name,
		model.URL,
		model.Provider,
		model.InputTokenCost,
		model.OutputTokenCost,
		model.CachedTokenCost,
		model.IsEmbeddingModel,
		model.IsDownloadable,
		defaultJSONObjectBytes(model.Capabilities),
		isEnabled,
		defaultJSONObjectBytes(model.ModelInfo),
		model.CreatorName,
		model.ModifiedBy,
		model.Description,
		version.JSONSchemaVersion,
		version.ModelRevisionVersion,
	)
	if err != nil {
		slog.Error("dao_sqlite:UpsertHostedModel", "message", "failed to upsert hosted model", "error", err, "modelID", model.ID)
		return fmt.Errorf("failed to upsert hosted model")
	}

	return nil
}

func (s *SQLiteDAO) RenameProject(userID string, projectId string, name string) error {
	result, err := s.db.Exec("UPDATE project SET name = ? WHERE id = ? AND user_id = ?", name, projectId, userID)
	if err != nil {
		slog.Error("dao_sqlite:RenameProject", "message", "failed to rename project", "error", err, "userID", userID, "projectId", projectId, "name", name)
		return fmt.Errorf("failed to rename project, please try again")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("dao_sqlite:RenameProject", "message", "failed to get rows affected", "error", err, "userID", userID, "projectId", projectId, "name", name)
		return fmt.Errorf("failed to get rows affected, please try again")
	}
	if rowsAffected == 0 {
		slog.Error("dao_sqlite:RenameProject", "message", "project not found or permission denied", "userID", userID, "projectId", projectId, "name", name)
		return fmt.Errorf("project not found or permission denied")
	}
	return nil
}

func (s *SQLiteDAO) IsProjectNameExists(userID string, projectId string, name string) (bool, error) {
	var exists bool
	err := s.db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM project WHERE name = ? AND user_id = ? AND id != ?)", name, userID, projectId)
	if err != nil {
		slog.Error("dao_sqlite:IsProjectNameExists", "message", "failed to check if project name exists", "error", err, "userID", userID, "projectId", projectId, "name", name)
		return false, fmt.Errorf("failed to check if project name exists")
	}
	return exists, nil
}

type SQLiteSettingsDAO struct {
	db *sqlx.DB
}

func NewSQLiteSettingsDAO(sqliteUrl string) *SQLiteSettingsDAO {
	db, err := sqlx.Open("sqlite3", sqliteUrl)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// Set busy timeout to 30 seconds
	_, err = db.Exec("PRAGMA busy_timeout = 30000;")
	if err != nil {
		slog.Error("dao_sqlite:NewSQLiteSettingsDAO", "message", "failed to set busy timeout", "error", err)
	}

	// Enable WAL mode
	_, err = db.Exec("PRAGMA journal_mode = WAL;")
	if err != nil {
		slog.Error("dao_sqlite:NewSQLiteSettingsDAO", "message", "failed to set WAL mode", "error", err)
	}

	return &SQLiteSettingsDAO{db: db}
}

func NewSQLiteSettingsDAOWithDB(db *sqlx.DB) *SQLiteSettingsDAO {
	return &SQLiteSettingsDAO{db: db}
}

func (s *SQLiteSettingsDAO) GetSettingValue(settingName string) (string, error) {
	var dbSetting dbSettings
	err := s.db.Get(&dbSetting, "SELECT name, settings FROM settings WHERE name = ?", settingName)
	if err != nil {
		// Preserve sql.ErrNoRows so callers can distinguish between no rows and actual database errors
		if err == sql.ErrNoRows {
			slog.Error("dao_sqlite:GetSettingValue", "message", "no rows found", "error", err)
			return "", err
		}
		slog.Error("dao_sqlite:GetSettingValue", "message", "failed to get setting", "error", err)
		return "", err
	}
	return dbSetting.Settings, nil
}

func (s *SQLiteSettingsDAO) SetSettingValue(settingName string, settingValue string) error {
	query := `
        INSERT INTO settings (name, settings) VALUES (?, ?)
        ON CONFLICT(name) DO UPDATE SET settings = excluded.settings
    `

	_, err := s.db.Exec(query, settingName, settingValue)
	if err != nil {
		return fmt.Errorf("failed to upsert settings: %w", err)
	}
	return nil
}

func (s *SQLiteSettingsDAO) GetSettingsByPrefix(prefix string) (map[string]string, error) {
	var dbSettings []dbSettings

	escapedPrefix := strings.ReplaceAll(prefix, "\\", "\\\\")
	escapedPrefix = strings.ReplaceAll(escapedPrefix, "%", "\\%")
	escapedPrefix = strings.ReplaceAll(escapedPrefix, "_", "\\_")

	// Use LIKE 'prefix%' to find matching settings
	err := s.db.Select(&dbSettings, "SELECT name, settings FROM settings WHERE name LIKE ?", escapedPrefix+"%")
	if err != nil {
		slog.Error("dao_sqlite:GetSettingsByPrefix", "message", "failed to get settings", "error", err)
		return nil, err
	}

	result := make(map[string]string)
	for _, setting := range dbSettings {
		result[setting.Name] = setting.Settings
	}
	return result, nil
}

// GetChatMessageByID retrieves a specific chat message by its ID
func (s *SQLiteDAO) GetChatMessageByID(userID string, messageID string) (*ChatMessageRow, error) {
	var message ChatMessageRow
	err := s.db.Get(&message, `
		SELECT role, content, id, COALESCE(document_references, '') as document_references 
		FROM chat_messages 
		WHERE id = ? AND user_id = ?`, messageID, userID)
	if err != nil {
		slog.Error("dao_sqlite:GetChatMessageByID", "message", "failed to get chat message by id", "error", err, "userID", userID, "messageID", messageID)
		return nil, fmt.Errorf("failed to get chat message by id, please try again")
	}
	return &message, nil
}

// UpdateChatMessageDocumentReferences updates the document references for a specific message
func (s *SQLiteDAO) UpdateChatMessageDocumentReferences(userID string, messageID string, documentReferences string) error {
	_, err := s.db.Exec(`
		UPDATE chat_messages 
		SET document_references = ? 
		WHERE id = ? AND user_id = ?`, documentReferences, messageID, userID)
	if err != nil {
		slog.Error("dao_sqlite:UpdateChatMessageDocumentReferences", "message", "failed to update chat message document references", "error", err, "userID", userID, "messageID", messageID)
		return fmt.Errorf("failed to update chat message document references, please try again")
	}
	return nil
}

func (s *SQLiteDAO) GetModels() ([]*proto.ModelListInfo, error) {
	var models []Models

	err := s.db.Select(&models, "SELECT id, name, provider, url, COALESCE(input_token_cost, 0) as input_token_cost, COALESCE(output_token_cost, 0) as output_token_cost, COALESCE(cached_token_cost, 0) as cached_token_cost, COALESCE(capabilities, '{}') AS capabilities, COALESCE(is_embedding_model, 0) as is_embedding_model, COALESCE(is_downloaded, 0) as is_downloaded, COALESCE(is_downloadable, 0) as is_downloadable FROM shared_models_metadata")
	if err != nil {
		slog.Error("dao_sqlite:GetModels", "message", "failed to get models", "error", err)
		return nil, fmt.Errorf("failed to get models")
	}

	var result []*proto.ModelListInfo
	for _, m := range models {
		// Parse capabilities JSON
		capabilities, err := ParseCapabilities(m.Capabilities)
		if err != nil {
			slog.Error("dao_sqlite:GetModels", "message", "failed to parse capabilities for model", "error", err, "modelID", m.ID)
			return nil, fmt.Errorf("failed to parse capabilities for model")
		}

		result = append(result, &proto.ModelListInfo{
			Id:               m.ID,
			Label:            m.Name,
			Provider:         m.Provider,
			Url:              m.URL,
			InputTokenCost:   m.InputTokenCost,
			OutputTokenCost:  m.OutputTokenCost,
			CachedTokenCost:  m.CachedTokenCost,
			Capabilities:     capabilities,
			IsDownloadable:   m.IsDownloadable,
			IsDownloaded:     m.IsDownloaded,
			IsEmbeddingModel: m.IsEmbeddingModel,
		})
	}
	return result, nil
}
