package dao

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	proto "sortedstartup/chatservice/proto"

	// sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteDAO implements the DAO interface using SQLite and sqlx
type SQLiteDAO struct {
	db *sqlx.DB
}

// NewSQLiteDAO creates a new SQLite DAO instance
func NewSQLiteDAO(sqliteUrl string) (*SQLiteDAO, error) {
	// sqlite_vec.Auto()

	db, err := sqlx.Open("sqlite3", sqliteUrl)
	if err != nil {
		return nil, err
	}

	return &SQLiteDAO{db: db}, nil
}

// CreateChat creates a new chat with the given ID and name
func (s *SQLiteDAO) CreateChat(userID string, chatId string, name string, projectID string) error {
	if projectID == "" || projectID == "null" {
		_, err := s.db.Exec("INSERT INTO chat_list (chat_id, name, user_id) VALUES (?, ?, ?)", chatId, name, userID)
		return err
	} else {
		_, err := s.db.Exec("INSERT INTO chat_list (chat_id, name, project_id, user_id) VALUES (?, ?, ?, ?)", chatId, name, projectID, userID)
		return err
	}
}

func (s *SQLiteDAO) GetChatName(userID string, chatId string) (string, error) {
	var name string
	err := s.db.Get(&name, "SELECT name FROM chat_list WHERE chat_id = ? AND user_id = ?", chatId, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get chat name: %w", err)
	}
	return name, nil
}

func (s *SQLiteDAO) SaveChatName(userID string, chatId string, name string) error {
	_, err := s.db.Exec("UPDATE chat_list SET name = ? WHERE chat_id = ? AND user_id = ?", name, chatId, userID)
	if err != nil {
		return fmt.Errorf("failed to get chat name: %w", err)
	}
	return nil
}

// AddChatMessage adds a message to a chat
func (s *SQLiteDAO) AddChatMessage(userID string, chatId string, role string, content string, ragEnabled bool) (string, error) {
	result, err := s.db.Exec("INSERT INTO chat_messages (chat_id, role, content, user_id, rag_enabled) VALUES (?, ?, ?, ?, ?)", chatId, role, content, userID, ragEnabled)
	if err != nil {
		return "", err
	}

	messageId, err := result.LastInsertId()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%d", messageId), nil
}

