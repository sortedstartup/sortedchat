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
func (s *SQLiteDAO) AddChatMessage(userID string, chatId string, role string, content string) error {
	_, err := s.db.Exec("INSERT INTO chat_messages (chat_id, role, content, user_id) VALUES (?, ?, ?, ?)", chatId, role, content, userID)
	return err
}

// GetChatMessages retrieves all messages for a given chat
func (s *SQLiteDAO) GetChatMessages(userID string, chatId string) ([]ChatMessageRow, error) {
	// todo : do we need to order by time?
	var messages []ChatMessageRow
	err := s.db.Select(&messages, "SELECT role, content, id FROM chat_messages WHERE chat_id = ? AND user_id = ?", chatId, userID)
	return messages, err
}

type ChatInfoRow struct {
	Id   string `db:"chat_id"`
	Name string `db:"name"`
}

// GetChatList retrieves all chats for a user
func (s *SQLiteDAO) GetChatList(userID string, projectID string) ([]*proto.ChatInfo, error) {
	var chats []ChatInfoRow
	var err error

	if projectID == "" || projectID == "null" {
		err = s.db.Select(&chats, "SELECT chat_id, name FROM chat_list WHERE project_id IS NULL AND user_id = ?", userID)
	} else {
		err = s.db.Select(&chats, "SELECT chat_id, name FROM chat_list WHERE project_id = ? AND user_id = ?", projectID, userID)
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

func (s *SQLiteDAO) AddChatMessageWithTokens(userID string, chatId string, role string, content string, model string, inputTokens int, outputTokens int) (int64, error) {
	result, err := s.db.Exec(`
		INSERT INTO chat_messages (chat_id, role, content, model, input_token_count, output_token_count, user_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		chatId, role, content, model, inputTokens, outputTokens, userID)
	if err != nil {
		return 0, err
	}

	messageId, err := result.LastInsertId()
	return messageId, err
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
				'\n-----\n'
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
        SELECT id,project_id,docs_id,start_byte,end_byte
        FROM rag_chunks
        WHERE project_id = ? AND user_id = ?
        AND id IN (
            SELECT id
            FROM rag_chunks_vec
            WHERE embedding MATCH ?
            ORDER BY distance
            LIMIT 2
        )
    `, projectID, userID, embedding)
	return chunks, err
}

func (s *SQLiteDAO) IsMainBranch(userID string, source_chat_id string) (bool, error) {
	var isMainBranch bool
	err := s.db.Get(&isMainBranch, `SELECT is_main_branch FROM chat_list WHERE chat_id = ? AND user_id = ?`, source_chat_id, userID)
	return isMainBranch, err
}

func (s *SQLiteDAO) BranchChat(userID string, source_chat_id string, parent_message_id string, new_chat_id string, branch_name string, project_id string) error {
	_, err := s.db.Exec(`INSERT INTO chat_list (chat_id, name, project_id, parent_chat_id, parent_message_id, is_main_branch, user_id)
						 VALUES (?, ?, ?, ?, ?, FALSE, ?)`, new_chat_id, branch_name, project_id, source_chat_id, parent_message_id, userID)
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

func (s *SQLiteDAO) GetChatBranches(userID string, chatId string, isMain bool) ([]*proto.ChatInfo, error) {
	var chats []ChatInfoRow
	var err error

	if isMain {
		err = s.db.Select(&chats, `SELECT chat_id, name FROM chat_list WHERE parent_chat_id = ? AND user_id = ?`, chatId, userID)
	} else {
		err = s.db.Select(&chats, `
			SELECT c1.chat_id, c1.name 
			FROM chat_list c1
			JOIN chat_list c2 ON c1.parent_chat_id = c2.chat_id
			WHERE c2.chat_id = ? AND c1.user_id = ? AND c2.user_id = ?
		`, chatId, userID, userID)
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

type dbSettings struct {
	Name     string `db:"name" json:"name"`
	Settings string `db:"settings" json:"settings"`
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
