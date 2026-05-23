package dao

import (
	proto "sortedstartup/chatservice/proto"
)

type DAO interface {
	// Chat CRUD
	CreateChat(userID string, chatId string, name string, projectID string) error
	GetChatName(userID string, chatId string) (string, error)
	SaveChatName(userID string, chatId string, name string) error
	AddChatMessage(userID string, chatId string, role string, content string, contentImage string, model string, inputTokens int, outputTokens int, cachedTokens int, references string, ragEnabled bool, toolInfo *ChatMessageToolInfo) (string, error)
	AddChatMessageWithTokens(userID string, chatId string, role string, content string, contentImage string, model string, inputTokens int, outputTokens int, cachedTokens int, searchCost float64, braveSearchCount int, scrapeAPIUsageTime float64, references string, ragEnabled bool, toolInfo *ChatMessageToolInfo) (MessageSummary, error)
	GetModelByID(modelID string) (*Models, error)
	GetChatMessages(userID string, chatId string) ([]ChatMessageRow, error)
	IsChatDeleted(chatId string, userID string) (bool, error)
	GetChatMetadata(userID string, chatId string) (ChatInfoRow, error)

	GetModels() ([]*proto.ModelListInfo, error)

	// GetChatList retrieves all chats for a user
	GetChatList(userID string, projectID string, softDeleted bool) ([]*proto.ChatInfo, error)

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
	BranchChat(userID string, source_chat_id string, parent_message_id string, new_chat_id string, branch_name string) error
	GetChatBranches(userID string, chatId string, isMain bool) ([]ChatInfoRow, error)

	// RAG Document Reference methods
	GetChatMessageByID(userID string, messageID string) (*ChatMessageRow, error)
	UpdateChatMessageDocumentReferences(userID string, messageID string, documentReferences string) error
	DeleteDocument(userID string, projectID string, docID string) error
	SoftDeleteChat(userID string, chatId string) error
	DeleteChat(userID string, chatId string) error
	RestoreChat(userID string, chatId string) error
	RenameChat(userID string, chatId string, name string) error
	RenameProject(userID string, projectId string, name string) error
	IsNameExists(userID string, chatId string, name string) (bool, error)
	IsProjectNameExists(userID string, projectId string, name string) (bool, error)
	UpsertModel(modelID string, name string, url string, provider string, inputTokenCost float64, outputTokenCost float64, cachedTokenCost float64, isEmbeddingModel bool) error
}

type SettingsDAO interface {
	GetSettingValue(settingName string) (string, error)
	SetSettingValue(settingName string, settingValue string) error
	// This is mainly needed to make like at UI layer simple
	// for e.g. DB has settings 'provider.openai', 'openai.gemini',
	// UI can ask for all LLM provider settings - 'provider.*'
	// return -
	//  {
	//    'provider.openai': setting json (proto message type struct ...,
	//    'provider.gemini': setting json (proto message type struct ...
	// }
	GetSettingsByPrefix(prefix string) (map[string]string, error)
}

type AgentDAO interface {
	CreateAgent(agent AgentRow) error
	UpdateAgent(agent AgentRow) error
	GetAgents() ([]AgentRow, error)
	GetAgent(agentID string) (*AgentRow, error)

	CreateSession(session AgentSessionRow) error
	GetSession(sessionID string) (*AgentSessionRow, error)
	GetAgentSessions(agentID string) ([]AgentSessionRow, error)

	AddAgentMessage(message AgentMessageRow) error
	GetAgentMessages(sessionID string) ([]AgentMessageRow, error)

	// Agent file operations
	SaveAgentFile(agentID, docsID, fileName, filePath string, fileSize int64, userID string) error
	GetAgentFiles(agentID string) ([]AgentDocumentRow, error)
	GetAgentFileByPath(agentID, filePath string) (*AgentDocumentRow, error)
	DeleteAgentFile(agentID, docsID string) error
}
