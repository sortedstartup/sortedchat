package dao

import (
	proto "sortedstartup/chatservice/proto"
)

type DAO interface {
	// Chat CRUD
	CreateChat(chatId string, name string, projectID string) error
	AddChatMessage(chatId string, role string, content string) error
	AddChatMessageWithTokens(chatId string, role string, content string, model string, inputTokens int, outputTokens int) error
	GetChatMessages(chatId string) ([]ChatMessageRow, error)

	// GetChatList retrieves all chats
	GetChatList(projectID string) ([]*proto.ChatInfo, error)

	// Model operations
	GetModels() ([]proto.ModelListInfo, error)

	// Search operations
	SearchChatMessages(query string) ([]proto.SearchResult, error)

	//Project Operations
	CreateProject(id string, name string, description string, additionalData string) (string, error)
	GetProjects() ([]ProjectRow, error)
	FileSave(project_id string, docs_id string, file_name string, fileSize int64) error
	FilesList(project_id string) ([]DocumentListRow, error)
	GetFileMetadata(docsId string) (*DocumentListRow, error)

	// SaveRAGChunk saves a chunk to rag_chunks table
	SaveRAGChunk(chunkID, projectID, docsID string, startByte, endByte int) error
	SaveRAGChunkEmbedding(chunkID string, embedding []float64) error
	GetTopSimilarRAGChunks(embedding string, projectID string) ([]RAGChunkRow, error)
}

type SettingsDAO interface {
	GetSettings() (*proto.Settings, error)
	SetSettings(settings *proto.Settings) error
}