func (s *SQLiteDAO) GetChatMessages(userID string, chatId string) ([]ChatMessageRow, error) {
	var messages []ChatMessageRow
	query := `
        SELECT
            cm.role,
            cm.content,
            cm.id,
            COALESCE(cm.document_references, '') as document_references,
            (cm.rag_enabled = 1) as rag_enabled,
            COALESCE(cm.model, '') as model,
            COALESCE(cm.input_token_count, 0) as input_token_count,
            COALESCE(cm.output_token_count, 0) as output_token_count,
            COALESCE(mm.input_token_cost, 0) * COALESCE(cm.input_token_count, 0) +
            COALESCE(mm.output_token_cost, 0) * COALESCE(cm.output_token_count, 0) as cost
        FROM chat_messages cm
        LEFT JOIN model_metadata mm ON cm.model = mm.id
        WHERE cm.chat_id = ? AND cm.user_id = ?
        ORDER BY cm.id
    `
	err := s.db.Select(&messages, query, chatId, userID)
	return messages, err
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

	err := s.db.Select(&chats, query, args...)
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

func (s *SQLiteDAO) AddChatMessageWithTokens(userID string, chatId string, role string, content string, model string, inputTokens int, outputTokens int, references string, ragEnabled bool) (MessageSummary, error) {

	result, err := s.db.Exec(`
		INSERT INTO chat_messages (chat_id, role, content, model, input_token_count, output_token_count, user_id, document_references, rag_enabled)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		chatId, role, content, model, inputTokens, outputTokens, userID, references, ragEnabled)
	if err != nil {
		return MessageSummary{}, err
	}

	messageId, err := result.LastInsertId()
	if err != nil {
		fmt.Println("errsanskar", err)
		return MessageSummary{}, err
	}

	var inputTokenCost, outputTokenCost float64

	err = s.db.QueryRow(`
        SELECT input_token_cost, output_token_cost
        FROM model_metadata
        WHERE id = ?
    `, model).Scan(&inputTokenCost, &outputTokenCost)
	if err != nil {
		return MessageSummary{}, err
	}

	cost := inputTokenCost*float64(inputTokens) + outputTokenCost*float64(outputTokens)

	summary := MessageSummary{
		MessageId:        fmt.Sprintf("%d", messageId),
		Model:            model,
		InputTokenCount:  inputTokens,
		OutputTokenCount: outputTokens,
		Cost:             cost,
	}

	return summary, err
}

// GetModels retrieves all available models
func (s *SQLiteDAO) GetModels() ([]proto.ModelListInfo, error) {
	var models []struct {
		ID   string `db:"id"`
		Name string `db:"name"`
	}
	err := s.db.Select(&models, "SELECT id, name FROM model_metadata")
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
		return nil, err
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
		return "", err
	}
	return id, nil
}

// GetProjectList retrieves all projects for a user
func (s *SQLiteDAO) GetProjects(userID string) ([]ProjectRow, error) {
	var projects []ProjectRow
	err := s.db.Select(&projects, `SELECT id, name, description, additional_data, created_at, updated_at FROM project WHERE user_id = ?`, userID)
	return projects, err
}

func (s *SQLiteDAO) FileSave(userID string, project_id string, docs_id string, file_name string, file_size int64) error {
	size_kb := file_size / 1024
	_, err := s.db.Exec("INSERT INTO project_docs (project_id, docs_id, file_name,file_size,embedding_status, user_id) VALUES (?, ?, ?, ?, ?, ?)", project_id, docs_id, file_name, size_kb, int32(proto.Embedding_Status_STATUS_QUEUED), userID)
	return err
}

func (s *SQLiteDAO) UpdateEmbeddingStatus(docs_id string, status int32) error {
	_, err := s.db.Exec("UPDATE project_docs SET embedding_status = ? WHERE docs_id = ?", status, docs_id)
	return err
}

func (s *SQLiteDAO) FetchErrorDocs(userID string, project_id string) ([]string, error) {
	var docs_list []string
	err := s.db.Select(&docs_list, "SELECT docs_id FROM project_docs WHERE project_id = ? AND embedding_status = ? AND user_id = ?", project_id, int32(proto.Embedding_Status_STATUS_ERROR), userID)
	if err != nil {
		fmt.Print("fetchErrorDocs dao", err)
		return nil, fmt.Errorf("failed to check embedding status: %w", err)
	}
	fmt.Println("fetchErrorDocs dao", docs_list)
	return docs_list, nil
}

func (s *SQLiteDAO) TotalUsedSize(userID string, projectID string) (int64, error) {
	var total int64
	err := s.db.Get(&total, `
		SELECT COALESCE(SUM(file_size), 0)
		FROM project_docs
		WHERE project_id = ? AND user_id = ?
	`, projectID, userID)
	return total, err
}

func (s *SQLiteDAO) FilesList(userID string, project_id string) ([]DocumentListRow, error) {
	var files []DocumentListRow
	err := s.db.Select(&files, `
		SELECT id, project_id, docs_id, file_name, created_at, updated_at,embedding_status
		FROM project_docs
		WHERE project_id = ? AND user_id = ?
	`, project_id, userID)
	return files, err
}

func (s *SQLiteDAO) GetFileMetadata(docsId string) (*DocumentListRow, error) {
	var doc DocumentListRow
	err := s.db.Get(&doc, `SELECT * FROM project_docs WHERE docs_id = ?`, docsId)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// SaveRAGChunk saves a chunk to rag_chunks table
func (s *SQLiteDAO) SaveRAGChunk(userID string, chunkID, projectID, docsID string, startByte, endByte int) error {
	_, err := s.db.Exec(`
		INSERT INTO rag_chunks (id, project_id, docs_id, start_byte, end_byte, user_id)
		VALUES (?, ?, ?, ?, ?, ?)
	`, chunkID, projectID, docsID, startByte, endByte, userID)
	return err
}

func (s *SQLiteDAO) SaveRAGChunkEmbedding(chunkID string, vector []float64) error {
	arr, err := json.Marshal(vector)
	if err != nil {
		return fmt.Errorf("failed: %w", err)
	}

	_, err = s.db.Exec("INSERT INTO rag_chunks_vec (id, embedding) VALUES (?, ?)", chunkID, string(arr))
	return err
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
	return chunks, err
}

func (s *SQLiteDAO) IsMainBranch(userID string, source_chat_id string) (bool, error) {
	var isMainBranch bool
	err := s.db.Get(&isMainBranch, `SELECT is_main_branch FROM chat_list WHERE chat_id = ? AND user_id = ?`, source_chat_id, userID)
	return isMainBranch, err
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
		return err
	}

	//copy messages up to branch point
	_, err = s.db.Exec(`INSERT INTO chat_messages (chat_id, role, content, model, error, input_token_count, output_token_count, created_at, user_id)
						SELECT ?, role, content, model, error, input_token_count, output_token_count, created_at, ?
						FROM chat_messages 
						WHERE chat_id = ? AND id <= ? AND user_id = ?
						ORDER BY id;`, new_chat_id, userID, source_chat_id, parent_message_id, userID)
	return err
}

func (s *SQLiteDAO) GetChatBranches(userID string, chatId string, isMain bool) ([]ChatInfoRow, error) {
	var chats []ChatInfoRow
	var err error

	if isMain {
		err = s.db.Select(&chats, `SELECT chat_id, name FROM chat_list WHERE parent_chat_id = ?`, chatId)
	} else {
		err = s.db.Select(&chats, `
			SELECT c1.chat_id, c1.name 
			FROM chat_list c1
			JOIN chat_list c2 ON c1.chat_id = c2.parent_chat_id
			WHERE c2.chat_id = ?
		`, chatId)
	}

	if err != nil {
		return nil, err
	}

	return chats, nil
}

func (s *SQLiteDAO) DeleteDocument(userID string, projectID string, docID string) error {
	// Start a transaction to ensure all operations succeed or fail together
	tx, err := s.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("transaction rollback failed: %v (original error: %v)", rbErr, err)
			}
		}
	}()
	// Delete from rag_chunks_vec first (embeddings)
	_, err = tx.Exec("DELETE FROM rag_chunks_vec WHERE id IN (SELECT id FROM rag_chunks WHERE project_id = ? AND docs_id = ? AND user_id = ?)", projectID, docID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete from rag_chunks_vec: %w", err)
	}

	// Delete from rag_chunks (document chunks)
	_, err = tx.Exec("DELETE FROM rag_chunks WHERE project_id = ? AND docs_id = ? AND user_id = ?", projectID, docID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete from rag_chunks: %w", err)
	}

	// Delete from project_docs (document metadata)
	_, err = tx.Exec("DELETE FROM project_docs WHERE project_id = ? AND docs_id = ? AND user_id = ?", projectID, docID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete from project_docs: %w", err)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *SQLiteDAO) SoftDeleteChat(userID string, chatId string) error {
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
	return err
}

func (s *SQLiteDAO) DeleteChat(userID string, chatId string) error {
	_, err := s.db.Exec(`
        WITH RECURSIVE chat_hierarchy AS (
            SELECT chat_id FROM chat_list WHERE chat_id = ? AND user_id = ?
            UNION ALL
            SELECT c.chat_id FROM chat_list c
            JOIN chat_hierarchy h ON c.parent_chat_id = h.chat_id
			WHERE c.user_id = ?
        )
        DELETE FROM chat_list
        WHERE chat_id IN (SELECT chat_id FROM chat_hierarchy)
        AND user_id = ?;
    `, chatId, userID, userID, userID)
	return err
}

func (s *SQLiteDAO) RestoreChat(userID string, chatId string) error {
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
	return err
}

func (s *SQLiteDAO) IsChatDeleted(chatId string, userID string) (bool, error) {
	var isDeleted bool
	err := s.db.Get(&isDeleted, "SELECT soft_deleted FROM chat_list WHERE chat_id = ? AND user_id = ?", chatId, userID)
	if err != nil {
		return false, err
	}
	return isDeleted, err
}

type SQLiteSettingsDAO struct {
	db *sqlx.DB
}

func NewSQLiteSettingsDAO(sqliteUrl string) *SQLiteSettingsDAO {
	db, err := sqlx.Open("sqlite3", sqliteUrl)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	return &SQLiteSettingsDAO{db: db}
}

func (s *SQLiteSettingsDAO) GetSettingValue(settingName string) (string, error) {
	var dbSetting dbSettings
	err := s.db.Get(&dbSetting, "SELECT name, settings FROM settings WHERE name = ?", settingName)
	if err != nil {
		// Preserve sql.ErrNoRows so callers can distinguish between no rows and actual database errors
		if err == sql.ErrNoRows {
			return "", err
		}
		return "", fmt.Errorf("failed to get setting '%s' from database: %w", settingName, err)
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

// GetChatMessageByID retrieves a specific chat message by its ID
func (s *SQLiteDAO) GetChatMessageByID(userID string, messageID string) (*ChatMessageRow, error) {
	var message ChatMessageRow
	err := s.db.Get(&message, `
		SELECT role, content, id, COALESCE(document_references, '') as document_references 
		FROM chat_messages 
		WHERE id = ? AND user_id = ?`, messageID, userID)
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// UpdateChatMessageDocumentReferences updates the document references for a specific message
func (s *SQLiteDAO) UpdateChatMessageDocumentReferences(userID string, messageID string, documentReferences string) error {
	_, err := s.db.Exec(`
		UPDATE chat_messages 
		SET document_references = ? 
		WHERE id = ? AND user_id = ?`, documentReferences, messageID, userID)
	return err
}
