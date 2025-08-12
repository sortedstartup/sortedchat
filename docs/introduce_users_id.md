# Introduce User ID Support Plan

## Overview
This document outlines the plan to introduce user ID support in the chat service by separating the API and service layers, and making necessary database changes.

## Goals
1. **Separate Concerns**: Split `api.go` into API layer (handling gRPC requests) and Service layer (business logic)
2. **User Context**: Make all service methods accept a `user_id` parameter for multi-user support
3. **Database Changes**: Update database schema and DAO methods to support user isolation
4. **Backward Compatibility**: Hardcode `user_id = "0"` in API layer to maintain existing functionality

## Implementation Plan

### 1. File Structure Changes
- **Current**: `backend/chatservice/api/api.go` (867 lines - monolithic)
- **New Structure**:
  - `backend/chatservice/api/api.go` - gRPC API handlers only
  - `backend/chatservice/service/service.go` - Business logic and service methods
  - `backend/chatservice/service/settings_service.go` - Settings service logic

### 2. Service Layer Methods (user_id parameter)
All service methods will be updated to accept `user_id` as the first parameter:

**Chat Operations:**
- `CreateChat(userID, chatId, name, projectID) error`
- `GetChatList(userID, projectID) ([]*proto.ChatInfo, error)`
- `GetChatMessages(userID, chatId) ([]ChatMessageRow, error)`
- `AddChatMessage(userID, chatId, role, content) error`
- `SearchChatMessages(userID, query) ([]proto.SearchResult, error)`

**Project Operations:**
- `CreateProject(userID, id, name, description, additionalData) (string, error)`
- `GetProjects(userID) ([]ProjectRow, error)`
- `ListDocuments(userID, projectID) ([]DocumentListRow, error)`

**Other Operations:**
- `GenerateChatName(userID, chatId, message, model) (string, error)`
- `BranchChat(userID, sourceChatId, messageId, branchName, projectId) (string, error)`

### 3. Database Schema Changes

#### 3.1 New Migration: `9_add_user_id.up.sql`
```sql
-- Add user_id column to main tables
ALTER TABLE chat_list ADD COLUMN user_id TEXT DEFAULT '0' NOT NULL;
ALTER TABLE chat_messages ADD COLUMN user_id TEXT DEFAULT '0' NOT NULL;
ALTER TABLE project ADD COLUMN user_id TEXT DEFAULT '0' NOT NULL;
ALTER TABLE project_docs ADD COLUMN user_id TEXT DEFAULT '0' NOT NULL;
ALTER TABLE rag_chunks ADD COLUMN user_id TEXT DEFAULT '0' NOT NULL;

-- Add indexes for performance
CREATE INDEX idx_chat_list_user_id ON chat_list(user_id);
CREATE INDEX idx_chat_messages_user_id ON chat_messages(user_id);
CREATE INDEX idx_project_user_id ON project(user_id);
CREATE INDEX idx_project_docs_user_id ON project_docs(user_id);
CREATE INDEX idx_rag_chunks_user_id ON rag_chunks(user_id);

-- Compound indexes for common queries
CREATE INDEX idx_chat_list_user_project ON chat_list(user_id, project_id);
CREATE INDEX idx_chat_messages_user_chat ON chat_messages(user_id, chat_id);
```

#### 3.2 Updated DAO Interface
All DAO methods will be updated to include `user_id` parameter and filter results accordingly.

### 4. API Layer Changes
- API handlers will hardcode `user_id = "0"` 
- All API methods will call corresponding service methods with user_id
- No proto file changes required (as specified)

### 5. Affected Components

#### 5.1 Core Tables
- `chat_list` - User's chat conversations
- `chat_messages` - Messages within chats  
- `project` - User's projects
- `project_docs` - Documents uploaded to projects
- `rag_chunks` - RAG embeddings for documents

#### 5.2 Service Methods
- **ChatService**: All chat-related operations
- **SettingsService**: User-specific settings (future enhancement)
- **RAG Operations**: Project-scoped embeddings with user context

### 6. Implementation Steps

1. **Create Service Layer** (`service.go`)
   - Extract business logic from `api.go`
   - Add `user_id` parameter to all methods
   - Implement user-scoped operations

2. **Database Migration**
   - Create migration script to add `user_id` columns
   - Add appropriate indexes
   - Set default value '0' for existing data

3. **Update DAO Layer**
   - Modify DAO interface to accept `user_id`
   - Update all SQL queries to filter by `user_id`
   - Ensure data isolation between users

4. **Refactor API Layer**
   - Keep only gRPC handler functions
   - Hardcode `user_id = "0"` 
   - Delegate to service layer methods

5. **Testing & Verification**
   - Verify existing functionality works with hardcoded user_id
   - Test data isolation (future users won't see each other's data)
   - Performance testing with new indexes

### 7. Benefits

- **Multi-user Ready**: Foundation for multiple users
- **Clean Architecture**: Clear separation of API and business logic
- **Data Isolation**: Users will only see their own data
- **Backward Compatible**: Existing functionality preserved
- **Performance**: Proper indexing for user-scoped queries

### 8. Future Enhancements

- Authentication/Authorization layer
- User management APIs
- User-specific settings
- Usage analytics per user
- Resource quotas per user

## Files to be Modified/Created

### New Files:
- `backend/chatservice/service/service.go`
- `backend/chatservice/service/settings_service.go` 
- `backend/chatservice/dao/db/sqlite/scripts/migrations/9_add_user_id.up.sql`

### Modified Files:
- `backend/chatservice/api/api.go` (significantly reduced)
- `backend/chatservice/dao/dao.go` (interface updates)
- `backend/chatservice/dao/dao_sqlite.go` (implementation updates)
- `backend/chatservice/dao/models.go` (add user_id fields)

## Notes
- No changes to proto files as per requirements
- All existing API behavior preserved
- Foundation laid for future multi-user features
- Proper error handling maintained
- Database constraints ensure data integrity
