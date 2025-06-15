package dao

type ChatMessageRow struct {
	Role    string `db:"role" json:"role"`
	Content string `db:"content" json:"content"`
}

type ProjectRow struct {
	ID             int64  `db:"id"`
	Name           string `db:"name"`
	Description    string `db:"description"`
	AdditionalData string `db:"additional_data"`
	CreatedAt      string `db:"created_at"`
	UpdatedAt      string `db:"updated_at"`
}

type DocumentListRow struct {
	ID        int64  `db:"id"`
	ProjectID int64  `db:"project_id"`
	DocsID    string `db:"docs_id"`
	FileName  string `db:"file_name"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}
