package dao

import (
	"encoding/json"
	proto "sortedstartup/chatservice/proto"
)

type ChatMessageRow struct {
	Role               string  `db:"role" json:"role"`
	Content            string  `db:"content" json:"content"`
	Id                 string  `db:"id" json:"id"`
	DocumentReferences string  `db:"document_references" json:"document_references"`
	RagEnabled         bool    `db:"rag_enabled" json:"rag_enabled"`
	Model              string  `db:"model" json:"model"`
	InputTokenCount    int     `db:"input_token_count" json:"input_token_count"`
	OutputTokenCount   int     `db:"output_token_count" json:"output_token_count"`
	CachedTokenCount   int     `db:"cached_token_count" json:"cached_token_count"`
	Cost               float64 `db:"cost" json:"cost"`
}

type MessageSummary struct {
	MessageId        string  `db:"message_id" json:"message_id"`
	Model            string  `db:"model" json:"model"`
	InputTokenCount  int     `db:"input_token_count" json:"input_token_count"`
	OutputTokenCount int     `db:"output_token_count" json:"output_token_count"`
	CachedTokenCount int     `db:"cached_token_count" json:"cached_token_count"`
	Cost             float64 `db:"cost" json:"cost"`
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
	ID         string  `db:"id"`
	ProjectID  string  `db:"project_id"`
	DocsID     string  `db:"docs_id"`
	StartByte  int     `db:"start_byte"`
	EndByte    int     `db:"end_byte"`
	Source     *string `db:"source"`
	Similarity float64 `db:"similarity"`
}

type ChatInfoRow struct {
	Id               string  `db:"chat_id"`
	Name             string  `db:"name"`
	Cost             float64 `db:"cost"`
	InputTokenCount  int     `db:"input_token_count"`
	OutputTokenCount int     `db:"output_token_count"`
	CachedTokenCount int     `db:"cached_token_count"`
}

type Models struct {
	ID              string  `db:"id"`
	Name            string  `db:"name"`
	Provider        string  `db:"provider"`
	URL             string  `db:"url"`
	InputTokenCost  float32 `db:"input_token_cost"`
	OutputTokenCost float32 `db:"output_token_cost"`
	Capabilities    string  `db:"capabilities"` // JSON string from SQLite
}

// Intermediate struct for JSON parsing
type CapabilitiesJSON struct {
	Text     CapabilityJSON `json:"text"`
	Audio    CapabilityJSON `json:"audio"`
	Video    CapabilityJSON `json:"video"`
	Image    CapabilityJSON `json:"image"`
	Realtime bool           `json:"realtime"`
}

type CapabilityJSON struct {
	Input  bool `json:"input"`
	Output bool `json:"output"`
}

type dbSettings struct {
	Name     string `db:"name"`
	Settings string `db:"settings"`
}

func parseCapabilities(capabilitiesJSON string) (*proto.ModelCapabilities, error) {
	var caps CapabilitiesJSON

	err := json.Unmarshal([]byte(capabilitiesJSON), &caps)
	if err != nil {
		return nil, err
	}

	return &proto.ModelCapabilities{
		Text: &proto.Capability{
			Input:  caps.Text.Input,
			Output: caps.Text.Output,
		},
		Audio: &proto.Capability{
			Input:  caps.Audio.Input,
			Output: caps.Audio.Output,
		},
		Video: &proto.Capability{
			Input:  caps.Video.Input,
			Output: caps.Video.Output,
		},
		Image: &proto.Capability{
			Input:  caps.Image.Input,
			Output: caps.Image.Output,
		},
		Realtime: caps.Realtime,
	}, nil
}
