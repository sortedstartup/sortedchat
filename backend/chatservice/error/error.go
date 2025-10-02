package errors

import "fmt"

// Common error messages used across the application
const (
	// Authentication errors
	ErrFailedToGetUserID = "failed to get user ID"

	// Chat service errors
	ErrFailedToGetSettings        = "failed to get settings"
	ErrFailedToSetSettings        = "failed to set settings"
	ErrFailedToInitializeChat     = "failed to initialize ChatService"
	ErrFailedToGenerateChatName   = "failed to generate chat name"
	ErrFailedToGetHistory         = "failed to get chat message history"
	ErrFailedToGetChatList        = "failed to get chat list"
	ErrFailedToCreateChat         = "failed to create chat"
	ErrFailedToListModels         = "failed to list models"
	ErrFailedToSearchChat         = "failed to search chat"
	ErrFailedToCreateProject      = "failed to create project"
	ErrFailedToGetProjects        = "failed to get projects"
	ErrFailedToFetchDocuments     = "failed to fetch documents"
	ErrFailedToSubmitEmbeddingJob = "failed to submit generate embeddings job"
	ErrFailedToBranchChat         = "failed to branch a chat"
	ErrFailedToListChatBranch     = "failed to list chat branch"
	ErrFailedToDeleteDocument     = "failed to delete document"
	ErrFailedToDeleteChat         = "failed to delete chat"
	ErrFailedToRestoreChat        = "failed to restore chat"
	ErrFailedToRenameChat         = "failed to rename chat"

	// Database errors
	ErrFailedToSaveFile              = "failed to save file"
	ErrFailedToUpdateEmbeddingStatus = "failed to update embedding status"
	ErrFailedToFetchErrorDocs        = "failed to fetch error docs"

	// HTTP errors
	ErrMethodNotAllowed   = "Method not allowed"
	ErrMissingProjectID   = "Missing project_id"
	ErrFileNotProvided    = "File not provided"
	ErrUserIDNotFound     = "User ID not found"
	ErrFailedToUploadFile = "Failed to upload file"
)

// ErrorWithContext creates an error with additional context
func ErrorWithContext(baseErr error, context string) error {
	return fmt.Errorf("%s: %w", context, baseErr)
}

// NewError creates a new error with the given message
func NewError(message string) error {
	return fmt.Errorf(message)
}
