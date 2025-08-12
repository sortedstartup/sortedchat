package dao

import (
	proto "sortedstartup/chatservice/proto"
)

type DAO interface {
	// Chat CRUD
	CreateChat(userID string, chatId string, name string, projectID string) error
	GetChatName(userID string, chatId string) (string, error)
	SaveChatName(userID string, chatId string, name string) error
	AddChatMessage(userID string, chatId string, role string, content string) error
	AddChatMessageWithTokens(userID string, chatId string, role string, content string, model string, inputTokens int, outputTokens int) (int64, error)
	GetChatMessages(userID string, chatId string) ([]ChatMessageRow, error)

	// GetChatList retrieves all chats for a user
	GetChatList(userID string, projectID string) ([]*proto.ChatInfo, error)

	// Model operations
	GetModels() ([]proto.ModelListInfo, error)

	// Search operations
	SearchChatMessages(userID string, query string) ([]proto.SearchResult, error)

	//Project Operations
	CreateProject(userID string, id string, name string, description string, additionalData string) (string, error)
	GetProjects(userID string) ([]ProjectRow, error)
	FileSave(userID string, project_id string, docs_id string, file_name string, fileSize int64) error
	UpdateEmbeddingStatus(docs_id string, status int32) error
	FetchErrorDocs(userID string, project_id string) ([]string, error)
	FilesList(userID string, project_id string) ([]DocumentListRow, error)
	GetFileMetadata(docsId string) (*DocumentListRow, error)
	TotalUsedSize(userID string, projectID string) (int64, error)

	// SaveRAGChunk saves a chunk to rag_chunks table
	SaveRAGChunk(userID string, chunkID, projectID, docsID string, startByte, endByte int) error
	SaveRAGChunkEmbedding(chunkID string, embedding []float64) error
	GetTopSimilarRAGChunks(userID string, embedding string, projectID string) ([]RAGChunkRow, error)

	IsMainBranch(userID string, source_chat_id string) (bool, error)
	BranchChat(userID string, source_chat_id string, parent_message_id string, new_chat_id string, branch_name string, project_id string) error
	GetChatBranches(userID string, chatId string, isMain bool) ([]*proto.ChatInfo, error)
}

type SettingsDAO interface {
	GetSettingValue(settingName string) (string, error)
	SetSettingValue(settingName string, settingValue string) error
}
