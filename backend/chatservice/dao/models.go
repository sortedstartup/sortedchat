package dao

import (
	"encoding/json"
	"fmt"
	"sortedstartup/chatservice/proto"
	"strings"
)

type ChatMessageRow struct {
	Role               string  `db:"role" json:"role"`
	Content            string  `db:"content" json:"content"`
	ContentImage       string  `db:"content_image" json:"content_image"`
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
	ID               string  `db:"id"`
	Name             string  `db:"name"`
	Provider         string  `db:"provider"`
	URL              string  `db:"url"`
	InputTokenCost   float32 `db:"input_token_cost"`
	OutputTokenCost  float32 `db:"output_token_cost"`
	Capabilities     string  `db:"capabilities"` // JSON string from SQLite
	IsDownloadable   bool    `db:"is_downloadable"`
	IsDownloaded     bool    `db:"is_downloaded"`
	IsEmbeddingModel bool    `db:"is_embedding_model"`
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

func ParseCapabilities(capabilitiesJSON string) (*proto.ModelCapabilities, error) {
	if strings.TrimSpace(capabilitiesJSON) == "" {
		return &proto.ModelCapabilities{}, nil
	}
	var caps CapabilitiesJSON
	dec := json.NewDecoder(strings.NewReader(capabilitiesJSON))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&caps); err != nil {
		return nil, fmt.Errorf("parse capabilities: %w", err)
	}

	toProto := func(c CapabilityJSON) *proto.Capability {
		if !c.Input && !c.Output {
			return nil
		}
		return &proto.Capability{Input: c.Input, Output: c.Output}
	}

	return &proto.ModelCapabilities{
		Text:     toProto(caps.Text),
		Audio:    toProto(caps.Audio),
		Video:    toProto(caps.Video),
		Image:    toProto(caps.Image),
		Realtime: caps.Realtime,
	}, nil
}

type AgentRow struct {
	ID           string `db:"id" json:"id"`
	Name         string `db:"name" json:"name"`
	Description  string `db:"description" json:"description"`
	SystemPrompt string `db:"system_prompt" json:"system_prompt"`
	Provider     string `db:"provider" json:"provider"`
	Model        string `db:"model" json:"model"`
	LocalTools   string `db:"local_tools" json:"local_tools"`
	MCPServers   string `db:"mcp_servers" json:"mcp_servers"` // JSON string
	CreatedAt    string `db:"created_at" json:"created_at"`
	UpdatedAt    string `db:"updated_at" json:"updated_at"`
}

type AgentSessionRow struct {
	ID              string  `db:"id" json:"id"`
	AgentID         string  `db:"agent_id" json:"agent_id"`
	UserID          string  `db:"user_id" json:"user_id"`
	Status          string  `db:"status" json:"status"`
	Title           *string `db:"title" json:"title"`
	ParentSessionID *string `db:"parent_session_id" json:"parent_session_id"`
	CreatedAt       string  `db:"created_at" json:"created_at"`
	UpdatedAt       string  `db:"updated_at" json:"updated_at"`
}

type AgentMessageRow struct {
	ID               string  `db:"id" json:"id"`
	SessionID        string  `db:"session_id" json:"session_id"`
	SequenceNumber   int     `db:"sequence_number" json:"sequence_number"`
	Role             string  `db:"role" json:"role"`
	Type             string  `db:"type" json:"type"`
	Content          string  `db:"content" json:"content"`
	ToolName         *string `db:"tool_name" json:"tool_name"`
	ToolCallID       *string `db:"tool_call_id" json:"tool_call_id"`
	ToolArgs         *string `db:"tool_args" json:"tool_args"`
	ThoughtSignature *string `db:"thought_signature" json:"thought_signature"`
	CreatedAt        string  `db:"created_at" json:"created_at"`
	Success          bool    `db:"success" json:"success"`
	ErrorMessage     *string `db:"error_message" json:"error_message"`
	RunTimeMs        int64   `db:"run_time_ms" json:"run_time_ms"`
}

type AgentFSOperationRow struct {
	ID           string  `db:"id" json:"id"`
	AgentID      string  `db:"agent_id" json:"agent_id"`
	SessionID    *string `db:"session_id" json:"session_id"`
	Operation    string  `db:"operation" json:"operation"`
	Path         string  `db:"path" json:"path"`
	Success      bool    `db:"success" json:"success"`
	ErrorMessage *string `db:"error_message" json:"error_message"`
	FileSize     *int64  `db:"file_size" json:"file_size"`
	CreatedAt    string  `db:"created_at" json:"created_at"`
}

type AgentDocumentRow struct {
	ID         string `db:"id" json:"id"`
	AgentID    string `db:"agent_id" json:"agent_id"`
	DocsID     string `db:"docs_id" json:"docs_id"`
	FileName   string `db:"file_name" json:"file_name"`
	FilePath   string `db:"file_path" json:"file_path"`
	FileSize   int64  `db:"file_size" json:"file_size"`
	UploadedBy string `db:"uploaded_by" json:"uploaded_by"`
	CreatedAt  string `db:"created_at" json:"created_at"`
	UpdatedAt  string `db:"updated_at" json:"updated_at"`
}
