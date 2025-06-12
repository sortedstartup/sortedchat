package dao

import (
	proto "sortedstartup/chatservice/proto"
)

type DAO interface {
	// Chat CRUD
	CreateChat(chatId string, name string) error
	AddChatMessage(chatId string, role string, content string) error
	AddChatMessageWithTokens(chatId string, role string, content string, model string, inputTokens int, outputTokens int) error
	GetChatMessages(chatId string) ([]ChatMessageRow, error)

	// GetChatList retrieves all chats
	GetChatList() ([]*proto.ChatInfo, error)

	// Model operations
	GetModels() ([]proto.ModelListInfo, error)

	// Search operations
	SearchChatMessages(query string) ([]proto.SearchResult, error)

	//Project Operations
	CreateProject(name string, description string, additionalData string) (int64, error)
	GetProjects() ([]ProjectRow, error)
}
