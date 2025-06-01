package dao

type ChatMessageRow struct {
	Role    string `db:"role" json:"role"`
	Content string `db:"content" json:"content"`
}
