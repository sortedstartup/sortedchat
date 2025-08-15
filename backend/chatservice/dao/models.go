package dao

type ChatMessageRow struct {
	Role    string `db:"role" json:"role"`
	Content string `db:"content" json:"content"`
	Id      string `db:"id" json:"id"`
}

type ProjectRow struct {
	ID             string `db:"id"`
	Name           string `db:"name"`
	Description    string `db:"description"`
	AdditionalData string `db:"additional_data"`
	CreatedAt      string `db:"created_at"`
	UpdatedAt      string `db:"updated_at"`
}

type DocumentListRow struct {
	ID              int64  `db:"id"`
	ProjectID       string `db:"project_id"`
	DocsID          string `db:"docs_id"`
	FileName        string `db:"file_name"`
	FileSize        string `db:"file_size"`
	CreatedAt       string `db:"created_at"`
	UpdatedAt       string `db:"updated_at"`
	EmbeddingStatus int32  `db:"embedding_status"`
	User            string `db:"user_id"`
}

type RAGChunkRow struct {
	ID        string `db:"id"`
	ProjectID string `db:"project_id"`
	DocsID    string `db:"docs_id"`
	StartByte int    `db:"start_byte"`
	EndByte   int    `db:"end_byte"`
	Source    string `db:"source"`
}

type ChatInfoRow struct {
	Id   string `db:"chat_id"`
	Name string `db:"name"`
}

type dbSettings struct {
	Name     string `db:"name"`
	Settings string `db:"settings"`
}
