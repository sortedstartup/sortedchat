package dao

import (
	proto "sortedstartup/chatservice/proto"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteDAO implements the DAO interface using SQLite and sqlx
type SQLiteDAO struct {
	db *sqlx.DB
}

// NewSQLiteDAO creates a new SQLite DAO instance
func NewSQLiteDAO(sqliteUrl string) (*SQLiteDAO, error) {
	db, err := sqlx.Open("sqlite3", sqliteUrl)
	if err != nil {
		return nil, err
	}

	return &SQLiteDAO{db: db}, nil
}

// CreateChat creates a new chat with the given ID and name
func (s *SQLiteDAO) CreateChat(chatId string, name string) error {
	_, err := s.db.Exec("INSERT INTO chat_list (chat_id, name) VALUES (?, ?)", chatId, name)
	return err
}

// AddChatMessage adds a message to a chat
func (s *SQLiteDAO) AddChatMessage(chatId string, role string, content string) error {
	_, err := s.db.Exec("INSERT INTO chat_messages (chat_id, role, content) VALUES (?, ?, ?)", chatId, role, content)
	return err
}

// GetChatMessages retrieves all messages for a given chat
func (s *SQLiteDAO) GetChatMessages(chatId string) ([]ChatMessageRow, error) {
	// todo : do we need to order by time?
	var messages []ChatMessageRow
	err := s.db.Select(&messages, "SELECT role, content FROM chat_messages WHERE chat_id = ?", chatId)
	return messages, err
}

type ChatInfoRow struct {
	Id   string `db:"chat_id"`
	Name string `db:"name"`
}

// GetChatList retrieves all chats
func (s *SQLiteDAO) GetChatList() ([]*proto.ChatInfo, error) {
	var chats []ChatInfoRow
	err := s.db.Select(&chats, "SELECT chat_id, name FROM chat_list")

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

// AddChatMessageWithTokens adds a message with token counts and model info
func (s *SQLiteDAO) AddChatMessageWithTokens(chatId string, role string, content string, model string, inputTokens int, outputTokens int) error {
	_, err := s.db.Exec(`
		INSERT INTO chat_messages (chat_id, role, content, model, input_token_count, output_token_count)
		VALUES (?, ?, ?, ?, ?, ?)`,
		chatId, role, content, model, inputTokens, outputTokens)
	return err
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
func (s *SQLiteDAO) SearchChatMessages(query string) ([]proto.SearchResult, error) {
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
			fts.chat_messages_fts MATCH ?
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

	err := s.db.Select(&rows, searchSQL, query)
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
func (s *SQLiteDAO) CreateProject(id string, name string, description string, additionalData string) (string, error) {
	_, err := s.db.Exec(`
		INSERT INTO project (id, name, description, additional_data, created_at, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, id, name, description, additionalData)
	if err != nil {
		return "", err
	}
	if err != nil {
		return "", err
	}
	return id, nil
}

// GetProjectList retrieves all projects
func (s *SQLiteDAO) GetProjects() ([]ProjectRow, error) {
	var projects []ProjectRow
	err := s.db.Select(&projects, `SELECT id, name, description, additional_data, created_at, updated_at FROM project`)
	return projects, err
}

func (s *SQLiteDAO) FileSave(project_id string, docs_id string, file_name string, file_size int64) error {
	size_kb := file_size / 1024
	_, err := s.db.Exec("INSERT INTO project_docs (project_id, docs_id, file_name,file_size) VALUES (?, ?, ?, ?)", project_id, docs_id, file_name, size_kb)
	return err
}

func (s *SQLiteDAO) TotalUsedSize(projectID string) (int64, error) {
	var total int64
	err := s.db.Get(&total, `
		SELECT total(file_size)
		FROM project_docs
		WHERE project_id = ?
	`, projectID)
	return total, err
}

func (s *SQLiteDAO) FilesList(project_id string) ([]DocumentListRow, error) {
	var files []DocumentListRow
	err := s.db.Select(&files, `
		SELECT id, project_id, docs_id, file_name, created_at, updated_at
		FROM project_docs
		WHERE project_id = ?
	`, project_id)
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
